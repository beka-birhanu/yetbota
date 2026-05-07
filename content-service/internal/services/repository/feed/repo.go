package feed

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	"github.com/redis/go-redis/v9"
)

// RedisRepository stores feed items in a Redis sorted set.
//
// Key format:
//
//	feed:{userID}
//
// Member  -> postID
// Score   -> ranking score
//
// Items are ordered by score.
type RedisRepository struct {
	client *redis.Client
	prefix string
}

type Config struct {
	RDB    *redis.Client `validate:"required"`
	Prefix string
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewRedisRepository(c *Config) (feed.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &RedisRepository{
		client: c.RDB,
		prefix: c.Prefix,
	}, nil
}

func (r *RedisRepository) Count(ctx context.Context, userID string) (int64, error) {
	key := r.key(userID)

	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return 0, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: count failed: %v", err),
		}
	}

	return n, nil
}

func (r *RedisRepository) AddBulk(ctx context.Context, userID string, items []*feed.FeedItem) error {
	if len(items) == 0 {
		return nil
	}

	members := make([]redis.Z, 0, len(items))

	for _, item := range items {
		members = append(members, redis.Z{
			Score:  item.Score,
			Member: item.PostID,
		})
	}

	err := r.client.ZAdd(ctx, r.key(userID), members...).Err()
	if err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: add bulk failed: %v", err),
		}
	}
	return nil
}

func (r *RedisRepository) List(ctx context.Context, userID string, opts *feed.ListOptions) ([]*feed.FeedItem, error) {
	key := r.key(userID)

	// Descending order (highest score first)
	values, err := r.client.ZRevRangeByScoreWithScores(ctx, key, r.mapListOptions(opts)).Result()
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: list failed: %v", err),
		}
	}

	items := make([]*feed.FeedItem, 0, len(values))

	for _, value := range values {
		postID, ok := value.Member.(string)
		if !ok {
			continue
		}

		items = append(items, &feed.FeedItem{
			PostID: postID,
			Score:  value.Score,
		})
	}

	return items, nil
}
