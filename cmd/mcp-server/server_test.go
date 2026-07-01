package main

import (
	"testing"

	"github.com/eshaffer321/monarch-go/v2/pkg/monarch"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestServerInitialization verifies that the server can initialize without panicking
// This catches jsonschema validation errors and other startup issues
func TestServerInitialization(t *testing.T) {
	// Create a mock client (we don't need a real token for this test)
	client := &monarch.Client{}

	// Create MCP server
	impl := &mcp.Implementation{
		Name:    "monarch-money",
		Version: "1.0.0",
	}

	server := mcp.NewServer(impl, nil)

	// This should not panic - if it does, the test fails
	// This catches jsonschema tag errors, tool registration issues, etc.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Server initialization panicked: %v", r)
		}
	}()

	registerTools(server, client)

	t.Log("✓ Server initialized successfully without panicking")
}
