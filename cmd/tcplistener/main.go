package main

import (
	"fmt"
	"net"

	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	panicIfErr(err)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		panicIfErr(err)
		fmt.Println("connection accepted")
		req, err := request.RequestFromReader(conn)
		panicIfErr(err)
		fmt.Printf(`Request line:
- Method: %s
- Target: %s
- Version: %s
`,
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for h, k := range req.Headers {
			fmt.Printf("- %s: %s\n", h, k)
		}
	}
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
