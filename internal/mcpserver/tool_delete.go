package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"redis-clone/internal/redisclient"
)

type deleteArgs struct {
	Key string `json:"key" jsonschema:"the key to delete"`
}

type deleteResult struct {
	Deleted bool `json:"deleted"`
}

func registerDelete(server *mcp.Server, client *redisclient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete",
		Description: "Delete a key. Returns deleted=false if the key didn't exist.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, args deleteArgs) (*mcp.CallToolResult, deleteResult, error) {
		v, err := client.Do("DEL", args.Key)
		if err != nil {
			return nil, deleteResult{}, err
		}
		return nil, deleteResult{Deleted: v.Integer > 0}, nil
	})
}
