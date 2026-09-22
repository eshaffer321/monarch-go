package types

import (
	"context"
	"net/http"
	"time"
)

// Session represents an authenticated session
type Session struct {
	Token      string    `json:"token"`
	Cookie     string    `json:"cookie,omitempty"`
	UserID     string    `json:"userId"`
	Email      string    `json:"email"`
	ExpiresAt  time.Time `json:"expiresAt"`
	DeviceUUID string    `json:"deviceUuid"`
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries int           `json:"maxRetries"`
	RetryWait  time.Duration `json:"retryWait"`
	MaxWait    time.Duration `json:"maxWait"`
}

// Hooks provides lifecycle hooks for requests
type Hooks struct {
	OnRequest  func(ctx context.Context, req *http.Request)
	OnResponse func(ctx context.Context, resp *http.Response, duration time.Duration)
	OnError    func(ctx context.Context, err error)
}
