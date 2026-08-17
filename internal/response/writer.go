package response

import (
	"fmt"
	"io"

	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/headers"
)

const (
	WriterStatusInitiated = iota
	WriterStatusLineDone
	WriterStatusHeadersDone
)

type Writer struct {
	w     io.Writer
	state int
}

func NewWritter(w io.Writer) *Writer {
	return &Writer{
		w:     w,
		state: WriterStatusInitiated,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != WriterStatusInitiated {
		return fmt.Errorf("response writer have wrong order")
	}
	w.state = WriterStatusLineDone
	WriteStatusLine(w.w, statusCode)
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != WriterStatusLineDone {
		return fmt.Errorf("response writer have wrong order")
	}
	err := WriteHeaders(w.w, headers)
	if err != nil {
		return err
	}
	w.state = WriterStatusHeadersDone
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != WriterStatusHeadersDone {
		return 0, fmt.Errorf("response writer have wrong order")
	}
	return w.w.Write(p)
}

// default response for internal error
func (w *Writer) InternalError(err error) {
	msg := err.Error()
	errW := w.WriteStatusLine(StatusCodeInternalServerErr)
	if errW != nil {
		fmt.Println("err write", errW)
	}
	errW = w.WriteHeaders(GetDefaultHeaders(len(msg)))
	if errW != nil {
		fmt.Println("err write", errW)
	}
	_, errW = w.WriteBody([]byte(msg))
	if errW != nil {
		fmt.Println("err write", errW)
	}
}

// default response for success
func (w *Writer) OK() {
	msg := "OK"
	errW := w.WriteStatusLine(StatusCodeSuccess)
	if errW != nil {
		fmt.Println("err write", errW)
	}
	errW = w.WriteHeaders(GetDefaultHeaders(len(msg)))
	if errW != nil {
		fmt.Println("err write", errW)
	}
	_, errW = w.WriteBody([]byte(msg))
	if errW != nil {
		fmt.Println("err write", errW)
	}
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	len := len(p)
	hex := fmt.Sprintf("%X", len)
	w.WriteBody([]byte(hex))
	w.WriteBody([]byte("\r\n"))

	w.WriteBody(p)
	w.WriteBody([]byte("\r\n"))
	return 0, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	// w.WriteBody([]byte("0\r\n\r\n")) // used for when no trailers
	w.WriteBody([]byte("0\r\n"))
	return 0, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for k, v := range h {
		w.WriteBody([]byte(k + ": " + v))
		w.WriteBody([]byte("\r\n"))
	}
	w.WriteBody([]byte("\r\n"))
	return nil
}
