package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/greboid/irc-bot/v5/plugins"
	"github.com/kouhin/envflag"
	"github.com/mmcdole/gofeed"
)

var (
	rpcHost  = flag.String("host", "localhost:8001", "Host and port to connect to RPC server on")
	rpcToken = flag.String("token", "isedjfiuwserfuesd", "Token to use to authenticate RPC requests")
)

var (
	channel = flag.String("channel", "#news", "Channel to send messages to")

	feeds = map[string]string{
		"HN":         "https://news.ycombinator.com/rss",
		"Lobsters":   "https://lobste.rs/rss",
		"Tilde.news": "https://tilde.news/rss",
		"BBC":        "https://feeds.bbci.co.uk/news/rss.xml",
	}

	seen      = map[string]bool{}
	lastCheck = map[string]time.Time{}
)

func main() {
	if err := envflag.Parse(); err != nil {
		log.Fatalf("Unable to parse config: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client, err := plugins.NewHelper(*rpcHost, *rpcToken)
	if err != nil {
		log.Fatalf("Failed to connec to RPC: %v\n", err)
	}

	defer cancel()

	timer := time.NewTicker(time.Minute)

	for {
		select {
		case _ = <-timer.C:
			checkForNewItems(ctx, client, selectSite())
		}
	}
}

func checkForNewItems(ctx context.Context, helper *plugins.PluginHelper, site string) {
	log.Printf("Updating site %s...\n", site)

	_, hasPrevious := lastCheck[site]

	parser := gofeed.NewParser()
	parser.Client = &http.Client{
		Transport: transport{},
	}

	i, err := items(parser, site)
	if err != nil {
		log.Printf("Failed to update %s: %v\n", site, err)
		lastCheck[site] = time.Now().Add(time.Minute * 10)
		return
	}

	for _, item := range i {
		if !hasPrevious {
			seen[item.Link] = true
		} else if !seen[item.Link] {
			filtered := filter(item.Link)
			if filtered != "" {
				announce(ctx, helper, site, item.Title, filtered)
			}
			seen[item.Link] = true
		}
	}

	lastCheck[site] = time.Now()
}

func filter(link string) string {
	if strings.HasPrefix(link, "https://www.bbc.co.uk/sport") {
		return ""
	}
	return strings.TrimSuffix(link, "?at_medium=RSS&at_campaign=KARANGA")
}

func announce(ctx context.Context, helper *plugins.PluginHelper, source, title, link string) {
	if err := helper.SendRelayMessageWithContext(ctx, *channel, fmt.Sprintf("news/%s", source), fmt.Sprintf("%s - %s", title, link)); err != nil {
		log.Panicf("Unable to send message: %v\n", err)
	}
}

func items(parser *gofeed.Parser, site string) ([]*gofeed.Item, error) {
	feed, err := parser.ParseURL(feeds[site])
	if err != nil {
		return nil, err
	}

	if len(feed.Items) > 50 {
		return feed.Items[:50], nil
	}
	return feed.Items, nil
}

func selectSite() (site string) {
	bestTime := time.Now()
	for s := range feeds {
		t, ok := lastCheck[s]
		if !ok || t.Before(bestTime) {
			bestTime = t
			site = s
		}
	}
	return
}

// transport is a HTTP transport that updates the user agent sent in requests.
// (Reddit severely rate limits default user agents.)
type transport struct {
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "FeedsToIrc/1.0 (by Chris Smith; https://chameth.com)")
	return http.DefaultTransport.RoundTrip(req)
}
