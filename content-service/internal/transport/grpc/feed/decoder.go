package feed

import (
	"context"

	pbfeed "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/feed/v1"
	feedSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/feed"
)

func decodeListReq(_ context.Context, req any) (any, error) {
	in := req.(*pbfeed.ListRequest)
	return &feedSvc.ListFeedRequest{
		Cursor: in.GetCursor(),
		PageSize: func() int {
			if l := int(in.GetLimit()); l > 0 {
				return l
			}
			return defaultPageSize
		}(),
	}, nil
}

func decodeMarkAsSeenReq(_ context.Context, req any) (any, error) {
	in := req.(*pbfeed.MarkAsSeenRequest)
	return &feedSvc.MarkViewedRequest{PostIDs: in.GetPostIds()}, nil
}
