package resp

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

func Parse(reader *bufio.Reader) (RespValue, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return RespValue{}, err
	}

	switch line[0] {
	case '$':
		return parseBulkStr(line, reader)
	case '*':
		return parseArray(line, reader)
	case '-':
		return parseError(line, reader)
	case ':':
		return parseInt(line, reader)
	case '+':
		return parseSimpleStr(line, reader)
	default:
		return RespValue{}, ErrMalformedRESP
	}
}

func parseBulkStr(line string, reader *bufio.Reader) (RespValue, error) {
	// $<length>\r\n<data>\r\n
	strLen, err := getLenHelper(line)
	if err != nil {
		return RespValue{}, ErrMalformedRESP
	}

	if strLen < 0 {
		return NewNullBulkString(), nil
	}

	data := make([]byte, strLen)
	if _, err := io.ReadFull(reader, data); err != nil {
		return RespValue{}, ErrUnexpectedEOF
	}
	crlf := make([]byte, 2)
	if _, err := io.ReadFull(reader, crlf); err != nil {
		return RespValue{}, ErrUnexpectedEOF
	}

	return NewBulkString(string(data)), nil
}

func removeCharPrefixAndCRLF(str string) string {
	return strings.TrimSuffix(str, "\r\n")[1:]
}

func parseArray(line string, reader *bufio.Reader) (RespValue, error) {
	// *<number-of-elements>\r\n<element-1>...<element-n>
	lengthStr := removeCharPrefixAndCRLF(line)
	arrLen, err := strconv.Atoi(lengthStr)
	if err != nil {
		return RespValue{}, ErrMalformedRESP
	}

	if arrLen < 0 {
		return NewNullArr(), nil
	}

	result := make([]RespValue, arrLen)
	for i := 0; i < arrLen; i++ {
		// mutually recursive call on array element; each element parser
		// already consumes its own trailing CRLF
		curData, err := Parse(reader)
		if err != nil {
			return RespValue{}, err
		}

		result[i] = curData
	}

	return RespValue{
		Type:  Array,
		Array: result,
	}, nil
}

func parseError(line string, reader *bufio.Reader) (RespValue, error) {
	errorStr := removeCharPrefixAndCRLF(line)
	return NewError(errorStr), nil
}

func parseInt(line string, reader *bufio.Reader) (RespValue, error) {
	intStr := removeCharPrefixAndCRLF(line)
	i, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return RespValue{}, ErrTypeIsNotInt
	}
	return NewInteger(i), nil
}

func parseSimpleStr(line string, reader *bufio.Reader) (RespValue, error) {
	simpleStr := removeCharPrefixAndCRLF(line)
	return NewSimpleString(simpleStr), nil
}

func getLenHelper(line string) (int, error) {
	lengthStr := strings.TrimSuffix(line, "\r\n")[1:]
	strLen, err := strconv.Atoi(lengthStr)
	if err != nil {
		return 0, err
	}
	return strLen, nil
}

func doesMutate(cmd string) (bool, error) {
	switch strings.ToUpper(cmd) {
	case "PING":
		return false, nil
	case "ECHO":
		return false, nil
	case "GET":
		return false, nil
	case "SET":
		return true, nil
	case "EXISTS":
		return false, nil
	case "DEL":
		return true, nil
	case "FLUSH":
		return true, nil
	case "KEYS":
		return false, nil
	case "INFO":
		return false, nil
	default:
		return false, ErrUnknownCommand
	}
}

func ParseSimple(reader *bufio.Reader) (*Request, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	tokens := strings.Fields(line)
	if len(tokens) == 0 {
		return nil, ErrEmptyRequest
	}

	// normalize cmd
	command := strings.ToUpper(tokens[0])
	args := tokens[1:]

	mutate, err := doesMutate(command)
	if err != nil {
		return nil, err
	}
	return &Request{
		Command: command,
		Args:    args,
		Mutates: mutate,
	}, nil
}

// * is RESP, anythign else is inline text protocol
func ReadRequest(reader *bufio.Reader) (*Request, error) {
	b, err := reader.Peek(1)
	if err != nil {
		return nil, err
	}

	if b[0] != '*' {
		return ParseSimple(reader)
	}

	v, err := Parse(reader)
	if err != nil {
		return nil, err
	}

	return requestFromArray(v)
}

// requestFromArray converts a parsed RESP array into a Request
// requirement: every element of the array must be a bulk string.
func requestFromArray(v RespValue) (*Request, error) {
	if v.Null || len(v.Array) == 0 {
		return nil, ErrEmptyRequest
	}

	for _, elem := range v.Array {
		if elem.Type != BulkString || elem.Null {
			return nil, ErrMalformedRESP
		}
	}

	args := make([]string, len(v.Array)-1)
	for i, elem := range v.Array[1:] {
		args[i] = elem.String
	}

	cmd := strings.ToUpper(v.Array[0].String)
	mutate, err := doesMutate(cmd)
	if err != nil {
		return nil, err
	}
	return &Request{
		Command: cmd,
		Args:    args,
		Mutates: mutate,
	}, nil
}
