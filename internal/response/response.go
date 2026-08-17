package response

import (
	"fmt"
	"io"

	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/headers"
)

type StatusCode int

var StatusCodeSuccess StatusCode = 200
var StatusCodeBadRequest StatusCode = 400
var StatusCodeInternalServerErr StatusCode = 500

func WriteStatusLine(w io.Writer, statusCode StatusCode) {
	switch statusCode {
	case StatusCodeSuccess:
		w.Write([]byte("HTTP/1.1 200 OK"))
	case StatusCodeBadRequest:
		w.Write([]byte("HTTP/1.1 400 Bad Request"))
	case StatusCodeInternalServerErr:
		w.Write([]byte("HTTP/1.1 500 Internal Server Error"))
	default:
		w.Write([]byte("HTTP/1.1 200 "))
	}
	w.Write([]byte("\r\n"))
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for h, k := range headers {
		fieldLine := fmt.Sprintf("%s: %s\r\n", h, k)
		_, err := w.Write([]byte(fieldLine))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}
