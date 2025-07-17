package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type scrapeResult struct {
	URL     string `json:"url"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

func main() {
	fmt.Println("======Starting web scraper======")

	websiteUrls := []string{
		"https://golang.org/pkg/",
		"https://news.ycombinator.com/",
		"https://duckduckgo.com/",
		"https://www.worldometers.info/",
		"https://old.reddit.com/r/golang/",
		"https://www.allrecipes.com/",
		"https://books.toscrape.com/",
		"https://www.imdb.com/chart/top/",
		"https://genius.com/",
		"https://webscraper.io/test-sites/e-commerce/static",
		"https://forecast.weather.gov/",
		"https://remoteok.io/",
		"https://www.google.com/search?q=golang", // ⚠️ Google blocks scrapers fast
		"https://stackoverflow.com/tags",
		"https://slickdeals.net/",
		"https://www.meetup.com/find/events/",
		"https://ocw.mit.edu/",
		"https://www.theverge.com/",
		"https://techcrunch.com/",
		"https://developer.mozilla.org/en-US/",
	}

	var wg sync.WaitGroup
	jobs := make(chan string, len(websiteUrls))
	results := make(chan scrapeResult, len(websiteUrls))
	var client = &http.Client{
		Timeout: time.Second * 5,
	}

	var contentReceived []scrapeResult

	maxWorkers := 20

	for i := range maxWorkers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// fmt.Printf("Worker %d started.\n", id)
			var result scrapeResult
			for url := range jobs {
				result = fetchUrls(url, results, client)
			}
			if result.Error != "" {
				fmt.Printf("Worker %d encountered an error: %s\n", id, result.Error)
			} else {
				fmt.Printf("Worker %d successfully fetched content from %s\n", id, result.URL)
			}
		}(i + 1)
	}

	for _, websiteUrl := range websiteUrls {
		jobs <- websiteUrl
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		contentReceived = append(contentReceived, result)
	}

	fmt.Printf("Handled %d URLs\n", len(contentReceived))
	fmt.Println("======Ending web scraper======")
}

func fetchUrls(url string, results chan<- scrapeResult, client *http.Client) scrapeResult {
	resp, err := client.Get(url)
	var result scrapeResult
	if err != nil {
		result = scrapeResult{URL: url, Error: fmt.Sprintf("Error while fetching %s: %s", url, err.Error())}
		results <- result
		return result
	}
	//goland:noinspection ALL
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		result = scrapeResult{URL: url, Error: fmt.Sprintf("bad status code: %d", resp.StatusCode)}
		results <- result
		return result
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result = scrapeResult{URL: url, Error: fmt.Sprintf("failed to read response body: %v", err)}
		results <- result
		return result
	}

	result = scrapeResult{URL: url, Content: string(body)}
	results <- result
	return result
}
