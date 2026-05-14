-- +goose Up
create table if not exists user_seen_posts (
    user_id text        not null,
    post_id text        not null,
    seen_at timestamptz not null default now(),
    primary key (user_id, post_id)
);

-- +goose Down
drop table if exists user_seen_posts;
