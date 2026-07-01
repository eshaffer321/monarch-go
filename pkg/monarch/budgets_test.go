package monarch

import (
	"context"
	"testing"
	"time"

	"github.com/eshaffer321/monarch-go/v2/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBudgetService_List(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &budgetService{client: client}

	// This test should FAIL before the fix and PASS after the fix
	// We're testing that the service sends the correct parameter names

	// Mock a successful response when correct parameters are used
	mockResponse := `{
		"budgetData": {
			"monthlyAmountsByCategory": [
				{
					"category": {
						"id": "cat-1",
						"name": "Food"
					},
					"monthlyAmounts": [
						{
							"month": "2025-08-01",
							"plannedAmount": 500.00,
							"actualAmount": 250.00,
							"remainingAmount": 250.00,
							"percentComplete": 50.0,
							"transactionCount": 10
						}
					]
				}
			]
		}
	}`

	// Set up expectation that startMonth/endMonth should be sent
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(vars map[string]interface{}) bool {
			// The service should send startDate and endDate as variable names
			// (but they get mapped to startMonth/endMonth in the GraphQL query)
			_, hasStartDate := vars["startDate"]
			_, hasEndDate := vars["endDate"]

			return hasStartDate && hasEndDate
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute with date range
	ctx := context.Background()
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	budgets, err := service.List(ctx, startOfMonth, endOfMonth)

	// After fix: should succeed
	assert.NoError(t, err, "Should not return an error when using correct parameters")
	assert.NotNil(t, budgets, "Should return budgets")

	// Verify the mock was called with correct parameters
	mockTransport.AssertExpectations(t)
}

func TestBudgetService_List_CorrectParameters(t *testing.T) {
	// This test shows what SHOULD happen with correct parameters

	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &budgetService{client: client}

	// Mock successful response when using correct parameters
	mockResponse := `{
		"budgetData": [
			{
				"id": "budget-1",
				"amount": 500.00,
				"spent": 250.00,
				"remaining": 250.00,
				"percentageComplete": 50.0,
				"category": {
					"id": "cat-1",
					"name": "Groceries"
				}
			}
		]
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(vars map[string]interface{}) bool {
			// After fix, we should be sending startMonth/endMonth
			_, hasStartMonth := vars["startMonth"]
			_, hasEndMonth := vars["endMonth"]

			// This will fail until we fix the implementation
			return hasStartMonth && hasEndMonth
		}),
		mock.Anything,
	).Return(mockResponse, nil).Maybe()

	// For now, return error since we haven't fixed it yet
	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(nil, &Error{Code: "BAD_REQUEST", Message: "Invalid parameters"})

	// Execute
	ctx := context.Background()
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	budgets, err := service.List(ctx, startOfMonth, endOfMonth)

	// Currently this fails because we're using wrong parameter names
	assert.Error(t, err, "Should fail until we fix the parameter names")
	assert.Nil(t, budgets)
}

func TestBudgetService_ListWithGoals(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &budgetService{client: client}

	// Mock a successful response with both budgets and goals
	mockResponse := `{
		"budgetData": {
			"monthlyAmountsByCategory": [
				{
					"category": {
						"id": "cat-1",
						"name": "Food",
						"icon": "food",
						"group": {
							"id": "group-1",
							"name": "Living",
							"type": "expense"
						}
					},
					"monthlyAmounts": [
						{
							"month": "2025-08-01",
							"plannedCashFlowAmount": 500.00,
							"plannedSetAsideAmount": 0,
							"actualAmount": -250.00,
							"remainingAmount": 250.00,
							"previousMonthRolloverAmount": 0,
							"rolloverType": ""
						}
					]
				}
			]
		},
		"goalsV2": {
			"goals": [
				{
					"id": "goal-1",
					"name": "Emergency Fund",
					"type": "savings",
					"amount": 10000.00,
					"priority": 1,
					"targetDate": "2025-12-31T00:00:00Z",
					"targetAmount": 10000.00,
					"currentAmount": 5000.00,
					"imageUrl": "https://example.com/image.png",
					"accountId": "acc-1",
					"account": {
						"id": "acc-1",
						"displayName": "Savings Account"
					},
					"percentageComplete": 50.0,
					"monthlyContribution": 500.00,
					"createdAt": "2025-01-01T00:00:00Z",
					"updatedAt": "2025-08-01T00:00:00Z"
				}
			]
		}
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.MatchedBy(func(vars map[string]interface{}) bool {
			_, hasStartDate := vars["startDate"]
			_, hasEndDate := vars["endDate"]
			return hasStartDate && hasEndDate
		}),
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	budgetsWithGoals, err := service.ListWithGoals(ctx, startOfMonth, endOfMonth)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, budgetsWithGoals)
	assert.Len(t, budgetsWithGoals, 1)

	// Verify budget data
	budgetWithGoals := budgetsWithGoals[0]
	assert.NotNil(t, budgetWithGoals.Budget)
	assert.Equal(t, "cat-1", budgetWithGoals.Budget.CategoryID)
	assert.Equal(t, "Food", budgetWithGoals.Budget.Category.Name)
	assert.Equal(t, 500.00, budgetWithGoals.Budget.Amount)
	assert.Equal(t, 250.00, budgetWithGoals.Budget.Spent)
	assert.Equal(t, 250.00, budgetWithGoals.Budget.Remaining)

	// Verify goals data
	assert.NotNil(t, budgetWithGoals.Goals)
	assert.Len(t, budgetWithGoals.Goals, 1)

	goal := budgetWithGoals.Goals[0]
	assert.Equal(t, "goal-1", goal.ID)
	assert.Equal(t, "Emergency Fund", goal.Name)
	assert.Equal(t, "savings", goal.Type)
	assert.Equal(t, 10000.00, goal.TargetAmount)
	assert.Equal(t, 5000.00, goal.CurrentAmount)
	assert.Equal(t, 50.0, goal.PercentageComplete)
	assert.Equal(t, 500.00, goal.MonthlyContribution)
	assert.Equal(t, 1, goal.Priority)
	assert.NotNil(t, goal.Account)
	assert.Equal(t, "acc-1", goal.Account.ID)
	assert.Equal(t, "Savings Account", goal.Account.DisplayName)

	mockTransport.AssertExpectations(t)
}

func TestBudgetService_ListWithGoals_NoGoals(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &budgetService{client: client}

	// Mock response with budgets but no goals
	mockResponse := `{
		"budgetData": {
			"monthlyAmountsByCategory": [
				{
					"category": {
						"id": "cat-1",
						"name": "Food",
						"icon": "food",
						"group": {
							"id": "group-1",
							"name": "Living",
							"type": "expense"
						}
					},
					"monthlyAmounts": [
						{
							"month": "2025-08-01",
							"plannedCashFlowAmount": 500.00,
							"plannedSetAsideAmount": 0,
							"actualAmount": -250.00,
							"remainingAmount": 250.00,
							"previousMonthRolloverAmount": 0,
							"rolloverType": ""
						}
					]
				}
			]
		},
		"goalsV2": null
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	budgetsWithGoals, err := service.ListWithGoals(ctx, startOfMonth, endOfMonth)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, budgetsWithGoals)
	assert.Len(t, budgetsWithGoals, 1)

	// Verify budget data
	budgetWithGoals := budgetsWithGoals[0]
	assert.NotNil(t, budgetWithGoals.Budget)
	assert.Equal(t, "cat-1", budgetWithGoals.Budget.CategoryID)

	// Verify no goals
	assert.Nil(t, budgetWithGoals.Goals)

	mockTransport.AssertExpectations(t)
}

func TestBudgetService_ListWithGoals_EmptyResponse(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &budgetService{client: client}

	// Mock empty response
	mockResponse := `{
		"budgetData": null,
		"goalsV2": null
	}`

	mockTransport.On("Execute",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.Anything,
		mock.Anything,
	).Return(mockResponse, nil)

	// Execute
	ctx := context.Background()
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	budgetsWithGoals, err := service.ListWithGoals(ctx, startOfMonth, endOfMonth)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, budgetsWithGoals)
	assert.Empty(t, budgetsWithGoals)

	mockTransport.AssertExpectations(t)
}

func TestBudgetService_ListWithGoals_MultipleGoals(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	service := &budgetService{client: client}

	// Mock response with multiple goals
	mockResponse := `{
		"budgetData": {
			"monthlyAmountsByCategory": [
				{
					"category": {
						"id": "cat-1",
						"name": "Food",
						"icon": "food",
						"group": {
							"id": "group-1",
							"name": "Living",
							"type": "expense"
						}
					},
					"monthlyAmounts": [
						{
							"month": "2025-08-01",
							"plannedCashFlowAmount": 500.00,
							"plannedSetAsideAmount": 0,
							"actualAmount": -250.00,
							"remainingAmount": 250.00,
							"previousMonthRolloverAmount": 0,
							"rolloverType": ""
						}
					]
				}
			]
		},
		"goalsV2": {
			"goals": [
				{
					"id": "goal-1",
					"name": "Emergency Fund",
					"type": "savings",
					"amount": 10000.00,
					"priority": 1,
					"targetDate": "2025-12-31T00:00:00Z",
					"targetAmount": 10000.00,
					"currentAmount": 5000.00,
					"imageUrl": null,
					"accountId": "acc-1",
					"account": {
						"id": "acc-1",
						"displayName": "Savings Account"
					},
					"percentageComplete": 50.0,
					"monthlyContribution": 500.00,
					"createdAt": "2025-01-01T00:00:00Z",
					"updatedAt": "2025-08-01T00:00:00Z"
				},
				{
					"id": "goal-2",
					"name": "Vacation Fund",
					"type": "savings",
					"amount": 5000.00,
					"priority": 2,
					"targetDate": "2025-06-30T00:00:00Z",
					"targetAmount": 5000.00,
					"currentAmount": 2500.00,
					"imageUrl": "https://example.com/vacation.png",
					"accountId": null,
					"account": null,
					"percentageComplete": 50.0,
					"monthlyContribution": 250.00,
					"createdAt": "2025-01-01T00:00:00Z",
					"updatedAt": "2025-08-01T00:00:00Z"
				}
			]
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
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	budgetsWithGoals, err := service.ListWithGoals(ctx, startOfMonth, endOfMonth)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, budgetsWithGoals)
	assert.Len(t, budgetsWithGoals, 1)

	// Verify goals data
	budgetWithGoals := budgetsWithGoals[0]
	assert.NotNil(t, budgetWithGoals.Goals)
	assert.Len(t, budgetWithGoals.Goals, 2)

	// Check first goal
	assert.Equal(t, "goal-1", budgetWithGoals.Goals[0].ID)
	assert.Equal(t, "Emergency Fund", budgetWithGoals.Goals[0].Name)
	assert.NotNil(t, budgetWithGoals.Goals[0].Account)

	// Check second goal
	assert.Equal(t, "goal-2", budgetWithGoals.Goals[1].ID)
	assert.Equal(t, "Vacation Fund", budgetWithGoals.Goals[1].Name)
	assert.Nil(t, budgetWithGoals.Goals[1].Account)

	mockTransport.AssertExpectations(t)
}
