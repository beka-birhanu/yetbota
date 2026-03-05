package chunk

import "context"

// Chunk represents a piece of text and metadata for vector storage and retrieval.
type Chunk struct {
	SourceID   string            // e.g. location ID
	Text       string
	Metadata   map[string]string // e.g. category, name
	SegmentIdx int               // order within source document
}

// Chunker splits documents into chunks suitable for embedding and RAG.
type Chunker interface {
	Chunk(ctx context.Context, documentID string, text string, metadata map[string]string) ([]Chunk, error)
}