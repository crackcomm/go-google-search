package spider

import (
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"

	"github.com/PuerkitoBio/goquery"
	"github.com/crackcomm/crawl"
	"github.com/crackcomm/crawl/forms"
	prompt "github.com/segmentio/go-prompt"
)

// Spider - Google search results spider.
type Spider struct {
	crawl.Crawler
	Output func(*SearchResult) error
}

// SearchResult - Google search result.
type SearchResult struct {
	// Query - Search query.
	Query string `json:"query,omitempty"`
	// Page - Search results page.
	Page int `json:"page,omitempty"`
	// Results - List of URLs.
	Results []string `json:"results,omitempty"`
	// Engine - Result engine, "google".
	Engine string `json:"engine,omitempty"`
	// Source - URL source.
	Source string `json:"source,omitempty"`
}

var (
	// Google - Gather URLs from google.
	Google = "github.com/crackcomm/go-google-search/spider.Google"

	// GoogleCaptcha - Gather URLs from google.
	GoogleCaptcha = "github.com/crackcomm/go-google-search/spider.GoogleCaptcha"
)

// Register - Registers spider.
func (spider *Spider) Register() {
	spider.Crawler.Register(Google, spider.Google)
	spider.Crawler.Register(GoogleCaptcha, spider.GoogleCaptcha)
}

// Google - Crawls google.
func (spider *Spider) Google(ctx context.Context, resp *crawl.Response) (err error) {
	if resp.URL().Host == "ipv4.google.com" {
		return spider.solveCaptcha(ctx, resp)
	}

	uri := resp.URL().String()
	glog.V(2).Infof("Search: %q", uri)
	result := &SearchResult{
		Query:  getQueryValue(uri, "q"),
		Page:   getPageFromURL(uri),
		Source: uri,
		Engine: "google",
	}

	// Crawl all the results on a page
	resp.Query().Find("h3.r a").Each(func(_ int, link *goquery.Selection) {
		href, _ := link.Attr("href")
		if strings.HasPrefix(href, "/url") {
			href = getQueryValue(href, "q")
		} else if strings.HasPrefix(href, "/") {
			return
		}
		result.Results = append(result.Results, href)
	})

	// Send result to output
	if err := spider.Output(result); err != nil {
		return err
	}

	// Look for a next page URL
	var nextURL string
	resp.Query().Find(`a[href^="/search"]`).Each(func(_ int, link *goquery.Selection) {
		if !strings.HasPrefix(link.Text(), "N") {
			return
		}
		href, _ := link.Attr("href")
		if getQueryValue(href, "start") != "" {
			nextURL = href
		}
	})

	if nextURL == "" {
		glog.V(2).Info("Search for %q is done", result.Query)
		return
	}
	glog.V(2).Infof("Scheduling next page in 5s: %s", nextURL)

	// Wait for 5 seconds so we won't get blocked as fast
	<-time.After(time.Second * 5)

	// Schedule next page crawl
	spider.Crawler.Schedule(context.Background(), &crawl.Request{
		URL:       nextURL,
		Referer:   resp.URL().String(),
		Callbacks: crawl.Callbacks(Google),
	})
	return
}

// GoogleCaptcha - Crawls google.
func (spider *Spider) GoogleCaptcha(ctx context.Context, resp *crawl.Response) (err error) {
	body, err := resp.Bytes()
	if err != nil {
		return
	}
	err = ioutil.WriteFile("captcha.jpg", body, os.ModePerm)
	if err != nil {
		return
	}
	return
}

// solveCaptcha - Solves captcha (currently manual only).
func (spider *Spider) solveCaptcha(ctx context.Context, resp *crawl.Response) (err error) {
	imgSrc := crawl.Attr(resp, "src", "img")
	if imgSrc == "" {
		glog.Error("Cannot find captcha image")
		return
	}
	spider.Crawler.Schedule(context.Background(), &crawl.Request{
		URL:       imgSrc,
		Referer:   resp.URL().String(),
		Callbacks: crawl.Callbacks(GoogleCaptcha),
	})
	<-time.After(time.Second)
	text := prompt.String("Give me the captcha text")
	glog.Infof("Using captcha text: %q", text)
	err = os.Remove("captcha.jpg")
	if err != nil {
		return
	}
	form := forms.NewSelector(resp, "form")
	form.Values.Set("captcha", text)
	spider.Crawler.Schedule(context.Background(), &crawl.Request{
		URL:       form.Action,
		Query:     form.Values,
		Referer:   resp.URL().String(),
		Callbacks: crawl.Callbacks(Google),
	})
	return
}

func getPageFromURL(uri string) (page int) {
	if s := getQueryValue(uri, "start"); s != "" {
		start, _ := strconv.Atoi(s)
		page = start / 10
	}
	page++
	return
}

func getQueryValue(uri, key string) string {
	index := strings.Index(uri, "?")
	query, _ := url.ParseQuery(uri[index+1:])
	return query.Get(key)
}
