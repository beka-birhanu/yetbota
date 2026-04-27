package comment

import (
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
)

type AddRequest struct {
	PostID    string `validate:"required"`
	Comment   string `validate:"required"`
	IsAnswer  bool
	CommentID string
}

func (r *AddRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type AddResponse struct {
	Comment *dbmodels.Comment
}

type ReadRequest struct {
	ID string `validate:"required"`
}

func (r *ReadRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ReadResponse struct {
	Comment *dbmodels.Comment
}

type ListRequest struct {
	PostID    string
	CommentID string
}

func (r *ListRequest) Validate() error {
	if r.PostID == "" && r.CommentID == "" {
		return &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "post_id or comment_id required",
			ServiceMessage:    "post_id or comment_id required",
		}
	}
	return nil
}

type ListResponse struct {
	Comments dbmodels.CommentSlice
}

type DeleteRequest struct {
	ID string `validate:"required"`
}

func (r *DeleteRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type VoteRequest struct {
	CommentID string `validate:"required"`
	VoteType  string `validate:"required" oneof:"upvote downvote"`
}

func (r *VoteRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}

	return nil
}

type VoteResponse struct {
	Upvote   int
	Downvote int
}
