package feed

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	pbfeed "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/feed/v1"
	feedSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/feed"
)

func encodeMarkAsSeenRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pbfeed.MarkAsSeenResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}
	return &pbfeed.MarkAsSeenResponse{Code: "00", Success: true, Message: "success"}, nil
}

func encodeListRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pbfeed.ListResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}
	r, ok := resp.(*feedSvc.ListFeedResponse)
	if !ok || r == nil {
		return &pbfeed.ListResponse{
			Code:    fmt.Sprintf("%d", status.ServerError),
			Success: false,
			Message: "something went wrong",
		}, nil
	}

	posts := make([]*pbfeed.Post, 0, len(r.Posts))
	for _, p := range r.Posts {
		posts = append(posts, toProtoPost(p, r.Photos[p.ID]))
	}

	return &pbfeed.ListResponse{
		Code:    "00",
		Success: true,
		Message: "success",
		Data:    &pbfeed.ListResponseData{Posts: posts, Cursor: r.NextCursor},
	}, nil
}
