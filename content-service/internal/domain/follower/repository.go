package follower

import "context"

type UserWithDepth struct {
	UserID string
	Depth  int
}

type Repository interface {
	FollowerTree(ctx context.Context, authorID string, maxDepth int) ([]UserWithDepth, error)
	FollowersOf(ctx context.Context, userIDs []string) (map[string][]string, error)
}
