package storage

import (
	"context"
	"fmt"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	"github.com/redis/go-redis/v9"
)

type set struct {
	rdb    *redis.Client
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

func NewSet(c *Config) (domainStorage.Set, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &set{rdb: c.RDB, prefix: c.Prefix}, nil
}

func (s *set) Add(ctx context.Context, key string, ttl int64) error {
	key = s.cookKey(key)
	if err := s.rdb.Set(ctx, key, 1, time.Duration(ttl)*time.Second).Err(); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis set: add failed: %v", err),
		}
	}
	return nil
}

func (s *set) Delete(ctx context.Context, key string) error {
	s.cookKey(key)
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis set: delete failed: %v", err),
		}
	}
	return nil
}

func (s *set) Exists(ctx context.Context, keys []string) (map[string]bool, error) {
	result := make(map[string]bool, len(keys))
	if len(keys) == 0 {
		return result, nil
	}

	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.IntCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Exists(ctx, s.cookKey(key))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis set: exists failed: %v", err),
		}
	}

	for i, key := range keys {
		result[key] = cmds[i].Val() > 0
	}
	return result, nil
}

func (s *set) cookKey(key string) string {
	return fmt.Sprintf("%s:%s", s.prefix, key)
}
