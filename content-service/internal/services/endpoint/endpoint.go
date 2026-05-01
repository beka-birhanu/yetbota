package endpoint

import (
	"github.com/go-kit/kit/endpoint"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	commentSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/comment"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
)

type Endpoints struct {
	PostAdd    endpoint.Endpoint
	PostRead   endpoint.Endpoint
	PostUpdate endpoint.Endpoint
	PostVote   endpoint.Endpoint
	PostList endpoint.Endpoint

	CommentAdd    endpoint.Endpoint
	CommentRead   endpoint.Endpoint
	CommentList   endpoint.Endpoint
	CommentDelete endpoint.Endpoint
	CommentVote   endpoint.Endpoint
}

type Config struct {
	PostService    postSvc.Service    `validate:"required"`
	CommentService commentSvc.Service `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewEndpoints(c *Config) (*Endpoints, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &Endpoints{
		PostAdd:    makePostAddEndpoint(c.PostService),
		PostRead:   makePostReadEndpoint(c.PostService),
		PostUpdate: makePostUpdateEndpoint(c.PostService),
		PostVote:   makePostVoteEndpoint(c.PostService),
		PostList: makePostListEndpoint(c.PostService),

		CommentAdd:    makeCommentAddEndpoint(c.CommentService),
		CommentRead:   makeCommentReadEndpoint(c.CommentService),
		CommentList:   makeCommentListEndpoint(c.CommentService),
		CommentDelete: makeCommentDeleteEndpoint(c.CommentService),
		CommentVote:   makeCommentVoteEndpoint(c.CommentService),
	}, nil
}
