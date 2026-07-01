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

func TestAccountService_GetBalances_JSONString(t *testing.T) {
	// Setup
	mockTransport := new(MockTransport)
	client := &Client{
		transport:   mockTransport,
		queryLoader: graphql.NewQueryLoader(),
		options:     &ClientOptions{},
		baseURL:     "https://api.test.com",
	}
	client.Accounts = &accountService{client: client}

	// This is the actual response from the API that's causing the unmarshal error
	// The recentBalances field is a JSON string, not an array of maps
	// Note: MockTransport expects just the data portion, not the full GraphQL response
	mockResponse := `{
		"accounts": [
			{
				"id": "123",
				"recentBalances": "[{\"date\":\"2025-08-01\",\"balance\":1000.50},{\"date\":\"2025-08-02\",\"balance\":1100.75}]"
			},
			{
				"id": "456",
				"recentBalances": "[{\"date\":\"2025-08-01\",\"balance\":2000.00}]"
			}
		]
	}`

	// Mock the GraphQL call
	mockTransport.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Run(func(args mock.Arguments) {
		// Verify the request
		query := args.Get(1).(string)
		assert.Contains(t, query, "GetAccountRecentBalances")

		variables := args.Get(2).(map[string]interface{})
		assert.Equal(t, "2025-08-01", variables["startDate"])
	})

	startDate := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
	balances, err := client.Accounts.GetBalances(context.Background(), &startDate)

	// This test should pass after we fix the unmarshaling
	require.NoError(t, err)
	t.Logf("Got %d balances: %+v", len(balances), balances)
	require.Len(t, balances, 3) // Total of 3 balance entries across both accounts

	// Check first balance
	assert.Equal(t, "123", balances[0].AccountID)
	assert.Equal(t, "2025-08-01", balances[0].Date.Format("2006-01-02"))
	assert.Equal(t, 1000.50, balances[0].Balance)

	// Check second balance
	assert.Equal(t, "123", balances[1].AccountID)
	assert.Equal(t, "2025-08-02", balances[1].Date.Format("2006-01-02"))
	assert.Equal(t, 1100.75, balances[1].Balance)

	// Check third balance
	assert.Equal(t, "456", balances[2].AccountID)
	assert.Equal(t, "2025-08-01", balances[2].Date.Format("2006-01-02"))
	assert.Equal(t, 2000.00, balances[2].Balance)

	mockTransport.AssertExpectations(t)
}
