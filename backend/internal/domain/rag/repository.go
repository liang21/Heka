package rag

import (
	"context"

	"github.com/liang21/heka/internal/domain/shared"
)

type ChunkRepository interface {
	CreateBatch(ctx context.Context, chunks []*DocumentChunk) error
	FindByFile(ctx context.Context, fileID shared.ID) ([]*DocumentChunk, error)
	DeleteByFile(ctx context.Context, fileID shared.ID) error
}

type VectorRepository interface {
	Upsert(ctx context.Context, chunks []*DocumentChunk, embeddings [][]float32) error
	DeleteByFile(ctx context.Context, fileID shared.ID) error
	Search(ctx context.Context, projectID shared.ID, query []float32, topK int) ([]*SearchResult, error)
}

type SearchResult struct {
	ChunkID shared.ID
	Content string
	Score   float32
}
