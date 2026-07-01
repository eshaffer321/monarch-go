package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eshaffer321/monarch-go/v2/internal/types"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
)

const (
	graphQLEndpoint = "/graphql"
	authEndpoint    = "/auth/login/"

	authHeaderKey = "Authorization"
	csrfKey       = "csrftoken"
	contentType   = "application/json"
)

// GraphQLTransport handles GraphQL communication
type GraphQLTransport struct {
	baseURL     string
	httpClient  *http.Client
	retryClient *retryablehttp.Client
	headers     map[string]string
	session     *types.Session
	logger      types.Logger
	hooks       *types.Hooks
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage       `json:"data,omitempty"`
	Errors []*types.GraphQLError `json:"errors,omitempty"`
}

// NewGraphQLTransport creates a new GraphQL transport
func NewGraphQLTransport(opts *Options) *GraphQLTransport {
	if opts == nil {
		opts = &Options{}
	}

	// Set defaults
	if opts.BaseURL == "" {
		opts.BaseURL = types.DefaultBaseURL
	}

	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{
			Timeout: types.DefaultTimeout,
		}
	}

	// Create retry client if configured
	var retryClient *retryablehttp.Client
	if opts.RetryConfig != nil {
		retryClient = retryablehttp.NewClient()
		retryClient.HTTPClient = opts.HTTPClient
		retryClient.RetryMax = opts.RetryConfig.MaxRetries
		retryClient.RetryWaitMin = opts.RetryConfig.RetryWait
		retryClient.RetryWaitMax = opts.RetryConfig.MaxWait

		if opts.Logger != nil {
			retryClient.Logger = &retryLogger{logger: opts.Logger}
		}
	}

	// Set default headers
	headers := map[string]string{
		"Accept":          contentType,
		"Content-Type":    contentType,
		"Client-Platform": "web",
		"User-Agent":      types.UserAgent,
		"Origin":          "https://app.monarch.com",
		"Referer":         "https://app.monarch.com/",
	}

	// Merge custom headers
	for k, v := range opts.Headers {
		headers[k] = v
	}

	return &GraphQLTransport{
		baseURL:     opts.BaseURL,
		httpClient:  opts.HTTPClient,
		retryClient: retryClient,
		headers:     headers,
		logger:      opts.Logger,
		hooks:       opts.Hooks,
	}
}

// Execute executes a GraphQL query
func (t *GraphQLTransport) Execute(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	// Check authentication
	if t.session == nil || t.session.Token == "" {
		return types.ErrNotAuthenticated
	}

	// Check session expiry
	if !t.session.ExpiresAt.IsZero() && time.Now().After(t.session.ExpiresAt) {
		return types.ErrSessionExpired
	}

	// Create request — include operationName for servers that require it
	opName := ""
	for _, prefix := range []string{"mutation ", "query ", "subscription "} {
		if idx := strings.Index(query, prefix); idx >= 0 {
			rest := query[idx+len(prefix):]
			if end := strings.IndexAny(rest, "( {"); end > 0 {
				opName = rest[:end]
			}
			break
		}
	}

	req := &GraphQLRequest{
		Query:         query,
		Variables:     variables,
		OperationName: opName,
	}

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request")
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", t.baseURL+graphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	// Set headers
	for k, v := range t.headers {
		httpReq.Header.Set(k, v)
	}

	// Set auth header
	if t.session != nil && t.session.Token != "" {
		httpReq.Header.Set(authHeaderKey, fmt.Sprintf("Token %s", t.session.Token))
	}

	// Add device UUID if in session
	if t.session != nil && t.session.DeviceUUID != "" {
		httpReq.Header.Set("device-uuid", t.session.DeviceUUID)
	}

	// Call request hook
	if t.hooks != nil && t.hooks.OnRequest != nil {
		t.hooks.OnRequest(ctx, httpReq)
	}

	// Log request
	if t.logger != nil {
		t.logger.Debug("GraphQL request", "query", truncateQuery(query), "variables", variables)
	}

	// Execute request
	start := time.Now()
	resp, err := t.doRequest(httpReq)
	duration := time.Since(start)

	if err != nil {
		if t.hooks != nil && t.hooks.OnError != nil {
			t.hooks.OnError(ctx, err)
		}
		return err
	}
	defer resp.Body.Close()

	// Call response hook
	if t.hooks != nil && t.hooks.OnResponse != nil {
		t.hooks.OnResponse(ctx, resp, duration)
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response")
	}

	// Log response
	if t.logger != nil {
		t.logger.Debug("GraphQL response", "status", resp.StatusCode, "duration", duration, "size", len(respBody))
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return t.handleHTTPError(resp.StatusCode, respBody)
	}

	// Parse response
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return errors.Wrap(err, "failed to parse response")
	}

	// Check for GraphQL errors
	if len(gqlResp.Errors) > 0 {
		return &types.GraphQLErrors{Errors: gqlResp.Errors}
	}

	// Unmarshal data
	if result != nil && len(gqlResp.Data) > 0 {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return errors.Wrap(err, "failed to unmarshal result")
		}
	}

	return nil
}

