// package redisclient is a minimal RESP client for the redis-clone server
package redisclient

import (
	"bufio"
	"fmt"
	"net"
	"sync"

	"redis-clone/internal/resp"
)

// ErrServer wraps a RESP error reply from the server
type ErrServer struct {
	Msg string
}

func (e *ErrServer) Error() string {
	return e.Msg
}

// Client is a single-connection RESP client, reused across calls
type Client struct {
	addr string

	mu     sync.Mutex
	conn   net.Conn
	reader *bufio.Reader
}

// New returns a client for the server at addr, dialed lazily on first use
func New(addr string) *Client {
	return &Client{addr: addr}
}

// Do sends a RESP command and returns the reply, reconnecting on I/O errors
func (c *Client) Do(cmd string, args ...string) (resp.RespValue, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		if err := c.dialLocked(); err != nil {
			return resp.RespValue{}, err
		}
	}

	if err := c.sendLocked(cmd, args); err != nil {
		c.closeLocked()
		return resp.RespValue{}, fmt.Errorf("redisclient: write: %w", err)
	}

	v, err := resp.Parse(c.reader)
	if err != nil {
		c.closeLocked()
		return resp.RespValue{}, fmt.Errorf("redisclient: read reply: %w", err)
	}

	if v.Type == resp.Error {
		return resp.RespValue{}, &ErrServer{Msg: v.String}
	}

	return v, nil
}

// Close closes the underlying connection, if one is open
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	c.reader = nil
	return err
}

func (c *Client) dialLocked() error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return fmt.Errorf("redisclient: dial %s: %w", c.addr, err)
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	return nil
}

func (c *Client) sendLocked(cmd string, args []string) error {
	elems := make([]resp.RespValue, len(args)+1)
	elems[0] = resp.NewBulkString(cmd)
	for i, a := range args {
		elems[i+1] = resp.NewBulkString(a)
	}
	req := resp.RespValue{Type: resp.Array, Array: elems}

	_, err := c.conn.Write([]byte(resp.Encode(req)))
	return err
}

func (c *Client) closeLocked() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.reader = nil
	}
}
