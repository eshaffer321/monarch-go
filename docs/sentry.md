# Sentry Integration

The Monarch Go client includes built-in support for Sentry error tracking, providing automatic error capture with rich context for debugging.

## Features

- **Automatic Error Capture**: All GraphQL errors are automatically sent to Sentry with full context
- **GraphQL Context**: Query, variables, and operation names are included in error reports
- **Performance Tracking**: Request duration is tracked for all GraphQL operations
- **Context Propagation**: Supports Sentry hub context for request-scoped data
- **Custom Configuration**: Full control over Sentry client options

## Quick Start

### Basic Setup

```go
client, err := monarch.NewClient(&monarch.ClientOptions{
    Token:     "your-token",
    SentryDSN: "your-sentry-dsn",
})
if err != nil {
    log.Fatal(err)
}
defer client.Close() // Important: flush Sentry events
```

### Advanced Configuration

```go
client, err := monarch.NewClient(&monarch.ClientOptions{
    Token: "your-token",
    SentryOptions: &sentry.ClientOptions{
        Dsn:         "your-sentry-dsn",
        Environment: "production",
        Release:     "v1.0.0",
        Debug:       false,
        SampleRate:  0.2, // Capture 20% of transactions
        BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
            // Scrub sensitive data
            if event.Request != nil {
                event.Request.Cookies = ""
                event.Request.Headers = nil
            }
            return event
        },
    },
})
```

## Error Context

When errors occur, the following context is automatically captured:

### GraphQL Operations
- **Operation Name**: Extracted from the query (e.g., `GetAccounts`, `CreateTransaction`)
- **Query**: The full GraphQL query text
- **Variables**: Query variables (be careful with sensitive data)
- **Duration**: Time taken for the request

### Example Error in Sentry

```json
{
  "tags": {
    "graphql.operation": "GetTransactions"
  },
  "contexts": {
    "graphql": {
      "query": "query GetTransactions($limit: Int) { ... }",
      "variables": {
        "limit": 100
      },
      "duration": "250ms"
    }
  }
}
```

## Using with Context

You can add custom context to errors using Sentry's hub:

```go
ctx := context.Background()

// Create a hub for this request
hub := sentry.CurrentHub().Clone()
ctx = sentry.SetHubOnContext(ctx, hub)

// Add user context
hub.ConfigureScope(func(scope *sentry.Scope) {
    scope.SetUser(sentry.User{
        ID:    "user123",
        Email: "user@example.com",
    })
    scope.SetTag("feature", "transaction_export")
    scope.SetLevel(sentry.LevelWarning)
})

// Use the client - context will be included in errors
transactions, err := client.Transactions.Query().Execute(ctx)
```

## Manual Error Capture

While errors are automatically captured, you can also manually capture with additional context:

```go
accounts, err := client.Accounts.List(ctx)
if err != nil {
    sentry.WithScope(func(scope *sentry.Scope) {
        scope.SetTag("retry_attempt", "3")
        scope.SetContext("business_logic", map[string]interface{}{
            "user_action": "monthly_reconciliation",
            "account_count": 5,
        })
        sentry.CaptureException(err)
    })
    return err
}
```

## Breadcrumbs

Add breadcrumbs to track user actions leading to an error:

```go
sentry.AddBreadcrumb(&sentry.Breadcrumb{
    Category: "auth",
    Message:  "User authenticated",
    Level:    sentry.LevelInfo,
})

sentry.AddBreadcrumb(&sentry.Breadcrumb{
    Category: "navigation",
    Message:  "Navigated to transactions",
    Level:    sentry.LevelInfo,
})

// If an error occurs later, these breadcrumbs will be included
```

## Performance Considerations

1. **Sampling**: Use `SampleRate` to control the percentage of errors sent to Sentry
2. **BeforeSend**: Filter or modify events before sending
3. **Sensitive Data**: Be careful not to send passwords, tokens, or PII to Sentry

```go
SentryOptions: &sentry.ClientOptions{
    // Only capture 10% of transactions in production
    SampleRate: 0.1,
    
    // Filter out specific errors
    BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
        // Don't send context timeout errors
        if event.Exception != nil && len(event.Exception) > 0 {
            if strings.Contains(event.Exception[0].Value, "context deadline exceeded") {
                return nil // Don't send this event
            }
        }
        return event
    },
}
```

## Cleanup

Always call `Close()` on the client when your application shuts down:

```go
client, err := monarch.NewClient(opts)
if err != nil {
    log.Fatal(err)
}
defer client.Close() // Flushes pending events to Sentry
```

The `Close()` method will flush any pending Sentry events with a 2-second timeout.

## Environment Variables

You can use environment variables for configuration:

```go
client, err := monarch.NewClient(&monarch.ClientOptions{
    Token:     os.Getenv("MONARCH_TOKEN"),
    SentryDSN: os.Getenv("SENTRY_DSN"),
})
```

## Testing

During testing, you might want to disable Sentry:

```go
var sentryDSN string
if os.Getenv("ENV") != "test" {
    sentryDSN = os.Getenv("SENTRY_DSN")
}

client, err := monarch.NewClient(&monarch.ClientOptions{
    Token:     token,
    SentryDSN: sentryDSN, // Will be empty in test environment
})
```

## Troubleshooting

### Events Not Appearing in Sentry

1. Ensure you're calling `client.Close()` before your application exits
2. Check that your DSN is correct
3. Enable debug mode to see what's being sent:

```go
SentryOptions: &sentry.ClientOptions{
    Dsn:   "your-dsn",
    Debug: true, // Enable debug output
}
```

### Too Many Events

Use sampling and filtering to reduce the number of events:

```go
SentryOptions: &sentry.ClientOptions{
    SampleRate: 0.1, // Only 10% of events
    BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
        // Custom filtering logic
        return event
    },
}
```