package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidHeaderFormat = errors.New("invalid header format")
	ErrInvalidFieldFormat  = errors.New("invalid field format")
	crlf                   = []byte("\r\n")
	specialChars           = []byte("!#$%&'*+-.^_`|~")
)

type Headers map[string]string

func (h Headers) Get(k string) (string, bool) {
	k = strings.ToLower(k)
	val, exist := h[k]

	return val, exist
}

func (h Headers) Set(k string, v string) {
	k = strings.ToLower(k) // make all header to lower case
	val, exist := h[k]
	if !exist {
		h[k] = v
	} else {
		h[k] = fmt.Sprintf("%s, %s", val, v)
	}
}

func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Remove(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}

func NewHeaders() Headers {
	return make(Headers)
}

func isValidHeaderKeyByte(b byte) bool {
	if b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' {
		return true
	}
	return bytes.IndexByte(specialChars, b) != -1
}

func isValidHeaderKey(key []byte) bool {
	if len(key) == 0 {
		return false
	}
	for _, b := range key {
		if !isValidHeaderKeyByte(b) {
			return false
		}
	}
	return true
}

func parseHeader(data []byte) (string, string, error) {
	parts := bytes.SplitN(data, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", ErrInvalidFieldFormat
	}

	key := parts[0]
	val := bytes.TrimSpace(parts[1])
	if !isValidHeaderKey(key) {
		return "", "", ErrInvalidFieldFormat
	}

	return string(key), string(val), nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		// invalid byte, could be incomplete byte
		idx := bytes.Index(data[read:], crlf)
		if idx == -1 {
			return read, false, nil
		}

		// empty header
		if idx == 0 {
			done = true
			read += len(crlf)
			return read, done, nil
		}
		k, v, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}
		read += idx + len(crlf)
		h.Set(k, v)
	}
}
