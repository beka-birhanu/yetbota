package storage

import "context"

type Set interface {
	Add(ctx context.Context, key string, ttl int64) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, keys []string) (map[string]bool, error)
}
