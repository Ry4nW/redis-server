package main

import (
	"context"
	"flag"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"redis-clone/internal/mcpserver"
)

func main() {
	redisAddr := flag.String("redis-addr", "localhost:8091", "address of the redis-clone server to connect to")
	flag.Parse()

	server := mcpserver.New(*redisAddr)

	// Claude Code launches MCP servers as a subprocess over stdio
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("mcp server failed: %v", err)
	}
}
