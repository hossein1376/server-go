package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	HTTPv1        = "HTTP/1.1"
	ContentType   = "Content-Type"
	ContentLength = "Content-Length"
	UserAgent     = "User-Agent"
)

const (
	StatusOK         = 200
	StatusNotFound   = 404
	StatusBadGateway = 502
)

type Request struct {
	Method  string
	URI     *url.URL
	Version string
	Headers Header
	Body    []byte
}

type Response struct {
	Status  int
	Headers Header
	Body    []byte
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
		return writeConn(c, Response{Status: StatusBadGateway, Body: []byte(err.Error())})
	}
	req, err := parseRequest(string(buf[:n]))
	if err != nil {
		fmt.Println("Error parsing request:", err.Error())
		return writeConn(c, Response{Status: StatusBadGateway, Body: []byte(err.Error())})
	}

	var (
		status int
		body   []byte
		header = make(Header)
	)

	switch {
	case req.URI.Path == "/":
		status = StatusOK
	case req.URI.Path == "/user-agent":
		status = StatusOK
		body = []byte(req.Headers.Get(UserAgent))
	case strings.HasPrefix(req.URI.Path, "/echo"):
		path := strings.Split(req.URI.Path, "/")
		status = StatusOK
		body = []byte(path[2])

	case strings.HasPrefix(req.URI.Path, "/files"):
		path := strings.Split(req.URI.Path, "/")
		body, status, err = getFile(path)
		if err != nil {
			fmt.Println("Error getting file:", err.Error())
			return writeConn(c, Response{Status: StatusBadGateway})
		}

	default:
		status = StatusNotFound
	}

	header.Set(ContentType, "text/plain")
	header.Set(ContentLength, strconv.Itoa(len(body)))

	return writeConn(c, Response{Status: status, Body: body, Headers: header})
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

	uri, err := url.Parse(reqLine[1])
	if err != nil {
		return nil, fmt.Errorf("invalid request line url: %s", reqLine[1])
	}

	var (
		headers = make(map[string]string)
		body    = make([]string, 0)
	)
	if len(parts) > 1 {
		for i, header := range parts[1:] {
			if header == "" {
				if i+2 < len(parts) {
					body = parts[i+2:]
				}
				break
			}
			headerParts := strings.SplitN(header, ":", 2)
			if len(headerParts) != 2 {
				return nil, fmt.Errorf("invalid header line parts: %d", len(headerParts))
			}
			headers[strings.ToLower(headerParts[0])] = strings.TrimLeft(headerParts[1], " ")
		}
	}

	return &Request{
		Method:  reqLine[0],
		URI:     uri,
		Version: reqLine[2],
		Headers: headers,
		Body:    []byte(strings.Join(body, "\r\n")),
	}, nil
}

func writeConn(conn net.Conn, resp Response) error {
	b := strings.Builder{}
	b.WriteString(HTTPv1)
	b.WriteString(" ")
	b.WriteString(strconv.Itoa(resp.Status))
	b.WriteString(" ")
	b.WriteString(StatusText(resp.Status))
	b.WriteString("\r\n")

	for k, v := range resp.Headers {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(v)
		b.WriteString("\r\n")
	}

	b.WriteString("\r\n")
	if len(resp.Body) != 0 {
		b.Write(resp.Body)
		b.WriteString("\r\n")
	}

	_, err := conn.Write([]byte(b.String()))
	if err != nil {
		return fmt.Errorf("writing to connection: %w", err)
	}
	return nil
}

func StatusText(code int) string {
	switch code {
	case StatusOK:
		return "OK"
	case StatusNotFound:
		return "Not Found"
	case StatusBadGateway:
		return "Bad Gateway"
	default:
		return ""
	}
}

func getFile(path []string) ([]byte, int, error) {
	if len(path) != 3 {
		return nil, StatusNotFound, nil
	}

	b, err := os.ReadFile(path[2])
	if err != nil {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return nil, StatusNotFound, nil
		default:
			return nil, StatusBadGateway, fmt.Errorf("readng file file: %w", err)
		}
	}

	return b, StatusOK, nil
}
