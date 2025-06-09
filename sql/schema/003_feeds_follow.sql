-- +goose Up
CREATE TABLE feeds_follow (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    user_id UUID,
    feeds_id UUID,
    FOREIGN KEY (feeds_id) REFERENCES feeds (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    UNIQUE (user_id, feeds_id)
);

-- +goose Down
DROP TABLE feeds_follow;