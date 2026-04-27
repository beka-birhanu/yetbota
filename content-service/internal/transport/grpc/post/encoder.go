package post

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/v1"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
)

func encodeAddRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.AddResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}
	r, ok := resp.(*postSvc.AddResponse)
	if ok {
		return &pb.AddResponse{
			Code:    "00",
			Success: true,
			Message: "Post created successfully",
			Data:    postToProto(r.Post, r.Photos),
		}, nil
	}
	return &pb.AddResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeReadRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.ReadResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}
	r, ok := resp.(*postSvc.ReadResponse)
	if ok {
		return &pb.ReadResponse{
			Code:    "00",
			Success: true,
			Message: "Post read successfully",
			Data:    postToProto(r.Post, r.Photos),
		}, nil
	}
	return &pb.ReadResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeUpdateRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.UpdateResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}
	r, ok := resp.(*postSvc.UpdateResponse)
	if ok {
		return &pb.UpdateResponse{
			Code:    "00",
			Success: true,
			Message: "Post updated successfully",
			Data:    postToProto(r.Post, nil),
		}, nil
	}
	return &pb.UpdateResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeVoteRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.VotePostResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}
	r, ok := resp.(*postSvc.PostVoteResponse)
	if ok {
		return &pb.VotePostResponse{
			Code:     "00",
			Success:  true,
			Message:  "Vote recorded",
			Likes:    int32(r.Likes),
			Dislikes: int32(r.Dislikes),
		}, nil
	}
	return &pb.VotePostResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}
