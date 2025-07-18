# A Concurrent Web Crawler

This is a simple but powerful web crawler written in Go. It's designed to concurrently fetch and parse web pages to discover new URLs, starting from a predefined list of seed websites. It leverages Go's powerful concurrency primitives (goroutines and channels) to perform crawling tasks in parallel, making it fast and efficient.

---

## 🚀 Features

- **Concurrent Crawling**  
  Utilizes a worker pool of goroutines to fetch multiple URLs simultaneously.

- **Channel-Based Work Distribution**  
  Uses buffered channels to manage the queue of URLs to be crawled and to collect the results.

- **Link Extraction**  
  Parses the HTML of each page to find all `<a>` tags and extract their `href` attributes.

- **Relative to Absolute URL Resolution**  
  Intelligently converts relative links (e.g., `/about`) into absolute URLs (e.g., `https://example.com/about`).

- **Duplicate Prevention**  
  Keeps track of visited URLs to avoid redundant work and infinite loops.

---

## 🧠 How It Works

The crawler operates on a **producer-consumer model** using a worker pool.

### 1. Initialization
The crawler starts with a seed list of URLs. A `visitedUrls` map is created to track every URL that has been queued for crawling.

### 2. Job Distribution
A `jobs` channel is populated with the initial seed URLs.

### 3. Worker Pool
A fixed number of worker goroutines (e.g., 50) are launched. Each worker listens for a URL on the `jobs` channel.

### 4. Crawling
When a worker receives a URL, it:

- Sends an HTTP GET request to the URL.
- Parses the HTML response body.
- Extracts all unique links.
- Resolves them into absolute URLs.

### 5. Result Collection
The results (the source URL and any new URLs found) are sent to a `results` channel.

### 6. Processing New Links
The main goroutine reads from the `results` channel. For each newly discovered URL, it:

- Checks if it has already been visited.
- If not, marks it as visited and adds it to the `jobs` channel for a worker to pick up.

### 7. Shutdown
The process continues until there are no more jobs in the queue and all workers are idle. A `sync.WaitGroup` is used to track the completion of all crawling tasks before the program exits gracefully.
