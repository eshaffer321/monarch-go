package main

import (
	"context"
	"fmt"
	"log"

	"github.com/eshaffer321/monarch-go/v2/pkg/monarch"
)

// This example demonstrates transaction deletion and the hideFromReports workaround
func main() {
	// Initialize client
	client, err := monarch.NewClient(&monarch.ClientOptions{
		Token: "your-session-token",
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	ctx := context.Background()

	// Example transaction ID
	transactionID := "225271673011145605"

	// ========================================
	// OPTION 1: Delete Transaction (Recommended)
	// ========================================
	// As of v1.0.1, the Delete method now uses the correct GraphQL mutation format
	// that matches the Python client and Monarch API expectations.
	//
	// This should work for:
	// ✅ Manually created transactions
	// ❓ Bank-imported transactions (may still return BAD_REQUEST)

	err = client.Transactions.Delete(ctx, transactionID)
	if err != nil {
		// If deletion fails, check if it's a structured error
		if apiErr, ok := err.(*monarch.Error); ok {
			log.Printf("Delete failed: %s - %s", apiErr.Code, apiErr.Message)

			// If deletion is not allowed (e.g., bank-imported transactions),
			// fall back to hideFromReports workaround
			if apiErr.Code == "BAD_REQUEST" {
				log.Println("Bank-imported transactions cannot be deleted. Using hideFromReports workaround...")
				if err := hideFromReportsWorkaround(client, ctx, transactionID); err != nil {
					log.Fatalf("Workaround failed: %v", err)
				}
				log.Println("Transaction successfully hidden from reports")
			}
		} else {
			log.Fatalf("Delete failed: %v", err)
		}
	} else {
		log.Println("Transaction successfully deleted")
	}

	// ========================================
	// OPTION 2: Hide from Reports (Alternative)
	// ========================================
	// If deletion fails or you want to preserve the transaction for record-keeping,
	// you can hide it from reports instead. This is useful for:
	//
	// • Duplicate transactions (like multiple Walmart deliveries for one order)
	// • Transactions you want to exclude from budgets/reports
	// • Bank-imported transactions that cannot be deleted
	//
	// Hidden transactions:
	// ✅ Still visible in transaction history
	// ✅ Can be unhidden later
	// ❌ Excluded from budget calculations
	// ❌ Excluded from reports and summaries

	hideFromReportsExample(client, ctx, transactionID)
}

// hideFromReportsWorkaround demonstrates hiding a transaction from reports
func hideFromReportsWorkaround(client *monarch.Client, ctx context.Context, transactionID string) error {
	hideFromReports := true
	params := &monarch.UpdateTransactionParams{
		HideFromReports: &hideFromReports,
	}

	_, err := client.Transactions.Update(ctx, transactionID, params)
	return err
}

// hideFromReportsExample shows the full workflow for hiding transactions
func hideFromReportsExample(client *monarch.Client, ctx context.Context, transactionID string) {
	// Example: Hide a transaction from reports
	hideFromReports := true
	params := &monarch.UpdateTransactionParams{
		HideFromReports: &hideFromReports,
	}

	transaction, err := client.Transactions.Update(ctx, transactionID, params)
	if err != nil {
		log.Fatalf("Failed to hide transaction: %v", err)
	}

	fmt.Printf("Transaction %s is now hidden from reports: %v\n",
		transaction.ID, transaction.HideFromReports)

	// Example: Unhide a transaction (restore to reports)
	showInReports := false
	params = &monarch.UpdateTransactionParams{
		HideFromReports: &showInReports,
	}

	transaction, err = client.Transactions.Update(ctx, transactionID, params)
	if err != nil {
		log.Fatalf("Failed to unhide transaction: %v", err)
	}

	fmt.Printf("Transaction %s is now visible in reports: %v\n",
		transaction.ID, !transaction.HideFromReports)
}

// consolidateWalmartDeliveriesExample demonstrates a real-world use case:
// Consolidating multiple Walmart delivery transactions into one
func consolidateWalmartDeliveriesExample(client *monarch.Client, ctx context.Context) {
	// Example: Walmart split one order into 3 deliveries
	// $6.56 + $14.39 + $106.95 = $127.90 total

	primaryTransactionID := "225271673011145601" // The $6.56 transaction
	extraTransactionIDs := []string{
		"225271673011145605", // $14.39
		"225271673011145606", // $106.95
	}

	// Step 1: Update the primary transaction to the full amount with itemized splits
	totalAmount := 127.90
	note := "Consolidated Walmart order (3 deliveries)"

	updateParams := &monarch.UpdateTransactionParams{
		Amount: &totalAmount,
		Notes:  &note,
	}

	primaryTxn, err := client.Transactions.Update(ctx, primaryTransactionID, updateParams)
	if err != nil {
		log.Fatalf("Failed to update primary transaction: %v", err)
	}

	// Step 2: Add category splits to itemize the purchase
	splits := []*monarch.TransactionSplit{
		{Amount: 45.00, CategoryID: "cat-groceries", Notes: "Food items"},
		{Amount: 30.00, CategoryID: "cat-household", Notes: "Cleaning supplies"},
		{Amount: 52.90, CategoryID: "cat-personal", Notes: "Personal care"},
	}

	err = client.Transactions.UpdateSplits(ctx, primaryTransactionID, splits)
	if err != nil {
		log.Fatalf("Failed to update splits: %v", err)
	}

	log.Printf("Updated primary transaction %s to $%.2f with %d splits\n",
		primaryTxn.ID, totalAmount, len(splits))

	// Step 3: Try to delete extra transactions, fall back to hiding if deletion fails
	for _, extraID := range extraTransactionIDs {
		err := client.Transactions.Delete(ctx, extraID)
		if err != nil {
			log.Printf("Delete failed for %s: %v", extraID, err)

			// Fall back to hiding from reports
			log.Printf("Hiding transaction %s from reports instead...", extraID)
			if err := hideFromReportsWorkaround(client, ctx, extraID); err != nil {
				log.Fatalf("Failed to hide transaction %s: %v", extraID, err)
			}
			log.Printf("Transaction %s hidden successfully", extraID)
		} else {
			log.Printf("Transaction %s deleted successfully", extraID)
		}
	}

	log.Println("Walmart order consolidation complete!")
}
