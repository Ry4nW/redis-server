// package mcpserver turns MCP tool calls into RESP commands against redis-clone
package mcpserver

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"redis-clone/internal/redisclient"
)

// New builds an MCP server with all tools registered
func New(redisAddr string) *mcp.Server {
	client := redisclient.New(redisAddr)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "redis-clone",
		Version: "0.1.0",
	}, nil)

	// each tool lives in its own tool_*.go and registers itself
	registerGet(server, client)
	registerSet(server, client)
	registerDelete(server, client)
	registerKeys(server, client)
	registerInfo(server, client)

	return server
}
