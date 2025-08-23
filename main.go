package main

import (
	"fmt"
	"io"
	"os"
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

			if len(parts) == 1 {
				currentLine += parts[0]
			} else if len(parts) == 2 {
				out <- (currentLine + parts[0])
				currentLine = parts[1]
			} else {
				fmt.Println("Error: More than 2 split parts found")
				return
			}
		}

		if currentLine != "" {
			out <- currentLine
		}
	}()

	return out
}

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	lines := getLinesChannel(file)
	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}
