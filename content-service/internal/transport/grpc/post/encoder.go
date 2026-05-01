package post

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/post/v1"
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
		return &pb.VoteResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}
	r, ok := resp.(*postSvc.PostVoteResponse)
	if ok {
		return &pb.VoteResponse{
			Code:     "00",
			Success:  true,
			Message:  "Vote recorded",
			Likes:    int32(r.Likes),
			Dislikes: int32(r.Dislikes),
		}, nil
	}
	return &pb.VoteResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeListRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.ListResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}
	r, ok := resp.(*postSvc.ListResponse)
	if !ok {
		return &pb.ListResponse{
			Code:    fmt.Sprintf("%d", status.ServerError),
			Success: false,
			Message: "something went wrong",
		}, nil
	}

	posts := make([]*pb.Post, 0, len(r.Posts))
	for _, p := range r.Posts {
		photos := r.Photos[p.ID]
		posts = append(posts, postToProto(p, photos))
	}

	return &pb.ListResponse{
		Code:     "00",
		Success:  true,
		Message:  "Posts retrieved successfully",
		Data:     posts,
		Total:    r.Total,
		Page:     int32(r.Page),
		PageSize: int32(r.PageSize),
	}, nil
}
