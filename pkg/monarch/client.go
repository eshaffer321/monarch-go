package monarch

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/eshaffer321/monarch-go/v2/internal/graphql"
	"github.com/eshaffer321/monarch-go/v2/internal/transport"
	internalTypes "github.com/eshaffer321/monarch-go/v2/internal/types"
	"github.com/getsentry/sentry-go"
)

const (
	// DefaultBaseURL is the default Monarch API base URL
	DefaultBaseURL = "https://api.monarch.com"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// UserAgent is the user agent string
	UserAgent = "monarch-go/2.0.0"
)

// Client is the main Monarch API client
type Client struct {
	// Service interfaces
	Accounts     AccountService
	Transactions TransactionService
	Tags         TagService
	Budgets      BudgetService
	Cashflow     CashflowService
	Recurring    RecurringService
	Institutions InstitutionService
	Admin        AdminService
	Auth         AuthService
	Subscription SubscriptionService

	// Internal fields
	baseURL     string
	httpClient  *http.Client
	transport   Transport
	options     *ClientOptions
	session     *Session
	queryLoader *graphql.QueryLoader
}

// ClientOptions configures the client
type ClientOptions struct {
	// BaseURL overrides the default API base URL
	BaseURL string

	// HTTPClient allows using a custom HTTP client
	HTTPClient *http.Client

	// Timeout sets the HTTP client timeout
	Timeout time.Duration

	// Token provides direct authentication token
	Token string

	// Cookie provides browser-session-cookie authentication (e.g. copied from
	// DevTools: "sessionid=...; csrftoken=..."). Preferred over Token when both
	// are set — Monarch's bearer-token login is now blocked by bot protection
	// for scripted clients, while a copied session cookie still works.
	Cookie string

	// SessionFile path for session persistence
	SessionFile string

	// UserAgent overrides the default user-agent string for auth requests
	UserAgent string

	// Logger for debug logging
	Logger Logger

	// RetryConfig configures retry behavior
	RetryConfig *internalTypes.RetryConfig

	// RateLimiter for rate limiting
	RateLimiter RateLimiter

	// Hooks for observability
	Hooks *internalTypes.Hooks

	// SentryDSN enables Sentry error tracking when set
	SentryDSN string

	// SentryOptions allows custom Sentry configuration
	SentryOptions *sentry.ClientOptions
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// RateLimiter interface for rate limiting
type RateLimiter interface {
	Wait(ctx context.Context) error
}

// Transport handles HTTP/GraphQL communication
type Transport interface {
	Execute(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error
	SetAuth(token string)
	SetCookie(cookie string)
	SetSession(session *internalTypes.Session)
}

// NewClient creates a new Monarch client
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		opts = &ClientOptions{}
	}

	// Initialize Sentry if DSN is provided
	if opts.SentryDSN != "" || opts.SentryOptions != nil {
		sentryOpts := sentry.ClientOptions{}

		// Use provided options if available, otherwise create new ones
		if opts.SentryOptions != nil {
			sentryOpts = *opts.SentryOptions
		}

		// Override DSN if provided separately
		if opts.SentryDSN != "" {
			sentryOpts.Dsn = opts.SentryDSN
		}

		// Set default environment if not provided
		if sentryOpts.Environment == "" {
			sentryOpts.Environment = "production"
		}

		// Initialize Sentry
		if err := sentry.Init(sentryOpts); err != nil {
			// Log error but don't fail client creation
			if opts.Logger != nil {
				opts.Logger.Error("Failed to initialize Sentry", "error", err)
			}
		}
	}

	// Set defaults
	if opts.BaseURL == "" {
		opts.BaseURL = DefaultBaseURL
	}

	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}

	if opts.Timeout > 0 {
		opts.HTTPClient.Timeout = opts.Timeout
	}

	// Create transport using the internal package
	transportOpts := &transport.Options{
		BaseURL:     opts.BaseURL,
		HTTPClient:  opts.HTTPClient,
		RetryConfig: opts.RetryConfig,
		Logger:      opts.Logger,
		Hooks:       opts.Hooks,
	}
	trans := transport.NewGraphQLTransport(transportOpts)

	// Set auth if provided — cookie takes precedence over token (see Cookie doc comment)
	if opts.Token != "" {
		trans.SetAuth(opts.Token)
	}
	if opts.Cookie != "" {
		trans.SetCookie(opts.Cookie)
	}

	// Create client
	c := &Client{
		baseURL:     opts.BaseURL,
		httpClient:  opts.HTTPClient,
		transport:   trans,
		options:     opts,
		queryLoader: graphql.NewQueryLoader(),
	}

	// Initialize services
	c.initServices()

	// Load session if file specified
	if opts.SessionFile != "" {
		if err := c.loadSession(opts.SessionFile); err != nil && opts.Logger != nil {
			opts.Logger.Warn("Failed to load session", "error", err)
		}
	}

	return c, nil
}

// NewClientWithToken creates a client with an auth token
func NewClientWithToken(token string) (*Client, error) {
	return NewClient(&ClientOptions{
		Token: token,
	})
}

// NewClientWithCookie creates a client authenticated with a browser session
// cookie (e.g. copied from DevTools) instead of a bearer token.
func NewClientWithCookie(cookie string) (*Client, error) {
	return NewClient(&ClientOptions{
		Cookie: cookie,
	})
}

