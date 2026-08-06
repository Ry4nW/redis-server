package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"redis-clone/internal/redisclient"
)

type setArgs struct {
	Key   string `json:"key" jsonschema:"the key to set"`
	Value string `json:"value" jsonschema:"the value to store"`
}

type setResult struct {
	OK bool `json:"ok"`
}

func registerSet(server *mcp.Server, client *redisclient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set",
		Description: "Set a key to a value, overwriting any existing value at that key.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args setArgs) (*mcp.CallToolResult, setResult, error) {
		if _, err := client.Do("SET", args.Key, args.Value); err != nil {
			return nil, setResult{}, err
		}
		return nil, setResult{OK: true}, nil
	})
}
