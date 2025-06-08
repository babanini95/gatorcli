package commands

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (rss *RSSFeed, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return &RSSFeed{}, err
	}
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error getting response: %v\n", err)
		return &RSSFeed{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response's body: %v\n", err)
		return &RSSFeed{}, err
	}

	err = xml.Unmarshal(data, &rss)
	if err != nil {
		fmt.Printf("Error parsing xml: %v\n", err)
		return &RSSFeed{}, err
	}

	return rss, nil
}

func handlerAgg(s *state, cmd command) error {
	rss, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

	fmt.Printf(
		"%v\n%v\n%v\n",
		html.UnescapeString(rss.Channel.Title),
		html.UnescapeString(rss.Channel.Description),
		rss.Channel.Link,
	)

	for _, item := range rss.Channel.Item {
		fmt.Printf(
			"%v\n%v\n%v\n%v\n",
			html.UnescapeString(item.Title),
			html.UnescapeString(item.Description),
			item.PubDate,
			item.Link,
		)
	}
	return nil
}
