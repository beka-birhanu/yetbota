package user

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/identity/v1"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/identity-service/internal/services/endpoint"
	gkgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pb.UnimplementedUserServiceServer
	list          gkgrpc.Handler
	read          gkgrpc.Handler
	readPublic    gkgrpc.Handler
	update        gkgrpc.Handler
	updateSelf    gkgrpc.Handler
	register      gkgrpc.Handler
	delete        gkgrpc.Handler
	deleteSelf    gkgrpc.Handler
	uploadProfile gkgrpc.Handler
	checkMobile   gkgrpc.Handler
	follow        gkgrpc.Handler
	unfollow      gkgrpc.Handler
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
		list: gkgrpc.NewServer(
			c.E.UserList, decodeListReq, encodeListRes,
		),
		read: gkgrpc.NewServer(
			c.E.UserRead, decodeReadReq, encodeReadRes,
		),
		readPublic: gkgrpc.NewServer(
			c.E.UserReadPublic, decodeReadPublicReq, encodeReadPublicRes,
		),
		update: gkgrpc.NewServer(
			c.E.UserUpdate, decodeUpdateReq, encodeUpdateRes,
		),
		updateSelf: gkgrpc.NewServer(
			c.E.UserUpdateSelf, decodeUpdateSelfReq, encodeUpdateSelfRes,
		),
		register: gkgrpc.NewServer(
			c.E.UserRegister, decodeRegisterReq, encodeRegisterRes,
		),
		delete: gkgrpc.NewServer(
			c.E.UserDelete, decodeDeleteReq, encodeDeleteRes,
		),
		deleteSelf: gkgrpc.NewServer(
			c.E.UserDeleteSelf, decodeDeleteSelfReq, encodeDeleteSelfRes,
		),
		uploadProfile: gkgrpc.NewServer(
			c.E.UserUploadProfile, decodeUploadProfileReq, encodeUploadProfileRes,
		),
		checkMobile: gkgrpc.NewServer(
			c.E.UserCheckMobile, decodeCheckMobileReq, encodeCheckMobileRes,
		),
		follow: gkgrpc.NewServer(
			c.E.UserFollow, decodeFollowReq, encodeFollowRes,
		),
		unfollow: gkgrpc.NewServer(
			c.E.UserUnfollow, decodeUnfollowReq, encodeUnfollowRes,
		),
	}, nil
}

func (h *Handler) RegisterService(srv grpc.ServiceRegistrar) {
	pb.RegisterUserServiceServer(srv, h)
}

func deadlineExceeded(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return status.Error(codes.Canceled, "The client canceled the request!")
	}
	return nil
}

func (h *Handler) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.list.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ListResponse), nil
}

func (h *Handler) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.read.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ReadResponse), nil
}

func (h *Handler) ReadPublic(ctx context.Context, req *pb.ReadPublicRequest) (*pb.ReadPublicResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.readPublic.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ReadPublicResponse), nil
}

func (h *Handler) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.update.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.UpdateResponse), nil
}

func (h *Handler) UpdateSelf(ctx context.Context, req *pb.UpdateSelfRequest) (*pb.UpdateSelfResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.updateSelf.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.UpdateSelfResponse), nil
}

func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.register.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.RegisterResponse), nil
}

func (h *Handler) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.delete.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.DeleteResponse), nil
}

func (h *Handler) DeleteSelf(ctx context.Context, req *pb.DeleteSelfRequest) (*pb.DeleteSelfResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.deleteSelf.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.DeleteSelfResponse), nil
}

func (h *Handler) UploadProfile(ctx context.Context, req *pb.UploadProfileRequest) (*pb.UploadProfileResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.uploadProfile.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.UploadProfileResponse), nil
}

func (h *Handler) CheckMobile(ctx context.Context, req *pb.CheckMobileRequest) (*pb.CheckMobileResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.checkMobile.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.CheckMobileResponse), nil
}

func (h *Handler) Follow(ctx context.Context, req *pb.FollowRequest) (*pb.FollowResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.follow.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.FollowResponse), nil
}

func (h *Handler) Unfollow(ctx context.Context, req *pb.UnfollowRequest) (*pb.UnfollowResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.unfollow.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.UnfollowResponse), nil
}
