package robotstxt

import (
	"bufio"
	"errors"
	"net/http"
	"strconv"
	"strings"

	_ "golang.org/x/lint"
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
	scanner := bufio.NewScanner(resp.Body)

	err := r.parseRules(scanner)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// FromString creates a new instance of RobotsData from string.
func FromString(text string) (*RobotsData, error) {
	r := RobotsData{}

	re := strings.NewReader(text)
	scanner := bufio.NewScanner(re)

	err := r.parseRules(scanner)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// FromBytes creates a new instance of RobotsData from file.
func FromBytes(bytes []byte) (*RobotsData, error) {
	return FromString(string(bytes))
}

// GetUserAgent returns rules for particular UserAgent.
func (rb *RobotsData) GetUserAgent(userAgent string) (*UserAgent, error) {
	ua, ok := rb.UserAgents[userAgent]
	if !ok {
		return nil, ErrorNoSuchUserAgent
	}
	return &ua, nil
}

// GetCrawlDelay returns crawl delay for particular UserAgent.
func (rb *RobotsData) GetCrawlDelay(userAgent string) (*int, error) {
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
func (rb *RobotsData) IsAllowed(userAgent string, URL string) bool {
	applicableRules := rb.getApplicableRules(userAgent)

	// Check the rules from most specific to the least specific
	for _, rule := range applicableRules {
		if strings.HasPrefix(URL, rule.Path) {
			return rule.Allow
		}
	}

	return true
}

// getApplicableRules retrieves rules for a specific user-agent.
func (rb *RobotsData) getApplicableRules(userAgent string) []Rule {
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

func (rb *RobotsData) parseRules(scanner *bufio.Scanner) error {
	rb.UserAgents = make(map[string]UserAgent)

	var currentUserAgent string
	rules := make(map[string][]Rule)
	delays := make(map[string]*int)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		rule, val := parseLine(line)
		if rule == userAgentRuleKey {
			currentUserAgent = val
			if _, exists := rules[currentUserAgent]; !exists {
				rules[currentUserAgent] = []Rule{}
			}

			delays[currentUserAgent] = nil
		}

		if rule == allowRuleKey {
			if currentUserAgent == "" {
				return ErrorMissingUserAgent
			}

			rules[currentUserAgent] = append(rules[currentUserAgent], Rule{
				Allow: true,
				Path:  val,
			})
		}

		if rule == disallowRuleKey {
			if currentUserAgent == "" {
				return ErrorMissingUserAgent
			}

			rules[currentUserAgent] = append(rules[currentUserAgent], Rule{
				Allow: false,
				Path:  val,
			})
		}

		if rule == crawlDelayRuleKey {
			if currentUserAgent == "" {
				return ErrorMissingUserAgent
			}

			res, _ := strconv.Atoi(val)
			delays[currentUserAgent] = &res
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
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return unknownRuleKey, ""
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

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
