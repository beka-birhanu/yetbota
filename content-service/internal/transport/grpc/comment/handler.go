package comment

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/comment/v1"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	gkgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

type Handler struct {
	pb.UnimplementedCommentServiceServer
	add    gkgrpc.Handler
	read   gkgrpc.Handler
	list   gkgrpc.Handler
	delete gkgrpc.Handler
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
		add:    gkgrpc.NewServer(c.E.CommentAdd, decodeAddReq, encodeAddRes),
		read:   gkgrpc.NewServer(c.E.CommentRead, decodeReadReq, encodeReadRes),
		list:   gkgrpc.NewServer(c.E.CommentList, decodeListReq, encodeListRes),
		delete: gkgrpc.NewServer(c.E.CommentDelete, decodeDeleteReq, encodeDeleteRes),
	}, nil
}

func (h *Handler) RegisterService(srv grpc.ServiceRegistrar) {
	pb.RegisterCommentServiceServer(srv, h)
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
