package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eshaffer321/monarch-go/v2/pkg/monarch"
)

// This example demonstrates all available services and their methods
func main() {
	// Create client
	client, err := monarch.NewClient(&monarch.ClientOptions{
		Token:       "your-token-here",
		SessionFile: ".monarch_session.json",
		Timeout:     30 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Demonstrate each service
	demonstrateAccountService(ctx, client)
	demonstrateTransactionService(ctx, client)
	demonstrateTagService(ctx, client)
	demonstrateBudgetService(ctx, client)
	demonstrateCashflowService(ctx, client)
	demonstrateRecurringService(ctx, client)
	demonstrateInstitutionService(ctx, client)
	demonstrateAdminService(ctx, client)
}

func demonstrateAccountService(ctx context.Context, client *monarch.Client) {
	fmt.Println("\n=== Account Service Demo ===")

	// List all accounts
	accounts, err := client.Accounts.List(ctx)
	if err != nil {
		log.Printf("Error listing accounts: %v", err)
		return
	}
	fmt.Printf("Found %d accounts\n", len(accounts))

	if len(accounts) > 0 {
		accountID := accounts[0].ID

		// Get single account
		account, err := client.Accounts.Get(ctx, accountID)
		if err != nil {
			log.Printf("Error getting account: %v", err)
		} else {
			fmt.Printf("Account: %s, Balance: $%.2f\n", account.DisplayName, account.CurrentBalance)
		}

		// Get account holdings (for investment accounts)
		holdings, err := client.Accounts.GetHoldings(ctx, accountID)
		if err != nil {
			log.Printf("Error getting holdings: %v", err)
		} else {
			fmt.Printf("Found %d holdings\n", len(holdings))
		}

		// Get account history
		history, err := client.Accounts.GetHistory(ctx, accountID)
		if err != nil {
			log.Printf("Error getting history: %v", err)
		} else {
			fmt.Printf("Account history has %d balance entries\n", len(history.Balances))
		}
	}

	// Get account types
	types, err := client.Accounts.GetTypes(ctx)
	if err != nil {
		log.Printf("Error getting account types: %v", err)
	} else {
		fmt.Printf("Available account types: %d\n", len(types))
	}

	// Get recent balances
	balances, err := client.Accounts.GetBalances(ctx, nil)
	if err != nil {
		log.Printf("Error getting balances: %v", err)
	} else {
		fmt.Printf("Retrieved %d balance records\n", len(balances))
	}

	// Create manual account example (commented out to avoid creating real accounts)
	/*
		newAccount, err := client.Accounts.Create(ctx, &monarch.CreateAccountParams{
			AccountType:       "other_asset",
			AccountSubtype:    "cash",
			IsAsset:          true,
			AccountName:      "Test Savings",
			CurrentBalance:   1000.00,
			IncludeInNetWorth: true,
		})
		if err != nil {
			log.Printf("Error creating account: %v", err)
		} else {
			fmt.Printf("Created account: %s\n", newAccount.ID)
		}
	*/

	// Refresh accounts
	if len(accounts) > 0 {
		job, err := client.Accounts.Refresh(ctx, accounts[0].ID)
		if err != nil {
			log.Printf("Error refreshing accounts: %v", err)
		} else {
			fmt.Printf("Refresh job started: %s\n", job.ID())

			// Wait for completion (with timeout)
			err = job.Wait(ctx, 30*time.Second)
			if err != nil {
				log.Printf("Refresh timeout or error: %v", err)
			} else {
				fmt.Println("Refresh completed successfully")
			}
		}
	}
}

func demonstrateTransactionService(ctx context.Context, client *monarch.Client) {
	fmt.Println("\n=== Transaction Service Demo ===")

	// Simple query
	transactions, err := client.Transactions.Query().
		Limit(10).
		Execute(ctx)
	if err != nil {
		log.Printf("Error querying transactions: %v", err)
		return
	}
	fmt.Printf("Found %d transactions (total: %d)\n", len(transactions.Transactions), transactions.TotalCount)

	// Complex query with builder pattern
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	filtered, err := client.Transactions.Query().
		Between(startDate, endDate).
		WithMinAmount(50).
		WithMaxAmount(500).
		Search("food").
		Limit(20).
		Execute(ctx)
	if err != nil {
		log.Printf("Error with filtered query: %v", err)
	} else {
		fmt.Printf("Filtered transactions: %d\n", len(filtered.Transactions))
	}

	// Get transaction details
	if len(transactions.Transactions) > 0 {
		txnID := transactions.Transactions[0].ID

		details, err := client.Transactions.Get(ctx, txnID)
		if err != nil {
			log.Printf("Error getting transaction details: %v", err)
		} else {
			fmt.Printf("Transaction: %s on %s for $%.2f\n",
				details.Merchant,
				details.Date.Format("2006-01-02"),
				details.Amount)
		}

		// Get splits if it's a split transaction
		if details.IsSplitTransaction {
			splits, err := client.Transactions.GetSplits(ctx, txnID)
			if err != nil {
				log.Printf("Error getting splits: %v", err)
			} else {
				fmt.Printf("Transaction has %d splits\n", len(splits))
			}
		}
	}

	// Get transaction summary
	summary, err := client.Transactions.GetSummary(ctx)
	if err != nil {
		log.Printf("Error getting summary: %v", err)
	} else {
		fmt.Printf("Transaction Summary - Income: $%.2f, Expenses: $%.2f\n",
			summary.SumIncome,
			summary.SumExpense)
	}

	// Get categories
	categories, err := client.Transactions.Categories().List(ctx)
	if err != nil {
		log.Printf("Error getting categories: %v", err)
	} else {
		fmt.Printf("Found %d transaction categories\n", len(categories))
	}

	// Get category groups
	groups, err := client.Transactions.Categories().GetGroups(ctx)
	if err != nil {
		log.Printf("Error getting category groups: %v", err)
	} else {
		fmt.Printf("Found %d category groups\n", len(groups))
	}

	// Stream transactions (for large datasets)
	fmt.Println("Streaming transactions...")
	txnChan, errChan := client.Transactions.Query().
		Between(startDate, endDate).
		Stream(ctx)

	count := 0
streamLoop:
	for {
		select {
		case _, ok := <-txnChan:
			if !ok {
				break streamLoop
			}
			count++
			if count >= 10 {
				break streamLoop // Just demo first 10
			}
		case err := <-errChan:
			if err != nil {
				log.Printf("Stream error: %v", err)
				break streamLoop
			}
		}
	}
	fmt.Printf("Streamed %d transactions\n", count)
}

func demonstrateTagService(ctx context.Context, client *monarch.Client) {
	fmt.Println("\n=== Tag Service Demo ===")

	// List all tags
	tags, err := client.Tags.List(ctx)
	if err != nil {
		log.Printf("Error listing tags: %v", err)
		return
	}
	fmt.Printf("Found %d tags\n", len(tags))

	// Create a new tag (commented out to avoid creating real tags)
	/*
		newTag, err := client.Tags.Create(ctx, "Vacation", "#FF5733")
		if err != nil {
			log.Printf("Error creating tag: %v", err)
		} else {
			fmt.Printf("Created tag: %s\n", newTag.Name)
		}
	*/

	// Set tags on a transaction (example)
	/*
		if len(tags) > 0 {
			err = client.Tags.SetTransactionTags(ctx, "transaction-id", tags[0].ID)
			if err != nil {
				log.Printf("Error setting transaction tags: %v", err)
			}
		}
	*/
}

func demonstrateBudgetService(ctx context.Context, client *monarch.Client) {
	fmt.Println("\n=== Budget Service Demo ===")

	// Get budgets for current month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	budgets, err := client.Budgets.List(ctx, startOfMonth, endOfMonth)
	if err != nil {
		log.Printf("Error listing budgets: %v", err)
		return
	}
	fmt.Printf("Found %d budgets for current month\n", len(budgets))

	for _, budget := range budgets[:min(3, len(budgets))] {
		fmt.Printf("  - %s: $%.2f budgeted, $%.2f spent (%.1f%%)\n",
			budget.Category.Name,
			budget.Amount,
			budget.Spent,
			budget.PercentageComplete)
	}

	// Set budget amount (example - commented out)
	/*
		if len(budgets) > 0 {
			err = client.Budgets.SetAmount(ctx, budgets[0].ID, 500.00, false, startOfMonth)
			if err != nil {
				log.Printf("Error setting budget amount: %v", err)
			}
		}
	*/
}

func demonstrateCashflowService(ctx context.Context, client *monarch.Client) {
	fmt.Println("\n=== Cashflow Service Demo ===")

	// Get cashflow for last month
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0)

	cashflow, err := client.Cashflow.Get(ctx, &monarch.CashflowParams{
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     10,
	})
	if err != nil {
		log.Printf("Error getting cashflow: %v", err)
		return
	}

	fmt.Printf("Cashflow for last month:\n")
	if cashflow.Summary != nil {
		fmt.Printf("  Income: $%.2f\n", cashflow.Summary.Income)
		fmt.Printf("  Expenses: $%.2f\n", cashflow.Summary.Expense)
		fmt.Printf("  Savings: $%.2f\n", cashflow.Summary.Savings)
	}

	// Get cashflow summary
	summary, err := client.Cashflow.GetSummary(ctx, &monarch.CashflowSummaryParams{
		StartDate: startDate,
		EndDate:   endDate,
		Interval:  "week",
	})
	if err != nil {
		log.Printf("Error getting cashflow summary: %v", err)
	} else {
		fmt.Printf("Cashflow summary: Income: $%.2f, Expenses: $%.2f\n", summary.Income, summary.Expense)
	}
}

func demonstrateRecurringService(ctx context.Context, client *monarch.Client) {
	fmt.Println("\n=== Recurring Service Demo ===")

	recurring, err := client.Recurring.List(ctx)
	if err != nil {
		log.Printf("Error listing recurring transactions: %v", err)
		return
	}

	fmt.Printf("Found %d recurring transactions\n", len(recurring))
	for _, r := range recurring[:min(3, len(recurring))] {
		fmt.Printf("  - %s: $%.2f %s (next: %s)\n",
			r.Merchant,
			r.Amount,
			r.Frequency,
			r.NextDate.Format("2006-01-02"))
	}
}

func demonstrateInstitutionService(ctx context.Context, client *monarch.Client) {
	fmt.Println("\n=== Institution Service Demo ===")

	institutions, err := client.Institutions.List(ctx)
	if err != nil {
		log.Printf("Error listing institutions: %v", err)
		return
	}

	fmt.Printf("Connected to %d institutions\n", len(institutions))
	for _, inst := range institutions {
		fmt.Printf("  - %s (Status: %s)\n", inst.Name, inst.Status)
	}
}

func demonstrateAdminService(ctx context.Context, client *monarch.Client) {
	fmt.Println("\n=== Admin Service Demo ===")

	// Get subscription details
	subscription, err := client.Admin.GetSubscription(ctx)
	if err != nil {
		log.Printf("Error getting subscription: %v", err)
		return
	}

	fmt.Printf("Subscription: %s (Status: %s)\n", subscription.PlanType, subscription.Status)
	fmt.Printf("Features: %v\n", subscription.Features)

	// Upload balance history example (commented out)
	/*
		csvData := []byte("date,balance\n2024-01-01,1000.00\n2024-01-02,1050.00")
		err = client.Admin.UploadBalanceHistory(ctx, "account-id", csvData)
		if err != nil {
			log.Printf("Error uploading balance history: %v", err)
		}
	*/
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
