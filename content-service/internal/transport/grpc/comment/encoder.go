package comment

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/comment/v1"
	commentSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/comment"
)

func encodeAddRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.AddResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}
	r, ok := resp.(*commentSvc.AddResponse)
	if ok {
		return &pb.AddResponse{
			Code:    "00",
			Success: true,
			Message: "Comment created successfully",
			Data:    commentToProto(r.Comment),
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
	r, ok := resp.(*commentSvc.ReadResponse)
	if ok {
		return &pb.ReadResponse{
			Code:    "00",
			Success: true,
			Message: "Comment read successfully",
			Data:    commentToProto(r.Comment),
		}, nil
	}
	return &pb.ReadResponse{
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
	r, ok := resp.(*commentSvc.ListResponse)
	if ok {
		comments := make([]*pb.Comment, 0, len(r.Comments))
		for _, c := range r.Comments {
			comments = append(comments, commentToProto(c))
		}
		return &pb.ListResponse{
			Code:    "00",
			Success: true,
			Message: "Comments listed successfully",
			Data: &pb.ListResponseData{
				Data:     comments,
				Total:    r.Total,
				Page:     int32(r.Page),
				PageSize: int32(r.PageSize),
			},
		}, nil
	}
	return &pb.ListResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeDeleteRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.DeleteResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}
	return &pb.DeleteResponse{
		Code:    "00",
		Success: true,
		Message: "Comment deleted successfully",
	}, nil
}
