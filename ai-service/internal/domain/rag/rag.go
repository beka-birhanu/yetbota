package rag

import "context"

type Citation struct {
	SourceID   string
	Text       string
	Score      float32
}

type ChatResponse struct {
	Answer     string
	Citations  []Citation
}

type RAGChat interface {
	Chat(ctx context.Context, query string) (*ChatResponse, error)
}