-- name: CreatePost :exec
INSERT INTO posts (
        id,
        created_at,
        updated_at,
        title,
        url,
        description,
        published_at,
        feed_id
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetPostsForUser :many
SELECT *
FROM posts p
WHERE p.feed_id IN (
        SELECT ff.feed_id
        FROM feeds_follow ff
        WHERE ff.user_id = $1
    )
ORDER BY p.published_at DESC
LIMIT $2;