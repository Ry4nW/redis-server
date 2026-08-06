package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"redis-clone/internal/redisclient"
)

type getArgs struct {
	Key string `json:"key" jsonschema:"the key to look up"`
}

type getResult struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
	Found bool   `json:"found"`
}

func registerGet(server *mcp.Server, client *redisclient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get",
		Description: "Get the value stored at a key. Returns found=false if the key doesn't exist.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args getArgs) (*mcp.CallToolResult, getResult, error) {
		v, err := client.Do("GET", args.Key)
		if err != nil {
			return nil, getResult{}, err
		}
		if v.Null {
			return nil, getResult{Key: args.Key, Found: false}, nil
		}
		return nil, getResult{Key: args.Key, Value: v.String, Found: true}, nil
	})
}
