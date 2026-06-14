package main

import (
	"fmt"
	"httpserver/internal/request"
	"net"
)

func main() {
	listner := createListener()
	defer func() {
		listner.Close()
		fmt.Println("Connection closed")
	}()

	for {
		con, err := listner.Accept()
		if err != nil {
			fmt.Println("Connection refused")
			continue
		}
		fmt.Println("Connection accepted")

		req, err := request.RequestFromReader(con)
		if err != nil {
			fmt.Println("Failed to parse request:", err)
			con.Close()
			continue
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		con.Close()
	}
}

func createListener() net.Listener {
	listener, _ := net.Listen("tcp", "127.0.0.1:42069")
	return listener
}