// initServices initializes all service implementations
func (c *Client) initServices() {
	// Create service implementations
	c.Accounts = &accountService{client: c}
	c.Transactions = newTransactionService(c)
	c.Tags = &tagService{client: c}
	c.Budgets = &budgetService{client: c}
	c.Cashflow = &cashflowService{client: c}
	c.Recurring = &recurringService{client: c}
	c.Institutions = &institutionService{client: c}
	c.Subscription = &subscriptionService{client: c}
	c.Admin = &adminService{client: c}
	c.Auth = newAuthService(c)
}

// WithContext returns a new client with the given context
func (c *Client) WithContext(ctx context.Context) *Client {
	// This allows for context-specific client instances
	// Useful for request-scoped configuration
	newClient := *c
	return &newClient
}

// SetToken sets the authentication token
func (c *Client) SetToken(token string) {
	c.transport.SetAuth(token)
	if c.session == nil {
		c.session = &Session{}
	}
	c.session.Token = token
}

// GetSession returns the current session
func (c *Client) GetSession() *Session {
	return c.session
}

// loadQuery loads a GraphQL query from the embedded filesystem
func (c *Client) loadQuery(queryPath string) string {
	query, err := c.queryLoader.Load(queryPath)
	if err != nil {
		// This should never happen in production as queries are embedded
		panic(fmt.Sprintf("failed to load query %s: %v", queryPath, err))
	}
	return query
}

// loadSession loads session from file
func (c *Client) loadSession(path string) error {
	if c.Auth != nil {
		return c.Auth.LoadSession(path)
	}
	return nil
}

// executeGraphQL executes a GraphQL query
func (c *Client) executeGraphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	// Add hooks
	if c.options.Hooks != nil && c.options.Hooks.OnRequest != nil {
		// Create pseudo request for hook
		req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/graphql", nil)
		c.options.Hooks.OnRequest(ctx, req)
	}

	// Rate limiting
	if c.options.RateLimiter != nil {
		if err := c.options.RateLimiter.Wait(ctx); err != nil {
			// Capture rate limiter errors in Sentry
			if hub := sentry.GetHubFromContext(ctx); hub != nil {
				hub.CaptureException(err)
			} else {
				sentry.CaptureException(err)
			}
			return fmt.Errorf("rate limiter: %w", err)
		}
	}

	// Execute query
	start := time.Now()
	err := c.transport.Execute(ctx, query, variables, result)
	duration := time.Since(start)

	// Capture errors in Sentry
	if err != nil {
		// Add context to Sentry
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("graphql.operation", extractOperationName(query))
				scope.SetContext("graphql", map[string]interface{}{
					"query":     query,
					"variables": variables,
					"duration":  duration.String(),
				})
				hub.CaptureException(err)
			})
		} else {
			sentry.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("graphql.operation", extractOperationName(query))
				scope.SetContext("graphql", map[string]interface{}{
					"query":     query,
					"variables": variables,
					"duration":  duration.String(),
				})
				sentry.CaptureException(err)
			})
		}
	}

	// Response hook
	if c.options.Hooks != nil && c.options.Hooks.OnResponse != nil {
		// Create pseudo response for hook
		resp := &http.Response{StatusCode: 200}
		if err != nil {
			resp.StatusCode = 500
		}
		c.options.Hooks.OnResponse(ctx, resp, duration)
	}

	// Error hook
	if err != nil && c.options.Hooks != nil && c.options.Hooks.OnError != nil {
		c.options.Hooks.OnError(ctx, err)
	}

	return err
}

// Close flushes any pending Sentry events and performs cleanup
func (c *Client) Close() {
	// Flush Sentry events with a 2 second timeout
	sentry.Flush(2 * time.Second)
}

// extractOperationName extracts the GraphQL operation name from a query
func extractOperationName(query string) string {
	// Simple extraction - looks for "query OperationName" or "mutation OperationName"
	// This is a basic implementation; you might want to use a proper GraphQL parser
	for _, prefix := range []string{"query ", "mutation ", "subscription "} {
		if idx := findOperationName(query, prefix); idx != "" {
			return idx
		}
	}
	return "unknown"
}

// findOperationName finds the operation name after a given prefix
func findOperationName(query, prefix string) string {
	idx := 0
	for {
		pos := indexAt(query, prefix, idx)
		if pos == -1 {
			return ""
		}

		// Skip the prefix
		start := pos + len(prefix)

		// Find the end of the operation name (space, parenthesis, or brace)
		end := start
		for end < len(query) {
			ch := query[end]
			if ch == ' ' || ch == '(' || ch == '{' || ch == '\n' || ch == '\r' {
				break
			}
			end++
		}

		if end > start {
			name := query[start:end]
			// Validate it's a valid operation name (starts with letter or underscore)
			if len(name) > 0 && (isLetter(name[0]) || name[0] == '_') {
				return name
			}
		}

		idx = pos + 1
	}
}

// indexAt finds the index of substr in s starting at position start
func indexAt(s, substr string, start int) int {
	if start >= len(s) {
		return -1
	}
	idx := stringIndex(s[start:], substr)
	if idx == -1 {
		return -1
	}
	return start + idx
}

// stringIndex is a simple string index function
func stringIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// isLetter checks if a byte is a letter
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}
