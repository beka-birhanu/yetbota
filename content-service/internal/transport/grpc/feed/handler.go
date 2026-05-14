package feed

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	pbfeed "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/feed/v1"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	gkgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

type Handler struct {
	pbfeed.UnimplementedFeedServiceServer
	list       gkgrpc.Handler
	markAsSeen gkgrpc.Handler
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
		list:       gkgrpc.NewServer(c.E.FeedGet, decodeListReq, encodeListRes),
		markAsSeen: gkgrpc.NewServer(c.E.FeedMarkViewed, decodeMarkAsSeenReq, encodeMarkAsSeenRes),
	}, nil
}

func (h *Handler) RegisterService(srv grpc.ServiceRegistrar) {
	pbfeed.RegisterFeedServiceServer(srv, h)
}

func (h *Handler) List(ctx context.Context, req *pbfeed.ListRequest) (*pbfeed.ListResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.list.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pbfeed.ListResponse), nil
}

func (h *Handler) MarkAsSeen(ctx context.Context, req *pbfeed.MarkAsSeenRequest) (*pbfeed.MarkAsSeenResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.markAsSeen.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pbfeed.MarkAsSeenResponse), nil
}
