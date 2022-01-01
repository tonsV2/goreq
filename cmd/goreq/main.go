package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/alecthomas/chroma/quick"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type outputOptions struct {
	HideHeaders *bool
	HideBody    *bool
	Raw         *bool
}

func main() {
	hideHeaders := flag.Bool("hideHeaders", false, "Show HTTP response headers")
	hideBody := flag.Bool("hideBody", false, "Don't display HTTP response body")
	raw := flag.Bool("raw", false, "No syntax highlighting")
	flag.Parse()

	outputOptions := outputOptions{
		hideHeaders,
		hideBody,
		raw,
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

	flag.Usage()
	os.Exit(1)
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
			os.Exit(1)
		}

		u, err := url.Parse(request.RequestURI)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
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
			os.Exit(1)
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

		b, err := io.ReadAll(response.Body)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
			return
		}

		if !*options.HideBody {
			body := string(bytes.TrimSpace(b))
			if *options.Raw {
				fmt.Println(body)
			} else {
				contentType := response.Header["Content-Type"][0]
				lexer := getLexer(contentType)
				err := quick.Highlight(os.Stdout, body, lexer, "terminal", "monokai")
				if err != nil {
					_, _ = fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
		}

		if i < len(responses)-1 {
			fmt.Println("###")
			fmt.Println()
		}
	}
}

func getLexer(contentType string) string {
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	split := strings.Split(mediatype, "/")
	return split[1]
}
