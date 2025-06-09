-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListFeeds :many
SELECT feeds.name,
    feeds.url,
    users.name AS user_name
FROM users
    RIGHT JOIN feeds ON users.id = feeds.user_id;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feeds_follow (id, created_at, updated_at, user_id, feeds_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT inserted_feed_follow.*,
    feeds.name AS feeds_name,
    users.name AS user_name
FROM inserted_feed_follow
    INNER JOIN feeds ON inserted_feed_follow.feeds_id = feeds.id
    INNER JOIN users ON feeds.user_id = users.id;

-- name: GetFeedByUrl :one
SELECT *
FROM feeds
WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT *
FROM feeds_follow
WHERE feeds_follow.user_id = $1;