// SetAuth sets the authentication token
func (t *GraphQLTransport) SetAuth(token string) {
	if t.session == nil {
		t.session = &types.Session{}
	}
	t.session.Token = token
}

// SetSession sets the session
func (t *GraphQLTransport) SetSession(session *types.Session) {
	t.session = session
}

// doRequest executes the HTTP request with retry if configured
func (t *GraphQLTransport) doRequest(req *http.Request) (*http.Response, error) {
	if t.retryClient != nil {
		// Convert to retryable request
		retryReq, err := retryablehttp.FromRequest(req)
		if err != nil {
			return nil, err
		}
		return t.retryClient.Do(retryReq)
	}
	return t.httpClient.Do(req)
}

// handleHTTPError handles HTTP errors
func (t *GraphQLTransport) handleHTTPError(statusCode int, body []byte) error {
	// Try to parse error response
	var errResp struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    string `json:"error_code"`
	}

	_ = json.Unmarshal(body, &errResp)

	// Map status codes to errors
	switch statusCode {
	case http.StatusUnauthorized:
		if errResp.Code == "MFA_REQUIRED" {
			return types.ErrMFARequired
		}
		return types.ErrNotAuthenticated
	case http.StatusForbidden:
		return types.ErrNotAuthenticated
	case http.StatusNotFound:
		return types.ErrNotFound
	case http.StatusTooManyRequests:
		return types.ErrRateLimited
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return types.ErrTimeout
	case http.StatusBadRequest:
		msg := errResp.Message
		if msg == "" {
			msg = errResp.Error
		}
		if msg == "" {
			// Include raw body for debugging when no structured error found
			msg = string(body)
			if len(msg) > 500 {
				msg = msg[:500]
			}
		}
		return &types.Error{
			Code:       "BAD_REQUEST",
			Message:    msg,
			StatusCode: statusCode,
		}
	default:
		if statusCode >= 500 {
			// Build informative error message for server errors
			msg := errResp.Message
			if msg == "" {
				msg = errResp.Error
			}

			// Create base message with status code and description
			baseMsg := fmt.Sprintf("server error: %d", statusCode)
			if desc := httpStatusDescription(statusCode); desc != "" {
				baseMsg = fmt.Sprintf("server error: %d (%s)", statusCode, desc)
			}

			// Append parsed error message if available
			if msg != "" {
				baseMsg = fmt.Sprintf("%s: %s", baseMsg, msg)
			}

			return &types.Error{
				Code:       "SERVER_ERROR",
				Message:    baseMsg,
				StatusCode: statusCode,
				Err:        types.ErrServerError,
			}
		}
		return &types.Error{
			Code:       "HTTP_ERROR",
			Message:    fmt.Sprintf("HTTP error: %d", statusCode),
			StatusCode: statusCode,
		}
	}
}

// httpStatusDescription returns a human-readable description for common HTTP status codes.
// This helps users understand errors like 525 (SSL Handshake Failed) which are Cloudflare-specific.
func httpStatusDescription(statusCode int) string {
	descriptions := map[int]string{
		500: "Internal Server Error",
		501: "Not Implemented",
		502: "Bad Gateway",
		503: "Service Unavailable",
		504: "Gateway Timeout",
		520: "Web Server Error",
		521: "Web Server Is Down",
		522: "Connection Timed Out",
		523: "Origin Is Unreachable",
		524: "A Timeout Occurred",
		525: "SSL Handshake Failed",
		526: "Invalid SSL Certificate",
		527: "Railgun Error",
		530: "Origin DNS Error",
	}
	return descriptions[statusCode]
}

// truncateQuery truncates long queries for logging
func truncateQuery(query string) string {
	const maxLen = 100
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}

// Options for GraphQL transport
type Options struct {
	BaseURL     string
	HTTPClient  *http.Client
	Headers     map[string]string
	RetryConfig *types.RetryConfig
	Logger      types.Logger
	Hooks       *types.Hooks
}

// retryLogger adapts our logger to retryablehttp
type retryLogger struct {
	logger types.Logger
}

func (l *retryLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, keysAndValues...)
}

func (l *retryLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

func (l *retryLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, keysAndValues...)
}

func (l *retryLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(msg, keysAndValues...)
}
