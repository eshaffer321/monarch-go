# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.1.0] - 2026-09-21

### Added
- Added cookie-based authentication via `ClientOptions.Cookie` / `NewClientWithCookie`, for use
  with a browser-copied Monarch session cookie (`sessionid=...; csrftoken=...`). Preferred over
  `Token` when both are set — Monarch's bearer-token/scripted login path is increasingly blocked
  by bot protection, while a copied session cookie continues to work and reportedly lasts longer.
  Cookie auth sends `Cookie` and, when present, `X-CSRFToken` (extracted from the cookie's
  `csrftoken` value) instead of the `Authorization: Token` header.

## [2.0.0] - 2026-07-01

### Changed
- ⚠️ **Breaking**: Module path renamed from `github.com/eshaffer321/monarchmoney-go` to
  `github.com/eshaffer321/monarch-go/v2`, matching Monarch's rebrand from "Monarch Money" to
  "Monarch" (the `/v2` suffix is required by Go's module versioning rules for a v2+ release).
  Consumers must update their import paths and `go get` command:
  `go get github.com/eshaffer321/monarch-go/v2`.
- Rebranded docs, comments, and the client `UserAgent` string (now `monarch-go/2.0.0`) from
  "Monarch Money" to "Monarch."

## [1.1.0] - 2026-05-21

### Added
- Added configurable auth `UserAgent` support via `ClientOptions.UserAgent`.
- Added `AuthService.LoginWithEmailOTP` for email one-time-password login flows.
- Added investment holdings management:
  - `Accounts.SearchSecurities`
  - `Accounts.CreateHolding`
  - `Accounts.CreateHoldingByTicker`
  - `Accounts.UpdateHoldingQuantity`
  - `Accounts.DeleteHolding`
- Added manual investments account creation with initial holdings via `Accounts.CreateInvestmentsAccount`.
- Added GraphQL `operationName` extraction and submission in transport requests.

### Changed
- Updated auth request headers to use the current Monarch app origin and referer.
- `Transactions.Create` now uses Monarch's current create transaction mutation shape, including `merchantName` instead of the previous `merchant` object.
- `Transactions.Create` no longer performs a follow-up `Get` after creating a transaction. It now returns the created transaction ID plus request-derived fields to avoid false failures from Monarch's `getTransaction` endpoint.
- `Transactions.Create` rounds amounts to two decimal places before sending them to Monarch.
- `Accounts.GetHoldings` now uses the account-based holdings endpoint and has a more resilient price fallback chain.
- `Accounts.UpdateHoldingQuantity` now uses the direct `updateHolding` mutation instead of delete-and-recreate, preserving holding metadata.
- `Accounts.CreateHoldingByTicker` now requires a case-insensitive exact ticker match instead of silently using the first search result.

### Fixed
- Fixed `Transactions.Create` by matching Monarch's current mutation schema and response format.
- Fixed `Transactions.Create` to require `CategoryID` locally with a clear error before calling Monarch, avoiding Monarch's vague backend `BAD_REQUEST`.
- Improved `Transactions.Create` mutation error messages by parsing `fieldErrors`.
- Improved `BAD_REQUEST` transport errors by including the raw response body when Monarch does not provide a structured message.

## [1.0.5] - 2026-01-17

### Fixed
- **CRITICAL**: Updated API base URL from `api.monarchmoney.com` to `api.monarch.com`
  - Monarch Money migrated their API to a new domain
  - The old domain was returning 525 SSL Handshake Failed errors
  - This fix restores API connectivity

## [1.0.4] - 2026-01-17

### Improved
- Enhanced 5xx server error messages with human-readable status code descriptions
  - Error messages now include descriptions like "SSL Handshake Failed" for 525 errors
  - Common Cloudflare error codes (520-530) are now explained in error messages
  - JSON error responses from the server are now included in 5xx error messages
  - Example: `server error: 525` is now `server error: 525 (SSL Handshake Failed)`
- Added comprehensive tests for HTTP error handling

### Added
- New `httpStatusDescription()` helper function for translating HTTP status codes to descriptions
- Test coverage for transport layer error handling

## [1.0.3] - 2025-11-26

### Fixed
- Removed unsupported `WithMerchant()` filter from transaction queries to prevent `400 BAD_REQUEST` errors
  - The MonarchMoney GraphQL API does not support filtering by merchant name via `TransactionFilterInput`
  - Users should use `Search()` method instead, which searches across merchant names and other text fields
  - Aligns Go client with Python reference implementation

### Breaking Changes
- ⚠️ Removed `WithMerchant()` method from `TransactionQueryBuilder` interface
  - Migration: Replace `.WithMerchant("name")` with `.Search("name")`
  - The `Search()` method provides equivalent functionality

## [1.0.2] - 2025-10-26

### Fixed
- **CRITICAL**: Fixed `Transactions.Get()` method failing with `BAD_REQUEST` error by:
  - Adding required `redirectPosted` parameter to GraphQL query (defaults to `true` matching Python client behavior)
  - Correcting field name from `splits` to `splitTransactions` in query response
  - Removing invalid fields (`reviewStatus`, `createdAt`, `updatedAt`, `originalDescription`) that don't exist in Monarch API
  - Adding `hasSplitTransactions` field to properly detect split transactions
  - Adding `__typename` fields following GraphQL best practices
- Enhanced `UpdateSplits()` mutation response to include parent transaction `amount` field for verification

### Notes
- `Transactions.Update()` was working correctly all along - the issue was that `Get()` couldn't fetch transactions to verify updates
- All tests pass with updated GraphQL schema

## [1.0.1] - 2025-10-25

### Fixed
- Fixed `Transactions.Delete()` method returning `BAD_REQUEST` error by updating GraphQL mutation format to match Python client and Monarch API expectations. The mutation now uses `DeleteTransactionMutationInput` with `transactionId` field wrapped in `input` parameter instead of passing UUID directly.

### Added
- Added comprehensive test coverage for `Transactions.Delete()` including error handling scenarios
- Added example documentation for transaction deletion and `hideFromReports` workaround (`examples/transaction_deletion/main.go`)

### Documentation
- Documented `HideFromReports` field as an alternative to deletion for bank-imported transactions that cannot be deleted
- Added detailed examples for consolidating multi-delivery orders (e.g., Walmart split deliveries)

## [1.0.0] - 2025-10-20

### Added

**Core Services**
- Full authentication support (Login, MFA, TOTP, session management)
- Account service with listing, creation, updates, and refresh capabilities
- Transaction service with querying, filtering, updates, splits, and streaming
- Budget service with listing, updates, and goals (goalsV2) support
- Cashflow service with summary and category-level details
- Institution service for managing financial institution connections
- Subscription service for managing Monarch Money subscriptions

**Advanced Features**
- MCP (Model Context Protocol) server for AI integration with Monarch Money
- Goals tracking with rollover amounts and types
- Transaction splits support with comprehensive error handling
- Sentry integration for error tracking and monitoring
- Rate limiting support with configurable limits
- Retry logic with exponential backoff
- Hooks for observability (OnRequest, OnResponse, OnError)
- Session management with file-based persistence

**Developer Experience**
- Comprehensive test coverage with mocked responses
- Full GoDoc documentation
- CI/CD with GitHub Actions
- Code coverage reporting with Codecov
- Go Report Card integration
- Example implementations demonstrating all features
- Structured error handling with custom error types

**Infrastructure**
- GraphQL query loader for efficient query management
- Connection pooling and smart caching
- Context support throughout for cancellation and timeouts
- Custom date parsing to handle multiple API date formats

### Changed
- N/A (Initial release)

### Fixed
- N/A (Initial release)

### Security
- N/A (Initial release)

## [0.0.0] - Development

All development work leading up to the v1.0.0 release.

[Unreleased]: https://github.com/eshaffer321/monarch-go/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/eshaffer321/monarch-go/compare/v1.0.5...v1.1.0
[1.0.5]: https://github.com/eshaffer321/monarch-go/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/eshaffer321/monarch-go/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/eshaffer321/monarch-go/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/eshaffer321/monarch-go/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/eshaffer321/monarch-go/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/eshaffer321/monarch-go/releases/tag/v1.0.0
