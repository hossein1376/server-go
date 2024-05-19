package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	HTTPv1   = "HTTP/1.1"
	StatusOK = "200 OK"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection:", err.Error())
	}

	err = serve(conn)
	if err != nil {
		fmt.Println("Error writing to connection:", err.Error())
	}
}

func serve(c net.Conn) error {
	msg := strings.Builder{}
	msg.WriteString(HTTPv1)
	msg.WriteString(" ")
	msg.WriteString(StatusOK)
	msg.WriteString("\r\n\r\n")

	_, err := c.Write([]byte(msg.String()))
	return err
}
