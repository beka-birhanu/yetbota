package post

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/post/v1"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	gkgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

type Handler struct {
	pb.UnimplementedPostServiceServer
	add    gkgrpc.Handler
	read   gkgrpc.Handler
	update gkgrpc.Handler
	vote   gkgrpc.Handler
	list   gkgrpc.Handler
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
		add:    gkgrpc.NewServer(c.E.PostAdd, decodeAddReq, encodeAddRes),
		read:   gkgrpc.NewServer(c.E.PostRead, decodeReadReq, encodeReadRes),
		update: gkgrpc.NewServer(c.E.PostUpdate, decodeUpdateReq, encodeUpdateRes),
		vote:   gkgrpc.NewServer(c.E.PostVote, decodeVoteReq, encodeVoteRes),
		list:   gkgrpc.NewServer(c.E.PostList, decodeListReq, encodeListRes),
	}, nil
}

func (h *Handler) RegisterService(srv grpc.ServiceRegistrar) {
	pb.RegisterPostServiceServer(srv, h)
}

func (h *Handler) Add(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.add.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.AddResponse), nil
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

func (h *Handler) Vote(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.vote.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.VoteResponse), nil
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
