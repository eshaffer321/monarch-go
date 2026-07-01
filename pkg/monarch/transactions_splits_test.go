package monarch

import (
	"context"
	"testing"

	"github.com/eshaffer321/monarch-go/v2/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionService_GetSplits(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock response
	response := `{
		
			"getTransaction": {
				"id": "test-txn-123",
				"amount": 100.50,
				"category": {
					"id": "cat-1",
					"name": "Groceries"
				},
				"merchant": {
					"id": "merch-1",
					"name": "Test Store"
				},
				"splitTransactions": [
					{
						"id": "split-1",
						"amount": 50.25,
						"notes": "Half for groceries",
						"category": {
							"id": "cat-1",
							"name": "Groceries"
						},
						"merchant": {
							"id": "merch-1",
							"name": "Test Store"
						}
					},
					{
						"id": "split-2",
						"amount": 50.25,
						"notes": "Half for household",
						"category": {
							"id": "cat-2",
							"name": "Household"
						}
					}
				]
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	splits, err := client.Transactions.GetSplits(context.Background(), "test-txn-123")

	assert.NoError(t, err)
	assert.Len(t, splits, 2)
	assert.Equal(t, "split-1", splits[0].ID)
	assert.Equal(t, 50.25, splits[0].Amount)
	assert.Equal(t, "Half for groceries", splits[0].Notes)
	assert.Equal(t, "Groceries", splits[0].Category.Name)
	
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_UpdateSplits_New(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock response
	response := `{
		
			"updateTransactionSplit": {
				"transaction": {
					"id": "test-txn-123",
					"hasSplitTransactions": true,
					"splitTransactions": [
						{
							"id": "split-1",
							"amount": 60.00,
							"notes": "Updated split"
						},
						{
							"id": "split-2",
							"amount": 40.50
						}
					]
				},
				"errors": []
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	splits := []*TransactionSplit{
		{
			Amount:     60.00,
			CategoryID: "cat-1",
			Notes:      "Updated split",
			Merchant:   &Merchant{Name: "Store A"},
		},
		{
			Amount:     40.50,
			CategoryID: "cat-2",
		},
	}

	err := client.Transactions.UpdateSplits(context.Background(), "test-txn-123", splits)
	assert.NoError(t, err)
	
	// Verify the mutation was called
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_GetSummary(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock response - note that aggregates is an array
	response := `{
		"aggregates": [{
			"summary": {
				"count": 5498,
				"sumIncome": 651023.79,
				"sumExpense": -478055.57,
				"avg": 31.47,
				"sum": 172968.22
			}
		}]
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	summary, err := client.Transactions.GetSummary(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, 5498, summary.Count)
	assert.Equal(t, 651023.79, summary.SumIncome)
	assert.Equal(t, -478055.57, summary.SumExpense)
	
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_GetSummary_EmptyResult(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock empty response
	response := `{
		"aggregates": []
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	summary, err := client.Transactions.GetSummary(context.Background())

	assert.Error(t, err)
	assert.Nil(t, summary)
	assert.Contains(t, err.Error(), "no transaction summary data available")
	
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_UpdateSplits_ErrorHandling_SingleObject(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock response with a single error object (not an array)
	// This reproduces the actual API response that causes the unmarshal error
	response := `{
		"updateTransactionSplit": {
			"transaction": null,
			"errors": {
				"message": "Transaction already has splits",
				"code": "TRANSACTION_HAS_SPLITS",
				"fieldErrors": []
			}
		}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	splits := []*TransactionSplit{
		{
			Amount:     30.00,
			CategoryID: "cat-1",
		},
		{
			Amount:     30.63,
			CategoryID: "cat-2",
		},
	}

	err := client.Transactions.UpdateSplits(context.Background(), "220449306160236661", splits)
	
	// Should get a proper error, not an unmarshal error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Transaction already has splits")
	
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_UpdateSplits_ErrorHandling_ValidationError(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock response with validation error details
	response := `{
		"updateTransactionSplit": {
			"transaction": null,
			"errors": {
				"message": "Validation failed",
				"code": "VALIDATION_ERROR",
				"fieldErrors": [
					{
						"field": "amount",
						"messages": ["Split amounts must equal transaction amount"]
					}
				]
			}
		}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	splits := []*TransactionSplit{
		{
			Amount:     30.00,
			CategoryID: "cat-1",
		},
		{
			Amount:     20.00,
			CategoryID: "cat-2",
		},
	}

	err := client.Transactions.UpdateSplits(context.Background(), "test-txn-123", splits)
	
	// Should get a proper error with field details
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Validation failed")
	assert.Contains(t, err.Error(), "amount: Split amounts must equal transaction amount")
	
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_UpdateSplits_ErrorHandling_ArrayFormat(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock response with errors as an array (standard GraphQL format)
	response := `{
		"updateTransactionSplit": {
			"transaction": null,
			"errors": [
				{
					"message": "Invalid split configuration",
					"code": "INVALID_SPLIT",
					"fieldErrors": []
				}
			]
		}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	splits := []*TransactionSplit{
		{
			Amount:     60.63,
			CategoryID: "cat-1",
		},
	}

	err := client.Transactions.UpdateSplits(context.Background(), "test-txn-123", splits)
	
	// Should handle array format correctly
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid split configuration")
	
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_UpdateSplits_Success_EmptyErrorArray(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	// Mock successful response with empty errors array
	response := `{
		"updateTransactionSplit": {
			"transaction": {
				"id": "test-txn-123",
				"hasSplitTransactions": true,
				"splitTransactions": [
					{
						"id": "split-1",
						"amount": 30.00,
						"category": {
							"id": "cat-1",
							"name": "Groceries"
						}
					},
					{
						"id": "split-2",
						"amount": 30.63,
						"category": {
							"id": "cat-2",
							"name": "Shopping"
						}
					}
				]
			},
			"errors": []
		}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	splits := []*TransactionSplit{
		{
			Amount:     30.00,
			CategoryID: "cat-1",
		},
		{
			Amount:     30.63,
			CategoryID: "cat-2",
		},
	}

	err := client.Transactions.UpdateSplits(context.Background(), "test-txn-123", splits)
	
	// Should succeed with empty errors array
	assert.NoError(t, err)
	
	mockTransport.AssertExpectations(t)
}