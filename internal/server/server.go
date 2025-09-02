package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/Orestistsira/http-from-scratch/internal/request"
	"github.com/Orestistsira/http-from-scratch/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h *HandlerError) WriteError(w io.Writer) {
	headers := response.GetDefaultHeaders(len(h.Message))
	response.WriteStatusLine(w, h.StatusCode)
	response.WriteHeaders(w, headers)
	response.WriteBody(w, []byte(h.Message))
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func WriteHandlerError(w io.Writer, handlerErr *HandlerError) error {
	_, err := fmt.Fprintf(w, "%d %s", handlerErr.StatusCode, handlerErr.Message)
	return err
}

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

	r, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.HTTP_400,
			Message:    err.Error(),
		}
		hErr.WriteError(conn)
		return
	}

	writer := bytes.NewBuffer([]byte{})

	hErr := s.handler(writer, r)
	if hErr != nil {
		hErr.WriteError(conn)
		return
	}

	body := writer.Bytes()
	headers := response.GetDefaultHeaders(len(body))
	response.WriteStatusLine(conn, response.HTTP_200)
	response.WriteHeaders(conn, headers)
	response.WriteBody(conn, body)
}
