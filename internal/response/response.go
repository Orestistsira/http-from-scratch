package response

import (
	"fmt"
	"io"

	"github.com/Orestistsira/http-from-scratch/internal/headers"
)

var ErrStatusLine = fmt.Errorf("error writing status line")
var ErrUnknownCode = fmt.Errorf("error unknown code")
var ErrWriterStatusInit = fmt.Errorf("error cannot write status-line. Follow order: status-line -> headers -> body")
var ErrWriterStatusHeaders = fmt.Errorf("error cannot write headers. Follow order: status-line -> headers -> body")
var ErrWriterStatusBody = fmt.Errorf("error cannot write body. Follow order: status-line -> headers -> body")

type StatusCode int

const (
	HTTP_200 StatusCode = 200
	HTTP_400 StatusCode = 400
	HTTP_500 StatusCode = 500
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return *h
}

func GetHTMLHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/html")
	return *h
}

type WriterStatus string

const (
	WriterStatusInit    WriterStatus = "init"
	WriterStatusHeaders WriterStatus = "headers"
	WriterStatusBody    WriterStatus = "body"
)

type Writer struct {
	writer       io.Writer
	writerStatus WriterStatus
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer:       w,
		writerStatus: WriterStatusInit,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerStatus != WriterStatusInit {
		return ErrWriterStatusInit
	}

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

	_, err := w.writer.Write(statusLine)
	if err == nil {
		w.writerStatus = WriterStatusHeaders
	}
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerStatus != WriterStatusHeaders {
		return ErrWriterStatusHeaders
	}

	b := []byte{}
	headers.ForEach(func(name, value string) {
		b = fmt.Appendf(b, "%s: %s\r\n", name, value)
	})

	// Write an extra CRLF to indicate end of headers
	b = fmt.Appendf(b, "\r\n")

	_, err := w.writer.Write(b)
	if err == nil {
		w.writerStatus = WriterStatusBody
	}
	return err
}

func (w *Writer) WriteBody(b []byte) error {
	if w.writerStatus != WriterStatusBody {
		return ErrWriterStatusBody
	}

	_, err := w.writer.Write(b)
	return err
}
