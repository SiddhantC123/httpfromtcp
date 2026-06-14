package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test 1: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"]) // Changed to lowercase "host"!
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test 2: Valid single header with extra whitespace in the value
	headers = NewHeaders()
	data = []byte("Accept:    */* \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "*/*", headers["accept"])
	assert.Equal(t, 17, n) // <--- CHANGE THIS FROM 20 TO 17
	assert.False(t, done)

	// Test 3: Invalid spacing header (Space before colon)
	headers = NewHeaders()
	data = []byte("Host : localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test 4: Valid done (The final blank line)
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test 5: Capital letters in key are converted to lowercase
	headers = NewHeaders()
	data = []byte("HoSt: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test 6: Invalid character in key
	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\n\r\n") // @ is not allowed!
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test 7: Multiple values for the same header key
	headers = NewHeaders()
	// Manually insert a starting header as requested by the assignment
	headers["set-person"] = "lane-loves-go"

	// Parse a new header with the EXACT SAME key
	data = []byte("Set-Person: prime-loves-zig\r\n\r\n")
	n, done, err = headers.Parse(data)

	require.NoError(t, err)
	// Assert that they were combined with a comma!
	assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
	assert.Equal(t, 29, n)
	assert.False(t, done)
}
