package monarch

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/eshaffer321/monarch-go/v2/internal/graphql"
	internalTypes "github.com/eshaffer321/monarch-go/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTransport is a mock implementation of the Transport interface
type MockTransport struct {
	mock.Mock
}

func (m *MockTransport) Execute(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	args := m.Called(ctx, query, variables, result)

	// If mock provides result data, unmarshal it
	if args.Get(0) != nil {
		resultJSON := args.Get(0).(string)
		if err := json.Unmarshal([]byte(resultJSON), result); err != nil {
			return err
		}
	}

	return args.Error(1)
}

func (m *MockTransport) SetAuth(token string) {
	m.Called(token)
}

func (m *MockTransport) SetSession(session *internalTypes.Session) {
	m.Called(session)
}

func TestAccountService_List(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &accountService{client: client}

	// Mock response
	mockResponse := `{
		"accounts": [
			{
				"id": "acc-123",
				"displayName": "Test Checking",
				"currentBalance": 1500.50,
				"isAsset": true,
				"type": {
					"name": "depository",
					"display": "Depository"
				}
			},
			{
				"id": "acc-456", 
				"displayName": "Test Savings",
				"currentBalance": 5000.00,
				"isAsset": true,
				"type": {
					"name": "depository",
					"display": "Depository"
				}
			}
		]
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	accounts, err := service.List(ctx)

	// Assert
	require.NoError(t, err)
	assert.Len(t, accounts, 2)
	assert.Equal(t, "acc-123", accounts[0].ID)
	assert.Equal(t, "Test Checking", accounts[0].DisplayName)
	assert.Equal(t, 1500.50, accounts[0].CurrentBalance)

	mockTransport.AssertExpectations(t)
}

func TestAccountService_Get(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &accountService{client: client}

	// Mock response for List (Get calls List internally)
	mockResponse := `{
		"accounts": [
			{
				"id": "acc-123",
				"displayName": "Test Account",
				"currentBalance": 1000.00
			},
			{
				"id": "acc-456",
				"displayName": "Another Account",
				"currentBalance": 2000.00
			}
		]
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(mockResponse, nil)

	// Test: Account found
	ctx := context.Background()
	account, err := service.Get(ctx, "acc-123")

	require.NoError(t, err)
	assert.Equal(t, "acc-123", account.ID)
	assert.Equal(t, "Test Account", account.DisplayName)

	// Test: Account not found
	account, err = service.Get(ctx, "non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	assert.Nil(t, account)

	mockTransport.AssertExpectations(t)
}

func TestAccountService_Create(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &accountService{client: client}

	// Mock create response
	createResponse := `{
		"createManualAccount": {
			"account": {
				"id": "new-acc-789"
			},
			"errors": []
		}
	}`

	// Mock get response (for fetching full details)
	getResponse := `{
		"accounts": [
			{
				"id": "new-acc-789",
				"displayName": "New Test Account",
				"currentBalance": 1000.00,
				"isAsset": true
			}
		]
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			// Verify the create mutation variables
			input, ok := v["input"].(map[string]interface{})
			return ok && input["name"] == "New Test Account"
		}),
		mock.Anything,
	).Return(createResponse, nil).Once()

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(getResponse, nil).Once()

	// Execute
	ctx := context.Background()
	params := &CreateAccountParams{
		AccountType:       "other_asset",
		AccountSubtype:    "cash",
		IsAsset:           true,
		AccountName:       "New Test Account",
		CurrentBalance:    1000.00,
		IncludeInNetWorth: true,
	}

	account, err := service.Create(ctx, params)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "new-acc-789", account.ID)
	assert.Equal(t, "New Test Account", account.DisplayName)
	assert.Equal(t, 1000.00, account.CurrentBalance)

	mockTransport.AssertExpectations(t)
}

func TestAccountService_Update(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &accountService{client: client}

	// Mock response
	mockResponse := `{
		"updateAccount": {
			"account": {
				"id": "acc-123",
				"displayName": "Updated Name",
				"currentBalance": 2000.00,
				"includeInNetWorth": false
			},
			"errors": []
		}
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			input, ok := v["input"].(map[string]interface{})
			return ok && input["id"] == "acc-123"
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	newName := "Updated Name"
	newBalance := 2000.00
	includeInNetWorth := false

	params := &UpdateAccountParams{
		DisplayName:       &newName,
		CurrentBalance:    &newBalance,
		IncludeInNetWorth: &includeInNetWorth,
	}

	account, err := service.Update(ctx, "acc-123", params)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", account.DisplayName)
	assert.Equal(t, 2000.00, account.CurrentBalance)
	assert.False(t, account.IncludeInNetWorth)

	mockTransport.AssertExpectations(t)
}

func TestAccountService_Delete(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &accountService{client: client}

	// Test: Successful deletion
	successResponse := `{
		"deleteAccount": {
			"deleted": true,
			"errors": []
		}
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			return v["id"] == "acc-123"
		}),
		mock.Anything,
	).Return(successResponse, nil).Once()

	ctx := context.Background()
	err := service.Delete(ctx, "acc-123")

	assert.NoError(t, err)

	// Test: Failed deletion
	failResponse := `{
		"deleteAccount": {
			"deleted": false,
			"errors": [
				{
					"code": "CANNOT_DELETE",
					"message": "Account has pending transactions"
				}
			]
		}
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(failResponse, nil).Once()

	err = service.Delete(ctx, "acc-456")

	assert.Error(t, err)
	var apiErr *Error
	assert.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "CANNOT_DELETE", apiErr.Code)

	mockTransport.AssertExpectations(t)
}

func TestAccountService_RefreshAndWait(t *testing.T) {
	t.Skip("Test needs proper mock setup for status checks")
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.Accounts = &accountService{client: client}
	service := &accountService{client: client}

	// Mock refresh request
	refreshResponse := `{
		"requestAccountsRefresh": true
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			ids, ok := v["accountIds"].([]string)
			return ok && len(ids) == 1 && ids[0] == "acc-123"
		}),
		mock.Anything,
	).Return(refreshResponse, nil).Once()

	// Mock status checks - first in progress, then complete
	inProgressResponse := `{
		"accounts": [
			{
				"id": "acc-123",
				"credential": {
					"updateRequired": true
				}
			}
		]
	}`

	completeResponse := `{
		"accounts": [
			{
				"id": "acc-123",
				"credential": {
					"updateRequired": false
				}
			}
		]
	}`

	// First status check - still updating
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(inProgressResponse, nil).Once()

	// Second status check - complete
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(completeResponse, nil).Once()

	// Execute with short timeout for testing
	ctx := context.Background()
	err := service.RefreshAndWait(ctx, 5*time.Second, "acc-123")

	// Assert
	assert.NoError(t, err)
	mockTransport.AssertExpectations(t)
}

func TestAccountService_GetBalances(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &accountService{client: client}

	// Mock response
	mockResponse := `{
		"accounts": [
			{
				"id": "acc-123",
				"recentBalances": [
					{"date": "2024-01-01", "balance": 1000.00},
					{"date": "2024-01-02", "balance": 1050.00},
					{"date": "2024-01-03", "balance": 1100.00}
				]
			},
			{
				"id": "acc-456",
				"recentBalances": [
					{"date": "2024-01-01", "balance": 5000.00},
					{"date": "2024-01-02", "balance": 4950.00}
				]
			}
		]
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			startDate, ok := v["startDate"].(string)
			return ok && startDate != ""
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	balances, err := service.GetBalances(ctx, &startDate)

	// Assert
	require.NoError(t, err)
	assert.Len(t, balances, 5) // 3 + 2 balance entries
	assert.Equal(t, "acc-123", balances[0].AccountID)
	assert.Equal(t, 1000.00, balances[0].Balance)

	mockTransport.AssertExpectations(t)
}
