package embedding

import "context"

// Embedder produces vector embeddings for text (e.g. for vector DB and retrieval).
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}