# Gator

A command-line RSS feed aggregator built with Go that allows users to manage RSS feeds, follow them, and browse aggregated posts through a simple CLI interface.

## Features

- User registration and authentication (coming soon)
- RSS feed management (add, follow, unfollow)
- Automatic feed aggregation at specified intervals
- Post browsing with customizable limits
- PostgreSQL database storage
- Database migrations with Goose
- Type-safe SQL queries with sqlc

## Tech Stack

- **Go** - Core application language
- **PostgreSQL** - Database for storing users, feeds, and posts
- **Goose** - Database migration tool
- **sqlc** - SQL query code generation

## Prerequisites

- Go 1.19 or higher
- PostgreSQL database
- Goose migration tool
- sqlc code generator

## Installation

1. Clone the repository:
```bash
git clone https://github.com/babanini95/gatorcli.git
cd gatorcli
```

2. Install dependencies:
```bash
go mod download
```

3. Set up your PostgreSQL database and [configure](##Configuration)

4. Run database migrations:
```bash
goose up
```

5. Generate SQL code:
```bash
sqlc generate
```

6. Build the application:
```bash
go build -o gator
```

## Usage

### Available Commands

#### User Management

**Register a new user:**
```bash
gator register <username>
```

**Login as a user:**
```bash
gator login <username>
```

**List all registered users:**
```bash
gator users
```

#### Feed Management

**Add a new feed (requires login):**
```bash
gator addfeed <feed_name> <feed_url>
```

**List all feeds:**
```bash
gator feeds
```

**Follow a feed (requires login):**
```bash
gator follow <feed_url>
```

**List feeds you're following (requires login):**
```bash
gator following
```

**Unfollow a feed (requires login):**
```bash
gator unfollow <feed_url>
```

#### Content Aggregation

**Start feed aggregation (requires login):**
```bash
gator agg <interval>
```
Fetches posts from all followed feeds at the specified interval.

**Examples:**
- `gator agg 1m` - Aggregate every minute
- `gator agg 1h` - Aggregate every hour
- `gator agg 30s` - Aggregate every 30 seconds

**Browse saved posts (requires login):**
```bash
gator browse <limit>
```
Shows your saved posts with the specified limit.

#### Utility Commands

**Get help:**
```bash
gator help
```

**Reset database (development only):**
```bash
gator reset
```

## Example Workflow

1. Register a new user:
```bash
gator register john
```

2. Login:
```bash
gator login john
```

3. Add some RSS feeds:
```bash
gator addfeed "Tech News" https://example.com/tech/rss
gator addfeed "Go Blog" https://blog.golang.org/feed.atom
```

4. Follow the feeds:
```bash
gator follow https://example.com/tech/rss
gator follow https://blog.golang.org/feed.atom
```

5. Start aggregating posts every 5 minutes:
```bash
gator agg 5m
```

6. Browse your collected posts:
```bash
gator browse 10
```

## Configuration

(Coming Soon) Need to add details about configuration files, environment variables, and database connection settings

## Database Schema

The application uses PostgreSQL with the following main tables:
- `users` - User accounts
- `feeds` - RSS feed information
- `posts` - Aggregated posts from feeds
- `feeds_follows` - User-feed relationships

## Development

### Database Migrations

To create a new migration:
```bash
goose create migration_name sql
```

To run migrations:
```bash
goose up
```

### Generating SQL Code

After modifying SQL queries in the `sql/queries/` directory:
```bash
sqlc generate
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- Built with [Goose](https://github.com/pressly/goose) for database migrations
- Uses [sqlc](https://sqlc.dev/) for type-safe SQL queries
- Part of a course from [boot.dev](https://boot.dev)