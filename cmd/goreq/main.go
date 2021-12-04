package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type outputOptions struct {
	ShowHeaders *bool
	ShowBody    *bool
}

// @title goreq
// @version 0.1.0
func main() {
	showHeaders := flag.Bool("showHeaders", false, "Show HTTP response headers")
	showBody := flag.Bool("showBody", true, "Show HTTP response body")
	flag.Parse()

	outputOptions := outputOptions{
		ShowHeaders: showHeaders,
		ShowBody:    showBody,
	}

	b := readRequests()
	requests := parseRequests(b)
	responses := doRequests(requests)
	displayResponses(responses, outputOptions)
}

func readRequests() []byte {
	var b []byte
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalln(err)
		}
		b = input
	} else {
		log.Fatalln("Unable to read input from stdin")
	}
	return b
}

func parseRequests(b []byte) []*http.Request {
	rawRequests := bytes.Split(b, []byte("###\n"))
	requests := make([]*http.Request, len(rawRequests))
	for i, rawRequest := range rawRequests {
		newReader := bytes.NewReader(bytes.TrimPrefix(rawRequest, []byte("\n")))
		request, err := http.ReadRequest(bufio.NewReader(newReader))
		if err != nil {
			log.Fatalln(err)
		}

		u, err := url.Parse(request.RequestURI)
		if err != nil {
			log.Fatalln(err)
		}
		request.URL = u
		request.RequestURI = ""

		requests[i] = request
	}
	return requests
}

func doRequests(requests []*http.Request) []*http.Response {
	client := http.DefaultClient
	responses := make([]*http.Response, len(requests))
	for i, request := range requests {
		log.Println(request.Method, request.URL)
		response, err := client.Do(request)
		if err != nil {
			log.Fatalln(err)
		}
		responses[i] = response
	}
	return responses
}

func displayResponses(responses []*http.Response, options outputOptions) {
	for _, response := range responses {
		if response.StatusCode > 300 {
			log.Println("Error:", response)
			break
		}

		if *options.ShowHeaders {
			log.Printf("%+v", response.Header)
		}

		if b, err := io.ReadAll(response.Body); err == nil && *options.ShowBody {
			log.Println(string(bytes.TrimSpace(b)))
		}
	}
}
