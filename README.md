[![](https://github.com/demyanovs/robotstxt/actions/workflows/go.yml/badge.svg)](https://github.com/demyanovs/robotstxt/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/demyanovs/robotstxt.svg)](https://pkg.go.dev/github.com/demyanovs/robotstxt)

# RobotsTXT
RobotsTXT is a robots.txt parser written in Go. The parser reads and interprets robots.txt files into a struct. 

## Features
* Parse robots.txt files
* Extract rules for different user-agents
* Check if a specific URL is allowed or disallowed for a given user-agent
* Crawl-delay support

## Installation
```
go get github.com/demyanovs/robotstxt
```

## Usage
Here are examples of how to parse a robots.txt file.

From a response (from URL as an example):
```
resp, err := http.Get("https://www.example.com/robots.txt")
if err != nil {
    log.Fatal(err)
}

defer resp.Body.Close()

robots, err := robotstxt.FromResponse(resp)
```

From a string:
```
robots, err = robotstxt.FromString("User-agent: *\nDisallow: /search\nAllow: /search/about\nCrawl-delay: 5")
```

From a slice of bytes: 
```
robots, err := robotstxt.FromBytes([]byte("User-agent: *\nDisallow: /search\nAllow: /search/about\nCrawl-delay: 5"))
```

To get a specific user-agent:
```
userAgent, err := robots.UserAgent("*")
```

To check if a URL is allowed for a specific user-agent:
```
isAllowed := robots.IsAllowed("*", "/search")
```

To get the crawl delay for a specific user-agent:
```
crawlDealy, err := robots.CrawlDelay("*")
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](LICENSE.md)