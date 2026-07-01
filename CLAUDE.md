# Monarch Go Client - Developer Guide

## 🚀 Quick Development Commands

```bash
# Run all tests (most important command)
go test ./pkg/monarch/... -v

# Run with coverage  
go test ./pkg/monarch/... -coverprofile=coverage.out && go tool cover -html=coverage.out

# Run specific service tests
go test ./pkg/monarch -run TestAccountService -v
go test ./pkg/monarch -run TestTransaction -v

# Debug a failing test
go test ./pkg/monarch -run TestSpecificMethod -v -count=1

# Check for issues
go vet ./...
```

## 🧭 How to Navigate This Codebase

### When Adding New Features
1. **Start**: `pkg/monarch/interfaces.go` - add method to relevant service interface
2. **Implement**: `pkg/monarch/{service}.go` - implement the method
3. **GraphQL**: Save query in `internal/graphql/queries/{service}/`
4. **Types**: Add to `pkg/monarch/types.go` if needed
5. **Test**: Create test in `pkg/monarch/{service}_test.go`

### Key Files by Purpose
```
pkg/monarch/
├── client.go           # Main client + service initialization
├── interfaces.go       # ALL service method signatures
├── types.go           # Data structures (40+ types)
├── errors.go          # Error handling patterns
├── date.go            # Custom date parsing (important!)
└── accounts.go         # Example service implementation
    transactions.go     
    budgets.go         
    ...

internal/graphql/queries/   # GraphQL organized by service
├── accounts/
├── transactions/  
├── budgets/
...
```

## 🔧 Development Patterns

### Adding a New Service Method
```go
// 1. Add to interface (interfaces.go)
type AccountService interface {
    NewMethod(ctx context.Context, param string) (*Result, error)
}

// 2. Implement (accounts.go) 
func (s *accountService) NewMethod(ctx context.Context, param string) (*Result, error) {
    query := s.client.loadQuery("accounts/new_method.graphql")
    
    variables := map[string]interface{}{
        "param": param,
    }
    
    var result struct {
        Data *Result `json:"data"`
    }
    
    if err := s.client.executeGraphQL(ctx, query, variables, &result); err != nil {
        return nil, errors.Wrap(err, "failed to execute new method")
    }
    
    return result.Data, nil
}

// 3. Test (accounts_test.go)
func TestAccountService_NewMethod(t *testing.T) {
    mockTransport := new(MockTransport)
    client := &Client{
        transport:   mockTransport,
        queryLoader: graphql.NewQueryLoader(),
        options:     &ClientOptions{},
        baseURL:     "https://api.test.com",
    }
    client.initServices()

    response := `{"data": {"field": "value"}}`
    mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(response, nil)

    result, err := client.Accounts.NewMethod(context.Background(), "test")

    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockTransport.AssertExpectations(t)
}
```

### GraphQL Response Handling
```go
// GraphQL responses don't always match the schema exactly
// Common patterns:

// 1. Direct field mapping
var result struct {
    Accounts []*Account `json:"accounts"`
}

// 2. Nested response (common)
var result struct {
    GetAccount *Account `json:"getAccount"`
}

// 3. Array with single object (aggregates pattern)
var result struct {
    Aggregates []struct {
        Summary *TransactionSummary `json:"summary"`
    } `json:"aggregates"`
}
```

## 🐛 Common Issues & Solutions

### Test Failures
```bash
# GraphQL field mismatch
# Look for: `json: cannot unmarshal`
# Fix: Check GraphQL response format in test vs actual API

# Mock expectations failing  
# Look for: `mock: Unexpected call`
# Fix: Verify mock.Anything vs specific matchers

# Date parsing errors
# Look for: `parsing time` errors
# Fix: Use Date type, not time.Time for API dates
```

### Authentication Issues
```bash
# Session token problems
client := monarch.NewClient("your-session-token")

# Login with MFA
session, err := client.Auth.Login(ctx, "email", "password")
if err != nil {
    // Check for MFA challenge
    if mfaErr, ok := err.(*MFARequired); ok {
        session, err = client.Auth.LoginWithMFA(ctx, "email", "password", mfaErr.Token, "123456")
    }
}
```

### GraphQL Debugging
```go
// Enable GraphQL request logging (set in client options)
client := monarch.NewClientWithOptions("token", &monarch.ClientOptions{
    Debug: true,  // Logs all GraphQL requests
})
```

## 🧪 Testing Patterns

