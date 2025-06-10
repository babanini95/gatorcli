package commands

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/babanini95/gatorcli/internal/database"
	"github.com/google/uuid"
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

func handlerAgg(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("invalid argument")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenReqs.String())
	ticker := time.NewTicker(timeBetweenReqs)

	for ; ; <-ticker.C {
		scrapeFeeds(s, user)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		fmt.Println("Need more arguments!")
		os.Exit(1)
	}

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: sqlCurrentTime(),
		UpdatedAt: sqlCurrentTime(),
		Name: sql.NullString{
			String: cmd.arguments[0],
			Valid:  true,
		},
		Url: sql.NullString{
			String: cmd.arguments[1],
			Valid:  true,
		},
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
	}

	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		fmt.Printf("Failed to create feed: %v\n", err)
		os.Exit(1)
	}

	createFeedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: sqlCurrentTime(),
		UpdatedAt: sqlCurrentTime(),
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		FeedID: uuid.NullUUID{
			UUID:  feed.ID,
			Valid: true,
		},
	}
	_, err = s.db.CreateFeedFollow(context.Background(), createFeedFollowParams)
	if err != nil {
		fmt.Printf("Can not create feed follow: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%v\n", feed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.ListFeeds(context.Background())
	if err != nil {
		fmt.Printf("Failed to list all feeds: %v", err)
		os.Exit(1)
	}

	fmt.Println("FEED LIST")
	for i, feed := range feeds {
		fmt.Printf(
			`%v. - Feed Name : %s
   - URL	   : %s
   - User Name : %s
`,
			(i + 1), feed.Name.String, feed.Url.String, feed.UserName.String,
		)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		fmt.Println("invalid argument")
		os.Exit(1)
	}

	urlNullString := sql.NullString{
		String: cmd.arguments[0],
		Valid:  true,
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), urlNullString)
	if err != nil {
		os.Exit(1)
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: sqlCurrentTime(),
		UpdatedAt: sqlCurrentTime(),
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		FeedID: uuid.NullUUID{
			UUID:  feed.ID,
			Valid: true,
		},
	}

	row, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		fmt.Printf("Can not create feed follow: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Feed name: %s\nUser: %s\n", row.FeedName.String, row.UserName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feedsFollow, err := s.db.GetFeedFollowsForUser(
		context.Background(),
		uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
	)

	if err != nil {
		os.Exit(1)
	}

	for _, feed := range feedsFollow {
		fmt.Printf("%s\n", feed.FeedName.String)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		fmt.Println("Invalid arguments")
		os.Exit(1)
	}

	params := database.DeleteFeedFollowsByUrlParams{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Url: sql.NullString{
			String: cmd.arguments[0],
			Valid:  true,
		},
	}
	err := s.db.DeleteFeedFollowsByUrl(context.Background(), params)
	if err != nil {
		fmt.Printf("Failed to delete feeds follow:\n%v\n", err)
		os.Exit(1)
	}

	return nil
}

func scrapeFeeds(s *state, user database.User) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background(), uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	})
	if err != nil {
		return err
	}

	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sqlCurrentTime(),
		ID:            feed.ID,
	})
	if err != nil {
		return err
	}

	rss, err := fetchFeed(context.Background(), feed.Url.String)
	if err != nil {
		return err
	}
	fmt.Printf("\nFeed from %s\n", html.UnescapeString(rss.Channel.Title))
	for _, item := range rss.Channel.Item {
		fmt.Printf("- %s\n", html.UnescapeString(item.Title))
	}
	fmt.Println("----------------------------------------")

	return nil
}
