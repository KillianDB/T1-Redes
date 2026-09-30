package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"server/response"
)

func handleConnection(conn net.Conn) {
	var newReponse string
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading request:", err)
			return
		}

		fmt.Print(line)

		if line == "\r\n" {
			break
		}
	}

	// body := "Hello, world!"

	content, err := os.ReadFile("index.html")
	if err != nil {
		newReponse = response.BuildResponse(response.BadRequest, []byte(err.Error()), "text/plain")
	} else {
		newReponse = response.BuildResponse(response.OK, []byte(content), "text/html")
	}

	_, err = conn.Write([]byte(newReponse))
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Failed to listen on port:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}
