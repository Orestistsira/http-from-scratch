package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/Orestistsira/http-from-scratch/internal/response"
)

type Server struct {
	running  atomic.Bool
	listener net.Listener
}

func Serve(port int) (*Server, error) {
	// Create TCP listener
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: ln,
	}
	s.running.Store(true)

	// Start listening in a goroutine
	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	if s.running.Load() {
		s.running.Store(false)
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Exit loop if server is closed
			if !s.running.Load() {
				return
			}
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)

	err := response.WriteStatusLine(conn, response.HTTP_200)
	if err != nil {
		return
	}

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		return
	}
}
