package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var ErrMalformedHeaders = fmt.Errorf("malformed headers")
var ErrMalformedFieldName = fmt.Errorf("malformed field-name")

var separator = []byte("\r\n")

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: make(map[string]string),
	}
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)

	existingValue, exists := h.headers[name]
	if exists {
		h.headers[name] = existingValue + ", " + value
	} else {
		h.headers[name] = value
	}
}

func (h *Headers) IsEmpty() bool {
	return len(h.headers) == 0
}

func (h *Headers) ForEach(fn func(name, value string)) {
	for name, value := range h.headers {
		fn(name, value)
	}
}

func (h *Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, separator)
	switch idx {
	case -1:
		return 0, false, nil
	case 0:
		return len(separator), true, nil
	}

	headerData := data[:idx]

	colonIdx := bytes.Index(headerData, []byte(":"))
	if colonIdx == -1 {
		return 0, false, ErrMalformedHeaders
	}

	// Get name
	nameData := headerData[:colonIdx]
	if nameData[len(nameData)-1] == ' ' {
		return 0, false, ErrMalformedFieldName
	}

	name := bytes.TrimSpace(nameData)
	if len(name) == 0 {
		return 0, false, ErrMalformedFieldName
	}

	// Validate name characters
	for _, b := range name {
		switch {
		case b >= 'A' && b <= 'Z':
		case b >= 'a' && b <= 'z':
		case b >= '0' && b <= '9':
		case b == '!' || b == '#' || b == '$' || b == '%' || b == '&' ||
			b == '\'' || b == '*' || b == '+' || b == '-' || b == '.' ||
			b == '^' || b == '_' || b == '`' || b == '|' || b == '~':
		default:
			return 0, false, ErrMalformedFieldName
		}
	}

	fmt.Println(string(name))

	// Get value
	valueData := headerData[colonIdx+1:]
	value := bytes.TrimSpace(valueData)

	fmt.Println(string(value))

	h.Set(string(name), string(value))

	return idx + len(separator), false, nil
}
