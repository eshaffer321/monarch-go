package monarch

import (
	"context"
	"testing"

	"github.com/eshaffer321/monarch-go/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionCategoryService_List_New(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	response := `{
		
			"categories": [
				{
					"id": "cat-1",
					"name": "Groceries",
					"order": 1,
					"systemCategory": "FOOD",
					"isSystemCategory": false,
					"isDisabled": false,
					"group": {
						"id": "group-1",
						"name": "Essential Expenses",
						"type": "expense"
					}
				},
				{
					"id": "cat-2",
					"name": "Entertainment",
					"order": 2,
					"systemCategory": "",
					"isSystemCategory": false,
					"isDisabled": false,
					"group": {
						"id": "group-2",
						"name": "Variable Expenses",
						"type": "expense"
					}
				}
			]
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	categories, err := client.Transactions.Categories().List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, categories, 2)
	assert.Equal(t, "cat-1", categories[0].ID)
	assert.Equal(t, "Groceries", categories[0].Name)
	assert.Equal(t, "Essential Expenses", categories[0].Group.Name)
	assert.Equal(t, "cat-2", categories[1].ID)
	assert.Equal(t, "Entertainment", categories[1].Name)
	
	mockTransport.AssertExpectations(t)
}

func TestTransactionCategoryService_Create(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	response := `{
		
			"createCategory": {
				"category": {
					"id": "new-cat-123",
					"name": "Test Category",
					"order": 5,
					"systemCategory": "",
					"isSystemCategory": false,
					"isDisabled": false,
					"group": {
						"id": "group-1",
						"name": "Variable Expenses",
						"type": "expense"
					}
				},
				"errors": []
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.MatchedBy(func(vars map[string]interface{}) bool {
		input := vars["input"].(map[string]interface{})
		return input["name"] == "Test Category" &&
			input["group"] == "group-1" &&
			input["icon"] == "❓"
	}), mock.Anything).Return(response, nil)

	params := &CreateCategoryParams{
		Name:    "Test Category",
		GroupID: "group-1",
		Icon:    "❓",
	}

	category, err := client.Transactions.Categories().Create(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, category)
	assert.Equal(t, "new-cat-123", category.ID)
	assert.Equal(t, "Test Category", category.Name)
	assert.Equal(t, "Variable Expenses", category.Group.Name)
	
	mockTransport.AssertExpectations(t)
}

func TestTransactionCategoryService_Delete(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	response := `{
		
			"deleteCategory": {
				"deleted": true,
				"errors": []
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.MatchedBy(func(vars map[string]interface{}) bool {
		return vars["id"] == "cat-to-delete"
	}), mock.Anything).Return(response, nil)

	err := client.Transactions.Categories().Delete(context.Background(), "cat-to-delete")

	assert.NoError(t, err)
	mockTransport.AssertExpectations(t)
}

func TestTransactionCategoryService_Delete_NotDeleted(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	response := `{
		
			"deleteCategory": {
				"deleted": false,
				"errors": []
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	err := client.Transactions.Categories().Delete(context.Background(), "cat-to-delete")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category was not deleted")
	mockTransport.AssertExpectations(t)
}

func TestTransactionCategoryService_GetGroups(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.initServices()

	response := `{
		
			"categoryGroups": [
				{
					"id": "group-1",
					"name": "Income",
					"order": 1,
					"type": "income",
					"updatedAt": "2025-01-01T00:00:00Z",
					"createdAt": "2025-01-01T00:00:00Z"
				},
				{
					"id": "group-2",
					"name": "Essential Expenses",
					"order": 2,
					"type": "expense",
					"updatedAt": "2025-01-01T00:00:00Z",
					"createdAt": "2025-01-01T00:00:00Z"
				}
			]
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	groups, err := client.Transactions.Categories().GetGroups(context.Background())

	assert.NoError(t, err)
	assert.Len(t, groups, 2)
	assert.Equal(t, "group-1", groups[0].ID)
	assert.Equal(t, "Income", groups[0].Name)
	assert.Equal(t, "income", groups[0].Type)
	assert.Equal(t, "group-2", groups[1].ID)
	assert.Equal(t, "Essential Expenses", groups[1].Name)
	assert.Equal(t, "expense", groups[1].Type)
	
	mockTransport.AssertExpectations(t)
}