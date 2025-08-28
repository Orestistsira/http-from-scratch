package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string)

	go func() {
		defer f.Close()
		defer close(out)

		buffer := make([]byte, 8)
		currentLine := ""

		for {
			n, err := f.Read(buffer)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Error reading file:", err)
				return
			}

			parts := strings.Split(string(buffer[:n]), "\n")

			// For all parts except the last, emit completed lines
			for i := 0; i < len(parts)-1; i++ {
				line := currentLine + parts[i]
				out <- line
				currentLine = ""
			}

			// Save the last part for the next iteration
			currentLine += parts[len(parts)-1]
		}

		if currentLine != "" {
			out <- currentLine
		}
	}()

	return out
}

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
		fmt.Println("A connection has been accepted")

		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Println(line)
		}

		fmt.Println("Connection closed")
	}
}
