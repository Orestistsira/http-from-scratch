package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Orestistsira/http-from-scratch/internal/request"
	"github.com/Orestistsira/http-from-scratch/internal/response"
	"github.com/Orestistsira/http-from-scratch/internal/server"
)

const port = 42069

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

func main() {
	server, err := server.Serve(port, handler)
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
