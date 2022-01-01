package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type outputOptions struct {
	HideHeaders *bool
	HideBody    *bool
}

func main() {
	hideHeaders := flag.Bool("hideHeaders", false, "Show HTTP response headers")
	hideBody := flag.Bool("hideBody", false, "Don't display HTTP response body")
	flag.Parse()

	outputOptions := outputOptions{
		HideHeaders: hideHeaders,
		HideBody:    hideBody,
	}

	b := readRequests()
	requests := parseRequests(b)
	responses := doRequests(requests)
	displayResponses(responses, outputOptions)
}

func readRequests() []byte {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
		return removeShebang(input)
	}

	args := os.Args
	filename := args[len(args)-1]
	b, err := ioutil.ReadFile(filename)
	if err == nil {
		return removeShebang(b)
	}

	_, _ = fmt.Fprintln(os.Stderr, "Unable to read request")
	return nil
}

func removeShebang(b []byte) []byte {
	lines := strings.Split(string(b), "\n")
	// Remove shebang
	if strings.HasPrefix(lines[0], "#!") {
		lines = lines[1 : len(lines)-1]
	}
	request := []byte(strings.Join(lines, "\n"))
	return append(request, "\n\n"...)
}

func parseRequests(b []byte) []*http.Request {
	rawRequests := bytes.Split(b, []byte("###\n"))
	requests := make([]*http.Request, len(rawRequests))
	for i, rawRequest := range rawRequests {
		newReader := bytes.NewReader(bytes.TrimPrefix(rawRequest, []byte("\n")))
		request, err := http.ReadRequest(bufio.NewReader(newReader))
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}

		u, err := url.Parse(request.RequestURI)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
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
		response, err := client.Do(request)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
		responses[i] = response
	}
	return responses
}

func displayResponses(responses []*http.Response, options outputOptions) {
	for i := 0; i < len(responses); i++ {
		response := responses[i]
		if response.StatusCode > 300 {
			fmt.Println("Error:", response)
			break
		}

		fmt.Println(response.Proto, response.Status)

		if !*options.HideHeaders {
			headers := response.Header
			for k, v := range headers {
				for _, vv := range v {
					fmt.Printf("%s: %s\n", k, vv)
				}
			}
		}
		fmt.Println()

		if b, err := io.ReadAll(response.Body); err == nil && !*options.HideBody {
			fmt.Println(string(bytes.TrimSpace(b)))
		}

		if i < len(responses)-1 {
			fmt.Println("###")
			fmt.Println()
		}
	}
}