### TDD for Bug Fixes (IMPORTANT!)
When you find a bug, **always write a failing test first**:

```bash
# 1. Write a test that reproduces the bug
func TestAccountService_BugFix_Issue123(t *testing.T) {
    // Setup that reproduces the problematic scenario
    mockTransport := new(MockTransport)
    client := setupTestClient(mockTransport)
    
    // Mock the exact response that causes the bug
    response := `{"problematic": "response"}`
    mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(response, nil)
    
    // Call the method that's broken
    result, err := client.Accounts.ProblematicMethod(ctx, "test")
    
    // This should fail before the fix
    assert.NoError(t, err)
    assert.Equal(t, "expected", result.Field)
}

# 2. Run the test - it should FAIL
go test ./pkg/monarch -run TestAccountService_BugFix_Issue123 -v

# 3. Fix the bug in the implementation
# Edit pkg/monarch/accounts.go

# 4. Run the test again - it should PASS  
go test ./pkg/monarch -run TestAccountService_BugFix_Issue123 -v

# 5. Run all tests to ensure no regression
go test ./pkg/monarch/... -v
```

### Standard Test Structure
```go
func TestServiceName_MethodName(t *testing.T) {
    // Setup mock transport
    mockTransport := new(MockTransport)
    client := &Client{
        transport:   mockTransport,
        queryLoader: graphql.NewQueryLoader(),
        options:     &ClientOptions{},
        baseURL:     "https://api.test.com",
    }
    client.initServices()

    // Mock GraphQL response (NO "data" wrapper needed)
    response := `{
        "fieldName": "value"
    }`

    // Setup mock expectation
    mockTransport.On("Execute", 
        mock.Anything,           // context
        mock.Anything,           // query string  
        mock.Anything,           // variables
        mock.Anything,           // result pointer
    ).Return(response, nil)

    // Execute method
    result, err := client.ServiceName.MethodName(ctx, "param")

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockTransport.AssertExpectations(t)
}
```

### Testing with Specific Parameters
```go
mockTransport.On("Execute", 
    mock.Anything, 
    mock.Anything,
    mock.MatchedBy(func(vars map[string]interface{}) bool {
        return vars["accountId"] == "test-123"
    }),
    mock.Anything,
).Return(response, nil)
```

## 🔍 Key Architectural Decisions

### Why These Patterns Exist
- **Interface-first**: All services are interfaces for easy testing/mocking
- **Context everywhere**: All methods accept context.Context for cancellation
- **Structured errors**: Custom error types with codes, not generic errors  
- **GraphQL transport**: Single HTTP client with GraphQL query loading
- **No "data" wrapper in tests**: Monarch API doesn't use standard GraphQL response format
- **Custom Date type**: API returns dates in multiple formats, needs custom parsing

### Python Client Differences
- **Sessions**: JSON files instead of Python pickle
- **Error handling**: Structured errors instead of generic exceptions
- **Types**: Strong typing instead of `Dict[str, Any]`
- **Concurrency**: Goroutines instead of asyncio

## 🔗 Key Resources

