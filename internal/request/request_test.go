package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestParse(t *testing.T) {
	t.Run("Standard Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: bootdev-client\r\n\r\n",
			numBytesPerRead: 10,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

		// Assert Headers parsed correctly
		assert.Equal(t, "localhost:42069", r.Headers["host"])
		assert.Equal(t, "bootdev-client", r.Headers["user-agent"])
	})

	t.Run("Empty Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\n\r\n", // Immediately followed by blank line
			numBytesPerRead: 10,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)

		// Should have successfully parsed, but the headers map should be empty
		assert.Empty(t, r.Headers)
	})

	t.Run("Malformed Header", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n", // Missing the colon!
			numBytesPerRead: 10,
		}
		_, err := RequestFromReader(reader)

		// This should trigger your headers parser error and bubble up
		require.Error(t, err)
	})

	t.Run("Duplicate Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nAccept: text/html\r\nAccept: application/json\r\n\r\n",
			numBytesPerRead: 10,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)

		// Should combine both values with a comma and space
		assert.Equal(t, "text/html, application/json", r.Headers["accept"])
	})

	t.Run("Case Insensitive Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHoSt: localhost:42069\r\n\r\n", // Mixed case key
			numBytesPerRead: 10,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)

		// The key must be saved as entirely lowercase
		assert.Equal(t, "localhost:42069", r.Headers["host"])
	})

	t.Run("Missing End of Headers", func(t *testing.T) {
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069", // Missing the final \r\n\r\n
			numBytesPerRead: 10,
		}
		r, err := RequestFromReader(reader)

		// Because the stream hits EOF before finishing the header section,
		// depending on how strict the server is, this might either return the partial request or an error.
		// We just want to ensure it doesn't crash here.
		if err == nil {
			require.NotNil(t, r)
		}
	})

	t.Run("Standard Body", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "POST", r.RequestLine.Method)
		assert.Equal(t, "hello world!\n", string(r.Body))
	})

	t.Run("Empty Body, 0 reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 0\r\n" +
				"\r\n", // No body follows the blank line
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, 0, len(r.Body))
	})

	t.Run("Empty Body, no reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "GET / HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n", // No Content-Length header at all
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, 0, len(r.Body))
	})

	t.Run("Body shorter than reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content", // This is only 15 bytes, but we claimed 20!
			numBytesPerRead: 3,
		}
		_, err := RequestFromReader(reader)
		// Because the stream ends (EOF) before we get the full 20 bytes, it should error out.
		require.Error(t, err)
	})
}

// ==========================================
// Test Utilities
// ==========================================

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
