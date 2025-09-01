package main

import (
	"fmt"
	"net"

	"github.com/Orestistsira/http-from-scratch/internal/request"
)

func main() {
	// Listen on TCP port 42069 on all available unicast and
	// anycast IP addresses of the local system.
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println("Error setting up tcp listener:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server listening on :42069")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			return
		}
		defer conn.Close()

		fmt.Println("A connection has been accepted")

		r, err := request.RequestFromReader(conn)

		if err != nil {
			fmt.Println("Error parsing request:", err)
			continue
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Println("Connection closed")
	}
}
