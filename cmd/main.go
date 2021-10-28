// Package webCheck polls URLs for response and content validation.

// TODO:
// 1.  Make it work. DONE
// 2.  Grab response codes. DONE
// 3.  Check for key words in response.Body. DONE
// 4.  Take a file of URLs. DONE
// 4.1  Error checking.
// 4.2  Check for and add http:// if missing. Or work around.
// 5.  Add key words to data file. (?csv)
// 6.  Set up base round of tests.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Page struct holds the results of an http request.
type Page struct {
	URL      string
	Size     int
	Response int
	Body     []byte
}

// file var is the filename to get URLs from.
var file = flag.String("f", "-", "File to read")

// trigger var is the string to search the response.Body for.
var trigger string

// hasString returns a bool, is the term string in the response body?
func hasString(rb []byte, term string) bool {
	b := string(rb[:len(rb)])
	return strings.Contains(b, term)
}

// lines takes the file of URLs and returns a slice to iterate over.
func lines(file *os.File) (lines []string) {
	input := bufio.NewScanner(file)
	for input.Scan() {
		my_string := input.Text()
		if len(my_string) > 2 {
			lines = append(lines, my_string)
		}
	}
	return
}

func main() {
	const (
		defaultString	= "Domici"
		usageTrigger	= "pattern to search for"
	)
	flag.StringVar(&trigger, "trigger", defaultString, usageTrigger)
	flag.StringVar(&trigger, "t", "Guido", usageTrigger+" short version")

	flag.Parse()
	data, err := os.Open(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open failed: %v\n", err)
		os.Exit(1)
	}
	defer data.Close()
	urls := lines(data)
	results := make(chan Page)

	for _, url := range urls {
		go func(url string) {
			res, err := http.Get(url)
			if err != nil {
				panic(err)
			}
			defer res.Body.Close()

			bs, err := ioutil.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}
			resp := res.StatusCode

			results <- Page{
				URL:      url,
				Size:     len(bs),
				Response: resp,
				Body:     bs,
			}
		}(url)
	}

	for range urls {
		result := <-results
		fmt.Printf("For %s the response was: %d", result.URL, result.Response)
		if hasString(result.Body, trigger) {
			fmt.Printf(", and %s was found in the body", trigger)
		}
		fmt.Printf(".\n")
	}

}
