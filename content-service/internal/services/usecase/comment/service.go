package comment

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainComment "github.com/beka-birhanu/yetbota/content-service/internal/domain/comment"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
)

type Service interface {
	Add(ctx context.Context, ctxSess *ctxRP.Context, req *AddRequest) (*AddResponse, error)
	Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error)
	List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error)
	Delete(ctx context.Context, ctxSess *ctxRP.Context, req *DeleteRequest) error
}

type Config struct {
	CommentRepo domainComment.Repository `validate:"required"`
	PostRepo    domainPost.Repository    `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	commentRepo domainComment.Repository
	postRepo    domainPost.Repository
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &svc{commentRepo: cfg.CommentRepo, postRepo: cfg.PostRepo}, nil
}
