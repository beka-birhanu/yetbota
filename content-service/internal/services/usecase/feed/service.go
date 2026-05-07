package feed

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainFeed "github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	domainPhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/photo"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
)

type Service interface {
	ListFeed(ctx context.Context, ctxSess *ctxRP.Context, req *ListFeedRequest) (*ListFeedResponse, error)
	MarkViewed(ctx context.Context, ctxSess *ctxRP.Context, req *MarkViewedRequest) error
}

type Config struct {
	FeedRepo      domainFeed.Repository      `validate:"required"`
	SeenRepo      domainFeed.SeenRepository  `validate:"required"`
	PostRepo      domainPost.Repository      `validate:"required"`
	PostPhotoRepo domainPostphoto.Repository `validate:"required"`
	PhotoRepo     domainPhoto.Repository     `validate:"required"`
	SeenCache     storage.Set                `validate:"required"`
	SeenCacheTTL  int64                      `validate:"required,min=1"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	feedRepo      domainFeed.Repository
	seenRepo      domainFeed.SeenRepository
	postRepo      domainPost.Repository
	postPhotoRepo domainPostphoto.Repository
	photoRepo     domainPhoto.Repository
	seenCache     storage.Set
	seenCacheTTL  int64
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &svc{
		feedRepo:      cfg.FeedRepo,
		seenRepo:      cfg.SeenRepo,
		postRepo:      cfg.PostRepo,
		postPhotoRepo: cfg.PostPhotoRepo,
		photoRepo:     cfg.PhotoRepo,
		seenCache:     cfg.SeenCache,
		seenCacheTTL:  cfg.SeenCacheTTL,
	}, nil
}
