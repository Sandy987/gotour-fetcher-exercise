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

type SafeMap struct {
	sMap map[string]string
	mux  sync.Mutex
}

func (sm *SafeMap) Add(k, v string) {
	sm.mux.Lock()
	sm.sMap[k] = v
	sm.mux.Unlock()
}

func (sm *SafeMap) Read(k string) (string, bool) {
	sm.mux.Lock()
	defer sm.mux.Unlock()
	val, ok := sm.sMap[k]
	return val, ok
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, safeMap *SafeMap, done chan struct{}) {
	defer func(d chan struct{}) { d <- struct{}{} }(done)
	if depth <= 0 {
		return
	}

	// Check if we already have read this url
	_, ok := safeMap.Read(url)
	if ok {
		return
	}

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Add searched url to safemap
	safeMap.Add(url, body)

	fmt.Printf("found: %s %q\n", url, body)
	subDones := make([]chan struct{}, len(urls))
	for i, u := range urls {
		subDones[i] = make(chan struct{})
		go Crawl(u, depth-1, fetcher, safeMap, subDones[i])
	}

	for _, d := range subDones {
		<-d
	}
	return
}

func main() {
	safeMap := SafeMap{sMap: make(map[string]string)}
	done := make(chan struct{})
	go Crawl("http://golang.org/", 4, fetcher, &safeMap, done)
	<-done
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
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
