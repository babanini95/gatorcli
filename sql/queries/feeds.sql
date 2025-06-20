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
    INSERT INTO feeds_follow (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
    INNER JOIN feeds ON inserted_feed_follow.feed_id = feeds.id
    INNER JOIN users ON inserted_feed_follow.user_id = users.id;

-- name: GetFeedByUrl :one
SELECT *
FROM feeds
WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT feeds_follow.*,
    users.name AS user_name,
    feeds.name AS feed_name
FROM feeds_follow
    INNER JOIN users ON feeds_follow.user_id = users.id
    INNER JOIN feeds ON feeds_follow.feed_id = feeds.id
WHERE feeds_follow.user_id = $1;

-- name: DeleteFeedFollowsByUrl :exec
DELETE FROM feeds_follow
WHERE feeds_follow.user_id = $1
    AND feeds_follow.feed_id = (
        SELECT id
        FROM feeds
        WHERE url = $2
    );

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1,
    updated_at = $1
WHERE id = $2;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds f
WHERE f.id IN (
        SELECT ff.feed_id
        FROM feeds_follow ff
        WHERE ff.user_id = $1
    )
ORDER BY f.last_fetched_at ASC NULLS FIRST
LIMIT 1;