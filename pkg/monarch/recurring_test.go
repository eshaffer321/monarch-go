package monarch

import (
	"context"
	"testing"
	"time"

	"github.com/eshaffer321/monarch-go/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRecurringService_List(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.Recurring = &recurringService{client: client}

	// Mock response based on the new GraphQL query structure
	mockResponse := `{
		"recurringTransactionItems": [
			{
				"stream": {
					"id": "123",
					"frequency": "monthly",
					"amount": 15.99,
					"isApproximate": false,
					"merchant": {
						"id": "merch1",
						"name": "Netflix",
						"logoUrl": "https://example.com/netflix.png"
					}
				},
				"date": "2025-09-01T00:00:00Z",
				"amount": 15.99,
				"isPending": false,
				"account": {
					"id": "acc1",
					"displayName": "Checking"
				},
				"category": {
					"id": "cat1",
					"name": "Entertainment"
				}
			}
		]
	}`

	// Mock the GraphQL call
	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Run(func(args mock.Arguments) {
		// Verify the request
		query := args.Get(1).(string)
		assert.Contains(t, query, "Web_GetUpcomingRecurringTransactionItems")
		assert.Contains(t, query, "recurringTransactionItems")

		// Should have startDate and endDate variables
		variables := args.Get(2).(map[string]interface{})
		assert.NotNil(t, variables["startDate"])
		assert.NotNil(t, variables["endDate"])
	})

	// Execute
	transactions, err := client.Recurring.List(context.Background())

	// Verify
	require.NoError(t, err)
	require.Len(t, transactions, 1)

	// Check the transaction details
	assert.Equal(t, "123", transactions[0].ID)
	assert.Equal(t, 15.99, transactions[0].Amount)
	assert.Equal(t, "monthly", transactions[0].Frequency)
	assert.True(t, transactions[0].IsActive)
	assert.False(t, transactions[0].IsApproximate)

	// Check merchant
	assert.NotNil(t, transactions[0].Merchant)
	assert.Equal(t, "merch1", transactions[0].Merchant.ID)
	assert.Equal(t, "Netflix", transactions[0].Merchant.Name)

	// Check nested objects
	assert.NotNil(t, transactions[0].Category)
	assert.Equal(t, "cat1", transactions[0].Category.ID)
	assert.Equal(t, "Entertainment", transactions[0].Category.Name)

	assert.NotNil(t, transactions[0].Account)
	assert.Equal(t, "acc1", transactions[0].Account.ID)
	assert.Equal(t, "Checking", transactions[0].Account.DisplayName)

	mockTransport.AssertExpectations(t)
}

func TestRecurringService_ListWithDateRange(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.Recurring = &recurringService{client: client}

	startDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)

	// Mock response
	mockResponse := `{
		"recurringTransactionItems": []
	}`

	// Mock the GraphQL call
	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Run(func(args mock.Arguments) {
		// Verify the dates are passed correctly
		variables := args.Get(2).(map[string]interface{})
		assert.Equal(t, "2025-09-01", variables["startDate"])
		assert.Equal(t, "2025-10-01", variables["endDate"])
		assert.Nil(t, variables["filters"])
	})

	// Execute
	transactions, err := client.Recurring.ListWithDateRange(context.Background(), startDate, endDate)

	// Verify
	require.NoError(t, err)
	require.Len(t, transactions, 0)

	mockTransport.AssertExpectations(t)
}
