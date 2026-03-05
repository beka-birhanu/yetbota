package vector

import (
	"context"

	"github.com/beka-birhanu/yetbota/ai-service/internal/domain/chunk"
)

type ScoredChunk struct {
	Chunk chunk.Chunk
	Score float32
}

type VectorStore interface {
	Upsert(ctx context.Context, collection string, chunks []chunk.Chunk, embeddings [][]float32) error
	Search(ctx context.Context, collection string, queryEmbedding []float32, limit int) ([]ScoredChunk, error)
}