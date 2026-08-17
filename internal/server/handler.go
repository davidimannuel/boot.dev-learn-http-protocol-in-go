package server

import (
	"io"

	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/request"
	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Msg        string
}

func (he *HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	response.WriteHeaders(w, response.GetDefaultHeaders(len(he.Msg)))
	w.Write([]byte(he.Msg))
}

type Handler func(w *response.Writer, req *request.Request)
