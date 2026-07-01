package monarch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eshaffer321/monarch-go/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTransactionService_Query(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	// Mock response
	mockResponse := `{
		"allTransactions": {
			"totalCount": 150,
			"results": [
				{
					"id": "txn-001",
					"amount": -50.00,
					"date": "2024-01-15T00:00:00Z",
					"merchant": {
						"name": "Grocery Store",
						"id": "merch-123"
					},
					"category": {
						"id": "cat-food",
						"name": "Food & Dining"
					},
					"account": {
						"id": "acc-123",
						"displayName": "Checking"
					}
				},
				{
					"id": "txn-002",
					"amount": -25.50,
					"date": "2024-01-14T00:00:00Z",
					"merchant": {
						"name": "Coffee Shop",
						"id": "merch-456"
					},
					"category": {
						"id": "cat-food",
						"name": "Food & Dining"
					},
					"account": {
						"id": "acc-123",
						"displayName": "Checking"
					}
				}
			]
		}
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			filters, ok := v["filters"].(map[string]interface{})
			if !ok {
				return false
			}
			// Check builder pattern filters were applied
			return filters["startDate"] == "2024-01-01" &&
				filters["endDate"] == "2024-01-31" &&
				filters["search"] == "grocery" &&
				v["limit"] == 20 &&
				v["offset"] == 0
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute query with builder pattern
	ctx := context.Background()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	result, err := service.Query().
		Between(startDate, endDate).
		Search("grocery").
		WithMinAmount(10).
		WithMaxAmount(100).
		Limit(20).
		Execute(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 150, result.TotalCount)
	assert.Len(t, result.Transactions, 2)
	assert.True(t, result.HasMore)
	assert.Equal(t, 20, result.NextOffset)

	// Verify first transaction
	txn := result.Transactions[0]
	assert.Equal(t, "txn-001", txn.ID)
	assert.Equal(t, -50.00, txn.Amount)
	assert.Equal(t, "Grocery Store", txn.Merchant.Name)
	assert.Equal(t, "Food & Dining", txn.Category.Name)

	mockTransport.AssertExpectations(t)
}

func TestTransactionService_Stream(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	// Mock first batch
	batch1Response := `{
		"allTransactions": {
			"totalCount": 3,
			"results": [
				{"id": "txn-001", "amount": -10.00},
				{"id": "txn-002", "amount": -20.00}
			]
		}
	}`

	// Mock second batch (last)
	batch2Response := `{
		"allTransactions": {
			"totalCount": 3,
			"results": [
				{"id": "txn-003", "amount": -30.00}
			]
		}
	}`

	// Set up expectations for two batches
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			return v["offset"] == 0
		}),
		mock.Anything,
	).Return(batch1Response, nil).Once()

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			return v["offset"] == 2
		}),
		mock.Anything,
	).Return(batch2Response, nil).Once()

	// Execute streaming
	ctx := context.Background()
	txnChan, errChan := service.Query().
		Limit(2). // Small batch size for testing
		Stream(ctx)

	// Collect results
	var transactions []*Transaction
	done := false

	for !done {
		select {
		case txn, ok := <-txnChan:
			if !ok {
				done = true
				break
			}
			transactions = append(transactions, txn)
		case err := <-errChan:
			require.NoError(t, err)
		case <-time.After(2 * time.Second):
			t.Fatal("Stream timeout")
		}
	}

	// Assert
	assert.Len(t, transactions, 3)
	assert.Equal(t, "txn-001", transactions[0].ID)
	assert.Equal(t, "txn-002", transactions[1].ID)
	assert.Equal(t, "txn-003", transactions[2].ID)

	mockTransport.AssertExpectations(t)
}

func TestTransactionService_Get(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	// Mock response
	mockResponse := `{
		"getTransaction": {
			"id": "txn-123",
			"amount": -75.50,
			"date": "2024-01-15T00:00:00Z",
			"merchant": {
				"name": "Test Store",
				"id": "merch-789"
			},
			"category": {
				"id": "cat-shop",
				"name": "Shopping"
			},
			"notes": "Test purchase",
			"isSplitTransaction": true,
			"splits": [
				{
					"id": "split-1",
					"amount": -50.00,
					"notes": "Item 1",
					"category": {
						"id": "cat-1",
						"name": "Category 1"
					}
				},
				{
					"id": "split-2",
					"amount": -25.50,
					"notes": "Item 2",
					"category": {
						"id": "cat-2",
						"name": "Category 2"
					}
				}
			]
		}
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			return v["id"] == "txn-123"
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	details, err := service.Get(ctx, "txn-123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "txn-123", details.ID)
	assert.Equal(t, -75.50, details.Amount)
	assert.Equal(t, "Test Store", details.Merchant.Name)
	assert.True(t, details.IsSplitTransaction)
	assert.Len(t, details.Splits, 2)
	assert.Equal(t, -50.00, details.Splits[0].Amount)

	mockTransport.AssertExpectations(t)
}

func TestTransactionService_Create(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	// Mock create response
	createResponse := `{
		"createTransaction": {
			"transaction": {
				"id": "new-txn-456"
			},
			"errors": []
		}
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			input, ok := v["input"].(map[string]interface{})
			if !ok {
				return false
			}
			_, hasMerchant := input["merchantName"]
			return hasMerchant
		}),
		mock.Anything,
	).Return(createResponse, nil).Once()

	// Execute
	ctx := context.Background()
	params := &CreateTransactionParams{
		Date:       Date{Time: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)},
		AccountID:  "acc-123",
		Amount:     -100.00,
		Merchant:   &Merchant{Name: "New Store"},
		CategoryID: "cat-new",
		Notes:      "New transaction",
	}

	txn, err := service.Create(ctx, params)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "new-txn-456", txn.ID)
	assert.Equal(t, -100.00, txn.Amount)

	mockTransport.AssertExpectations(t)
}

func TestTransactionService_Create_RoundsAmount(t *testing.T) {
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			input, ok := v["input"].(map[string]interface{})
			if !ok {
				return false
			}
			amount, ok := input["amount"].(float64)
			return ok && amount == -99.99 // should be rounded from -99.9876
		}),
		mock.Anything,
	).Return(`{"createTransaction":{"transaction":{"id":"txn-1"},"errors":[]}}`, nil).Once()

	txn, err := service.Create(context.Background(), &CreateTransactionParams{
		Date:       Date{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		AccountID:  "acc-1",
		Amount:     -99.9876,
		CategoryID: "cat-1",
	})

	require.NoError(t, err)
	assert.Equal(t, -99.99, txn.Amount)
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_Create_ShouldUpdateBalance(t *testing.T) {
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	shouldUpdate := false
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			input, ok := v["input"].(map[string]interface{})
			if !ok {
				return false
			}
			val, exists := input["shouldUpdateBalance"]
			return exists && val == false
		}),
		mock.Anything,
	).Return(`{"createTransaction":{"transaction":{"id":"txn-2"},"errors":[]}}`, nil).Once()

	txn, err := service.Create(context.Background(), &CreateTransactionParams{
		Date:                Date{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		AccountID:           "acc-1",
		Amount:              50.0,
		CategoryID:          "cat-1",
		ShouldUpdateBalance: &shouldUpdate,
	})

	require.NoError(t, err)
	assert.Equal(t, "txn-2", txn.ID)
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_Create_MissingCategoryID(t *testing.T) {
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	_, err := service.Create(context.Background(), &CreateTransactionParams{
		Date:      Date{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		AccountID: "acc-1",
		Amount:    10.0,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "CategoryID is required")
}

func TestTransactionService_Create_FieldErrors(t *testing.T) {
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	errorResp := `{
		"createTransaction": {
			"transaction": null,
			"errors": [{
				"message": "Validation failed",
				"code": "INVALID_INPUT",
				"fieldErrors": [
					{"field": "amount", "messages": ["must be non-zero"]},
					{"field": "date", "messages": ["is required", "must be valid"]}
				]
			}]
		}
	}`

	mockTransport.On("Execute",
		mock.Anything, mock.AnythingOfType("string"),
		mock.Anything, mock.Anything,
	).Return(errorResp, nil).Once()

	_, err := service.Create(context.Background(), &CreateTransactionParams{
		Date:       Date{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		AccountID:  "acc-1",
		Amount:     0,
		CategoryID: "cat-1",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Validation failed")
	assert.Contains(t, err.Error(), "amount: must be non-zero")
	assert.Contains(t, err.Error(), "date: is required, must be valid")
}

func TestTransactionService_Create_NoTransactionReturned(t *testing.T) {
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	mockTransport.On("Execute",
		mock.Anything, mock.AnythingOfType("string"),
		mock.Anything, mock.Anything,
	).Return(`{"createTransaction":{"transaction":null,"errors":[]}}`, nil).Once()

	_, err := service.Create(context.Background(), &CreateTransactionParams{
		Date:       Date{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		AccountID:  "acc-1",
		Amount:     10.0,
		CategoryID: "cat-1",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no transaction returned")
}

func TestTransactionService_Create_GraphQLError(t *testing.T) {
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	mockTransport.On("Execute",
		mock.Anything, mock.AnythingOfType("string"),
		mock.Anything, mock.Anything,
	).Return("", errors.New("graphql error")).Once()

	_, err := service.Create(context.Background(), &CreateTransactionParams{
		Date:       Date{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		AccountID:  "acc-1",
		Amount:     10.0,
		CategoryID: "cat-1",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create transaction")
}

func TestTransactionCategoryService_List(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &transactionCategoryService{client: client}

	// Mock response
	mockResponse := `{
		"categories": [
			{
				"id": "cat-1",
				"name": "Food & Dining",
				"icon": "🍔",
				"order": 1,
				"group": {
					"id": "grp-1",
					"name": "Essential Expenses",
					"type": "expense"
				}
			},
			{
				"id": "cat-2",
				"name": "Transportation",
				"icon": "🚗",
				"order": 2,
				"group": {
					"id": "grp-1",
					"name": "Essential Expenses",
					"type": "expense"
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
	categories, err := service.List(ctx)

	// Assert
	require.NoError(t, err)
	assert.Len(t, categories, 2)
	assert.Equal(t, "Food & Dining", categories[0].Name)
	assert.Equal(t, "🍔", categories[0].Icon)
	assert.Equal(t, "Essential Expenses", categories[0].Group.Name)

	mockTransport.AssertExpectations(t)
}

func TestTransactionService_UpdateSplits_Old(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	// Mock response
	mockResponse := `{
		"updateTransactionSplits": {
			"transaction": {
				"id": "txn-123"
			},
			"errors": []
		}
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(v map[string]interface{}) bool {
			input, ok := v["input"].(map[string]interface{})
			if !ok {
				return false
			}
			splitData, ok := input["splitData"].([]map[string]interface{})
			return ok && len(splitData) == 2
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	splits := []*TransactionSplit{
		{
			Amount:     -30.00,
			Notes:      "Split 1",
			CategoryID: "cat-1",
			Merchant:   &Merchant{Name: "Store A"},
		},
		{
			Amount:     -20.00,
			Notes:      "Split 2",
			CategoryID: "cat-2",
			Merchant:   &Merchant{Name: "Store B"},
		},
	}

	err := service.UpdateSplits(ctx, "txn-123", splits)

	// Assert
	assert.NoError(t, err)
	mockTransport.AssertExpectations(t)
}

// TestTransactionService_Update_CorrectFieldNames verifies that Update() uses
// the correct GraphQL field names as per the Python reference implementation:
// - "category" (not "categoryId")
// - "name" (not "merchant")
// This test documents the expected API contract.
func TestTransactionService_Update_CorrectFieldNames(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	// Mock response
	mockResponse := `{
		"updateTransaction": {
			"transaction": {
				"id": "txn-123",
				"amount": -50.00,
				"date": "2024-01-15T00:00:00Z",
				"merchant": {
					"id": "merch-456",
					"name": "New Store"
				},
				"category": {
					"id": "cat-groceries",
					"name": "Groceries"
				},
				"notes": "Updated transaction",
				"hideFromReports": false,
				"needsReview": false
			},
			"errors": []
		}
	}`

	// This matcher verifies the CORRECT field names are used
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(variables map[string]interface{}) bool {
			input, ok := variables["input"].(map[string]interface{})
			if !ok {
				t.Logf("ERROR: input is not a map: %T", variables["input"])
				return false
			}

			// Verify transaction ID
			if input["id"] != "txn-123" {
				t.Logf("ERROR: Wrong transaction ID: %v", input["id"])
				return false
			}

			// CRITICAL: Verify "category" field is used (not "categoryId")
			categoryValue, hasCategoryField := input["category"]
			if !hasCategoryField {
				t.Logf("ERROR: Missing 'category' field. Found fields: %v", input)
				return false
			}
			if categoryValue != "cat-groceries" {
				t.Logf("ERROR: Wrong category value: %v", categoryValue)
				return false
			}

			// CRITICAL: Verify "name" field is used for merchant (not "merchant")
			nameValue, hasNameField := input["name"]
			if !hasNameField {
				t.Logf("ERROR: Missing 'name' field for merchant. Found fields: %v", input)
				return false
			}
			if nameValue != "New Store" {
				t.Logf("ERROR: Wrong merchant name value: %v", nameValue)
				return false
			}

			// Verify notes field
			if input["notes"] != "Updated note" {
				t.Logf("ERROR: Wrong notes value: %v", input["notes"])
				return false
			}

			t.Logf("SUCCESS: All field names are correct: %v", input)
			return true
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	categoryID := "cat-groceries"
	merchantName := "New Store"
	notes := "Updated note"

	transaction, err := service.Update(ctx, "txn-123", &UpdateTransactionParams{
		CategoryID: &categoryID,
		Merchant:   &merchantName,
		Notes:      &notes,
	})

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, "txn-123", transaction.ID)
	assert.Equal(t, "cat-groceries", transaction.Category.ID)
	assert.Equal(t, "New Store", transaction.Merchant.Name)
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_Delete(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	// Mock successful deletion response
	mockResponse := `{
		"deleteTransaction": {
			"deleted": true,
			"errors": []
		}
	}`

	// Verify the mutation uses input wrapper (matching Python client format)
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(variables map[string]interface{}) bool {
			// The Python client uses: {"input": {"transactionId": "..."}}
			input, ok := variables["input"].(map[string]interface{})
			if !ok {
				t.Logf("ERROR: Expected 'input' wrapper, got variables: %v", variables)
				return false
			}

			transactionID, ok := input["transactionId"].(string)
			if !ok {
				t.Logf("ERROR: Expected 'transactionId' field in input, got: %v", input)
				return false
			}

			if transactionID != "txn-123" {
				t.Logf("ERROR: Wrong transaction ID: %v", transactionID)
				return false
			}

			t.Logf("SUCCESS: Correct mutation format with input wrapper")
			return true
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	err := service.Delete(ctx, "txn-123")

	// Assert
	require.NoError(t, err)
	mockTransport.AssertExpectations(t)
}

func TestTransactionService_Delete_Error(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	// Mock error response (e.g., trying to delete bank-imported transaction)
	mockResponse := `{
		"deleteTransaction": {
			"deleted": false,
			"errors": [{
				"code": "BAD_REQUEST",
				"message": "Cannot delete bank-imported transactions"
			}]
		}
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	err := service.Delete(ctx, "txn-123")

	// Assert - should return structured error with both code and message
	require.Error(t, err)

	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "BAD_REQUEST", apiErr.Code)
	assert.Equal(t, "Cannot delete bank-imported transactions", apiErr.Message)

	mockTransport.AssertExpectations(t)
}

// TestTransactionService_Update_AmountOnly tests updating a transaction's amount.
// This reproduces the bug reported where Update() succeeds but the amount isn't actually updated.
func TestTransactionService_Update_AmountOnly(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := newTransactionService(client)

	// Mock response showing the updated amount
	mockResponse := `{
		"updateTransaction": {
			"transaction": {
				"id": "225271673011145604",
				"amount": -127.90,
				"date": "2024-01-15T00:00:00Z",
				"merchant": {
					"id": "merch-456",
					"name": "Test Store"
				},
				"category": {
					"id": "cat-123",
					"name": "Shopping"
				},
				"notes": "Multi-delivery order (3 charges: $6.56, $14.39, $106.95)",
				"hideFromReports": false,
				"needsReview": false
			},
			"errors": []
		}
	}`

	// This matcher verifies that the amount field is correctly passed
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(variables map[string]interface{}) bool {
			input, ok := variables["input"].(map[string]interface{})
			if !ok {
				t.Logf("ERROR: input is not a map: %T", variables["input"])
				return false
			}

			// Verify transaction ID
			if input["id"] != "225271673011145604" {
				t.Logf("ERROR: Wrong transaction ID: %v", input["id"])
				return false
			}

			// CRITICAL: Verify "amount" field is present and has the correct value
			amountValue, hasAmountField := input["amount"]
			if !hasAmountField {
				t.Logf("ERROR: Missing 'amount' field. Found fields: %v", input)
				return false
			}

			// Check the amount value (should be -127.90)
			var actualAmount float64
			switch v := amountValue.(type) {
			case float64:
				actualAmount = v
			case float32:
				actualAmount = float64(v)
			case int:
				actualAmount = float64(v)
			default:
				t.Logf("ERROR: Unexpected amount type: %T, value: %v", amountValue, amountValue)
				return false
			}

			if actualAmount != -127.90 {
				t.Logf("ERROR: Wrong amount value: %v (expected -127.90)", actualAmount)
				return false
			}

			t.Logf("SUCCESS: Amount field is correctly included: %v", input)
			return true
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute - update only the amount
	ctx := context.Background()
	amount := -127.90
	notes := "Multi-delivery order (3 charges: $6.56, $14.39, $106.95)"

	transaction, err := service.Update(ctx, "225271673011145604", &UpdateTransactionParams{
		Amount: &amount,
		Notes:  &notes,
	})

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, "225271673011145604", transaction.ID)
	assert.Equal(t, -127.90, transaction.Amount, "Amount should be updated to -127.90")
	assert.Equal(t, "Multi-delivery order (3 charges: $6.56, $14.39, $106.95)", transaction.Notes)
	mockTransport.AssertExpectations(t)
}
