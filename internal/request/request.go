package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var ErrMalformedRequestLine = fmt.Errorf("malformed request-line")

var SEPARATOR = "\r\n"

func parseRequestLine(s string) (*RequestLine, string, error) {
	idx := strings.Index(s, SEPARATOR)
	if idx == -1 {
		return nil, s, nil
	}

	startLine := s[:idx]
	restMsg := s[idx+len(SEPARATOR):]

	parts := strings.Split(startLine, " ")
	if len(parts) != 3 {
		return nil, restMsg, ErrMalformedRequestLine
	}

	method := parts[0]
	target := parts[1]

	for _, char := range method {
		if !unicode.IsUpper(char) {
			return nil, restMsg, ErrMalformedRequestLine
		}
	}

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 || versionParts[0] != "HTTP" || versionParts[1] != "1.1" {
		return nil, restMsg, ErrMalformedRequestLine
	}

	version := versionParts[1]

	return &RequestLine{HttpVersion: version, RequestTarget: target, Method: method}, restMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading request: %w", err)
	}

	requestLine, _, err := parseRequestLine(string(data))
	if err != nil {
		return nil, fmt.Errorf("error parsing request-line: %w", err)
	}

	return &Request{RequestLine: *requestLine}, nil
}
