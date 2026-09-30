package main

import (
	"bufio"
	"fmt"
	"net"
	"server/response"
)

func handleResponse(status response.Status, body string) string {
	response := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %d\r\n"+
			"Connection: close\r\n"+
			"\r\n"+
			"%s",
		status.Code,
		status.Message,
		len(body),
		body,
	)
	return response
}

func handleConnection(conn net.Conn) {
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

	body := "Hello, world!"

	response := handleResponse(response.OK, body)

	_, err := conn.Write([]byte(response))
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
