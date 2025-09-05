package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Orestistsira/http-from-scratch/internal/headers"
	"github.com/Orestistsira/http-from-scratch/internal/request"
	"github.com/Orestistsira/http-from-scratch/internal/response"
	"github.com/Orestistsira/http-from-scratch/internal/server"
)

const port = 42069
const bufferSize = 1024

func hashToStr(b []byte) string {
	out := ""
	for _, b := range b {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

// handler handles incoming HTTP requests
func handler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		b := `<html>
				<head>
					<title>400 Bad Request</title>
				</head>
				<body>
					<h1>Bad Request</h1>
					<p>Your request honestly kinda sucked.</p>
				</body>
			</html>`

		h := response.GetHTMLHeaders(len(b))
		w.WriteStatusLine(response.HTTP_400)
		w.WriteHeaders(h)
		w.WriteBody([]byte(b))
	case "/myproblem":
		b := `<html>
				<head>
					<title>500 Internal Server Error</title>
				</head>
				<body>
					<h1>Internal Server Error</h1>
					<p>Okay, you know what? This one is on me.</p>
				</body>
			</html>`

		h := response.GetHTMLHeaders(len(b))
		w.WriteStatusLine(response.HTTP_500)
		w.WriteHeaders(h)
		w.WriteBody([]byte(b))
	default:
		b := `<html>
				<head>
					<title>200 OK</title>
				</head>
				<body>
					<h1>Success!</h1>
					<p>Your request was an absolute banger.</p>
				</body>
			</html>`

		h := response.GetHTMLHeaders(len(b))
		w.WriteStatusLine(response.HTTP_200)
		w.WriteHeaders(h)
		w.WriteBody([]byte(b))
	}
}

func proxyHandler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget

	if !strings.HasPrefix(target, "/httpbin") {
		log.Println("Wrong prefix")
		return
	}

	target = strings.TrimPrefix(target, "/httpbin")

	res, err := http.Get("https://httpbin.org" + target)
	if err != nil {
		h := response.GetDefaultHeaders(0)
		w.WriteStatusLine(response.HTTP_500)
		w.WriteHeaders(h)
		return
	}

	h := response.GetChunkedWithTrailerHeaders()
	w.WriteStatusLine(response.HTTP_200)
	w.WriteHeaders(h)

	buffer := make([]byte, bufferSize)
	readIdx := 0

	for {
		if readIdx >= len(buffer) {
			newBuf := make([]byte, len(buffer)*2)
			copy(newBuf, buffer)
			buffer = newBuf
		}

		n, err := res.Body.Read(buffer[readIdx:])
		if err != nil {
			break
		}
		w.WriteChunkedBody(buffer[readIdx : readIdx+n])
		fmt.Printf("%s", buffer)

		readIdx += n
	}
	trailers := headers.NewHeaders()
	hash := sha256.Sum256(buffer[:readIdx])
	trailers.Set("X-Content-SHA256", hashToStr(hash[:]))
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(buffer[:readIdx])))
	w.WriteTrailers(*trailers)
}

func main() {
	server, err := server.Serve(port, proxyHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
