package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

var ErrMalformedRequestLine = fmt.Errorf("malformed request-line")
var ErrParserDoneStatus = fmt.Errorf("error trying to read data in a done state")
var ErrUnknownState = fmt.Errorf("error unknown state")

var separator = []byte("\r\n")

const bufferSize = 8

type ParserStatus string

const (
	ParserStatusInit = "init"
	ParserStatusDone = "done"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Status      ParserStatus
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

	switch r.Status {
	case ParserStatusInit:
		rl, n, err := parseRequestLine(data[read:])
		if err != nil {
			return 0, err
		}

		if n == 0 {
			break
		}

		r.RequestLine = *rl
		read += n

		r.Status = ParserStatusDone
	case ParserStatusDone:
		return 0, ErrParserDoneStatus
		// break outer
	default:
		return 0, ErrUnknownState
	}

	// outer:
	// 	for {
	// 		switch r.Status {
	// 		case ParserStatusInit:
	// 			rl, n, err := parseRequestLine(data[read:])
	// 			if err != nil {
	// 				return 0, err
	// 			}

	// 			if n == 0 {
	// 				break outer
	// 			}

	// 			r.RequestLine = *rl
	// 			read += n

	// 			r.Status = ParserStatusDone
	// 		case ParserStatusDone:
	// 			return 0, ErrParserDoneStatus
	// 			// break outer
	// 		default:
	// 			return 0, ErrUnknownState
	// 		}
	// 	}
	return read, nil
}

func (r *Request) isDone() bool {
	return r.Status == ParserStatusDone
}

func newRequest() *Request {
	return &Request{
		Status: ParserStatusInit,
	}
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
	request := newRequest()

	buffer := make([]byte, bufferSize)
	readIdx := 0

	for !request.isDone() {
		// Grow buffer if full
		if readIdx >= len(buffer) {
			newBuf := make([]byte, len(buffer)*2)
			copy(newBuf, buffer)
			buffer = newBuf
		}

		n, err := reader.Read(buffer[readIdx:])
		if err != nil {
			if err == io.EOF {
				request.Status = ParserStatusDone
				break
			}
			return nil, fmt.Errorf("error reading request: %w", err)
		}
		readIdx += n

		fmt.Println(string(buffer[:readIdx]))

		parsedIdx, err := request.parse(buffer[:readIdx])
		if err != nil {
			return nil, fmt.Errorf("error parsing request: %w", err)
		}

		copy(buffer, buffer[parsedIdx:readIdx])
		readIdx -= parsedIdx
	}

	return request, nil
}
