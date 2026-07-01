package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/eshaffer321/monarch-go/v2/pkg/monarch"
	"github.com/getsentry/sentry-go"
)

func main() {
	// Example 1: Using SentryDSN option
	client, err := monarch.NewClient(&monarch.ClientOptions{
		Token:     os.Getenv("MONARCH_TOKEN"),
		SentryDSN: os.Getenv("SENTRY_DSN"), // Your Sentry DSN
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close() // Important: flush Sentry events on exit

	// Example 2: Using custom Sentry options
	clientWithCustomSentry, err := monarch.NewClient(&monarch.ClientOptions{
		Token: os.Getenv("MONARCH_TOKEN"),
		SentryOptions: &sentry.ClientOptions{
			Dsn:         os.Getenv("SENTRY_DSN"),
			Environment: "development",
			Release:     "monarch-go@1.0.0",
			Debug:       true,
			SampleRate:  1.0,
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				// Custom event processing
				// For example, scrub sensitive data
				if event.Request != nil {
					event.Request.Cookies = ""
				}
				return event
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer clientWithCustomSentry.Close()

	// Example 3: Using Sentry with context
	ctx := context.Background()

	// Create a hub for this context
	hub := sentry.CurrentHub().Clone()
	ctx = sentry.SetHubOnContext(ctx, hub)

	// Add user context
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       "user123",
			Email:    "user@example.com",
			Username: "exampleuser",
		})
		scope.SetTag("service", "monarch-client")
		scope.SetLevel(sentry.LevelInfo)
	})

	// Use the client - errors will automatically be sent to Sentry
	accounts, err := client.Accounts.List(ctx)
	if err != nil {
		// Error is automatically captured by the client
		// but you can also manually capture with additional context
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("operation", "list_accounts")
			scope.SetContext("custom", map[string]interface{}{
				"retry_count": 3,
				"user_action": "viewing dashboard",
			})
			sentry.CaptureException(err)
		})

		log.Printf("Failed to list accounts: %v", err)
		return
	}

	fmt.Printf("Found %d accounts\n", len(accounts))

	// Example 4: Capture custom messages
	sentry.CaptureMessage("User successfully listed accounts")

	// Example 5: Breadcrumbs
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "auth",
		Message:  "User authenticated successfully",
		Level:    sentry.LevelInfo,
	})

	// Transactions will have GraphQL context automatically added
	transactions, err := client.Transactions.Query().
		Limit(10).
		Execute(ctx)
	if err != nil {
		// Error captured with GraphQL query context
		log.Printf("Failed to query transactions: %v", err)
		return
	}

	fmt.Printf("Found %d transactions\n", len(transactions.Transactions))
}
