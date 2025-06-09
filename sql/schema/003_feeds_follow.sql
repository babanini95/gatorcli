-- +goose Up
CREATE TABLE feeds_follow (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    user_id UUID,
    feed_id UUID,
    FOREIGN KEY (feed_id) REFERENCES feeds (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feeds_follow;