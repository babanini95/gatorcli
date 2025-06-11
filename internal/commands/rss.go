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
	"strconv"
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

func handlerBrowse(s *state, cmd command, user database.User) error {
	postLimit := 2
	var err error = nil
	if len(cmd.arguments) > 0 {
		postLimit, err = strconv.Atoi(cmd.arguments[0])
		if err != nil {
			return nil
		}
	}

	params := database.GetPostsForUserParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		Limit:  int32(postLimit),
	}
	posts, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return nil
	}

	for _, post := range posts {
		fmt.Printf("Title:%s\nDescription: %s\nURL: %s\n", post.Title.String, post.Description.String, post.Url.String)
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
		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   sqlCurrentTime(),
			UpdatedAt:   sqlCurrentTime(),
			Title:       sqlString(html.UnescapeString(item.Title)),
			Url:         sqlString(item.Link),
			Description: sqlString(html.UnescapeString(item.Description)),
			FeedID:      uuid.NullUUID{UUID: feed.ID, Valid: true},
		}

		pubAt, err := convertRssTimestamp(item.PubDate)
		if err != nil {
			return err
		}
		postParams.PublishedAt = sql.NullTime{Time: pubAt, Valid: true}

		if err = s.db.CreatePost(context.Background(), postParams); err != nil {
			return nil
		}
		fmt.Printf("Successfully save %s to database\n", html.UnescapeString(item.Title))
	}
	fmt.Println("----------------------------------------")

	return nil
}

func convertRssTimestamp(timeStamp string) (time.Time, error) {
	var rssTimeFormats = []string{
		time.RFC1123Z,                    // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,                     // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC3339,                     // "2006-01-02T15:04:05Z07:00"
		time.RFC822Z,                     // "02 Jan 06 15:04 -0700"
		time.RFC822,                      // "02 Jan 06 15:04 MST"
		"2006-01-02 15:04:05",            // MySQL datetime format
		"2006-01-02T15:04:05",            // Without timezone
		"Mon, 2 Jan 2006 15:04:05 -0700", // Single digit day
		"Mon, 2 Jan 2006 15:04:05 MST",   // Single digit day with MST
		"2006-01-02",                     // Date only
		"Jan 2, 2006",                    // Month Day, Year
		"January 2, 2006",                // Full month name
	}

	for _, timeFormat := range rssTimeFormats {
		if t, err := time.Parse(timeFormat, timeStamp); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("no time format available for %s", timeStamp)
}
