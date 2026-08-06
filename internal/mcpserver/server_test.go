package mcpserver

import (
	"bufio"
	"context"
	"net"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"redis-clone/internal/command"
	"redis-clone/internal/resp"
)

// startRedisClone runs a real redis-clone listener, no mocking
func startRedisClone(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	h := command.NewHandlers()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				reader := bufio.NewReader(conn)
				for {
					req, err := resp.ReadRequest(reader)
					if err != nil {
						return
					}
					out, _ := h.Dispatch(req)
					if _, err := conn.Write([]byte(out)); err != nil {
						return
					}
				}
			}()
		}
	}()

	return ln.Addr().String()
}

// connectClient wires an in-memory MCP client/server pair
func connectClient(t *testing.T, redisAddr string) *mcp.ClientSession {
	t.Helper()

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	server := New(redisAddr)
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	return session
}

func TestTools_SetThenGet(t *testing.T) {
	session := connectClient(t, startRedisClone(t))
	ctx := context.Background()

	_, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "set",
		Arguments: map[string]any{"key": "foo", "value": "bar"},
	})
	if err != nil {
		t.Fatalf("set: %v", err)
	}

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get",
		Arguments: map[string]any{"key": "foo"},
	})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	got, ok := res.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("expected structured content, got %#v", res.StructuredContent)
	}
	if got["value"] != "bar" || got["found"] != true {
		t.Fatalf("got %v", got)
	}
}

func TestTools_GetMissingKeyNotFound(t *testing.T) {
	session := connectClient(t, startRedisClone(t))
	ctx := context.Background()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "get",
		Arguments: map[string]any{"key": "missing"},
	})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	got := res.StructuredContent.(map[string]any)
	if got["found"] != false {
		t.Fatalf("expected found=false, got %v", got)
	}
}

func TestTools_DeleteAndKeys(t *testing.T) {
	session := connectClient(t, startRedisClone(t))
	ctx := context.Background()

	for _, key := range []string{"a", "b", "c"} {
		if _, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "set",
			Arguments: map[string]any{"key": key, "value": "1"},
		}); err != nil {
			t.Fatalf("set %s: %v", key, err)
		}
	}

	delRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "delete",
		Arguments: map[string]any{"key": "a"},
	})
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if delRes.StructuredContent.(map[string]any)["deleted"] != true {
		t.Fatalf("expected deleted=true, got %v", delRes.StructuredContent)
	}

	keysRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "keys",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("keys: %v", err)
	}
	got := keysRes.StructuredContent.(map[string]any)
	if got["count"] != float64(2) {
		t.Fatalf("expected 2 remaining keys, got %v", got)
	}
}

func TestTools_Info(t *testing.T) {
	session := connectClient(t, startRedisClone(t))
	ctx := context.Background()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "info",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	got := res.StructuredContent.(map[string]any)
	fields, ok := got["fields"].(map[string]any)
	if !ok || fields["role"] != "master" {
		t.Fatalf("expected role=master in fields, got %v", got)
	}
}
