package main

import (
	"context"
	"fmt"
	"time"

	"github.com/eshaffer321/monarch-go/v2/pkg/monarch"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// monarchTools holds the Monarch client and implements all tool handlers
type monarchTools struct {
	client *monarch.Client
}

// GetBudget tool - retrieves budget information for a specific month
type GetBudgetInput struct {
	Month string `json:"month" jsonschema:"Month in YYYY-MM format (e.g. 2025-10)"`
}

type BudgetEntry struct {
	Category       string  `json:"category" jsonschema:"Budget category name"`
	Group          string  `json:"group" jsonschema:"Budget category group"`
	Budgeted       float64 `json:"budgeted" jsonschema:"Budgeted amount for this category"`
	Spent          float64 `json:"spent" jsonschema:"Actual amount spent"`
	Remaining      float64 `json:"remaining" jsonschema:"Remaining budget amount"`
	RolloverAmount float64 `json:"rolloverAmount" jsonschema:"Amount rolled over from previous month"`
	RolloverType   string  `json:"rolloverType,omitempty" jsonschema:"Type of rollover (if applicable)"`
	Percentage     float64 `json:"percentage" jsonschema:"Percentage of budget spent"`
}

type GoalEntry struct {
	ID                 string  `json:"id" jsonschema:"Goal ID"`
	Name               string  `json:"name" jsonschema:"Goal name"`
	CurrentBalance     float64 `json:"currentBalance" jsonschema:"Current balance toward goal"`
	TargetBalance      float64 `json:"targetBalance" jsonschema:"Target balance for goal"`
	Type               string  `json:"type" jsonschema:"Goal type"`
	PaceType           string  `json:"paceType,omitempty" jsonschema:"Pace status (on_track, behind, etc.)"`
	OriginAccount      string  `json:"originAccount,omitempty" jsonschema:"Origin account name"`
	DestinationAccount string  `json:"destinationAccount,omitempty" jsonschema:"Destination account name"`
}

type GetBudgetOutput struct {
	Month   string        `json:"month" jsonschema:"Month of the budget data"`
	Budgets []BudgetEntry `json:"budgets" jsonschema:"List of budget entries for each category"`
	Goals   []GoalEntry   `json:"goals,omitempty" jsonschema:"List of goals associated with this budget"`
}

func (t *monarchTools) GetBudget(ctx context.Context, req *mcp.CallToolRequest, input GetBudgetInput) (*mcp.CallToolResult, GetBudgetOutput, error) {

	// Parse the month
	startDate, err := time.Parse("2006-01", input.Month)
	if err != nil {
		return nil, GetBudgetOutput{}, fmt.Errorf("invalid month format (expected YYYY-MM): %w", err)
	}

	// Calculate end date (last day of the month)
	endDate := startDate.AddDate(0, 1, -1)

	// Fetch budgets from Monarch
	budgets, err := t.client.Budgets.List(ctx, startDate, endDate)
	if err != nil {
		return nil, GetBudgetOutput{}, fmt.Errorf("failed to fetch budgets: %w", err)
	}

	// Convert to output format
	var entries []BudgetEntry
	for _, b := range budgets {
		entry := BudgetEntry{
			Category:       b.Category.Name,
			Group:          b.Category.Group.Name,
			Budgeted:       b.Amount,
			Spent:          b.Spent,
			Remaining:      b.Remaining,
			RolloverAmount: b.RolloverAmount,
			RolloverType:   b.RolloverType,
			Percentage:     b.PercentageComplete,
		}

		entries = append(entries, entry)
	}

	return nil, GetBudgetOutput{
		Month:   input.Month,
		Budgets: entries,
		Goals:   nil, // Goals not currently supported in budget query
	}, nil
}

// GetTransactions tool - queries transactions with optional filters
type GetTransactionsInput struct {
	StartDate string `json:"startDate,omitempty" jsonschema:"Start date in YYYY-MM-DD format (optional)"`
	EndDate   string `json:"endDate,omitempty" jsonschema:"End date in YYYY-MM-DD format (optional)"`
	Category  string `json:"category,omitempty" jsonschema:"Filter by category name (optional)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of transactions to return (default: 50)"`
}

type TransactionEntry struct {
	ID          string    `json:"id" jsonschema:"Transaction ID"`
	Date        time.Time `json:"date" jsonschema:"Transaction date"`
	Amount      float64   `json:"amount" jsonschema:"Transaction amount (negative for expenses)"`
	Merchant    string    `json:"merchant" jsonschema:"Merchant name"`
	Category    string    `json:"category,omitempty" jsonschema:"Transaction category"`
	Account     string    `json:"account" jsonschema:"Account name"`
	Notes       string    `json:"notes,omitempty" jsonschema:"Transaction notes"`
	Pending     bool      `json:"pending" jsonschema:"Whether transaction is pending"`
	Tags        []string  `json:"tags,omitempty" jsonschema:"Transaction tags"`
}

type GetTransactionsOutput struct {
	Transactions []TransactionEntry `json:"transactions" jsonschema:"List of transactions"`
	Count        int                `json:"count" jsonschema:"Number of transactions returned"`
}

func (t *monarchTools) GetTransactions(ctx context.Context, req *mcp.CallToolRequest, input GetTransactionsInput) (*mcp.CallToolResult, GetTransactionsOutput, error) {

	// Build query
	query := t.client.Transactions.Query()

	// Parse and apply date filters
	if input.StartDate != "" || input.EndDate != "" {
		var startDate, endDate time.Time
		var err error

		if input.StartDate != "" {
			startDate, err = time.Parse("2006-01-02", input.StartDate)
			if err != nil {
				return nil, GetTransactionsOutput{}, fmt.Errorf("invalid startDate format (expected YYYY-MM-DD): %w", err)
			}
		}

		if input.EndDate != "" {
			endDate, err = time.Parse("2006-01-02", input.EndDate)
			if err != nil {
				return nil, GetTransactionsOutput{}, fmt.Errorf("invalid endDate format (expected YYYY-MM-DD): %w", err)
			}
		}

		if !startDate.IsZero() && !endDate.IsZero() {
			query = query.Between(startDate, endDate)
		} else if !startDate.IsZero() {
			// Start date only - go to today
			query = query.Between(startDate, time.Now())
		} else if !endDate.IsZero() {
			// End date only - go back 30 days
			query = query.Between(endDate.AddDate(0, 0, -30), endDate)
		}
	}

	// Apply limit (default to 50)
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}
	query = query.Limit(limit)

	// Execute query
	result, err := query.Execute(ctx)
	if err != nil {
		return nil, GetTransactionsOutput{}, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	// Filter by category if specified
	var transactions []TransactionEntry
	for _, tx := range result.Transactions {
		// Apply category filter if specified
		if input.Category != "" {
			if tx.Category == nil || tx.Category.Name != input.Category {
				continue
			}
		}

		entry := TransactionEntry{
			ID:      tx.ID,
			Date:    tx.Date.Time,
			Amount:  tx.Amount,
			Pending: tx.Pending,
			Notes:   tx.Notes,
		}

		if tx.Merchant != nil {
			entry.Merchant = tx.Merchant.Name
		}

		if tx.Category != nil {
			entry.Category = tx.Category.Name
		}

		if tx.Account != nil {
			entry.Account = tx.Account.DisplayName
		}

		if len(tx.Tags) > 0 {
			for _, tag := range tx.Tags {
				entry.Tags = append(entry.Tags, tag.Name)
			}
		}

		transactions = append(transactions, entry)
	}

	return nil, GetTransactionsOutput{
		Transactions: transactions,
		Count:        len(transactions),
	}, nil
}

// GetAccounts tool - retrieves all accounts
type GetAccountsInput struct {
	// No input parameters needed
}

type AccountEntry struct {
	ID              string  `json:"id" jsonschema:"Account ID"`
	Name            string  `json:"name" jsonschema:"Account display name"`
	Balance         float64 `json:"balance" jsonschema:"Current account balance"`
	Type            string  `json:"type" jsonschema:"Account type (e.g. checking, savings, credit)"`
	Subtype         string  `json:"subtype,omitempty" jsonschema:"Account subtype"`
	Institution     string  `json:"institution,omitempty" jsonschema:"Financial institution name"`
	IsHidden        bool    `json:"isHidden" jsonschema:"Whether account is hidden"`
	IncludeInNetWorth bool `json:"includeInNetWorth" jsonschema:"Whether account is included in net worth calculation"`
}

type GetAccountsOutput struct {
	Accounts []AccountEntry `json:"accounts" jsonschema:"List of all accounts"`
	Count    int            `json:"count" jsonschema:"Number of accounts"`
}

func (t *monarchTools) GetAccounts(ctx context.Context, req *mcp.CallToolRequest, input GetAccountsInput) (*mcp.CallToolResult, GetAccountsOutput, error) {
	// Fetch accounts from Monarch
	accounts, err := t.client.Accounts.List(ctx)
	if err != nil {
		return nil, GetAccountsOutput{}, fmt.Errorf("failed to fetch accounts: %w", err)
	}

	// Convert to output format
	var entries []AccountEntry
	for _, acc := range accounts {
		entry := AccountEntry{
			ID:                acc.ID,
			Name:              acc.DisplayName,
			Balance:           acc.DisplayBalance,
			Type:              string(acc.Type.Name),
			IsHidden:          acc.IsHidden,
			IncludeInNetWorth: acc.IncludeInNetWorth,
		}

		if acc.Subtype != nil {
			entry.Subtype = string(acc.Subtype.Name)
		}

		if acc.Institution != nil {
			entry.Institution = acc.Institution.Name
		}

		entries = append(entries, entry)
	}

	return nil, GetAccountsOutput{
		Accounts: entries,
		Count:    len(entries),
	}, nil
}

// GetCategories tool - retrieves all transaction categories
type GetCategoriesInput struct {
	// No input parameters needed
}

type CategoryEntry struct {
	ID          string `json:"id" jsonschema:"Category ID"`
	Name        string `json:"name" jsonschema:"Category name"`
	Group       string `json:"group" jsonschema:"Category group name"`
	IsSystemCategory bool `json:"isSystemCategory" jsonschema:"Whether this is a system category"`
	IsDisabled  bool   `json:"isDisabled" jsonschema:"Whether category is disabled"`
}

type GetCategoriesOutput struct {
	Categories []CategoryEntry `json:"categories" jsonschema:"List of all categories"`
	Count      int             `json:"count" jsonschema:"Number of categories"`
}

func (t *monarchTools) GetCategories(ctx context.Context, req *mcp.CallToolRequest, input GetCategoriesInput) (*mcp.CallToolResult, GetCategoriesOutput, error) {
	// Fetch categories from Monarch
	categories, err := t.client.Transactions.Categories().List(ctx)
	if err != nil {
		return nil, GetCategoriesOutput{}, fmt.Errorf("failed to fetch categories: %w", err)
	}

	// Convert to output format
	var entries []CategoryEntry
	for _, cat := range categories {
		entry := CategoryEntry{
			ID:               cat.ID,
			Name:             cat.Name,
			IsSystemCategory: cat.IsSystemCategory,
			IsDisabled:       cat.IsDisabled,
		}

		if cat.Group != nil {
			entry.Group = cat.Group.Name
		}

		entries = append(entries, entry)
	}

	return nil, GetCategoriesOutput{
		Categories: entries,
		Count:      len(entries),
	}, nil
}

// GetTags tool - retrieves all tags
type GetTagsInput struct {
	// No input parameters needed
}

type TagEntry struct {
	ID    string `json:"id" jsonschema:"Tag ID"`
	Name  string `json:"name" jsonschema:"Tag name"`
	Color string `json:"color,omitempty" jsonschema:"Tag color (hex code)"`
	Order int    `json:"order" jsonschema:"Tag display order"`
}

type GetTagsOutput struct {
	Tags  []TagEntry `json:"tags" jsonschema:"List of all tags"`
	Count int        `json:"count" jsonschema:"Number of tags"`
}

func (t *monarchTools) GetTags(ctx context.Context, req *mcp.CallToolRequest, input GetTagsInput) (*mcp.CallToolResult, GetTagsOutput, error) {
	// Fetch tags from Monarch
	tags, err := t.client.Tags.List(ctx)
	if err != nil {
		return nil, GetTagsOutput{}, fmt.Errorf("failed to fetch tags: %w", err)
	}

	// Convert to output format
	var entries []TagEntry
	for _, tag := range tags {
		entry := TagEntry{
			ID:    tag.ID,
			Name:  tag.Name,
			Color: tag.Color,
			Order: tag.Order,
		}

		entries = append(entries, entry)
	}

	return nil, GetTagsOutput{
		Tags:  entries,
		Count: len(entries),
	}, nil
}
