package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eshaffer321/monarch-go/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleHTTPError_ServerError_IncludesResponseBody(t *testing.T) {
	transport := &GraphQLTransport{}

	tests := []struct {
		name           string
		statusCode     int
		responseBody   []byte
		expectedInMsg  string
		expectedCode   string
	}{
		{
			name:          "525 SSL Handshake Failed with HTML body",
			statusCode:    525,
			responseBody:  []byte(`<html><body>SSL Handshake Failed</body></html>`),
			expectedInMsg: "525",
			expectedCode:  "SERVER_ERROR",
		},
		{
			name:          "500 with JSON error message",
			statusCode:    500,
			responseBody:  []byte(`{"error": "Internal server error", "message": "Database connection failed"}`),
			expectedInMsg: "Database connection failed",
			expectedCode:  "SERVER_ERROR",
		},
		{
			name:          "502 Bad Gateway with empty body",
			statusCode:    502,
			responseBody:  []byte{},
			expectedInMsg: "502",
			expectedCode:  "SERVER_ERROR",
		},
		{
			name:          "503 Service Unavailable",
			statusCode:    503,
			responseBody:  []byte(`Service temporarily unavailable`),
			expectedInMsg: "503",
			expectedCode:  "SERVER_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transport.handleHTTPError(tt.statusCode, tt.responseBody)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedInMsg, "error should contain status code or message")

			// For JSON responses with message field, should include the message
			if tt.statusCode == 500 && len(tt.responseBody) > 0 {
				assert.Contains(t, err.Error(), "Database connection failed", "should include parsed error message")
			}
		})
	}
}

func TestHandleHTTPError_ServerError_IncludesStatusCodeDescription(t *testing.T) {
	transport := &GraphQLTransport{}

	tests := []struct {
		name         string
		statusCode   int
		expectedDesc string
	}{
		{"500 Internal Server Error", 500, "Internal Server Error"},
		{"502 Bad Gateway", 502, "Bad Gateway"},
		{"503 Service Unavailable", 503, "Service Unavailable"},
		{"525 SSL Handshake Failed", 525, "SSL Handshake Failed"},
		{"526 Invalid SSL Certificate", 526, "Invalid SSL Certificate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transport.handleHTTPError(tt.statusCode, []byte(`error page`))

			assert.Error(t, err)
			errMsg := err.Error()
			assert.Contains(t, errMsg, tt.expectedDesc, "error should include human-readable description")
		})
	}
}

func TestHandleHTTPError_BadRequest_RawBody(t *testing.T) {
	transport := &GraphQLTransport{}

	tests := []struct {
		name         string
		body         []byte
		expectedMsg  string
	}{
		{
			name:        "structured JSON error",
			body:        []byte(`{"message": "invalid field"}`),
			expectedMsg: "invalid field",
		},
		{
			name:        "no message, uses error field",
			body:        []byte(`{"error": "bad input"}`),
			expectedMsg: "bad input",
		},
		{
			name:        "no structured error, falls back to raw body",
			body:        []byte(`something went wrong with processing`),
			expectedMsg: "something went wrong with processing",
		},
		{
			name:        "empty body, raw body fallback is empty string",
			body:        []byte{},
			expectedMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transport.handleHTTPError(http.StatusBadRequest, tt.body)
			assert.Error(t, err)
			typedErr, ok := err.(*types.Error)
			require.True(t, ok)
			assert.Equal(t, "BAD_REQUEST", typedErr.Code)
			assert.Contains(t, typedErr.Message, tt.expectedMsg)
		})
	}
}

func TestNewGraphQLTransport_DefaultHeaders(t *testing.T) {
	transport := NewGraphQLTransport(nil)

	assert.Equal(t, "https://api.monarch.com", transport.baseURL)
	assert.Equal(t, "https://app.monarch.com", transport.headers["Origin"])
	assert.Equal(t, "https://app.monarch.com/", transport.headers["Referer"])
	assert.Equal(t, types.UserAgent, transport.headers["User-Agent"])
}

func TestNewGraphQLTransport_CustomHeaders(t *testing.T) {
	transport := NewGraphQLTransport(&Options{
		Headers: map[string]string{
			"X-Custom": "test-value",
		},
	})

	assert.Equal(t, "test-value", transport.headers["X-Custom"])
	// Default headers should still be present
	assert.Equal(t, "https://app.monarch.com", transport.headers["Origin"])
}

func TestExecute_OperationName(t *testing.T) {
	// Set up a test server that captures the request
	var receivedReq GraphQLRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"test": "ok"},
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	transport := NewGraphQLTransport(&Options{
		BaseURL: server.URL,
	})
	transport.session = &types.Session{
		Token:     "test-token",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{"mutation", "mutation CreateTransaction($input: Input!) { create { id } }", "CreateTransaction"},
		{"query", "query GetAccounts { accounts { id } }", "GetAccounts"},
		{"no operation name", "{ accounts { id } }", ""},
		{"mutation with parens", "mutation Common_Create($x: X) { create { id } }", "Common_Create"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receivedReq = GraphQLRequest{} // reset between tests
			var result map[string]interface{}
			err := transport.Execute(context.Background(), tt.query, nil, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, receivedReq.OperationName)
		})
	}
}

