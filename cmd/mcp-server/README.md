# Monarch MCP Server

A [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server that provides Claude Desktop and Claude Code access to your Monarch budget, transaction, and account data.

## Features

- **Budget Data**: Get monthly budget information including unique rollover tracking
- **Transactions**: Query transactions with flexible date/category filters
- **Accounts**: List all accounts with balances and institution details
- **Categories**: Browse all transaction categories
- **Tags**: Access transaction tags for better organization

### Unique Advantages Over Other Monarch MCP Servers

✨ **Rollover Tracking**: Unlike other implementations, this server exposes `rolloverAmount` and `rolloverType` fields - critical for accurate budget planning

✨ **Month Selection**: Query budgets for any month, not just the current month

✨ **Native Go**: Fast, single-binary deployment with no runtime dependencies

✨ **Feature-Complete Client**: Built on the `monarch-go` client with 100% API coverage

## Installation

### Prerequisites

- Go 1.22 or later
- A Monarch account with session token

### Getting Your Monarch Token

1. Log into [Monarch](https://app.monarch.com) in your browser
2. Open Developer Tools (F12 or Cmd+Option+I)
3. Go to Application/Storage → Cookies
4. Copy the value of the `session` cookie

### Building the Server

```bash
cd cmd/mcp-server
go mod download
go build -o monarch-mcp-server
```

This creates a single binary: `monarch-mcp-server`

### Installation for Claude Desktop

1. Build the server (see above)
2. Move the binary to a permanent location:
   ```bash
   sudo mv monarch-mcp-server /usr/local/bin/
   ```
3. Configure Claude Desktop (see below)

## Configuration

### Claude Desktop

Add this to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "monarch-money": {
      "command": "/usr/local/bin/monarch-mcp-server",
      "args": [],
      "env": {
        "MONARCH_TOKEN": "your-session-token-here"
      }
    }
  }
}
```

Replace `your-session-token-here` with your actual Monarch session token.

### Claude Code

Claude Code will automatically detect the MCP server if it's configured in Claude Desktop.

Alternatively, you can run it manually:

```bash
MONARCH_TOKEN="your-token" ./monarch-mcp-server
```

## Available Tools

### 1. `get_budget`

Get budget information for a specific month, including rollover amounts.

**Input:**
```json
{
  "month": "2025-10"
}
```

**Output:**
```json
{
  "month": "2025-10",
  "budgets": [
    {
      "category": "Groceries",
      "group": "Food & Dining",
      "budgeted": 800.00,
      "spent": 645.32,
      "remaining": 154.68,
      "rolloverAmount": 50.00,
      "rolloverType": "ADD_TO_BUDGET",
      "percentage": 80.67
    }
  ]
}
```

### 2. `get_transactions`

Query transactions with optional filters.

**Input:**
```json
{
  "startDate": "2025-10-01",
  "endDate": "2025-10-31",
  "category": "Groceries",
  "limit": 50
}
```

All fields are optional:
- `startDate`: YYYY-MM-DD format
- `endDate`: YYYY-MM-DD format
- `category`: Category name (exact match)
- `limit`: Max results (default: 50)

**Output:**
```json
{
  "transactions": [
    {
      "id": "txn_123",
      "date": "2025-10-15T00:00:00Z",
      "amount": -52.43,
      "merchant": "Whole Foods",
      "category": "Groceries",
      "account": "Chase Checking",
      "notes": "Weekly shopping",
      "pending": false,
      "tags": ["essential"]
    }
  ],
  "count": 1
}
```

### 3. `get_accounts`

Get all accounts with current balances.

**Input:** None

**Output:**
```json
{
  "accounts": [
    {
      "id": "acc_123",
      "name": "Chase Checking",
      "balance": 5234.56,
      "type": "checking",
      "subtype": "checking",
      "institution": "Chase",
      "isHidden": false,
      "includeInNetWorth": true
    }
  ],
  "count": 1
}
```

### 4. `get_categories`

Get all transaction categories.

**Input:** None

**Output:**
```json
{
  "categories": [
    {
      "id": "cat_123",
      "name": "Groceries",
      "group": "Food & Dining",
      "isSystemCategory": false,
      "isDisabled": false
    }
  ],
  "count": 1
}
```

### 5. `get_tags`

Get all available tags.

**Input:** None

**Output:**
```json
{
  "tags": [
    {
      "id": "tag_123",
      "name": "essential",
      "color": "#FF5733",
      "order": 1
    }
  ],
  "count": 1
}
```

## Example Usage with Claude

Once configured, you can ask Claude questions like:

- "What's my budget for October 2025? Show me categories that are over budget."
- "Find all grocery transactions from last month over $100"
- "What's my total spending on dining out this month?"
- "Show me all my account balances"
- "Which budget categories have rollover amounts?"

## Development

This server is designed to be easily extracted into its own repository. It uses the `monarch-go` client as a dependency via Go modules.

### Project Structure

```
cmd/mcp-server/
├── main.go          # Server initialization and registration
├── tools.go         # Tool implementations
├── go.mod           # Module dependencies
└── README.md        # This file
```

### Running in Development

```bash
cd cmd/mcp-server
MONARCH_TOKEN="your-token" go run .
```

### Adding New Tools

1. Define input/output structs in `tools.go` with `jsonschema` tags
2. Implement the tool handler method on `monarchTools`
3. Register the tool in `registerTools()` in `main.go`

See the [MCP Go SDK documentation](https://github.com/modelcontextprotocol/go-sdk) for more details.

## Troubleshooting

### "MONARCH_TOKEN environment variable is required"

Make sure you've set the `MONARCH_TOKEN` in your Claude Desktop configuration or environment.

### Session token expired

Monarch session tokens expire periodically. If you get authentication errors:
1. Log into Monarch in your browser
2. Get a fresh session token from cookies
3. Update your Claude Desktop configuration

### Server not appearing in Claude Desktop

1. Restart Claude Desktop completely
2. Check the configuration file path is correct
3. Verify the binary path in the config is absolute and correct
4. Check Claude Desktop logs for errors

## License

MIT

## Contributing

Issues and pull requests welcome at: https://github.com/eshaffer321/monarch-go

## Related Projects

- [monarch-go](https://github.com/eshaffer321/monarch-go) - The underlying Go client library
- [monarchmoney (Python)](https://github.com/hammem/monarchmoney) - Original Python implementation
- [MCP Specification](https://modelcontextprotocol.io) - Model Context Protocol documentation
