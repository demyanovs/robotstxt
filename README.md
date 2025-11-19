[![](https://github.com/demyanovs/robotstxt/actions/workflows/go.yml/badge.svg)](https://github.com/demyanovs/robotstxt/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/demyanovs/robotstxt.svg)](https://pkg.go.dev/github.com/demyanovs/robotstxt)

# RobotsTXT
A compliant robots.txt parser written in Go that follows the [robots.txt specification](https://developers.google.com/search/docs/crawling-indexing/robots/robots_txt). The parser reads and interprets robots.txt files into a convenient Go struct.

## Features
* Parse robots.txt from HTTP responses, strings, or bytes
* Extract and organize rules for different user-agents
* Check if URLs are allowed or disallowed for specific user-agents
* Crawl-delay support
* Sitemap extraction
* Most-specific rule matching according to specification

## Installation
```
go get github.com/demyanovs/robotstxt
```

## Usage

### Parsing robots.txt

**From an HTTP response:**

```go
resp, err := http.Get("https://www.example.com/robots.txt")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

robots, err := robotstxt.FromResponse(resp)
if err != nil {
    log.Fatal(err)
}
```

**From a string:**

```go
robotsTxt := `User-agent: *
Disallow: /admin
Allow: /admin/public
Crawl-delay: 5

Sitemap: https://example.com/sitemap.xml`

robots, err := robotstxt.FromString(robotsTxt)
if err != nil {
    log.Fatal(err)
}
```

**From bytes:**

```go
data := []byte("User-agent: *\nDisallow: /private\n")
robots, err := robotstxt.FromBytes(data)
if err != nil {
    log.Fatal(err)
}
```

### Checking URL Access

The `IsAllowed` method checks if a URL is allowed for a specific user-agent:

```go
// Case-insensitive user-agent matching
allowed := robots.IsAllowed("Googlebot", "/search/results")
fmt.Printf("Allowed: %v\n", allowed)

// Works with any case
allowed = robots.IsAllowed("googlebot", "/search/results") // same result
allowed = robots.IsAllowed("GOOGLEBOT", "/search/results") // same result
```

**Important:** The parser applies the **most specific (longest matching)** rule according to the robots.txt specification:

```go
// Given these rules:
// Disallow: /search
// Allow: /search/public

robots.IsAllowed("*", "/search")        // false - matches /search
robots.IsAllowed("*", "/search/admin")  // false - matches /search
robots.IsAllowed("*", "/search/public") // true  - matches /search/public (more specific)
```

### Getting User-Agent Rules

Retrieve all rules for a specific user-agent:

```go
userAgent, err := robots.UserAgent("Googlebot")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User-Agent: %s\n", userAgent.Name)
fmt.Printf("Crawl-Delay: %v\n", userAgent.CrawlDelay)

for _, rule := range userAgent.Rules {
    action := "Disallow"
    if rule.Allow {
        action = "Allow"
    }
    fmt.Printf("%s: %s\n", action, rule.Path)
}
```

### Getting Crawl Delay

```go
delay, err := robots.CrawlDelay("Googlebot")
if err != nil {
    log.Fatal(err)
}

if delay != nil {
    fmt.Printf("Crawl delay: %d seconds\n", *delay)
} else {
    fmt.Println("No crawl delay specified")
}
```

### Accessing Sitemaps

```go
robots, _ := robotstxt.FromString(robotsTxt)

for _, sitemap := range robots.Sitemaps {
    fmt.Printf("Sitemap: %s\n", sitemap)
}
```

## Advanced Examples

### Handling Multiple User-Agents

```go
robotsTxt := `User-agent: Googlebot
Crawl-delay: 2
Disallow: /private

User-agent: *
Crawl-delay: 5
Disallow: /admin`

robots, _ := robotstxt.FromString(robotsTxt)

// Googlebot uses its specific rules
fmt.Println(robots.IsAllowed("Googlebot", "/private")) // false
fmt.Println(robots.IsAllowed("Googlebot", "/admin"))   // true (no rule for /admin)

// Other bots use wildcard (*) rules
fmt.Println(robots.IsAllowed("OtherBot", "/admin"))    // false
fmt.Println(robots.IsAllowed("OtherBot", "/private"))  // true (no rule for /private)
```

### Complete Example

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/demyanovs/robotstxt"
)

func main() {
    robotsTxt := `# Example robots.txt
User-agent: Googlebot
Crawl-delay: 2
Disallow: /search
Allow: /search/about

User-agent: *
Disallow: /admin
Disallow: /private

Sitemap: https://example.com/sitemap.xml
Sitemap: https://example.com/sitemap2.xml`

    robots, err := robotstxt.FromString(robotsTxt)
    if err != nil {
        log.Fatal(err)
    }

    // Check various URLs
    urls := []string{"/", "/about", "/admin", "/search", "/search/about"}
    
    for _, url := range urls {
        allowed := robots.IsAllowed("Googlebot", url)
        fmt.Printf("Googlebot - %s: %v\n", url, allowed)
    }

    // Get crawl delay
    delay, _ := robots.CrawlDelay("googlebot")
    if delay != nil {
        fmt.Printf("\nCrawl delay for Googlebot: %d seconds\n", *delay)
    }

    // List all sitemaps
    fmt.Println("\nSitemaps:")
    for _, sitemap := range robots.Sitemaps {
        fmt.Printf("  - %s\n", sitemap)
    }
}
```

## Running Tests

```bash
go test -v
go test -cover
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](LICENSE.md)