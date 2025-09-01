package headers

import (
	"bytes"
	"fmt"
	"strings"
)

var ErrMalformedHeaders = fmt.Errorf("malformed headers")
var ErrMalformedFieldKey = fmt.Errorf("malformed field-key")

var separator = []byte("\r\n")

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
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

	// Get key
	keyData := headerData[:colonIdx]
	if keyData[len(keyData)-1] == ' ' {
		return 0, false, ErrMalformedFieldKey
	}

	key := bytes.TrimSpace(keyData)
	if len(key) == 0 {
		return 0, false, ErrMalformedFieldKey
	}

	// Validate key characters
	for _, b := range key {
		switch {
		case b >= 'A' && b <= 'Z':
		case b >= 'a' && b <= 'z':
		case b >= '0' && b <= '9':
		case b == '!' || b == '#' || b == '$' || b == '%' || b == '&' ||
			b == '\'' || b == '*' || b == '+' || b == '-' || b == '.' ||
			b == '^' || b == '_' || b == '`' || b == '|' || b == '~':
		default:
			return 0, false, ErrMalformedFieldKey
		}
	}

	fmt.Println(string(key))

	// Get value
	valueData := headerData[colonIdx+1:]
	value := bytes.TrimSpace(valueData)

	fmt.Println(string(value))

	h[strings.ToLower(string(key))] = string(value)

	return idx + len(separator), false, nil
}

func NewHeaders() Headers {
	return Headers{}
}
