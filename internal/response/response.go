package response

import (
	"fmt"
	"httpserver/internal/headers"
	"io"
)

// Fake enum
type StatusCode int
type writerState int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

const (
	stateStatusLine writerState = iota
	stateHeaders
	stateBody
)

type Writer struct {
	w     io.Writer
	state writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:     w,
		state: stateStatusLine,
	}
}

func (rw *Writer) WriteStatusLine(statusCode StatusCode) error {
	if rw.state != stateStatusLine {
		return fmt.Errorf("wrong order: status line")
	}

	err := WriteStatusLine(rw.w, statusCode)
	if err != nil {
		return err
	}

	rw.state = stateHeaders
	return nil
}

func (rw *Writer) WriteHeaders(h headers.Headers) error {
	if rw.state != stateHeaders {
		return fmt.Errorf("wrong order: headers")
	}

	err := WriteHeaders(rw.w, h)
	if err != nil {
		return err
	}

	rw.state = stateBody
	return nil
}

func (rw *Writer) WriteBody(p []byte) (int, error) {
	if rw.state != stateBody {
		return 0, fmt.Errorf("wrong order: body")
	}

	return rw.w.Write(p)
}

// 1. Write status line
func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reason string

	switch statusCode {
	case StatusOK:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternalServerError:
		reason = "Internal Server Error"
	default:
		reason = ""
	}

	line := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reason)

	_, err := w.Write([]byte(line))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for key, value := range h {
		line := fmt.Sprintf("%s: %s\r\n", key, value)
		if _, err := w.Write([]byte(line)); err != nil {
			return err
		}
	}

	// End of headers
	_, err := w.Write([]byte("\r\n"))
	return err
}
