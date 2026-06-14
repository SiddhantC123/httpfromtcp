package server

import (
	"httpserver/internal/request"
	"httpserver/internal/response"
	"io"
)

type Handler func(w *response.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func WriteError(w io.Writer, handlerErr *HandlerError) error {
	body := handlerErr.Message

	// Status line
	if err := response.WriteStatusLine(w, handlerErr.StatusCode); err != nil {
		return err
	}

	// Headers
	headers := response.GetDefaultHeaders(len(body))
	headers.Set("Content-Type", "text/html")

	if err := response.WriteHeaders(w, headers); err != nil {
		return err
	}

	// Body
	_, err := w.Write([]byte(body))
	return err
}
