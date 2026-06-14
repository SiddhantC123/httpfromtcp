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

// Parse reads a single header line, adds it to the map, and returns bytes consumed.
func (h Headers) Parse(data []byte) (int, bool, error) {
	// 1. Look for the CRLF (\r\n)
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}

	// 2. If the CRLF is at index 0, it means we hit the empty line!
	if idx == 0 {
		return 2, true, nil
	}

	// 3. Extract just the line and calculate total bytes consumed
	line := data[:idx]
	consumed := idx + 2

	// 4. Split the line into exactly 2 parts at the first colon
	parts := bytes.SplitN(line, []byte(":"), 2)
	if len(parts) != 2 {
		return 0, false, errors.New("invalid header format: missing colon")
	}

	keyRaw := string(parts[0])
	valRaw := string(parts[1])

	// 5. Ensure the key has NO leading or trailing spaces!
	if strings.TrimSpace(keyRaw) != keyRaw {
		return 0, false, errors.New("invalid header key: contains leading or trailing spaces")
	}

	if keyRaw == "" {
		return 0, false, errors.New("empty header key")
	}

	// 6. Validate every character in the key against the allowed HTTP token alphabet
	for i := 0; i < len(keyRaw); i++ {
		char := keyRaw[i]
		valid := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			strings.ContainsRune("!#$%&'*+-.^_`|~", rune(char)) // The allowed special characters!

		if !valid {
			return 0, false, errors.New("invalid character in header key")
		}
	}

	// 7. Convert the validated key to lowercase
	key := strings.ToLower(keyRaw)

	// 8. Values CAN have leading spaces (like " localhost"), so we trim those
	val := strings.TrimSpace(valRaw)

	// 9. THE NEW FIX: Check if it exists, and append if it does!
	if existingVal, exists := h[key]; exists {
		h[key] = existingVal + ", " + val
	} else {
		h[key] = val
	}

	return consumed, false, nil
}
