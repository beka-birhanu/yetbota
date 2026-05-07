package feed

import (
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
)

type ListFeedRequest struct {
	Cursor   string `validate:"omitempty,regexp=^cursor:[0-9]+(\\.[0-9]+)?,[a-f0-9-]+$"`
	PageSize int    `validate:"required,min=1,max=100"`
}

func (r *ListFeedRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ListFeedResponse struct {
	Posts      []*dbmodels.Post
	Photos     map[string][]*postSvc.OrderedPhoto
	NextCursor string
}

type MarkViewedRequest struct {
	PostIDs []string `validate:"required,min=1,dive,uuid"`
}

func (r *MarkViewedRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}
