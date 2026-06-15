package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}
func (h Headers) Set(key, value string) {
	h[strings.ToLower(key)] = value
}

func (h Headers) Parse(data []byte) (int, bool, error) {

	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}

	if idx == 0 {
		return 2, true, nil
	}

	line := data[:idx]
	consumed := idx + 2

	parts := bytes.SplitN(line, []byte(":"), 2)
	if len(parts) != 2 {
		return 0, false, errors.New("invalid header format: missing colon")
	}

	keyRaw := string(parts[0])
	valRaw := string(parts[1])

	if strings.TrimSpace(keyRaw) != keyRaw {
		return 0, false, errors.New("invalid header key: contains leading or trailing spaces")
	}

	if keyRaw == "" {
		return 0, false, errors.New("empty header key")
	}

	for i := 0; i < len(keyRaw); i++ {
		char := keyRaw[i]
		valid := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			strings.ContainsRune("!#$%&'*+-.^_`|~", rune(char))

		if !valid {
			return 0, false, errors.New("invalid character in header key")
		}
	}

	key := strings.ToLower(keyRaw)

	val := strings.TrimSpace(valRaw)

	if existingVal, exists := h[key]; exists {
		h[key] = existingVal + ", " + val
	} else {
		h[key] = val
	}

	return consumed, false, nil
}
