package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	ErrInvalidFormat = errors.New("invalid format")
)

const (
	ParserStateInitialized = iota
	ParserStateDone
)

type Request struct {
	ParserState int
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r Request) IsParserDone() bool {
	return r.ParserState == ParserStateDone
}

func (r *Request) ParserDone() {
	r.ParserState = ParserStateDone
}

var crlf = []byte("\r\n")

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{ParserState: ParserStateInitialized}
	buf := make([]byte, bufferSize) // buffer
	readToIndex := 0

	// loop until parser done
	for !req.IsParserDone() {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		// put data start from readToIndex
		nRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.ParserDone()
			}
			return req, err
		}
		readToIndex += nRead

		nParse, err := req.parse(buf[:readToIndex])
		if err != nil {
			return req, err
		}
		copy(buf, buf[nParse:])
		readToIndex -= nParse
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	nParser := 0

	switch r.ParserState {
	case ParserStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")

	case ParserStateInitialized:
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 { // no message process
			return 0, nil
		}
		nParser += n
		r.RequestLine = *rl
		r.ParserDone()

	default:
		return 0, fmt.Errorf("error: unknown state")
	}

	return 0, nil
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	sepIndex := bytes.Index(b, crlf)
	if sepIndex == -1 {
		return nil, 0, nil
	}

	dataConsumed := b[:sepIndex]

	nConsumed := len(dataConsumed) + len(crlf)

	reqlineParts := strings.Split(string(dataConsumed), " ")
	if len(reqlineParts) != 3 {
		return nil, 0, ErrInvalidFormat
	}

	httpParts := strings.Split(reqlineParts[2], "/")
	if len(httpParts) != 2 {
		return nil, 0, ErrInvalidFormat
	}

	return &RequestLine{
		HttpVersion:   httpParts[1],
		RequestTarget: reqlineParts[1],
		Method:        reqlineParts[0],
	}, nConsumed, nil
}
