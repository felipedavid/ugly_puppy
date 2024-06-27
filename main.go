package main

import (
	"fmt"
	"strings"
)

func main() {
	url, err := parseURL("http://google.com")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", url)
}

type URL struct {
	Scheme string
	Host   string
	Path   string
}

func parseURL(urlStr string) (*URL, error) {
	var url URL

	schemeAndRest := strings.Split(urlStr, "://")
	url.Scheme = schemeAndRest[0]

	urlStr = schemeAndRest[1]

	if !strings.Contains(urlStr, "/") {
		urlStr += "/"
	}

	hostAndPath := strings.Split(urlStr, "/")
	url.Host = hostAndPath[0]
	url.Path = hostAndPath[1]

	return &url, nil
}
