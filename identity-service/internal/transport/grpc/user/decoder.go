package user

import (
	"context"

	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/identity/v1"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
	userSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/user"
)

func decodeListReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.ListRequest)

	opts := &domainUser.Options{}
	if in.GetOptions() != nil {
		opts.FirstName = in.GetOptions().GetFirstName()
		opts.Surname = in.GetOptions().GetSurname()
		opts.Username = in.GetOptions().GetUsername()
		opts.Mobile = in.GetOptions().GetMobile()
		opts.Status = in.GetOptions().GetStatus()
		opts.Role = roleFromPb(in.GetOptions().GetRole())
	}

	pagination := &domainUser.Pagination{
		Limit: constants.DefaultPaginationLength,
	}
	if in.GetPagination() != nil {
		pagination.Limit = int(in.GetPagination().GetLimit())
		pagination.Page = int(in.GetPagination().GetPage())
	}

	sort := &domainUser.SortOption{
		Field:     domainUser.SortFieldRating,
		Direction: domainUser.SortDirectionDesc,
	}
	if in.GetSort() != nil {
		sort.Field = domainUser.SortField(in.GetSort().GetField())
		sort.Direction = domainUser.SortDirection(in.GetSort().GetDirection())
	}

	return &userSvc.ListRequest{
		Options:    opts,
		Pagination: pagination,
		Sort:       sort,
		Resolution: resolutionFromPb(in.GetResolution()),
	}, nil
}

func decodeReadReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.ReadRequest)
	return &userSvc.ReadRequest{
		ID:         in.GetId(),
		Resolution: resolutionFromPb(in.GetResolution()),
	}, nil
}

func decodeReadPublicReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.ReadPublicRequest)
	return &userSvc.ReadPublicRequest{
		ID:         in.GetId(),
		Resolution: resolutionFromPb(in.GetResolution()),
	}, nil
}

func decodeUpdateReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.UpdateRequest)
	return &userSvc.UpdateRequest{
		ID:     in.GetId(),
		Status: in.GetStatus(),
		Role:   roleFromPb(in.GetRole()),
	}, nil
}

func decodeUpdateSelfReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.UpdateSelfRequest)
	return &userSvc.UpdateSelfRequest{
		FirstName: in.GetFirstName(),
		LastName:  in.GetLastName(),
		Username:  in.GetUsername(),
	}, nil
}

func decodeRegisterReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.RegisterRequest)
	return &userSvc.RegisterRequest{
		ID:        in.GetId(),
		FirstName: in.GetFirstName(),
		LastName:  in.GetLastName(),
		Username:  in.GetUsername(),
		Mobile:    in.GetMobile(),
		Password:  in.GetPassword(),
		Random:    in.GetRandom(),
	}, nil
}

func decodeDeleteReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.DeleteRequest)
	return &userSvc.DeleteRequest{
		ID: in.GetId(),
	}, nil
}

func decodeDeleteSelfReq(_ context.Context, req any) (any, error) {
	_ = req.(*pb.DeleteSelfRequest)
	return &userSvc.DeleteSelfRequest{}, nil
}

func decodeUploadProfileReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.UploadProfileRequest)
	return &userSvc.UploadProfileRequest{
		Image: in.GetImage(),
	}, nil
}

func decodeCheckMobileReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.CheckMobileRequest)
	return &userSvc.CheckMobileRequest{
		Mobile: in.GetMobile(),
	}, nil
}

func decodeFollowReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.FollowRequest)
	return &userSvc.FollowRequest{
		FolloweeID: in.GetFolloweeId(),
	}, nil
}

func decodeUnfollowReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.UnfollowRequest)
	return &userSvc.UnfollowRequest{
		FolloweeID: in.GetFolloweeId(),
	}, nil
}
