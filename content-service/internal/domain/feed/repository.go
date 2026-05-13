package feed

import "context"

type FeedItem struct {
	PostID string  `json:"postID" validate:"required,uuid4"`
	Score  float64 `json:"score" validate:"required,min=0"`
}

type Repository interface {
	Count(ctx context.Context, userID string) (int64, error)
	List(ctx context.Context, userID string, opts *ListOptions) ([]*FeedItem, error)
	AddBulk(ctx context.Context, userID string, items []*FeedItem) error
	// AddBulkGT adds items to the feed only when the new score is greater than the existing score.
	AddBulkGT(ctx context.Context, userID string, items []*FeedItem) error
	// FanOutGT writes postID to each user's feed at the given score (ZADD GT).
	// A single pipelined round-trip covers all users.
	FanOutGT(ctx context.Context, postID string, userScores map[string]float64) error
	// AddRecipients records which users received a post (for future score-update propagation).
	AddRecipients(ctx context.Context, postID string, userIDs []string) error
}

type SeenRepository interface {
	AddBulk(ctx context.Context, userID string, postIDs []string) error
}

type ListOptions struct {
	MinScore float64
	MaxScore float64
	Limit    int
}
