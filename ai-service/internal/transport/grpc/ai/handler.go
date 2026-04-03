package ai

import (
	"context"

	"github.com/beka-birhanu/yetbota/ai-service/internal/services/endpoint"
	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/ai/v1"
	gkgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

type Handler struct {
	pb.UnimplementedEmbeddingServiceServer
	pb.UnimplementedVectorServiceServer
	pb.UnimplementedDuplicateServiceServer
	pb.UnimplementedRAGServiceServer

	embed          gkgrpc.Handler
	embedBatch     gkgrpc.Handler
	upsertChunks   gkgrpc.Handler
	search         gkgrpc.Handler
	findCandidates gkgrpc.Handler
	chat           gkgrpc.Handler
}

func NewHandler(e *endpoint.Endpoints) *Handler {
	return &Handler{
		embed:          gkgrpc.NewServer(e.Embed, decodeEmbedReq, encodeEmbedRes),
		embedBatch:     gkgrpc.NewServer(e.EmbedBatch, decodeEmbedBatchReq, encodeEmbedBatchRes),
		upsertChunks:   gkgrpc.NewServer(e.UpsertChunks, decodeUpsertChunksReq, encodeUpsertChunksRes),
		search:         gkgrpc.NewServer(e.Search, decodeSearchReq, encodeSearchRes),
		findCandidates: gkgrpc.NewServer(e.FindCandidates, decodeFindCandidatesReq, encodeFindCandidatesRes),
		chat:           gkgrpc.NewServer(e.Chat, decodeChatReq, encodeChatRes),
	}
}

func (h *Handler) RegisterService(srv grpc.ServiceRegistrar) {
	pb.RegisterEmbeddingServiceServer(srv, h)
	pb.RegisterVectorServiceServer(srv, h)
	pb.RegisterDuplicateServiceServer(srv, h)
	pb.RegisterRAGServiceServer(srv, h)
}

func (h *Handler) Embed(ctx context.Context, req *pb.EmbedRequest) (*pb.EmbedResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.embed.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.EmbedResponse), nil
}

func (h *Handler) EmbedBatch(ctx context.Context, req *pb.EmbedBatchRequest) (*pb.EmbedBatchResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.embedBatch.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.EmbedBatchResponse), nil
}

func (h *Handler) UpsertChunks(ctx context.Context, req *pb.UpsertChunksRequest) (*pb.UpsertChunksResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.upsertChunks.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.UpsertChunksResponse), nil
}

func (h *Handler) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.search.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.SearchResponse), nil
}

func (h *Handler) FindCandidates(ctx context.Context, req *pb.FindCandidatesRequest) (*pb.FindCandidatesResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.findCandidates.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.FindCandidatesResponse), nil
}

func (h *Handler) Chat(ctx context.Context, req *pb.ChatRequest) (*pb.ChatResponse, error) {
	if err := deadlineExceeded(ctx); err != nil {
		return nil, err
	}
	_, resp, err := h.chat.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ChatResponse), nil
}