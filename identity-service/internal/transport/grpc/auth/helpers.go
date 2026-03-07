package auth

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/identity/v1"
	authSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/auth"
)

func deadlineExceeded(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return status.Error(codes.Canceled, "The client canceled the request!")
	}
	return nil
}

func tokenDataFromLoginResponse(r *authSvc.LoginResponse) *pb.TokenData {
	return &pb.TokenData{
		AccessToken:     r.AccessToken,
		AccessTokenTtl:  durationpb.New(time.Duration(r.AccessTokenTTL) * time.Second),
		RefreshToken:    r.RefreshToken,
		RefreshTokenTtl: durationpb.New(time.Duration(r.RefreshTokenTTL) * time.Second),
	}
}

func tokenDataFromRefreshResponse(r *authSvc.RefreshResponse) *pb.TokenData {
	return &pb.TokenData{
		AccessToken:     r.AccessToken,
		AccessTokenTtl:  durationpb.New(time.Duration(r.AccessTokenTTL) * time.Second),
		RefreshToken:    r.RefreshToken,
		RefreshTokenTtl: durationpb.New(time.Duration(r.RefreshTokenTTL) * time.Second),
	}
}

func otpDataToProto(reqCount, maxReq, errCount, maxErr int32) *pb.OTPData {
	return &pb.OTPData{
		OtpReqCount: reqCount,
		MaxOtpReq:   maxReq,
		OtpErrCount: errCount,
		MaxOtpErr:   maxErr,
	}
}
