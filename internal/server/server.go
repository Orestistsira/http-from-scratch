package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/Orestistsira/http-from-scratch/internal/request"
	"github.com/Orestistsira/http-from-scratch/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h *HandlerError) WriteError(w *response.Writer) {
	headers := response.GetDefaultHeaders(len(h.Message))
	w.WriteStatusLine(h.StatusCode)
	w.WriteHeaders(headers)
	w.WriteBody([]byte(h.Message))
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	running  atomic.Bool
	listener net.Listener
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	// Create TCP listener
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener: ln,
		handler:  handler,
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

	writer := response.NewWriter(conn)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.HTTP_400,
			Message:    err.Error(),
		}
		hErr.WriteError(writer)
		return
	}

	s.handler(writer, r)
}
