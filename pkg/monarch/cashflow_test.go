package monarch

import (
	"context"
	"testing"
	"time"

	"github.com/eshaffer321/monarch-go/v2/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCashflowService_Get(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.Cashflow = &cashflowService{client: client}

	// Mock response based on the actual GraphQL query structure
	mockResponse := `{
		"byCategory": [
			{
				"groupBy": {
					"category": {
						"id": "cat1",
						"name": "Groceries",
						"group": {
							"id": "grp1",
							"type": "expense"
						}
					}
				},
				"summary": {
					"sum": -500.00
				}
			}
		],
		"byCategoryGroup": [
			{
				"groupBy": {
					"categoryGroup": {
						"id": "grp1",
						"name": "Food",
						"type": "expense"
					}
				},
				"summary": {
					"sum": -500.00
				}
			}
		],
		"byMerchant": [
			{
				"groupBy": {
					"merchant": {
						"id": "merch1",
						"name": "Walmart",
						"logoUrl": "https://example.com/walmart.png"
					}
				},
				"summary": {
					"sum": -200.00
				}
			}
		],
		"summary": [
			{
				"summary": {
					"sum": 2500.00,
					"sumIncome": 5000.00,
					"sumExpense": -2500.00,
					"savings": 2500.00,
					"savingsRate": 0.5
				}
			}
		]
	}`

	// Mock the GraphQL call
	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Run(func(args mock.Arguments) {
		// Verify the request
		query := args.Get(1).(string)
		assert.Contains(t, query, "Web_GetCashFlowPage")
		assert.Contains(t, query, "aggregates")

		// Check variables structure
		variables := args.Get(2).(map[string]interface{})
		filters := variables["filters"].(map[string]interface{})
		assert.Equal(t, "2025-08-01", filters["startDate"])
		assert.Equal(t, "2025-08-31", filters["endDate"])
	})

	// Execute
	params := &CashflowParams{
		StartDate: time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 8, 31, 0, 0, 0, 0, time.UTC),
	}
	cashflow, err := client.Cashflow.Get(context.Background(), params)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, cashflow)

	// Check summary
	assert.Equal(t, 5000.00, cashflow.Summary.Income)
	assert.Equal(t, 2500.00, cashflow.Summary.Expense)
	assert.Equal(t, 2500.00, cashflow.Summary.Savings)
	assert.Equal(t, 0.5, cashflow.Summary.SavingsRate)

	// Check categories
	require.Len(t, cashflow.ByCategory, 1)
	assert.Equal(t, "cat1", cashflow.ByCategory[0].Category.ID)
	assert.Equal(t, "Groceries", cashflow.ByCategory[0].Category.Name)
	assert.Equal(t, -500.00, cashflow.ByCategory[0].Amount)

	mockTransport.AssertExpectations(t)
}

func TestCashflowService_GetSummary(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.Cashflow = &cashflowService{client: client}

	// Mock response
	mockResponse := `{
		"summary": [
			{
				"summary": {
					"sumIncome": 5000.00,
					"sumExpense": -2500.00,
					"savings": 2500.00,
					"savingsRate": 0.5
				}
			}
		]
	}`

	// Mock the GraphQL call
	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Run(func(args mock.Arguments) {
		// Verify the request
		query := args.Get(1).(string)
		assert.Contains(t, query, "Web_GetCashFlowSummary")

		// Check variables structure
		variables := args.Get(2).(map[string]interface{})
		filters := variables["filters"].(map[string]interface{})
		assert.Equal(t, "2025-08-01", filters["startDate"])
		assert.Equal(t, "2025-08-31", filters["endDate"])
	})

	// Execute
	params := &CashflowSummaryParams{
		StartDate: time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 8, 31, 0, 0, 0, 0, time.UTC),
	}
	summary, err := client.Cashflow.GetSummary(context.Background(), params)

	// Verify
	require.NoError(t, err)
	require.NotNil(t, summary)

	assert.Equal(t, 5000.00, summary.Income)
	assert.Equal(t, 2500.00, summary.Expense)
	assert.Equal(t, 2500.00, summary.Savings)
	assert.Equal(t, 0.5, summary.SavingsRate)

	mockTransport.AssertExpectations(t)
}
