package main

import (
	"context"
	"log"
	"os"

	"github.com/eshaffer321/monarch-go/v2/pkg/monarch"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Get Monarch token from environment
	token := os.Getenv("MONARCH_TOKEN")
	if token == "" {
		log.Fatal("MONARCH_TOKEN environment variable is required")
	}

	// Initialize Monarch client
	client, err := monarch.NewClient(&monarch.ClientOptions{
		Token: token,
	})
	if err != nil {
		log.Fatalf("failed to initialize Monarch client: %v", err)
	}

	// Create MCP server with v1.0.0 API
	impl := &mcp.Implementation{
		Name:    "monarch-money",
		Version: "1.0.0",
	}

	server := mcp.NewServer(impl, nil)

	// Register all tools
	registerTools(server, client)

	// Run server over stdio transport (for Claude Desktop)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func registerTools(server *mcp.Server, client *monarch.Client) {
	// Create tools instance with client
	tools := &monarchTools{client: client}

	// Register all tools using v1.0.0 API
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_budget",
		Description: "Get budget information for a specific month, including rollover amounts. Returns budget entries with category names, budgeted amounts, actual spending, remaining amounts, and rollover data.",
	}, tools.GetBudget)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_transactions",
		Description: "Query transactions with optional filters for date range, category, and limit. Returns transaction details including date, amount, merchant, category, and notes.",
	}, tools.GetTransactions)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_accounts",
		Description: "Get all accounts with their current balances, types, and institution information.",
	}, tools.GetAccounts)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_categories",
		Description: "Get all available transaction categories organized by groups.",
	}, tools.GetCategories)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tags",
		Description: "Get all available transaction tags.",
	}, tools.GetTags)
}
