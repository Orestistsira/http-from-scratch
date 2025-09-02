package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/Orestistsira/http-from-scratch/internal/headers"
)

var ErrMalformedRequestLine = fmt.Errorf("malformed request-line")
var ErrUnknownState = fmt.Errorf("error unknown state")
var ErrBodyLength = fmt.Errorf("content-length does not match body length")

var separator = []byte("\r\n")

const bufferSize = 8

type ParserStatus string

const (
	ParserStatusInit    = "init"
	ParserStatusDone    = "done"
	ParserStatusHeaders = "headers"
	ParserStatusBody    = "body"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        []byte
	Status      ParserStatus
}

func newRequest() *Request {
	return &Request{
		Status:  ParserStatusInit,
		Headers: headers.NewHeaders(),
		Body:    []byte{},
	}
}

func (r *Request) getBodyContentLength() int {
	contentLenStr := r.Headers.Get("content-length")
	// No content-length in the headers -> No body
	if contentLenStr == "" {
		return 0
	}

	contentLen, err := strconv.Atoi(contentLenStr)
	if err != nil {
		return 0
	}
	return contentLen
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

	switch r.Status {
	case ParserStatusInit:
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			break
		}

		r.RequestLine = *rl
		read += n

		r.Status = ParserStatusHeaders
	case ParserStatusHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			break
		}

		read += n

		if done {
			r.Status = ParserStatusBody
		}
	case ParserStatusBody:
		contentLen := r.getBodyContentLength()
		// No content-length in the headers -> No body
		if contentLen == 0 {
			r.Status = ParserStatusDone
			break
		}

		if len(data) == 0 {
			if len(r.Body) < contentLen {
				return 0, ErrBodyLength
			}
		}

		r.Body = append(r.Body, data...)
		read += len(data)

		if len(r.Body) > contentLen {
			return 0, ErrBodyLength
		}

		if len(r.Body) == contentLen {
			r.Status = ParserStatusDone
		}
	case ParserStatusDone:
		break
	default:
		return 0, ErrUnknownState
	}
	return read, nil
}

func (r *Request) isDone() bool {
	return r.Status == ParserStatusDone
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, separator)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := data[:idx]
	read := idx + len(separator)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrMalformedRequestLine
	}

	method := string(parts[0])
	target := string(parts[1])

	for _, char := range method {
		if !unicode.IsUpper(char) {
			return nil, 0, ErrMalformedRequestLine
		}
	}

	versionParts := strings.Split(string(parts[2]), "/")
	if len(versionParts) != 2 || versionParts[0] != "HTTP" || versionParts[1] != "1.1" {
		return nil, 0, ErrMalformedRequestLine
	}

	version := versionParts[1]

	return &RequestLine{HttpVersion: version, RequestTarget: target, Method: method}, read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := newRequest()

	buffer := make([]byte, bufferSize)
	readIdx := 0

	for !r.isDone() {
		// Grow buffer if full
		if readIdx >= len(buffer) {
			newBuf := make([]byte, len(buffer)*2)
			copy(newBuf, buffer)
			buffer = newBuf
		}

		n, err := reader.Read(buffer[readIdx:])
		if err != nil {
			// NOTE: In real request we will have an error if we get an EOF. Now for testing when we get an EOF
			// we go to the body - in order to end the parsing if body is empty or if we have read all the body
			if err == io.EOF {
				r.Status = ParserStatusBody
			} else {
				return nil, fmt.Errorf("error reading request: %w", err)
			}
		}
		readIdx += n

		parsedIdx, err := r.parse(buffer[:readIdx])
		if err != nil {
			return nil, fmt.Errorf("error parsing request: %w", err)
		}

		copy(buffer, buffer[parsedIdx:readIdx])
		readIdx -= parsedIdx
	}

	return r, nil
}
