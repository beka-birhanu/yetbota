package auth

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/identity/v1"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/identity-service/internal/services/endpoint"
	gkgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

type Handler struct {
	pb.UnimplementedAuthServiceServer
	login             gkgrpc.Handler
	refresh           gkgrpc.Handler
	logout            gkgrpc.Handler
	generateMobileOTP gkgrpc.Handler
	validateOTP       gkgrpc.Handler
	newPassword       gkgrpc.Handler
	authorization     gkgrpc.Handler
	changePassword    gkgrpc.Handler
	changeMobile      gkgrpc.Handler
}

type Config struct {
	E *endpoint.Endpoints `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewHandler(c *Config) (*Handler, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &Handler{
		login: gkgrpc.NewServer(
			c.E.Login, decodeLoginReq, encodeLoginRes,
		),
		refresh: gkgrpc.NewServer(
			c.E.Refresh, decodeRefreshReq, encodeRefreshRes,
		),
		logout: gkgrpc.NewServer(
			c.E.Logout, decodeLogoutReq, encodeLogoutRes,
		),
		generateMobileOTP: gkgrpc.NewServer(
			c.E.GenerateMobileOTP, decodeGenerateMobileOTPReq, encodeGenerateMobileOTPRes,
		),
		validateOTP: gkgrpc.NewServer(
			c.E.ValidateOTP, decodeValidateOTPReq, encodeValidateOTPRes,
		),
		newPassword: gkgrpc.NewServer(
			c.E.NewPassword, decodeNewPasswordReq, encodeNewPasswordRes,
		),
		authorization: gkgrpc.NewServer(
			c.E.Authorization, decodeAuthorizationReq, encodeAuthorizationRes,
		),
		changePassword: gkgrpc.NewServer(
			c.E.ChangePassword, decodeChangePasswordReq, encodeChangePasswordRes,
		),
		changeMobile: gkgrpc.NewServer(
			c.E.ChangeMobile, decodeChangeMobileReq, encodeChangeMobileRes,
		),
	}, nil
}

func (h *Handler) RegisterService(srv grpc.ServiceRegistrar) {
	pb.RegisterAuthServiceServer(srv, h)
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.login.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.LoginResponse), nil
}

func (h *Handler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.refresh.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.RefreshResponse), nil
}

func (h *Handler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.logout.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.LogoutResponse), nil
}

func (h *Handler) GenerateMobileOTP(ctx context.Context, req *pb.GenerateMobileOTPRequest) (*pb.GenerateMobileOTPResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.generateMobileOTP.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.GenerateMobileOTPResponse), nil
}

func (h *Handler) ValidateOTP(ctx context.Context, req *pb.ValidateOTPRequest) (*pb.ValidateOTPResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.validateOTP.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ValidateOTPResponse), nil
}

func (h *Handler) NewPassword(ctx context.Context, req *pb.NewPasswordRequest) (*pb.NewPasswordResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.newPassword.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.NewPasswordResponse), nil
}

func (h *Handler) Authorization(ctx context.Context, req *pb.AuthorizationRequest) (*pb.AuthorizationResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.authorization.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.AuthorizationResponse), nil
}

func (h *Handler) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.changePassword.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ChangePasswordResponse), nil
}

func (h *Handler) ChangeMobile(ctx context.Context, req *pb.ChangeMobileRequest) (*pb.ChangeMobileResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.changeMobile.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ChangeMobileResponse), nil
}