func TestExecute_NotAuthenticated(t *testing.T) {
	transport := NewGraphQLTransport(nil)
	// No session set
	err := transport.Execute(context.Background(), "query { test }", nil, nil)
	assert.ErrorIs(t, err, types.ErrNotAuthenticated)
}

func TestExecute_SessionExpired(t *testing.T) {
	transport := NewGraphQLTransport(nil)
	transport.session = &types.Session{
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	err := transport.Execute(context.Background(), "query { test }", nil, nil)
	assert.ErrorIs(t, err, types.ErrSessionExpired)
}

func TestTruncateQuery(t *testing.T) {
	short := "query { test }"
	assert.Equal(t, short, truncateQuery(short))

	long := ""
	for i := 0; i < 200; i++ {
		long += "x"
	}
	truncated := truncateQuery(long)
	assert.Equal(t, 103, len(truncated)) // 100 + "..."
	assert.Contains(t, truncated, "...")
}

func TestSetAuth(t *testing.T) {
	transport := NewGraphQLTransport(nil)
	transport.SetAuth("my-token")
	assert.Equal(t, "my-token", transport.session.Token)
}

func TestSetSession(t *testing.T) {
	transport := NewGraphQLTransport(nil)
	session := &types.Session{Token: "tok", UserID: "uid"}
	transport.SetSession(session)
	assert.Equal(t, session, transport.session)
}

func TestSetCookie(t *testing.T) {
	transport := NewGraphQLTransport(nil)
	transport.SetCookie("sessionid=abc; csrftoken=def")
	assert.Equal(t, "sessionid=abc; csrftoken=def", transport.session.Cookie)
}

func TestExecute_NotAuthenticated_CookieOnlySessionPasses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"test":"ok"}}`))
	}))
	defer server.Close()

	transport := NewGraphQLTransport(&Options{BaseURL: server.URL})
	transport.session = &types.Session{Cookie: "sessionid=abc; csrftoken=def"}

	var result map[string]interface{}
	err := transport.Execute(context.Background(), "query { test }", nil, &result)
	assert.NoError(t, err)
}

func TestExecute_CookieAuth_SetsCookieAndCSRFHeaders(t *testing.T) {
	var gotCookie, gotCSRF, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		gotCSRF = r.Header.Get("X-CSRFToken")
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"test":"ok"}}`))
	}))
	defer server.Close()

	transport := NewGraphQLTransport(&Options{BaseURL: server.URL})
	transport.session = &types.Session{Cookie: "sessionid=abc123; csrftoken=xyz789; other=ignored"}

	var result map[string]interface{}
	err := transport.Execute(context.Background(), "query { test }", nil, &result)
	require.NoError(t, err)

	assert.Equal(t, "sessionid=abc123; csrftoken=xyz789; other=ignored", gotCookie)
	assert.Equal(t, "xyz789", gotCSRF)
	assert.Empty(t, gotAuth, "Authorization header should not be sent when using cookie auth")
}

func TestExecute_CookiePreferredOverToken(t *testing.T) {
	var gotCookie, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"test":"ok"}}`))
	}))
	defer server.Close()

	transport := NewGraphQLTransport(&Options{BaseURL: server.URL})
	transport.session = &types.Session{Token: "test-token", Cookie: "sessionid=abc123; csrftoken=xyz789"}

	var result map[string]interface{}
	err := transport.Execute(context.Background(), "query { test }", nil, &result)
	require.NoError(t, err)

	assert.Equal(t, "sessionid=abc123; csrftoken=xyz789", gotCookie)
	assert.Empty(t, gotAuth)
}

func TestExtractCookieValue(t *testing.T) {
	tests := []struct {
		name   string
		cookie string
		key    string
		want   string
	}{
		{"simple", "csrftoken=abc123", "csrftoken", "abc123"},
		{"multiple cookies", "sessionid=xyz; csrftoken=abc123; other=val", "csrftoken", "abc123"},
		{"missing key", "sessionid=xyz; other=val", "csrftoken", ""},
		{"empty cookie string", "", "csrftoken", ""},
		{"trailing semicolon", "sessionid=xyz; csrftoken=abc123;", "csrftoken", "abc123"},
		{"value containing equals", "csrftoken=abc=123==", "csrftoken", "abc=123=="},
		{"extra whitespace", "sessionid=xyz;   csrftoken=abc123", "csrftoken", "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractCookieValue(tt.cookie, tt.key))
		})
	}
}
