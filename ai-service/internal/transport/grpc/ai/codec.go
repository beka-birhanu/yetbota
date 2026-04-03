package ai

import (
	"context"

	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/ai/v1"
)

func decodeEmbedReq(_ context.Context, r any) (any, error) {
	return r.(*pb.EmbedRequest), nil
}

func encodeEmbedRes(_ context.Context, r any) (any, error) {
	return r.(*pb.EmbedResponse), nil
}

func decodeEmbedBatchReq(_ context.Context, r any) (any, error) {
	return r.(*pb.EmbedBatchRequest), nil
}

func encodeEmbedBatchRes(_ context.Context, r any) (any, error) {
	return r.(*pb.EmbedBatchResponse), nil
}

func decodeUpsertChunksReq(_ context.Context, r any) (any, error) {
	return r.(*pb.UpsertChunksRequest), nil
}

func encodeUpsertChunksRes(_ context.Context, r any) (any, error) {
	return r.(*pb.UpsertChunksResponse), nil
}

func decodeSearchReq(_ context.Context, r any) (any, error) {
	return r.(*pb.SearchRequest), nil
}

func encodeSearchRes(_ context.Context, r any) (any, error) {
	return r.(*pb.SearchResponse), nil
}

func decodeFindCandidatesReq(_ context.Context, r any) (any, error) {
	return r.(*pb.FindCandidatesRequest), nil
}

func encodeFindCandidatesRes(_ context.Context, r any) (any, error) {
	return r.(*pb.FindCandidatesResponse), nil
}

func decodeChatReq(_ context.Context, r any) (any, error) {
	return r.(*pb.ChatRequest), nil
}

func encodeChatRes(_ context.Context, r any) (any, error) {
	return r.(*pb.ChatResponse), nil
}