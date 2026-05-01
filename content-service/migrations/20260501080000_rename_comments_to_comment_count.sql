-- +goose Up
ALTER TABLE posts RENAME COLUMN comments TO comment_count;

-- +goose Down
ALTER TABLE posts RENAME COLUMN comment_count TO comments;
