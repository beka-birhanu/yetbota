package post

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainPhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/photo"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
)

type Service interface {
	Add(ctx context.Context, ctxSess *ctxRP.Context, req *AddRequest) (*AddResponse, error)
	Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error)
	Update(ctx context.Context, ctxSess *ctxRP.Context, req *UpdateRequest) (*UpdateResponse, error)
	Vote(ctx context.Context, ctxSess *ctxRP.Context, req *PostVoteRequest) (*PostVoteResponse, error)
	List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error)
}

type Config struct {
	PostRepo      domainPost.Repository      `validate:"required"`
	PostPhotoRepo domainPostphoto.Repository `validate:"required"`
	PhotoRepo     domainPhoto.Repository     `validate:"required"`
	Bucket        domainStorage.Bucket       `validate:"required"`
	BucketName    string                     `validate:"required"`
	BucketRegion  string                     `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	postRepo      domainPost.Repository
	postPhotoRepo domainPostphoto.Repository
	photoRepo     domainPhoto.Repository
	bucket        domainStorage.Bucket
	bucketName    string
	bucketRegion  string
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &svc{
		postRepo:      cfg.PostRepo,
		postPhotoRepo: cfg.PostPhotoRepo,
		photoRepo:     cfg.PhotoRepo,
		bucket:        cfg.Bucket,
		bucketName:    cfg.BucketName,
		bucketRegion:  cfg.BucketRegion,
	}, nil
}
