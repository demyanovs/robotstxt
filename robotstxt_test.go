package robotstxt

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	crawlDelay2 = 2
	crawlDelay5 = 5
	crawlDelay7 = 7
)

var xxx = ""
var robotsStr = "User-agent: *\nCrawl-delay: 5\nDisallow: /search/advanced\nAllow: /search/about\nDisallow: /groups\nAllow: /news\nAllow: /blog\nDisallow: /user\n\n# Comment\nUser-agent: OtherBot\nDisallow: /maps/api/\nAllow: /maps/\nDisallow: /maps/private\n\nUser-agent: spambot\nDisallow: /\n\nSitemap: https://www.example.com/sitemap1.xml\nSitemap: https://www.example.com/sitemap2.xml"
var robotsStrNotValid = "Disallow: /search/advanced\nAllow: /search/about\nDisallow: /groups\nAllow: /news\nAllow: /blog\nDisallow: /user\n\n# Comment\nUser-agent: OtherBot\nDisallow: /maps/api/\nAllow: /maps/\nDisallow: /maps/private\n\nUser-agent: spambot\nDisallow: /\n\nSitemap: https://www.example.com/sitemap1.xml\nSitemap: https://www.example.com/sitemap2.xml"

var robotsDataResult = RobotsData{
	UserAgents: map[string]UserAgent{
		"*": {
			Name:       "*",
			CrawlDelay: &crawlDelay7,
			Rules: []Rule{
				{
					Allow: false,
					Path:  "/search/advanced",
				},
				{
					Allow: true,
					Path:  "/search",
				},
				{
					Allow: false,
					Path:  "/blog/draft",
				},
				{
					Allow: false,
					Path:  "/admin",
				},
				{
					Allow: true,
					Path:  "/blog",
				},
				{
					Allow: false,
					Path:  "/user",
				},
			},
		},
		"Googlebot": {
			Name:       "Googlebot",
			CrawlDelay: &crawlDelay2,
			Rules: []Rule{
				{
					Allow: true,
					Path:  "/search",
				},
				{
					Allow: false,
					Path:  "/images",
				},
				{
					Allow: true,
					Path:  "/register/u1",
				},
				{
					Allow: false,
					Path:  "/register",
				},
			},
		},
		"Spambot": {
			Name: "Spambot",
			Rules: []Rule{
				{
					Allow: false,
					Path:  "/",
				},
			},
		},
	},
}

var robotsDataExpected = RobotsData{
	Sitemaps: []string{
		"https://www.example.com/sitemap1.xml",
		"https://www.example.com/sitemap2.xml",
	},
	UserAgents: map[string]UserAgent{
		"*": {
			Name:       "*",
			CrawlDelay: &crawlDelay5,
			Rules: []Rule{
				{
					Allow: false,
					Path:  "/search/advanced",
				},
				{
					Allow: true,
					Path:  "/search/about",
				},
				{
					Allow: false,
					Path:  "/groups",
				},
				{
					Allow: true,
					Path:  "/news",
				},
				{
					Allow: true,
					Path:  "/blog",
				},
				{
					Allow: false,
					Path:  "/user",
				},
			},
		},
		"OtherBot": {
			Name: "OtherBot",
			Rules: []Rule{
				{
					Allow: false,
					Path:  "/maps/api/",
				},
				{
					Allow: true,
					Path:  "/maps/",
				},
				{
					Allow: false,
					Path:  "/maps/private",
				},
			},
		},
		"spambot": {
			Name: "spambot",
			Rules: []Rule{
				{
					Allow: false,
					Path:  "/",
				},
			},
		},
	},
}

func TestFromResponse_Success(t *testing.T) {
	resp := http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(robotsStr)),
	}

	robots, err := FromResponse(&resp)

	require.NoError(t, err)
	require.Equal(t, &robotsDataExpected, robots)
}

func TestFromString_MissingUserAgentError(t *testing.T) {
	_, err := FromString(robotsStrNotValid)

	require.ErrorIs(t, ErrorMissingUserAgent, err)
}

func TestFromString_Success(t *testing.T) {
	robots, err := FromString(robotsStr)

	require.NoError(t, err)
	require.Equal(t, &robotsDataExpected, robots)
}

func TestFromBytes_Success(t *testing.T) {
	robots, err := FromBytes([]byte(robotsStr))

	require.NoError(t, err)
	require.Equal(t, &robotsDataExpected, robots)
}

func TestGetUserAgentRules_UnknownUserAgent_Error(t *testing.T) {
	_, err := robotsDataResult.GetUserAgent("unknownUserAgent")

	require.ErrorIs(t, ErrorNoSuchUserAgent, err)
}

func TestGetUserAgentRules_WildcardUserAgent_Success(t *testing.T) {
	userAgent, err := robotsDataResult.GetUserAgent("*")

	require.NoError(t, err)
	require.Equal(t, &UserAgent{
		Name:       "*",
		CrawlDelay: &crawlDelay7,
		Rules: []Rule{
			{
				Allow: false,
				Path:  "/search/advanced",
			},
			{
				Allow: true,
				Path:  "/search",
			},
			{
				Allow: false,
				Path:  "/blog/draft",
			},
			{
				Allow: false,
				Path:  "/admin",
			},
			{
				Allow: true,
				Path:  "/blog",
			},
			{
				Allow: false,
				Path:  "/user",
			},
		},
	}, userAgent)
}

func TestGetUserAgentRules_SpambotUserAgent_Success(t *testing.T) {
	userAgent, err := robotsDataResult.GetUserAgent("Spambot")

	require.NoError(t, err)
	require.Equal(t, &UserAgent{
		Name: "Spambot",
		Rules: []Rule{
			{
				Allow: false,
				Path:  "/",
			},
		},
	}, userAgent)
}

func TestGetCrawlDelay_UnknownUserAgent_Error(t *testing.T) {
	_, err := robotsDataResult.GetCrawlDelay("unknownUserAgent")

	require.ErrorIs(t, ErrorNoSuchUserAgent, err)
}

func TestGetCrawlDelay_WildcardUserAgent_Success(t *testing.T) {
	crawlDelay, err := robotsDataResult.GetCrawlDelay("*")

	require.NoError(t, err)
	require.Equal(t, 7, *crawlDelay)
}

func TestGetCrawlDelay_Nil_Success(t *testing.T) {
	crawlDelay, err := robotsDataResult.GetCrawlDelay("Spambot")

	var num *int

	require.NoError(t, err)
	require.Equal(t, num, crawlDelay)
}

func TestIsAllowed(t *testing.T) {
	t.Parallel()

	type tcase struct {
		userAgent string
		url       string
		isAllow   bool
	}

	tests := []tcase{
		{"*", "/admin", false},
		{"*", "/admin/edit", false},
		{"*", "/adm", true},

		{"Googlebot", "/search", true},
		{"Googlebot", "/admin", true},
		{"Googlebot", "/register", false},
		{"Googlebot", "/register/u1", true},

		{"Spambot", "/hello", false},
		{"Spambot", "/", false},

		{"Unknown", "/test", true},
	}

	for i, test := range tests {
		isAllowed := robotsDataResult.IsAllowed(test.userAgent, test.url)
		require.Equal(t, test.isAllow, isAllowed, fmt.Sprintf("case %d, url: %s", i, test.url))
	}
}
