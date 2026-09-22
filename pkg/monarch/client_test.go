package monarch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturedRequest records the auth-relevant headers a request arrived with.
type capturedRequest struct {
	authorization string
	cookie        string
	csrfToken     string
}

func newAuthCapturingServer(t *testing.T, capture *capturedRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture.authorization = r.Header.Get("Authorization")
		capture.cookie = r.Header.Get("Cookie")
		capture.csrfToken = r.Header.Get("X-CSRFToken")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"accounts":[]}}`))
	}))
}

func TestNewClientWithToken_SendsBearerAuthHeader(t *testing.T) {
	var captured capturedRequest
	server := newAuthCapturingServer(t, &captured)
	defer server.Close()

	client, err := NewClient(&ClientOptions{BaseURL: server.URL, Token: "test-token"})
	require.NoError(t, err)

	_, err = client.Accounts.List(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "Token test-token", captured.authorization)
	assert.Empty(t, captured.cookie)
}

func TestNewClientWithCookie_SendsCookieAndCSRFHeader(t *testing.T) {
	var captured capturedRequest
	server := newAuthCapturingServer(t, &captured)
	defer server.Close()

	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
		Cookie:  "sessionid=abc123; csrftoken=xyz789; other=ignored",
	})
	require.NoError(t, err)

	_, err = client.Accounts.List(context.Background())
	require.NoError(t, err)

	assert.Empty(t, captured.authorization)
	assert.Equal(t, "sessionid=abc123; csrftoken=xyz789; other=ignored", captured.cookie)
	assert.Equal(t, "xyz789", captured.csrfToken)
}

func TestNewClient_CookiePreferredOverToken(t *testing.T) {
	var captured capturedRequest
	server := newAuthCapturingServer(t, &captured)
	defer server.Close()

	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
		Token:   "test-token",
		Cookie:  "sessionid=abc123; csrftoken=xyz789",
	})
	require.NoError(t, err)

	_, err = client.Accounts.List(context.Background())
	require.NoError(t, err)

	assert.Empty(t, captured.authorization, "cookie auth should suppress the Authorization header")
	assert.Equal(t, "sessionid=abc123; csrftoken=xyz789", captured.cookie)
}

func TestNewClientWithCookie_Helper(t *testing.T) {
	// NewClientWithCookie is a thin wrapper around NewClient(&ClientOptions{Cookie: ...}),
	// whose header-sending behavior is covered by TestNewClientWithCookie_SendsCookieAndCSRFHeader.
	client, err := NewClientWithCookie("sessionid=abc123; csrftoken=xyz789")
	require.NoError(t, err)
	assert.NotNil(t, client)
}
