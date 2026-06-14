package main

import (
	"httpserver/internal/request"
	"httpserver/internal/response"
	"httpserver/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
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

		var body string
		var status response.StatusCode

		if req == nil {
			status = response.StatusBadRequest
			body = badRequestHTML
		} else {
			status = response.StatusOK
			body = successHTML
		}

		headers := response.GetDefaultHeaders(len(body))
		headers.Set("Content-Type", "text/html")

		w.WriteStatusLine(status)
		w.WriteHeaders(headers)
		w.WriteBody([]byte(body))

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