- **Python reference**: [Original Python client](https://github.com/hammem/monarchmoney) - check `monarchmoney/monarchmoney.py`
- **Example usage**: `examples/full_example.go`
- **GraphQL queries**: `internal/graphql/queries/`
- **All interfaces**: `pkg/monarch/interfaces.go`

## 📦 Release Process

**⚠️ CRITICAL: You MUST create a new release whenever code changes are pushed to main.**

This library is consumed as a versioned dependency from GitHub. Forgetting to create a release tag means users cannot access your changes.

### When to Release

Create a new release for:
- ✅ **New features** (minor version bump: v1.0.0 → v1.1.0)
- ✅ **Bug fixes** (patch version bump: v1.0.0 → v1.0.1)
- ✅ **Breaking changes** (major version bump: v1.0.0 → v2.0.0)
- ✅ **Documentation updates** that affect API usage
- ✅ **ANY code changes** that consumers might need

**❌ DO NOT** commit code changes to main without also creating a release tag.

### Pre-Release Checklist

Before creating a release, ensure:

1. **All tests pass**:
   ```bash
   go test ./pkg/monarch/... -v
   ```

2. **Code builds successfully**:
   ```bash
   go build ./...
   ```

3. **CHANGELOG.md is updated** with:
   - Version number and date in `## [X.Y.Z] - YYYY-MM-DD` format
   - Changes organized by category (Added/Changed/Fixed/Removed/Security)
   - Breaking changes clearly marked with ⚠️
   - Link to release at the bottom

4. **Version follows semantic versioning**:
   - **Major** (v1.0.0 → v2.0.0): Breaking API changes
   - **Minor** (v1.0.0 → v1.1.0): New features, backward compatible
   - **Patch** (v1.0.0 → v1.0.1): Bug fixes, backward compatible

### Release Steps

```bash
# 1. Update CHANGELOG.md
# Edit CHANGELOG.md to add new version section with all changes
# Move items from [Unreleased] to new [X.Y.Z] section

# 2. Commit the changelog
git add CHANGELOG.md
git commit -m "docs: Update CHANGELOG for vX.Y.Z"

# 3. Create annotated tag with detailed message
git tag -a vX.Y.Z -m "Release vX.Y.Z - Brief description

Added:
- Feature 1
- Feature 2

Fixed:
- Bug fix 1

Breaking Changes:
- None (or list them)
"

# 4. Push commit AND tag (tags are not pushed by default!)
git push && git push origin vX.Y.Z

# 5. Verify tag is visible on GitHub
git ls-remote --tags origin
# Should show: refs/tags/vX.Y.Z
```

### Example Release Message

```
Release v1.2.0 - Add transaction filtering and pagination

Added:
- Transaction filtering by date range
- Pagination support for large transaction lists
- New GetTransactionsByCategory method

Fixed:
- Authentication token refresh bug
- Memory leak in streaming API

Breaking Changes:
- None
```

### After Releasing

1. **Verify tag on GitHub**:
   - Visit https://github.com/eshaffer321/monarch-go/tags
   - Confirm your tag appears

2. **Test consuming the new version**:
   ```bash
   # In a test project
   go get github.com/eshaffer321/monarch-go@vX.Y.Z
   go mod tidy
   ```

3. **Update dependent projects**:
   - If you maintain projects using this library, update them
   - Run `go get -u github.com/eshaffer321/monarch-go@vX.Y.Z`

### Emergency Hotfix Releases

For critical bugs in production:

1. Create hotfix branch from the tagged release:
   ```bash
   git checkout -b hotfix/vX.Y.Z vX.Y.Z
   ```

2. Fix the bug and commit

3. Update CHANGELOG with patch version

4. Create patch version tag:
   ```bash
   git tag -a vX.Y.Z+1 -m "Hotfix vX.Y.Z+1 - Critical bug fix"
   git push origin hotfix/vX.Y.Z
   git push origin vX.Y.Z+1
   ```

5. Merge hotfix back to main

### Common Mistakes to Avoid

- ❌ **Forgetting to push tags**: `git push` doesn't push tags by default
  - ✅ Always use: `git push origin vX.Y.Z`

- ❌ **Using wrong module path**: Module path in go.mod MUST match GitHub repo URL
  - ✅ Should be: `module github.com/eshaffer321/monarch-go`

- ❌ **Skipping CHANGELOG updates**: Always document what changed
  - ✅ Update CHANGELOG.md BEFORE creating tag

- ❌ **Committing code without releasing**: Every code change needs a release
  - ✅ Commit → Update CHANGELOG → Tag → Push both

- ❌ **Using v2+ without /v2 in module path**: Go modules require suffix for major versions ≥2
  - ✅ For v2.0.0+, module path must be: `github.com/eshaffer321/monarch-go/v2`

- ❌ **Creating lightweight tags**: Use annotated tags with `-a` flag
  - ✅ Annotated tags include metadata and show up properly on GitHub

### Versioning Strategy for This Project

- **v1.x.x**: Current stable version, backward compatible improvements
- **v2.x.x**: Only when breaking changes are absolutely necessary
- **v0.x.x**: Pre-release versions (no longer used, we're at v1+)

### Why This Matters

Go modules use git tags for versioning. When someone runs:
```bash
go get github.com/eshaffer321/monarch-go@v1.2.0
```

Go fetches the code at that exact tag. **Without a tag, users cannot access your changes.**

## 🚨 Before You Start Coding

1. **Run the tests**: `go test ./pkg/monarch/... -v` (should all pass)
2. **For bugs**: **ALWAYS write a failing test first** to reproduce the issue (TDD)
3. **For new features**: Add to interface first, then implement, then test  
4. **Check coverage**: Current coverage is ~37%, don't make it worse
5. **Check Python client**: When in doubt, see how Python version works
6. **Use existing patterns**: Don't invent new ways of doing things

---

**Most Important**: This codebase has consistent patterns. Follow them, don't create new ones.