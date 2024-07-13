package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: ./uggly_puppy <url>\n")
		os.Exit(-1)
	}

	url, err := NewURL(os.Args[1])
	if err != nil {
		panic(err)
	}

	_, err = url.Request()
	if err != nil {
		panic(err)
	}
}

type URLScheme string

const (
	HTTP  URLScheme = "http"
	HTTPS URLScheme = "https"
)

type URL struct {
	Scheme URLScheme
	Host   string
	Port   int
	Path   string
}

var (
	ErrUnspecifiedScheme = errors.New("unspecified scheme")
	ErrEmptyDomain       = errors.New("empty domain")
)

func NewURL(urlStr string) (*URL, error) {
	url := &URL{}

	schemeAndRest := strings.Split(urlStr, "://")
	if len(schemeAndRest) < 2 {
		return nil, ErrUnspecifiedScheme
	}

	url.Scheme = URLScheme(schemeAndRest[0])
	switch url.Scheme {
	case HTTP:
		url.Port = 80
	case HTTPS:
		url.Port = 443
	}

	hostAndRest := strings.Split(schemeAndRest[1], "/")
	if len(hostAndRest) == 0 {
		return nil, ErrEmptyDomain
	}

	url.Host = hostAndRest[0]
	url.Path = "/"

	if len(hostAndRest) > 1 {
		url.Path += strings.Join(hostAndRest[1:], "/")
	}

	return url, nil
}

type Response struct {
}

func (u *URL) Request() (*Response, error) {
	res := &Response{}

	conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", u.Host, u.Port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	payload := fmt.Sprintf("GET %s HTTP/1.0\r\n", u.Path)
	payload += fmt.Sprintf("Host: %s\r\n\r\n", u.Host)

	reqBuf := []byte(payload)

	wLen, err := conn.Write(reqBuf)
	if err != nil {
		return nil, err
	}

	if wLen != len(reqBuf) {
		return nil, errors.New("unable to send all payload")
	}

	resReader := bufio.NewReader(conn)

	for {
		line, _, err := resReader.ReadLine()
		if err != nil {
			return nil, err
		}

		fmt.Println(string(line))
	}

	return res, nil
}
