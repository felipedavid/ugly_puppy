package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	inputUrl := ""
	if len(os.Args) >= 2 {
		inputUrl = os.Args[1]
	} else if len(os.Args) == 1 {
		inputUrl = "file:///home/batman/work/ugly_puppy/test.html"
	} else {
		fmt.Fprintf(os.Stderr, "Usage: ./puppy <url>\n\n")
		return
	}

	url, err := parseURL(inputUrl)
	if err != nil {
		panic(err)
	}

	if url.Scheme == "file" {
		buf, err := os.ReadFile(url.Path)
		if err != nil {
			panic(err)
		}

		if url.viewSource {
			_, err = os.Stdout.Write(buf)
			if err != nil {
				panic(err)
			}
		} else {
			err = renderHTML(os.Stdout, buf)
			if err != nil {
				panic(err)
			}
		}

		return
	}

	if url.Scheme == "data" {
		if url.MediaType == "text/plain" {
			fmt.Printf("%v", url.Data)
		} else if url.MediaType == "text/html" {
			if url.viewSource {
				_, err = os.Stdout.Write(url.Data)
				if err != nil {
					panic(err)
				}
			} else {
				err = renderHTML(os.Stdout, url.Data)
				if err != nil {
					panic(err)
				}
			}
		} else {
			panic(fmt.Sprintf("unsuppoted media type '%s'", url.MediaType))
		}
		return
	}

	var conn io.ReadWriteCloser
	if url.Scheme == "https" {
		conn, err = tls.Dial("tcp", fmt.Sprintf("%s:%s", url.Host, url.Port), nil)
	} else if url.Scheme == "http" {
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%s", url.Host, url.Port))
	} else {
		panic("unsupported url schema")
	}
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var req string
	req += fmt.Sprintf("GET %s HTTP/1.1\r\n", url.Path)
	req += fmt.Sprintf("Host: %s\r\n", url.Host)
	req += "Connection: close\r\n"
	req += "User-Agent: ugly-puppy\r\n"
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

	res.Body = body
	if url.viewSource {
		_, err = os.Stdout.Write(url.Data)
		if err != nil {
			panic(err)
		}
	} else {
		err = renderHTML(os.Stdout, url.Data)
		if err != nil {
			panic(err)
		}
	}
}

type Response struct {
	Version     string
	Status      string
	Explanation string
	Headers     map[string]string
	Body        []byte
}

type URL struct {
	Scheme string
	Host   string
	Port   string
	Path   string

	viewSource bool

	// Scheme = data
	MediaType string
	Data      []byte
}

func parseURL(urlStr string) (*URL, error) {
	var url URL

	viewSourcePrefix := "view-source:"
	if strings.HasPrefix(urlStr, viewSourcePrefix) {
		url.viewSource = true
		urlStr = urlStr[len(viewSourcePrefix):]
	}

	schemeAndRest := strings.Split(urlStr, ":")
	if len(schemeAndRest) == 2 {
		url.Scheme = schemeAndRest[0]
		urlStr = schemeAndRest[1]
		if urlStr[0] == '/' && urlStr[1] == '/' {
			urlStr = urlStr[2:]
		}

		if url.Scheme == "data" {
			mediaTypeAndData := strings.Split(urlStr, ",")
			url.MediaType = mediaTypeAndData[0]
			url.Data = []byte(mediaTypeAndData[1])
		}
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
		if url.Scheme == "https" {
			url.Port = "443"
		} else {
			url.Port = "80"
		}
	}

	return &url, nil
}

func renderHTML(w io.Writer, content []byte) error {
	inTag := false
	buf := make([]byte, 0, len(content))

	for i := 0; i < len(content); i++ {
		ch := content[i]

		switch ch {
		case '<':
			inTag = true
		case '>':
			inTag = false
		case '&':
			chToPrint := byte('&')
			if i+4 <= len(content) {
				possibleEscapeStr := content[i : i+4]
				if string(possibleEscapeStr) == "&lt;" {
					chToPrint = '<'
					i += 3
				} else if string(possibleEscapeStr) == "&gt;" {
					chToPrint = '>'
					i += 3
				}
			}
			buf = append(buf, chToPrint)
		default:
			if !inTag {
				buf = append(buf, content[i])
			}
		}
	}

	_, err := w.Write(buf)
	if err != nil {
		return err
	}

	return nil
}
