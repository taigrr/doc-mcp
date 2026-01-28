package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	docmcp "github.com/taigrr/doc-mcp/internal/mcp"
)

func main() {
	server := docmcp.NewServer()

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Failed to serve MCP server: %v", err)
	}
}
