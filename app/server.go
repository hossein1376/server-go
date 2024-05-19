package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	HTTPv1           = "HTTP/1.1"
	StatusOK         = "200 OK"
	StatusNotFound   = "404 Not Found"
	StatusBadGateway = "502 Bad Gateway"
)

type Request struct {
	method  string
	target  string
	version string
	headers map[string]string
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
		}

		go func() {
			err = serve(conn)
			if err != nil {
				fmt.Println("Error:", err.Error())
			}
		}()
	}
}

func serve(c net.Conn) error {
	defer c.Close()

	buf := make([]byte, 128)
	n, err := c.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return writeConn(c, StatusBadGateway)
	}
	req, err := parseRequest(string(buf[:n]))
	if err != nil {
		fmt.Println("Error parsing request:", err.Error())
		return writeConn(c, StatusBadGateway)
	}

	status := StatusOK
	if req.target != "/" {
		status = StatusNotFound
	}

	return writeConn(c, status)
}

func parseRequest(req string) (*Request, error) {
	parts := strings.Split(req, "\r\n")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid parts: %d", len(parts))
	}

	reqLine := strings.Split(parts[0], " ")
	if len(reqLine) != 3 {
		return nil, fmt.Errorf("invalid request line parts: %d", len(reqLine))
	}
	method := reqLine[0]
	target := reqLine[1]
	version := reqLine[2]

	headers := map[string]string{}
	if len(parts) > 1 {
		for _, header := range parts[1:] {
			if header == "" {
				continue
			}
			headerParts := strings.SplitN(header, ":", 2)
			if len(headerParts) != 2 {
				return nil, fmt.Errorf("invalid header line parts: %d", len(headerParts))
			}
			headers[headerParts[0]] = strings.TrimLeft(headerParts[1], " ")
		}
	}

	return &Request{method, target, version, headers}, nil
}

func writeConn(conn net.Conn, status string) error {
	msg := strings.Builder{}
	msg.WriteString(HTTPv1)
	msg.WriteString(" ")
	msg.WriteString(status)
	msg.WriteString("\r\n\r\n")

	_, err := conn.Write([]byte(msg.String()))
	if err != nil {
		return fmt.Errorf("writing to connection: %w", err)
	}
	return nil
}
