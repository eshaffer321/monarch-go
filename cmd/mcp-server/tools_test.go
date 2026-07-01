package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/eshaffer321/monarch-go/v2/pkg/monarch"
)

func TestGetAccountsTool(t *testing.T) {
	token := os.Getenv("MONARCH_TOKEN")
	if token == "" {
		t.Skip("MONARCH_TOKEN not set")
	}

	client, err := monarch.NewClient(&monarch.ClientOptions{
		Token: token,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tools := &monarchTools{client: client}

	callResult, output, err := tools.GetAccounts(context.Background(), nil, GetAccountsInput{})

	if err != nil {
		t.Fatalf("GetAccounts failed: %v", err)
	}

	if output.Count == 0 {
		t.Error("Expected at least one account")
	}

	t.Logf("✓ GetAccounts returned %d accounts (callResult=%v)", output.Count, callResult)

	// Pretty print first account
	if len(output.Accounts) > 0 {
		jsonData, _ := json.MarshalIndent(output.Accounts[0], "", "  ")
		t.Logf("First account:\n%s", string(jsonData))
	}
}

func TestGetCategoriesTo(t *testing.T) {
	token := os.Getenv("MONARCH_TOKEN")
	if token == "" {
		t.Skip("MONARCH_TOKEN not set")
	}

	client, err := monarch.NewClient(&monarch.ClientOptions{
		Token: token,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tools := &monarchTools{client: client}

	callResult, output, err := tools.GetCategories(context.Background(), nil, GetCategoriesInput{})

	if err != nil {
		t.Fatalf("GetCategories failed: %v", err)
	}

	if output.Count == 0 {
		t.Error("Expected at least one category")
	}

	t.Logf("✓ GetCategories returned %d categories (callResult=%v)", output.Count, callResult)
}

func TestGetTagsTool(t *testing.T) {
	token := os.Getenv("MONARCH_TOKEN")
	if token == "" {
		t.Skip("MONARCH_TOKEN not set")
	}

	client, err := monarch.NewClient(&monarch.ClientOptions{
		Token: token,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tools := &monarchTools{client: client}

	callResult, output, err := tools.GetTags(context.Background(), nil, GetTagsInput{})

	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	t.Logf("✓ GetTags returned %d tags (callResult=%v)", output.Count, callResult)
}

func TestGetBudgetTool(t *testing.T) {
	token := os.Getenv("MONARCH_TOKEN")
	if token == "" {
		t.Skip("MONARCH_TOKEN not set")
	}

	client, err := monarch.NewClient(&monarch.ClientOptions{
		Token: token,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tools := &monarchTools{client: client}

	// Test current month
	now := time.Now()
	month := now.Format("2006-01")

	callResult, output, err := tools.GetBudget(context.Background(), nil, GetBudgetInput{Month: month})

	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}

	if len(output.Budgets) == 0 {
		t.Error("Expected at least one budget entry")
	}

	t.Logf("✓ GetBudget returned %d budget entries for %s (callResult=%v)", len(output.Budgets), month, callResult)

	// Pretty print first budget entry
	if len(output.Budgets) > 0 {
		jsonData, _ := json.MarshalIndent(output.Budgets[0], "", "  ")
		t.Logf("First budget entry:\n%s", string(jsonData))
	}
}

func TestGetTransactionsTool(t *testing.T) {
	token := os.Getenv("MONARCH_TOKEN")
	if token == "" {
		t.Skip("MONARCH_TOKEN not set")
	}

	client, err := monarch.NewClient(&monarch.ClientOptions{
		Token: token,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tools := &monarchTools{client: client}

	// Test last 7 days
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	callResult, output, err := tools.GetTransactions(context.Background(), nil, GetTransactionsInput{
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     10,
	})

	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}

	t.Logf("✓ GetTransactions returned %d transactions (callResult=%v)", output.Count, callResult)

	// Pretty print first transaction
	if len(output.Transactions) > 0 {
		jsonData, _ := json.MarshalIndent(output.Transactions[0], "", "  ")
		t.Logf("First transaction:\n%s", string(jsonData))
	}
}
