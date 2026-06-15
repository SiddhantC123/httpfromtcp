package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"httpserver/internal/request"
	"httpserver/internal/response"
	"httpserver/internal/server"
)

const badRequestHTML = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const internalErrorHTML = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const successHTML = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

const port = 42069

func main() {
	handler := func(w *response.Writer, req *request.Request) *server.HandlerError {
		target := req.RequestLine.RequestTarget
		method := req.RequestLine.Method

		if method == "GET" && target == "/video" {
			videoBytes, err := os.ReadFile(("assets/mohit.mp4"))

			if err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    internalErrorHTML,
				}
			}

			videoHeaders := response.GetDefaultHeaders((len(videoBytes)))
			videoHeaders.Set("Content-Type", "video/mp4")

			if err := w.WriteStatusLine(response.StatusOK); err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    internalErrorHTML,
				}
			}

			if err := w.WriteHeaders(videoHeaders); err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    internalErrorHTML,
				}
			}

			if _, err := w.WriteBody(videoBytes); err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    internalErrorHTML,
				}
			}
			return nil

		}

		body := successHTML
		defaultHeaders := response.GetDefaultHeaders(len(body))
		defaultHeaders.Set("Content-Type", "text/html")

		if err := w.WriteStatusLine(response.StatusOK); err != nil {
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    internalErrorHTML,
			}
		}
		if err := w.WriteHeaders(defaultHeaders); err != nil {
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    internalErrorHTML,
			}
		}
		if _, err := w.WriteBody([]byte(body)); err != nil {
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    internalErrorHTML,
			}
		}

		return nil
	}

	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()

	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
