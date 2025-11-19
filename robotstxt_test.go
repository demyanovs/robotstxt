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
		"googlebot": {
			Name:       "googlebot",
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
		"otherbot": {
			Name: "otherbot",
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

func TestUserAgentRules_UnknownUserAgent_Error(t *testing.T) {
	_, err := robotsDataResult.UserAgent("unknownUserAgent")

	require.ErrorIs(t, ErrorNoSuchUserAgent, err)
}

func TestUserAgentRules_WildcardUserAgent_Success(t *testing.T) {
	userAgent, err := robotsDataResult.UserAgent("*")

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

func TestUserAgentRules_SpambotUserAgent_Success(t *testing.T) {
	userAgent, err := robotsDataResult.UserAgent("Spambot")

	require.NoError(t, err)
	require.Equal(t, &UserAgent{
		Name: "spambot",
		Rules: []Rule{
			{
				Allow: false,
				Path:  "/",
			},
		},
	}, userAgent)
}

func TestCrawlDelay_UnknownUserAgent_Error(t *testing.T) {
	_, err := robotsDataResult.CrawlDelay("unknownUserAgent")

	require.ErrorIs(t, ErrorNoSuchUserAgent, err)
}

func TestCrawlDelay_WildcardUserAgent_Success(t *testing.T) {
	crawlDelay, err := robotsDataResult.CrawlDelay("*")

	require.NoError(t, err)
	require.Equal(t, 7, *crawlDelay)
}

func TestCrawlDelay_Nil_Success(t *testing.T) {
	crawlDelay, err := robotsDataResult.CrawlDelay("Spambot")

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

func TestIsAllowed_MostSpecificRule(t *testing.T) {
	robots := &RobotsData{
		UserAgents: map[string]UserAgent{
			"*": {
				Rules: []Rule{
					{Path: "/search", Allow: false},
					{Path: "/search/about", Allow: true},
					{Path: "/search/about/team", Allow: false},
				},
			},
		},
	}

	require.False(t, robots.IsAllowed("*", "/search"))
	require.False(t, robots.IsAllowed("*", "/search/admin"))
	require.True(t, robots.IsAllowed("*", "/search/about"))
	require.True(t, robots.IsAllowed("*", "/search/about/company"))
	require.False(t, robots.IsAllowed("*", "/search/about/team"))
	require.False(t, robots.IsAllowed("*", "/search/about/team/leads"))
}

func TestIsAllowed_CaseInsensitiveUserAgent(t *testing.T) {
	robots := &RobotsData{
		UserAgents: map[string]UserAgent{
			"googlebot": {
				Rules: []Rule{{Path: "/admin", Allow: false}},
			},
		},
	}

	require.False(t, robots.IsAllowed("googlebot", "/admin"))
	require.False(t, robots.IsAllowed("Googlebot", "/admin"))
	require.False(t, robots.IsAllowed("GOOGLEBOT", "/admin"))
	require.False(t, robots.IsAllowed("GoogleBot", "/admin"))
}

func TestIsAllowed_URLNormalization(t *testing.T) {
	robots := &RobotsData{
		UserAgents: map[string]UserAgent{
			"*": {
				Rules: []Rule{{Path: "/admin", Allow: false}},
			},
		},
	}

	require.False(t, robots.IsAllowed("*", "admin"))
	require.False(t, robots.IsAllowed("*", "/admin"))
	require.False(t, robots.IsAllowed("*", "admin/users"))
}

func TestFromString_EmptyDisallow(t *testing.T) {
	text := "User-agent: *\nDisallow:\nDisallow: /admin"
	robots, err := FromString(text)

	require.NoError(t, err)
	require.Len(t, robots.UserAgents["*"].Rules, 1)
	require.Equal(t, "/admin", robots.UserAgents["*"].Rules[0].Path)
	require.False(t, robots.UserAgents["*"].Rules[0].Allow)
}

func TestFromString_InvalidCrawlDelay(t *testing.T) {
	text := "User-agent: *\nCrawl-delay: invalid\nDisallow: /admin"
	robots, err := FromString(text)

	require.NoError(t, err)
	require.Nil(t, robots.UserAgents["*"].CrawlDelay)
}

func TestFromString_Comments(t *testing.T) {
	text := `# Full line comment
User-agent: *
Disallow: /admin # inline comment
Allow: /public
# Another comment
Disallow: /private`

	robots, err := FromString(text)

	require.NoError(t, err)
	require.Len(t, robots.UserAgents["*"].Rules, 3)
	require.Equal(t, "/admin", robots.UserAgents["*"].Rules[0].Path)
	require.Equal(t, "/public", robots.UserAgents["*"].Rules[1].Path)
	require.Equal(t, "/private", robots.UserAgents["*"].Rules[2].Path)
}

func TestFromString_EmptyFile(t *testing.T) {
	robots, err := FromString("")

	require.NoError(t, err)
	require.Empty(t, robots.UserAgents)
	require.Empty(t, robots.Sitemaps)
}

func TestFromString_OnlyComments(t *testing.T) {
	text := "# Comment 1\n# Comment 2\n# Comment 3"
	robots, err := FromString(text)

	require.NoError(t, err)
	require.Empty(t, robots.UserAgents)
}

func TestFromString_MultipleSitemaps(t *testing.T) {
	text := `User-agent: *
Disallow: /admin
Sitemap: https://example.com/sitemap1.xml
Sitemap: https://example.com/sitemap2.xml
Sitemap: https://example.com/sitemap3.xml`

	robots, err := FromString(text)

	require.NoError(t, err)
	require.Len(t, robots.Sitemaps, 3)
	require.Equal(t, "https://example.com/sitemap1.xml", robots.Sitemaps[0])
	require.Equal(t, "https://example.com/sitemap2.xml", robots.Sitemaps[1])
	require.Equal(t, "https://example.com/sitemap3.xml", robots.Sitemaps[2])
}

func TestUserAgent_CaseInsensitive(t *testing.T) {
	robots := &RobotsData{
		UserAgents: map[string]UserAgent{
			"googlebot": {
				Name:  "googlebot",
				Rules: []Rule{{Path: "/admin", Allow: false}},
			},
		},
	}

	ua1, err1 := robots.UserAgent("googlebot")
	require.NoError(t, err1)
	require.Equal(t, "googlebot", ua1.Name)

	ua2, err2 := robots.UserAgent("Googlebot")
	require.NoError(t, err2)
	require.Equal(t, "googlebot", ua2.Name)

	ua3, err3 := robots.UserAgent("GOOGLEBOT")
	require.NoError(t, err3)
	require.Equal(t, "googlebot", ua3.Name)
}

func TestCrawlDelay_CaseInsensitive(t *testing.T) {
	delay := 5
	robots := &RobotsData{
		UserAgents: map[string]UserAgent{
			"googlebot": {
				Name:       "googlebot",
				CrawlDelay: &delay,
			},
		},
	}

	cd1, err1 := robots.CrawlDelay("googlebot")
	require.NoError(t, err1)
	require.Equal(t, 5, *cd1)

	cd2, err2 := robots.CrawlDelay("Googlebot")
	require.NoError(t, err2)
	require.Equal(t, 5, *cd2)

	cd3, err3 := robots.CrawlDelay("GOOGLEBOT")
	require.NoError(t, err3)
	require.Equal(t, 5, *cd3)
}

func TestIsAllowed_WildcardFallback(t *testing.T) {
	robots := &RobotsData{
		UserAgents: map[string]UserAgent{
			"*": {
				Rules: []Rule{{Path: "/admin", Allow: false}},
			},
			"googlebot": {
				Rules: []Rule{{Path: "/search", Allow: false}},
			},
		},
	}

	require.False(t, robots.IsAllowed("googlebot", "/search"))
	require.True(t, robots.IsAllowed("googlebot", "/admin"))

	require.False(t, robots.IsAllowed("unknownbot", "/admin"))
	require.True(t, robots.IsAllowed("unknownbot", "/search"))
}

func TestIsAllowed_NoRulesAllowAll(t *testing.T) {
	robots := &RobotsData{
		UserAgents: map[string]UserAgent{
			"googlebot": {
				Rules: []Rule{},
			},
		},
	}

	require.True(t, robots.IsAllowed("googlebot", "/admin"))
	require.True(t, robots.IsAllowed("googlebot", "/search"))
	require.True(t, robots.IsAllowed("unknownbot", "/anything"))
}

func TestFromString_MultipleUserAgentsOneBlock(t *testing.T) {
	text := `User-agent: googlebot
User-agent: bingbot
Disallow: /admin
Allow: /public`

	robots, err := FromString(text)

	require.NoError(t, err)

	require.Len(t, robots.UserAgents["googlebot"].Rules, 2)
	require.Len(t, robots.UserAgents["bingbot"].Rules, 2)

	require.Equal(t, "/admin", robots.UserAgents["googlebot"].Rules[0].Path)
	require.Equal(t, "/admin", robots.UserAgents["bingbot"].Rules[0].Path)
}

func TestIsAllowed_RootPath(t *testing.T) {
	robots := &RobotsData{
		UserAgents: map[string]UserAgent{
			"spambot": {
				Rules: []Rule{{Path: "/", Allow: false}},
			},
		},
	}

	require.False(t, robots.IsAllowed("spambot", "/"))
	require.False(t, robots.IsAllowed("spambot", "/admin"))
	require.False(t, robots.IsAllowed("spambot", "/anything/else"))
}
