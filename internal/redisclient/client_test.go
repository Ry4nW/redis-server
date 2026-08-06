package redisclient

import (
	"bufio"
	"errors"
	"net"
	"testing"

	"redis-clone/internal/command"
	"redis-clone/internal/resp"
)

// startTestServer runs a real listener, no mocking
func startTestServer(t *testing.T) string {
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

func TestClient_SetGetRoundTrip(t *testing.T) {
	c := New(startTestServer(t))
	defer c.Close()

	if _, err := c.Do("SET", "foo", "bar"); err != nil {
		t.Fatalf("SET: %v", err)
	}

	v, err := c.Do("GET", "foo")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	if v.String != "bar" {
		t.Fatalf("got %q, want %q", v.String, "bar")
	}
}

func TestClient_GetMissingKeyReturnsNullNotError(t *testing.T) {
	c := New(startTestServer(t))
	defer c.Close()

	v, err := c.Do("GET", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !v.Null {
		t.Fatalf("expected null bulk string, got %+v", v)
	}
}

func TestClient_ServerErrorBecomesErrServer(t *testing.T) {
	c := New(startTestServer(t))
	defer c.Close()

	// SET requires exactly 2 args
	_, err := c.Do("SET", "onlyonearg")
	if err == nil {
		t.Fatal("expected error for bad arg count")
	}

	var serverErr *ErrServer
	if !errors.As(err, &serverErr) {
		t.Fatalf("expected *ErrServer, got %T: %v", err, err)
	}
}

func TestClient_KeysGlobPattern(t *testing.T) {
	c := New(startTestServer(t))
	defer c.Close()

	mustDo(t, c, "SET", "user:1", "a")
	mustDo(t, c, "SET", "user:2", "b")
	mustDo(t, c, "SET", "other", "c")

	v, err := c.Do("KEYS", "user:*")
	if err != nil {
		t.Fatalf("KEYS: %v", err)
	}
	if len(v.Array) != 2 {
		t.Fatalf("expected 2 matches, got %d: %+v", len(v.Array), v.Array)
	}
}

func TestClient_DialErrorOnUnreachableAddr(t *testing.T) {
	c := New("127.0.0.1:1")
	defer c.Close()

	if _, err := c.Do("PING"); err == nil {
		t.Fatal("expected dial error, got nil")
	}
}

func mustDo(t *testing.T, c *Client, cmd string, args ...string) resp.RespValue {
	t.Helper()
	v, err := c.Do(cmd, args...)
	if err != nil {
		t.Fatalf("%s: %v", cmd, err)
	}
	return v
}
