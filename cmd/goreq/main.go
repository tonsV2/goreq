package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/alecthomas/chroma/quick"
	chromaStyles "github.com/alecthomas/chroma/styles"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type options struct {
	HideHeaders *bool
	HideBody    *bool
	Raw         *bool
	FailOnError *bool
	Style       *string
}

func main() {
	hideHeaders := flag.Bool("hideHeaders", false, "Show HTTP response headers")
	hideBody := flag.Bool("hideBody", false, "Don't display HTTP response body")
	raw := flag.Bool("raw", false, "No syntax highlighting")
	failOnError := flag.Bool("failOnError", false, "Return HTTP status code if it's bigger than 300")
	styles := flag.Bool("styles", false, "List all style options")
	style := flag.String("style", "monokai", "Specify output formatting style, use -styles to get a list of all options")
	flag.Parse()

	if *styles {
		names := chromaStyles.Names()
		fmt.Println(strings.Join(names, "\n"))
		return
	}

	opts := options{
		hideHeaders,
		hideBody,
		raw,
		failOnError,
		style,
	}

	b, err := readRequests()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	requests, err := parseRequests(b)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	responses, err := doRequests(requests, opts)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = displayResponses(responses, opts)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readRequests() ([]byte, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		return removeShebang(input), nil
	}

	args := os.Args
	filename := args[len(args)-1]
	b, err := ioutil.ReadFile(filename)
	if err == nil {
		return removeShebang(b), nil
	}

	flag.Usage()
	return nil, nil
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

func parseRequests(b []byte) ([]*http.Request, error) {
	rawRequests := bytes.Split(b, []byte("###\n"))
	requests := make([]*http.Request, len(rawRequests))
	for i, rawRequest := range rawRequests {
		newReader := bytes.NewReader(bytes.TrimPrefix(rawRequest, []byte("\n")))
		request, err := http.ReadRequest(bufio.NewReader(newReader))
		if err != nil {
			return nil, err
		}

		u, err := url.Parse(request.RequestURI)
		if err != nil {
			return nil, err
		}
		request.URL = u
		request.RequestURI = ""

		requests[i] = request
	}
	return requests, nil
}

func doRequests(requests []*http.Request, opts options) ([]*http.Response, error) {
	client := http.DefaultClient
	responses := make([]*http.Response, len(requests))
	for i, request := range requests {
		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}

		if response.StatusCode > 300 && *opts.FailOnError {
			os.Exit(response.StatusCode)
		}

		responses[i] = response
	}
	return responses, nil
}

func displayResponses(responses []*http.Response, opts options) error {
	for i := 0; i < len(responses); i++ {
		response := responses[i]
		if response.StatusCode > 300 {
			fmt.Println("Error:", response)
			break
		}

		fmt.Println(response.Proto, response.Status)

		if !*opts.HideHeaders {
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
			return err
		}

		if !*opts.HideBody {
			body := string(bytes.TrimSpace(b))
			if *opts.Raw {
				fmt.Println(body)
			} else {
				contentType := response.Header["Content-Type"][0]
				lexer, err := getLexer(contentType)
				if err != nil {
					return err
				}

				err = quick.Highlight(os.Stdout, body, lexer, "terminal", *opts.Style)
				if err != nil {
					return err
				}
			}
		}

		if i < len(responses)-1 {
			fmt.Println("###")
			fmt.Println()
		}
	}
	return nil
}

func getLexer(contentType string) (string, error) {
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}

	split := strings.Split(mediatype, "/")
	return split[1], nil
}
