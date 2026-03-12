package user

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/identity/v1"
	userSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/user"
)

func encodeListRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.ListResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	r, ok := resp.(*userSvc.ListResponse)
	if ok {
		return &pb.ListResponse{
			Code:    "00",
			Success: true,
			Message: "List Successful",
			Data:    listResponseToProto(r),
		}, nil
	}

	return &pb.ListResponse{
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

	r, ok := resp.(*userSvc.ReadResponse)
	if ok {
		return &pb.ReadResponse{
			Code:    "00",
			Success: true,
			Message: "Read Successful",
			Data:    readResponseToProto(r),
		}, nil
	}

	return &pb.ReadResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeReadPublicRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.ReadPublicResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	r, ok := resp.(*userSvc.ReadPublicResponse)
	if ok {
		return &pb.ReadPublicResponse{
			Code:    "00",
			Success: true,
			Message: "Read Successful",
			Data:    readPublicResponseToProto(r),
		}, nil
	}

	return &pb.ReadPublicResponse{
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

	r, ok := resp.(*userSvc.UpdateResponse)
	if ok {
		return &pb.UpdateResponse{
			Code:    "00",
			Success: true,
			Message: "Update Successful",
			Data:    userToPrivateUser(r.User, ""),
		}, nil
	}

	return &pb.UpdateResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeUpdateSelfRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.UpdateSelfResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	r, ok := resp.(*userSvc.UpdateSelfResponse)
	if ok {
		return &pb.UpdateSelfResponse{
			Code:    "00",
			Success: true,
			Message: "Update Successful",
			Data:    userToPrivateUser(r.User, ""),
		}, nil
	}

	return &pb.UpdateSelfResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeRegisterRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.RegisterResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	r, ok := resp.(*userSvc.RegisterResponse)
	if ok {
		return &pb.RegisterResponse{
			Code:    "00",
			Success: true,
			Message: "Register Successful",
			Data:    userToPrivateUser(r.User, ""),
		}, nil
	}

	return &pb.RegisterResponse{
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

	_, ok := resp.(*userSvc.DeleteResponse)
	if ok {
		return &pb.DeleteResponse{
			Code:    "00",
			Success: true,
			Message: "Delete Successful",
		}, nil
	}

	return &pb.DeleteResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeDeleteSelfRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.DeleteSelfResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	_, ok := resp.(*userSvc.DeleteSelfResponse)
	if ok {
		return &pb.DeleteSelfResponse{
			Code:    "00",
			Success: true,
			Message: "Delete Successful",
		}, nil
	}

	return &pb.DeleteSelfResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeUploadProfileRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.UploadProfileResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	r, ok := resp.(*userSvc.UploadProfileResponse)
	if ok {
		return &pb.UploadProfileResponse{
			Code:    "00",
			Success: true,
			Message: "Upload Profile Successful",
			Data:    r.URL,
		}, nil
	}

	return &pb.UploadProfileResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeCheckMobileRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.CheckMobileResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	r, ok := resp.(*userSvc.CheckMobileResponse)
	if ok {
		return &pb.CheckMobileResponse{
			Code:    "00",
			Success: true,
			Message: "Check Mobile Successful",
			Data:    r.Exists,
		}, nil
	}

	return &pb.CheckMobileResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeFollowRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.FollowResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	_, ok := resp.(*userSvc.FollowResponse)
	if ok {
		return &pb.FollowResponse{
			Code:    "00",
			Success: true,
			Message: "Follow Successful",
		}, nil
	}

	return &pb.FollowResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeUnfollowRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.UnfollowResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	_, ok := resp.(*userSvc.UnfollowResponse)
	if ok {
		return &pb.UnfollowResponse{
			Code:    "00",
			Success: true,
			Message: "Unfollow Successful",
		}, nil
	}

	return &pb.UnfollowResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}
