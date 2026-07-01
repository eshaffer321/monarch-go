package monarch

import (
	"context"
	"testing"

	"github.com/eshaffer321/monarch-go/v2/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInstitutionService_List(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.Institutions = &institutionService{client: client}

	// Mock response based on the actual GraphQL query structure
	mockResponse := `{
		"credentials": [
			{
				"id": "cred1",
				"updateRequired": false,
				"disconnectedFromDataProviderAt": null,
				"displayLastUpdatedAt": "2025-08-30T10:00:00Z",
				"dataProvider": "plaid",
				"institution": {
					"id": "inst1",
					"name": "Chase Bank",
					"url": "https://chase.com",
					"logoUrl": "https://example.com/chase.png",
					"primaryColor": "#0066CC"
				}
			},
			{
				"id": "cred2",
				"updateRequired": true,
				"disconnectedFromDataProviderAt": "2025-08-29T15:00:00Z",
				"displayLastUpdatedAt": "2025-08-29T14:00:00Z",
				"dataProvider": "finicity",
				"institution": {
					"id": "inst2",
					"name": "Bank of America",
					"url": "https://bankofamerica.com",
					"logoUrl": "https://example.com/boa.png",
					"primaryColor": "#E31837"
				}
			}
		]
	}`

	// Mock the GraphQL call
	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Run(func(args mock.Arguments) {
		// Verify the request
		query := args.Get(1).(string)
		assert.Contains(t, query, "credentials")
		assert.Contains(t, query, "institution")

		// Should have no variables for this query
		variables := args.Get(2)
		assert.Nil(t, variables)
	})

	// Execute
	institutions, err := client.Institutions.List(context.Background())

	// Verify
	require.NoError(t, err)
	require.Len(t, institutions, 2)

	// Check first institution
	assert.Equal(t, "inst1", institutions[0].ID)
	assert.Equal(t, "Chase Bank", institutions[0].Name)
	assert.Equal(t, "https://chase.com", institutions[0].URL)
	assert.Equal(t, "cred1", institutions[0].CredentialID)
	assert.False(t, institutions[0].UpdateRequired)
	assert.Equal(t, "plaid", institutions[0].DataProvider)

	// Check second institution
	assert.Equal(t, "inst2", institutions[1].ID)
	assert.Equal(t, "Bank of America", institutions[1].Name)
	assert.True(t, institutions[1].UpdateRequired)
	assert.Equal(t, "finicity", institutions[1].DataProvider)

	mockTransport.AssertExpectations(t)
}
