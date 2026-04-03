package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Endpoints struct {
	Embed          endpoint.Endpoint
	EmbedBatch     endpoint.Endpoint
	UpsertChunks   endpoint.Endpoint
	Search         endpoint.Endpoint
	FindCandidates endpoint.Endpoint
	Chat           endpoint.Endpoint
}

func NewEndpoints() *Endpoints {
	stub := func(name string) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			return nil, status.Errorf(codes.Unimplemented, "%s not implemented (phase 2+)", name)
		}
	}

	return &Endpoints{
		Embed:          stub("EmbeddingService.Embed"),
		EmbedBatch:     stub("EmbeddingService.EmbedBatch"),
		UpsertChunks:   stub("VectorService.UpsertChunks"),
		Search:         stub("VectorService.Search"),
		FindCandidates: stub("DuplicateService.FindCandidates"),
		Chat:           stub("RAGService.Chat"),
	}
}