package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type crawlResult struct {
	sourceUrl string
	foundUrls []string
	Error     string
}

func main() {
	fmt.Println("======Starting web crawler======")

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

	totalUrlsVisited := 0
	visitedUrls := make(map[string]bool)
	for _, url := range websiteUrls {
		visitedUrls[url] = false
	}

	var wg sync.WaitGroup          //This is for pending jobs, not workers
	jobs := make(chan string, 200) // Buffered channel size needs to be updated??
	results := make(chan crawlResult, 100)
	var client = &http.Client{
		Timeout: time.Second * 5,
	}

	maxWorkers := 30

	for i := range maxWorkers {
		go func(id int) {
			for url := range jobs {
				results <- crawl(url, client)
			}
			fmt.Printf("Worker %d finished processing jobs.\n", id)
		}(i + 1)
	}

	go func() {
		wg.Wait()
		close(jobs)
		close(results)
	}()

	wg.Add(len(websiteUrls)) // Add the number of initial URLs to the wait group
	for _, websiteUrl := range websiteUrls {
		visitedUrls[websiteUrl] = true
		jobs <- websiteUrl
	}

	for result := range results {
		totalUrlsVisited++
		if result.Error != "" {
			// fmt.Printf("Error while crawling %s: %s\n", result.sourceUrl, result.Error)
			wg.Done()
			continue
		}

		for _, foundUrl := range result.foundUrls {
			if !visitedUrls[foundUrl] {
				visitedUrls[foundUrl] = true
				wg.Add(1)
				go func(url string) {
					jobs <- url
				}(foundUrl)
			}
		}
		wg.Done()
	}

	fmt.Printf("Handled %d URLs\n", totalUrlsVisited)
	fmt.Println("======Ending web crawler======")
}

// crawl fetches the content of the given URL and extracts links from it.
func crawl(url string, client *http.Client) crawlResult {
	var result crawlResult

	resp, err := client.Get(url)
	if err != nil {
		result = crawlResult{sourceUrl: url, Error: fmt.Sprintf("Error while fetching %s: %s", url, err.Error())}
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		result = crawlResult{sourceUrl: url, Error: fmt.Sprintf("bad status code: %d", resp.StatusCode)}
		return result
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		result = crawlResult{sourceUrl: url, Error: fmt.Sprintf("Error parsing HTML from %s: %s", url, err.Error())}
		return result
	}

	var absoluteUrls []string
	foundUrls := extractLinks(doc)
	if len(foundUrls) >= 0 {
		absoluteUrls = resolveLinks(foundUrls, resp.Request.URL)
	}

	result = crawlResult{sourceUrl: url, foundUrls: absoluteUrls}
	return result
}

// extractLinks traverses the HTML document and extracts all links (href attributes of <a> tags).
func extractLinks(doc *html.Node) []string {
	var urls []string

	var recursiveExtract func(*html.Node)
	recursiveExtract = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					if attr.Val != "" {
						urls = append(urls, attr.Val)
					}
					break //because we just need to check href attribute
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			recursiveExtract(c)
		}
	}
	recursiveExtract(doc)
	if len(urls) > 0 {
		return urls
	}
	return []string{}
}

// resolveLinks converts relative links to absolute links based on the base URL.
func resolveLinks(relativeLinks []string, baseURL *url.URL) []string {
	var absoluteLinks []string

	for _, link := range relativeLinks {
		if link == "" {
			continue
		}
		parsedLink, err := url.Parse(link)
		if err != nil {
			continue
		}

		var absoluteLink *url.URL
		if !parsedLink.IsAbs() {
			absoluteLink = baseURL.ResolveReference(parsedLink)
		}

		// Only add links with http or https scheme to be safe
		if absoluteLink.Scheme == "http" || absoluteLink.Scheme == "https" {
			absoluteLinks = append(absoluteLinks, absoluteLink.String())
		}
	}
	return absoluteLinks
}
