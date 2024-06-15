package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type SafeFetch struct {
	v   map[string]bool
	mux sync.Mutex
}

func (s *SafeFetch) getValue(k string) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.v[k]
}

func (s *SafeFetch) setTrue(k string) {
	s.mux.Lock()
	s.v[k] = true
	s.mux.Unlock()

}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, status chan string, urlMap *SafeFetch) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	defer close(status)
	if depth <= 0 {
		return
	}

	if urlMap.getValue(url) {
		status <- "Already fetched url!"
		return
	}
	urlMap.setTrue(url)

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		status <- err.Error()
		return
	}
	status <- fmt.Sprintf("found: %s %q", url, body)
	statuses := make([]chan string, len(urls))
	for index, u := range urls {
		statuses[index] = make(chan string)
		go Crawl(u, depth-1, fetcher, statuses[index], urlMap)
	}

	// Wait for child goroutines.

	for i := range statuses {
		for s := range statuses[i] {
			status <- s
		}
	}
}

func main() {
	urlMap := SafeFetch{v: make(map[string]bool)}
	status := make(chan string)
	go Crawl("https://golang.org/", 4, fetcher, status, &urlMap)
	for s := range status {
		fmt.Println(s)
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
