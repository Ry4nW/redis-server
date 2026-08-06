package mcpserver

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"redis-clone/internal/redisclient"
)

type infoArgs struct{}

type infoResult struct {
	Raw    string            `json:"raw"`
	Fields map[string]string `json:"fields"`
}

func registerInfo(server *mcp.Server, client *redisclient.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "info",
		Description: "Get redis-clone server info: version, role, and key count.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ infoArgs) (*mcp.CallToolResult, infoResult, error) {
		v, err := client.Do("INFO")
		if err != nil {
			return nil, infoResult{}, err
		}

		return nil, infoResult{Raw: v.String, Fields: parseInfoFields(v.String)}, nil
	})
}

// parseInfoFields turns "field:value" lines into a map
func parseInfoFields(raw string) map[string]string {
	fields := map[string]string{}
	for _, line := range strings.Split(raw, "\r\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		fields[k] = v
	}
	return fields
}
