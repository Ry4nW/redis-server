package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"redis-clone/internal/redisclient"
)

type keysArgs struct {
	Pattern string `json:"pattern,omitempty" jsonschema:"glob pattern to match keys against, e.g. user:*, defaults to * (all keys)"`
}

type keysResult struct {
	Keys  []string `json:"keys"`
	Count int      `json:"count"`
}

func registerKeys(server *mcp.Server, client *redisclient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "keys",
		Description: `List keys matching a glob pattern. Defaults to "*" (all keys).`,
	}, func(_ context.Context, _ *mcp.CallToolRequest, args keysArgs) (*mcp.CallToolResult, keysResult, error) {
		pattern := args.Pattern
		if pattern == "" {
			pattern = "*"
		}

		v, err := client.Do("KEYS", pattern)
		if err != nil {
			return nil, keysResult{}, err
		}

		keys := make([]string, len(v.Array))
		for i, elem := range v.Array {
			keys[i] = elem.String
		}

		return nil, keysResult{Keys: keys, Count: len(keys)}, nil
	})
}
