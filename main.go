package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: ./puppy <url>\n\n")
		return
	}

	url, err := parseURL(os.Args[1])
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", url.Host, url.Port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var req string
	req += fmt.Sprintf("GET %s HTTP/1.0\r\n", url.Path)
	req += fmt.Sprintf("Host: %s\r\n", url.Host)
	req += "\r\n"

	_, err = conn.Write([]byte(req))
	if err != nil {
		panic(err)
	}

	var res Response
	res.Headers = make(map[string]string, 0)

	resReader := bufio.NewReader(conn)

	statusLine, _, err := resReader.ReadLine()
	if err != nil {
		panic(err)
	}

	statusLineWords := strings.Split(string(statusLine), " ")
	res.Version = statusLineWords[0]
	res.Status = statusLineWords[1]
	res.Explanation = statusLineWords[2]

	for {
		buf, _, err := resReader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			panic(err)
		}
		line := string(buf)

		if line == "" {
			break
		}

		headerValue := strings.Split(line, ":")
		header := headerValue[0]
		value := strings.TrimSpace(strings.Join(headerValue[1:], ""))

		res.Headers[header] = value
	}

	body, err := io.ReadAll(resReader)
	if err != nil && err.Error() != "EOF" {
		panic(err)
	}

	res.Body = string(body)

	fmt.Printf("%+v", res)
}

type Response struct {
	Version     string
	Status      string
	Explanation string
	Headers     map[string]string
	Body        string
}

type URL struct {
	Scheme string
	Host   string
	Port   string
	Path   string
}

func parseURL(urlStr string) (*URL, error) {
	var url URL

	schemeAndRest := strings.Split(urlStr, "://")
	if len(schemeAndRest) == 2 {
		url.Scheme = schemeAndRest[0]
		urlStr = schemeAndRest[1]
	} else {
		url.Scheme = "http"
	}

	if !strings.Contains(urlStr, "/") {
		urlStr += "/"
	}

	hostAndPath := strings.Split(urlStr, "/")
	url.Host = hostAndPath[0]
	if len(hostAndPath) >= 2 {
		url.Path = "/" + strings.Join(hostAndPath[1:], "/")
	} else {
		url.Path = "/"
	}

	if strings.Contains(url.Host, ":") {
		hostAndPort := strings.Split(url.Host, ":")
		url.Host = hostAndPort[0]
		url.Port = hostAndPort[1]
	} else {
		url.Port = "80"
	}

	return &url, nil
}
