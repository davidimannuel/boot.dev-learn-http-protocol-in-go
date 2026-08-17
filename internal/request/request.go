package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/headers"
)

var (
	ErrInvalidFormat  = errors.New("invalid format")
	ErrIncompleteBody = errors.New("incomplete request body: content-length exceeds actual content")
)

const (
	ParserStateInitialized = iota
	ParserStateParsingHeaders
	ParserStateParsingBody
	ParserStateDone
)

type Request struct {
	ParserState int
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
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

func (r *Request) ParsingHeaders() {
	r.ParserState = ParserStateParsingHeaders
}

func (r *Request) ParsingBody() {
	r.ParserState = ParserStateParsingBody
}

func (r *Request) ContentLength() int {
	if r.Headers == nil {
		return 0
	}

	val, _ := r.Headers.Get("Content-Length")
	valInt, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}

	return valInt
}

func (r *Request) ContentExists() bool {
	return r.ContentLength() > 0
}

var crlf = []byte("\r\n")

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		ParserState: ParserStateInitialized,
		Headers:     headers.NewHeaders(),
	}
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
				// the loop only keeps reading while the parser isn't done,
				// so hitting EOF here means the request was cut off mid-parse
				if req.ParserState == ParserStateParsingBody {
					return req, ErrIncompleteBody
				}
				return req, io.ErrUnexpectedEOF
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
		r.ParsingHeaders()

	case ParserStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if n == 0 { // no message process
			return 0, nil
		}
		nParser += n
		if done {
			// if content exists continue to parsing the body
			if r.ContentExists() {
				r.ParsingBody()
			} else { // if no content length move to done
				r.ParserDone()
			}
		}

	case ParserStateParsingBody:
		contentLength := r.ContentLength()
		if len(data) < contentLength { // body isn't complete yet
			return 0, nil
		}
		r.Body = append(r.Body, data[:contentLength]...) // copy body
		nParser = contentLength
		r.ParserDone()
	default:
		return 0, fmt.Errorf("error: unknown state")
	}

	return nParser, nil
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
