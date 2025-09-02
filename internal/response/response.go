package response

import (
	"fmt"
	"io"

	"github.com/Orestistsira/http-from-scratch/internal/headers"
)

var ErrStatusLine = fmt.Errorf("error writing status line")
var ErrUnknownCode = fmt.Errorf("error unknown code")

type StatusCode int

const (
	HTTP_200 StatusCode = 200
	HTTP_400 StatusCode = 400
	HTTP_500 StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var statusLine []byte

	switch statusCode {
	case HTTP_200:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case HTTP_400:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case HTTP_500:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		statusLine = []byte("HTTP/1.1 " + fmt.Sprintf("%d", statusCode) + "\r\n")
	}

	_, err := w.Write(statusLine)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return *h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	b := []byte{}
	headers.ForEach(func(name, value string) {
		b = fmt.Appendf(b, "%s: %s\r\n", name, value)
	})

	// Write an extra CRLF to indicate end of headers
	b = fmt.Appendf(b, "\r\n")

	_, err := w.Write(b)
	return err
}

func WriteBody(w io.Writer, b []byte) error {
	_, err := w.Write(b)
	return err
}
