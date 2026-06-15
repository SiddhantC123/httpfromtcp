package server

import (
	"fmt"
	"httpserver/internal/request"
	"httpserver/internal/response"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	handler  Handler
	isClosed atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: listener,
		handler:  handler,
	}

	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	// Mark the server as closed safely using atomic
	s.isClosed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			fmt.Println("Error accepting connection:", err)
			continue
		}

		fmt.Println("Connection established")

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		WriteError(conn, &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    "Bad Request\n",
		})
		return
	}

	rw := response.NewWriter(conn)

	handlerErr := s.handler(rw, req)
	if handlerErr != nil {
		WriteError(conn, handlerErr)
		return
	}
}
