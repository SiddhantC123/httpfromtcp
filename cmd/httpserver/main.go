package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpserver/internal/headers" // Imported for headers.NewHeaders()
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

		// --- NEW: PROXY HANDLER FOR HTTPBIN ---
		if strings.HasPrefix(target, "/httpbin") {
			// 1. Route parsing
			path := strings.TrimPrefix(target, "/httpbin")
			url := "https://httpbin.org" + path

			// 2. Make the request to the real httpbin
			resp, err := http.Get(url)
			if err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    internalErrorHTML,
				}
			}
			defer resp.Body.Close()

			// 3. Write 200 OK Status Line
			if err := w.WriteStatusLine(response.StatusOK); err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    internalErrorHTML,
				}
			}

			// 4. Setup chunked headers (NO Content-Length!)
			proxyHeaders := headers.NewHeaders()
			proxyHeaders.Set("Transfer-Encoding", "chunked")
			proxyHeaders.Set("Connection", "close")

			if err := w.WriteHeaders(proxyHeaders); err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalServerError,
					Message:    internalErrorHTML,
				}
			}

			// 5. Read from httpbin and write chunks back to the client
			// You can drop this buffer size to 32 while debugging to see more chunks!
			buf := make([]byte, 1024)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					// As requested: print 'n' on every call to Read
					fmt.Printf("Read %d bytes from httpbin\n", n)

					if _, writeErr := w.WriteChunkedBody(buf[:n]); writeErr != nil {
						fmt.Println("Error writing chunk:", writeErr)
						break
					}
				}

				if err != nil {
					if err == io.EOF {
						break // Normal end of stream
					}
					fmt.Println("Error reading from httpbin:", err)
					break
				}
			}

			// 6. Send the final 0-length chunk to finish the message
			if _, err := w.WriteChunkedBodyDone(); err != nil {
				fmt.Println("Error writing chunk done:", err)
			}

			return nil
		}
		// --------------------------------------

		// --- FALLBACK: DEFAULT HANDLER ---
		body := successHTML
		// Note: Changed variable name to defaultHeaders to prevent shadowing the 'headers' package
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
