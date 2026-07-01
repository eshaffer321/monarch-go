package monarch

import (
	"context"
	"testing"

	"github.com/eshaffer321/monarch-go/v2/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTagService_List(t *testing.T) {
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
		
			"householdTransactionTags": [
				{
					"id": "tag-1",
					"name": "Tax Deductible",
					"color": "#FF5733",
					"order": 1
				},
				{
					"id": "tag-2",
					"name": "Reimbursable",
					"color": "#33FF57",
					"order": 2
				}
			]
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	tags, err := client.Tags.List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, tags, 2)
	assert.Equal(t, "tag-1", tags[0].ID)
	assert.Equal(t, "Tax Deductible", tags[0].Name)
	assert.Equal(t, "#FF5733", tags[0].Color)
	assert.Equal(t, 1, tags[0].Order)
	assert.Equal(t, "tag-2", tags[1].ID)
	assert.Equal(t, "Reimbursable", tags[1].Name)
	assert.Equal(t, 2, tags[1].Order)
	
	mockTransport.AssertExpectations(t)
}

func TestTagService_Create(t *testing.T) {
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
		
			"createTransactionTag": {
				"tag": {
					"id": "new-tag-123",
					"name": "New Tag",
					"color": "#123456",
					"order": 5
				},
				"errors": []
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.MatchedBy(func(vars map[string]interface{}) bool {
		input := vars["input"].(map[string]interface{})
		return input["name"] == "New Tag" && input["color"] == "#123456"
	}), mock.Anything).Return(response, nil)

	tag, err := client.Tags.Create(context.Background(), "New Tag", "#123456")

	assert.NoError(t, err)
	assert.NotNil(t, tag)
	assert.Equal(t, "new-tag-123", tag.ID)
	assert.Equal(t, "New Tag", tag.Name)
	assert.Equal(t, "#123456", tag.Color)
	assert.Equal(t, 5, tag.Order)
	
	mockTransport.AssertExpectations(t)
}

func TestTagService_Create_WithError(t *testing.T) {
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
		
			"createTransactionTag": {
				"tag": null,
				"errors": [
					{
						"message": "Tag name already exists",
						"code": "DUPLICATE_TAG"
					}
				]
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(response, nil)

	tag, err := client.Tags.Create(context.Background(), "Duplicate Tag", "#000000")

	assert.Error(t, err)
	assert.Nil(t, tag)
	assert.Contains(t, err.Error(), "Tag name already exists")
	
	mockTransport.AssertExpectations(t)
}

func TestTagService_SetTransactionTags(t *testing.T) {
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
		
			"setTransactionTags": {
				"transaction": {
					"id": "txn-123",
					"tags": [
						{"id": "tag-1"},
						{"id": "tag-2"}
					]
				},
				"errors": []
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.MatchedBy(func(vars map[string]interface{}) bool {
		input := vars["input"].(map[string]interface{})
		tagIds := input["tagIds"].([]string)
		return input["transactionId"] == "txn-123" && 
			len(tagIds) == 2 && 
			tagIds[0] == "tag-1" && 
			tagIds[1] == "tag-2"
	}), mock.Anything).Return(response, nil)

	err := client.Tags.SetTransactionTags(context.Background(), "txn-123", "tag-1", "tag-2")

	assert.NoError(t, err)
	mockTransport.AssertExpectations(t)
}

func TestTagService_SetTransactionTags_RemoveAll(t *testing.T) {
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
		
			"setTransactionTags": {
				"transaction": {
					"id": "txn-123",
					"tags": []
				},
				"errors": []
			}
	}`

	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.MatchedBy(func(vars map[string]interface{}) bool {
		input := vars["input"].(map[string]interface{})
		tagIds := input["tagIds"].([]string)
		return input["transactionId"] == "txn-123" && len(tagIds) == 0
	}), mock.Anything).Return(response, nil)

	// Calling with no tag IDs removes all tags
	err := client.Tags.SetTransactionTags(context.Background(), "txn-123")

	assert.NoError(t, err)
	mockTransport.AssertExpectations(t)
}