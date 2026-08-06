# redis-server

A Redis server clone written in Go from scratch, no external dependencies.
Implements a raw TCP server, a wire protocol parser, and a command
dispatcher.

## Running it

```
make run
```

Starts on `:8091`. Change the port with `-port`:

```
go run ./cmd/server -port 6380
```

Then connect with nc/telnet and type commands, one per line:

```
$ nc localhost 8091
PING
+PONG
SET foo bar
+OK
GET foo
$3
bar
```

## Connect with redis-cli

Connect to the server using the official Redis CLI.

```
➜  redis-clone git:(main) ✗ redis-cli -p 8091
127.0.0.1:8091> SET hi "ryan"
OK
127.0.0.1:8091> GET hi
"ryan"
127.0.0.1:8091>
```

## Commands

`PING`, `ECHO <msg>`, `GET <key>`, `SET <key> <value>`, `EXISTS <key>`,
`DEL <key> [key ...]`, `FLUSH`. Command names are case insensitive.

## Protocol

The server reads whichever protocol the client sends. If a request starts
with `*` it's parsed as real RESP (arrays of bulk strings), the format
redis-cli and other real Redis clients use. Anything else is parsed as
plain space-separated text, one command per line, which is what makes the
nc/telnet usage above possible.

The inline text form has no support for quoted arguments, `SET foo "two
words"` just looks like too many args and errors out. RESP requests don't
have that limitation since each argument is length-prefixed.

Responses are always encoded in real RESP regardless of which form the
request came in as.

## Testing

```
make test
make race
make vet
make fmt
```

Single test:

```
go test ./internal/resp -run TestParseRESP_BulkString -v
go test ./internal/command -run TestHandleSetGet_RoundTrip -v
```

## Layout

- `internal/resp` - parsing. `ReadRequest` picks between `ParseSimple` (the
  inline protocol) and `Parse`/`Encode` (real RESP) based on the first byte
  of the request.
- `internal/command` - `Handlers.Dispatch` routes a request to a handler and
  encodes the result. Also holds the store, a map guarded by a mutex.
- `internal/server` - accepts connections, one goroutine per connection,
  each reading a request, dispatching it, and writing the response back.
- `cmd/mcp-server` - exposes this server as an MCP server, see its
  [README](cmd/mcp-server/README.md).

## TODO 
1. AOF persistence
   1. append every write command to log on disk
   2. write, fsync
   3. replay aof commands on startup
   4. configurable policies 
      1. always: log after every write
      2. everysec: once per sec
2. impl PX, EX options in SET
3. benchmark
   1. test diff policies and scenarios for aof, tradeoffs and optimizations
   2. profiling
   3. pipeline support
4. testing
   1. concurrent stress tests
   2. race tests


## License

GPLv3, see LICENSE.
