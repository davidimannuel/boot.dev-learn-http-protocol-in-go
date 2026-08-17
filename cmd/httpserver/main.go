package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/headers"
	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/request"
	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/response"
	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if req == nil {
		w.InternalError(fmt.Errorf("re struct is nil"))
		return
	}

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		// handlerHttpbin(w, req)
		handlerHttpbinEncoding(w, req)
		return
	}

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		handler400(w, req)
		return
	case "/myproblem":
		handler500(w, req)
		return
	case "/video":
		handlerVideo(w, req)
		return
	default:
		handler200(w, req)
		return
	}
}

func handlerVideo(w *response.Writer, r *request.Request) {
	f, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		w.InternalError(err)
		return
	}
	h := response.GetDefaultHeaders(len(f))
	h.Override("Content-Type", "video/mp4")
	w.WriteStatusLine(response.StatusCodeSuccess)
	w.WriteHeaders(h)
	w.WriteBody(f)
}

func handlerHttpbin(w *response.Writer, r *request.Request) {
	httpbinPath := strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/")
	httpbinPath = "https://httpbingo.org/" + httpbinPath
	fmt.Println("Proxying to", httpbinPath)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Remove("Content-Length")
	w.WriteStatusLine(response.StatusCodeSuccess)
	w.WriteHeaders(h)
	httpr, err := http.Get(httpbinPath)
	if err != nil {
		w.InternalError(err)
		return
	}
	buff := make([]byte, 1024)
	for {
		n, err := httpr.Body.Read(buff)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			w.InternalError(err)
			return
		}
		// fmt.Println("chunked:", n, string(buff))
		w.WriteChunkedBody(buff[:n])
	}
	w.WriteChunkedBodyDone()
}

func handlerHttpbinEncoding(w *response.Writer, r *request.Request) {
	httpbinPath := strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/")
	httpbinPath = "https://httpbingo.org/" + httpbinPath
	fmt.Println("Proxying to", httpbinPath)
	h := response.GetDefaultHeaders(0)
	h.Remove("Content-Length")
	h.Override("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	w.WriteStatusLine(response.StatusCodeSuccess)
	w.WriteHeaders(h)
	httpr, err := http.Get(httpbinPath)
	if err != nil {
		w.InternalError(err)
		return
	}
	buff := make([]byte, 1024)
	fulldata := []byte{}
	done := false
	for {
		n, err := httpr.Body.Read(buff)
		// EOF, and n > 0 could be together
		if err != nil {
			if errors.Is(err, io.EOF) {
				done = true
			}
			if n == 0 {
				w.InternalError(err)
				return
			}
		}
		fmt.Println("chunked:", n, string(buff[:n]))
		w.WriteChunkedBody(buff[:n])
		fulldata = append(fulldata, buff[:n]...)
		if done {
			break
		}
	}
	w.WriteChunkedBodyDone()
	lenData := len(fulldata)
	s256 := sha256.Sum256(fulldata)
	trailer := headers.NewHeaders()
	trailer.Set("X-Content-SHA256", fmt.Sprintf("%x", s256))
	trailer.Set("X-Content-Length", fmt.Sprintf("%d", lenData))
	w.WriteTrailers(trailer)
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeInternalServerErr)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}
