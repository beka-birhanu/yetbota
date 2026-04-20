package auth

import (
	"context"

	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/identity/v1"
	authSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/auth"
)

func decodeLoginReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.LoginRequest)
	return &authSvc.LoginRequest{
		Username: in.GetUsername(),
		Password: in.GetPassword(),
		Site:     in.GetSite(),
	}, nil
}

func decodeRefreshReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.RefreshRequest)
	return &authSvc.RefreshRequest{
		RefreshToken: in.GetRefreshToken(),
		Username:     in.GetUsername(),
	}, nil
}

func decodeLogoutReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.LogoutRequest)
	return &authSvc.LogoutRequest{
		RefreshToken: in.GetRefreshToken(),
		Username:     in.GetUsername(),
	}, nil
}

func decodeGenerateMobileOTPReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.GenerateMobileOTPRequest)
	return &authSvc.GenerateMobileOTPRequest{
		Mobile: in.GetMobile(),
		Random: in.GetRandom(),
	}, nil
}

func decodeValidateOTPReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.ValidateOTPRequest)
	return &authSvc.ValidateOTPRequest{
		Otp:    in.GetOtp(),
		Mobile: in.GetMobile(),
		Random: in.GetRandom(),
	}, nil
}

func decodeNewPasswordReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.NewPasswordRequest)
	return &authSvc.NewPasswordRequest{
		Password: in.GetPassword(),
		Random:   in.GetRandom(),
		Mobile:   in.GetMobile(),
	}, nil
}

func decodeAuthorizationReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.AuthorizationRequest)
	return &authSvc.AuthorizationRequest{
		Resource: in.GetResource(),
		Action:   in.GetAction(),
	}, nil
}

func decodeChangePasswordReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.ChangePasswordRequest)
	return &authSvc.ChangePasswordRequest{
		CurrentPassword: in.GetCurrentPassword(),
		NewPassword:     in.GetNewPassword(),
	}, nil
}

func decodeChangeMobileReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.ChangeMobileRequest)
	return &authSvc.ChangeMobileRequest{
		NewMobile: in.GetNewMobile(),
		Random:    in.GetRandom(),
	}, nil
}
