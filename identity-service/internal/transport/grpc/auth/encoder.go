package auth

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/identity/v1"
	authSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/auth"
)

func encodeLoginRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.LoginResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	r, ok := resp.(*authSvc.LoginResponse)
	if ok {
		return &pb.LoginResponse{
			Code:    "00",
			Success: true,
			Message: "Login Successful",
			Data:    tokenDataFromLoginResponse(r),
		}, nil
	}

	return &pb.LoginResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeRefreshRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.RefreshResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	r, ok := resp.(*authSvc.RefreshResponse)
	if ok {
		return &pb.RefreshResponse{
			Code:    "00",
			Success: true,
			Message: "Refresh Successful",
			Data:    tokenDataFromRefreshResponse(r),
		}, nil
	}

	return &pb.RefreshResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeLogoutRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.LogoutResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	_, ok := resp.(*authSvc.LogoutResponse)
	if ok {
		return &pb.LogoutResponse{
			Code:    "00",
			Success: true,
			Message: "Logout Successful",
		}, nil
	}

	return &pb.LogoutResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeGenerateMobileOTPRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.GenerateMobileOTPResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	r, ok := resp.(*authSvc.GenerateMobileOTPResponse)
	if ok {
		return &pb.GenerateMobileOTPResponse{
			Code:    "00",
			Success: true,
			Message: "OTP Generated",
			Data:    otpDataToProto(r.OtpReqCount, r.MaxOtpReq, r.OtpErrCount, r.MaxOtpErr),
		}, nil
	}

	return &pb.GenerateMobileOTPResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeValidateOTPRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.ValidateOTPResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	r, ok := resp.(*authSvc.ValidateOTPResponse)
	if ok {
		return &pb.ValidateOTPResponse{
			Code:    "00",
			Success: true,
			Message: "OTP Validated",
			Data:    otpDataToProto(r.OtpReqCount, r.MaxOtpReq, r.OtpErrCount, r.MaxOtpErr),
		}, nil
	}

	return &pb.ValidateOTPResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeNewPasswordRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.NewPasswordResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	_, ok := resp.(*authSvc.NewPasswordResponse)
	if ok {
		return &pb.NewPasswordResponse{
			Code:    "00",
			Success: true,
			Message: "Password Updated",
		}, nil
	}

	return &pb.NewPasswordResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeAuthorizationRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.AuthorizationResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	_, ok := resp.(*authSvc.AuthorizationResponse)
	if ok {
		return &pb.AuthorizationResponse{
			Code:    "00",
			Success: true,
			Message: "Authorized",
		}, nil
	}

	return &pb.AuthorizationResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeChangePasswordRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.ChangePasswordResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	_, ok := resp.(*authSvc.ChangePasswordResponse)
	if ok {
		return &pb.ChangePasswordResponse{
			Code:    "00",
			Success: true,
			Message: "Password Changed",
		}, nil
	}

	return &pb.ChangePasswordResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}

func encodeChangeMobileRes(_ context.Context, resp any) (any, error) {
	if errResp, ok := resp.(*toddlerr.Error); ok {
		return &pb.ChangeMobileResponse{
			Code:    fmt.Sprintf("%d", errResp.PublicStatusCode),
			Success: false,
			Message: errResp.PublicMessage,
		}, nil
	}

	_, ok := resp.(*authSvc.ChangeMobileResponse)
	if ok {
		return &pb.ChangeMobileResponse{
			Code:    "00",
			Success: true,
			Message: "Mobile Changed",
		}, nil
	}

	return &pb.ChangeMobileResponse{
		Code:    fmt.Sprintf("%d", status.ServerError),
		Success: false,
		Message: "something went wrong",
	}, nil
}
