package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/request"
	"github.com/davidimannuel/boot.dev-learn-http-protocol-in-go/internal/response"
)

type Server struct {
	l       net.Listener
	h       Handler
	isClose atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	srv := &Server{
		l: listener,
		h: handler,
	}
	srv.isClose.Store(false)
	go srv.listen()
	return srv, nil
}

func (s *Server) Close() error {
	s.isClose.Store(true)
	return s.l.Close()
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println("failed parse request", err)
	}

	rw := response.NewWritter(conn)
	s.h(rw, req)
}

func (s *Server) listen() {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			if s.isClose.Load() {
				fmt.Println("Server is shutdown")
				return
			}
			fmt.Println("error when accept connection", err)
			continue
		}
		go s.handle(conn)
	}
}
