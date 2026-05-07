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
}

type ListOptions struct {
	MinScore float64
	MaxScore float64
	Limit    int
}

type SeenRepository interface {
	AddBulk(ctx context.Context, userID string, postIDs []string) error
}
