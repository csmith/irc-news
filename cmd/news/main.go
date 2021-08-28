package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/csmith/ircplugins"
	"github.com/kouhin/envflag"
	"github.com/mmcdole/gofeed"
)

var (
	channel = flag.String("channel", "#news", "Channel to send messages to")

	feeds = map[string]string{
		"HN":         "https://news.ycombinator.com/rss",
		"Lobsters":   "https://lobste.rs/rss",
		"Tilde.news": "https://tilde.news/rss",
		"laarc":      "https://www.laarc.io/rss",
		"TrueReddit": "https://www.reddit.com/r/TrueReddit/.rss",
		"Pinboard":   "https://feeds.pinboard.in/rss/popular/",
	}

	seen      = map[string]bool{}
	lastCheck = map[string]time.Time{}
)

func main() {
	if err := envflag.Parse(); err != nil {
		log.Fatalf("Unable to parse config: %v\n", err)
	}

	client, err := ircplugins.NewClient()
	if err != nil {
		log.Fatalf("Failed to connec to RPC: %v\n", err)
	}

	defer func() {
		_ = client.Close()
	}()

	timer := time.NewTicker(time.Minute)

	for {
		select {
		case _ = <-timer.C:
			checkForNewItems(client, selectSite())
		}
	}
}

func checkForNewItems(client *ircplugins.RpcClient, site string) {
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
			log.Printf("New item: %s (%s)\n", item.Title, item.Link)
			if err := client.Send(*channel, fmt.Sprintf("[%s] %s - %s", site, item.Title, item.Link)); err != nil {
				log.Panicf("Unable to send message: %v\n", err)
			}
			seen[item.Link] = true
		}
	}

	lastCheck[site] = time.Now()
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
