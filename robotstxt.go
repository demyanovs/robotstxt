// Package robotstxt reads and interprets robots.txt files into a struct.
package robotstxt

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// RobotsData represents a robots.txt file.
type RobotsData struct {
	Sitemaps   []string
	UserAgents map[string]UserAgent
}

// Rule represents a rule in robots.txt file.
type Rule struct {
	Path  string
	Allow bool
}

// UserAgent represents rules for a user agent in robots.txt file.
type UserAgent struct {
	Name       string
	CrawlDelay *int
	Rules      []Rule
}

type robotsRuleKey string

const (
	userAgentRuleKey  robotsRuleKey = "user-agent"
	allowRuleKey      robotsRuleKey = "allow"
	disallowRuleKey   robotsRuleKey = "disallow"
	crawlDelayRuleKey robotsRuleKey = "crawl-delay"
	sitemapRuleKey    robotsRuleKey = "sitemap"
	unknownRuleKey    robotsRuleKey = "unknown"
)

var rulesKeysSlice = []robotsRuleKey{
	userAgentRuleKey,
	allowRuleKey,
	disallowRuleKey,
	crawlDelayRuleKey,
	sitemapRuleKey,
}

var (
	// ErrorNoSuchUserAgent is returned when there is no such user agent in UserAgents.
	ErrorNoSuchUserAgent = errors.New("no such user agent")
	// ErrorMissingUserAgent is returned when there is no user agent in robots.txt file.
	ErrorMissingUserAgent = errors.New("missing user agent")
)

// FromResponse creates a new instance of RobotsData from an HTTP response.
func FromResponse(resp *http.Response) (*RobotsData, error) {
	r := RobotsData{}

	err := r.parseRules(resp.Body)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// FromString creates a new instance of RobotsData from string.
func FromString(text string) (*RobotsData, error) {
	r := RobotsData{}
	reader := strings.NewReader(text)

	err := r.parseRules(reader)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// FromBytes creates a new instance of RobotsData from file.
func FromBytes(bytes []byte) (*RobotsData, error) {
	return FromString(string(bytes))
}

// UserAgent returns rules for particular UserAgent.
func (rb *RobotsData) UserAgent(userAgent string) (*UserAgent, error) {
	userAgent = strings.ToLower(userAgent)
	ua, ok := rb.UserAgents[userAgent]
	if !ok {
		return nil, ErrorNoSuchUserAgent
	}
	return &ua, nil
}

// CrawlDelay returns crawl delay for particular UserAgent.
func (rb *RobotsData) CrawlDelay(userAgent string) (*int, error) {
	userAgent = strings.ToLower(userAgent)
	ua, ok := rb.UserAgents[userAgent]
	if !ok {
		return nil, ErrorNoSuchUserAgent
	}

	if ua.CrawlDelay == nil {
		return nil, nil
	}

	return ua.CrawlDelay, nil
}

// IsAllowed checks if the URL is allowed for the user agent.
// It applies the most specific (longest matching) rule according to robots.txt specification.
func (rb *RobotsData) IsAllowed(userAgent string, URL string) bool {
	if !strings.HasPrefix(URL, "/") {
		URL = "/" + URL
	}

	applicableRules := rb.applicableRules(userAgent)

	// Find the most specific (longest) matching rule
	var matchedRuleIndex = -1
	maxMatchLength := 0

	for i, rule := range applicableRules {
		if strings.HasPrefix(URL, rule.Path) && len(rule.Path) > maxMatchLength {
			matchedRuleIndex = i
			maxMatchLength = len(rule.Path)
		}
	}

	// If a matching rule is found, return its permission
	if matchedRuleIndex != -1 {
		return applicableRules[matchedRuleIndex].Allow
	}

	// Default: allow access if no rules match
	return true
}

// applicableRules retrieves rules for a specific user-agent.
// User-agent matching is case-insensitive according to robots.txt specification.
func (rb *RobotsData) applicableRules(userAgent string) []Rule {
	userAgent = strings.ToLower(userAgent)

	// Exact match
	if u, exists := rb.UserAgents[userAgent]; exists {
		return u.Rules
	}

	// Wildcard user-agent
	if u, exists := rb.UserAgents["*"]; exists {
		return u.Rules
	}

	return []Rule{}
}

func (rb *RobotsData) parseRules(reader io.Reader) error {
	rb.UserAgents = make(map[string]UserAgent)

	var currentUserAgents []string
	rules := make(map[string][]Rule)
	delays := make(map[string]*int)
	blockStarted := false // Flag indicating that we are processing a rules block

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		rule, val := parseLine(line)

		if rule == userAgentRuleKey {
			ua := strings.ToLower(val)

			if blockStarted {
				currentUserAgents = []string{ua}
				blockStarted = false
			} else {
				currentUserAgents = append(currentUserAgents, ua)
			}

			if _, exists := rules[ua]; !exists {
				rules[ua] = []Rule{}
			}
			if _, exists := delays[ua]; !exists {
				delays[ua] = nil
			}
		}

		if rule == allowRuleKey {
			if len(currentUserAgents) == 0 {
				return ErrorMissingUserAgent
			}

			blockStarted = true

			for _, ua := range currentUserAgents {
				rules[ua] = append(rules[ua], Rule{
					Allow: true,
					Path:  val,
				})
			}
		}

		if rule == disallowRuleKey {
			// Empty disallow means "allow all", so skip it
			if val == "" {
				continue
			}

			if len(currentUserAgents) == 0 {
				return ErrorMissingUserAgent
			}

			blockStarted = true

			for _, ua := range currentUserAgents {
				rules[ua] = append(rules[ua], Rule{
					Allow: false,
					Path:  val,
				})
			}
		}

		if rule == crawlDelayRuleKey {
			if len(currentUserAgents) == 0 {
				return ErrorMissingUserAgent
			}

			blockStarted = true

			res, err := strconv.Atoi(val)
			if err != nil {
				// Skip invalid crawl-delay values
				continue
			}

			for _, ua := range currentUserAgents {
				delays[ua] = &res
			}
		}

		if rule == sitemapRuleKey {
			rb.Sitemaps = append(rb.Sitemaps, val)
		}
	}

	for u, rule := range rules {
		rb.UserAgents[u] = UserAgent{
			Name:       u,
			Rules:      rule,
			CrawlDelay: delays[u],
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func parseLine(line string) (robotsRuleKey, string) {
	// Ignore comments and empty strings
	if line == "" || strings.HasPrefix(line, "#") {
		return unknownRuleKey, ""
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return unknownRuleKey, ""
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Remove inline comments from value
	if idx := strings.Index(value, "#"); idx != -1 {
		value = strings.TrimSpace(value[:idx])
	}

	for _, ruleKey := range rulesKeysSlice {
		if strings.ToLower(key) == ruleKey.toString() {
			return ruleKey, value
		}
	}

	return unknownRuleKey, ""
}

func (rlk robotsRuleKey) toString() string {
	return string(rlk)
}
