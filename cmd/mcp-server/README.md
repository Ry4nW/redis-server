# mcp-server

Exposes the redis-clone server as an MCP server, so Claude Code (or any
MCP client) can read and write keys as tools instead of shell commands.

It's a thin adapter: it holds no data itself, it just translates tool
calls into RESP commands and sends them to a running redis-clone
instance over TCP. See `internal/mcpserver` for the tool implementations
and `internal/redisclient` for the RESP client.

## 1. Start the redis-clone server

```
make run
```

Starts on `:8091` by default. See the top-level README for details.

## 2. Start the MCP server

```
go run ./cmd/mcp-server
```

By default it connects to `localhost:8091`. Point it at a different
address with `-redis-addr`:

```
go run ./cmd/mcp-server -redis-addr localhost:6380
```

The MCP server speaks JSON-RPC over stdio, so running it directly in a
terminal won't print anything, it's waiting for a client to connect on
stdin/stdout. That's expected.

## 3. Connect Claude Code to it

```
claude mcp add redis-clone -- go run /path/to/redis-clone/cmd/mcp-server -redis-addr localhost:8091
```

Or build a binary first and point at that instead of `go run`:

```
go build -o mcp-server ./cmd/mcp-server
claude mcp add redis-clone -- /path/to/mcp-server -redis-addr localhost:8091
```

Once added, Claude Code launches the server itself as a subprocess when
needed, there's nothing to keep running manually.

## 4. Tools

- `get(key)` - returns `{key, value, found}`. `found` is `false` for a
  missing key rather than an error.
- `set(key, value)` - returns `{ok}`.
- `delete(key)` - returns `{deleted}`.
- `keys(pattern)` - glob match over key names, defaults to `*`. Returns
  `{keys, count}`.
- `info()` - server version, role, key count. Returns both `{raw}` (the
  text as redis-clone sends it) and `{fields}` (parsed into a map).

## Testing

```
go test ./internal/redisclient/...
go test ./internal/mcpserver/...
```

Both spin up a real redis-clone instance in-process (a real TCP
listener backed by `command.Handlers`) rather than mocking anything, so
they're exercising the actual wire protocol.

To test manually, start both servers as above, then either use
`claude mcp add` and try a tool from a Claude Code session, or write a
small Go client using `mcp.CommandTransport` from the SDK to drive
`cmd/mcp-server` as a subprocess directly.
