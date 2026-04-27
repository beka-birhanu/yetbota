package comment

import (
	"context"

	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/comment/v1"
	commentSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/comment"
)

func decodeAddReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.AddRequest)
	return &commentSvc.AddRequest{
		PostID:    in.GetPostId(),
		Comment:   in.GetComment(),
		IsAnswer:  in.GetIsAnswer(),
		CommentID: in.GetCommentId(),
	}, nil
}

func decodeReadReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.ReadRequest)
	return &commentSvc.ReadRequest{
		ID: in.GetId(),
	}, nil
}

func decodeListReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.ListRequest)
	return &commentSvc.ListRequest{
		PostID:    in.GetPostId(),
		CommentID: in.GetCommentId(),
	}, nil
}

func decodeDeleteReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.DeleteRequest)
	return &commentSvc.DeleteRequest{
		ID: in.GetId(),
	}, nil
}

func decodeVoteReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.VoteCommentRequest)
	return &commentSvc.VoteRequest{
		CommentID: in.GetCommentId(),
		VoteType:  mapCommentVoteTypeFromProto(in.GetVoteType()),
	}, nil
}
