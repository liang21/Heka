package rag

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

type DocumentChunk struct {
	ID        shared.ID
	FileID    shared.ID
	Content   string
	Index     int
	Tokens    int
	CreatedAt time.Time
}

type VectorEmbedding struct {
	ID        shared.ID
	ChunkID   shared.ID
	Model     string
	Dimension int
	MilvusID  string
	CreatedAt time.Time
}

type ChunkConfig struct {
	MaxTokens    int
	Overlap      int
	MinChunkSize int
}

func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		MaxTokens:    500,
		Overlap:      50,
		MinChunkSize: 100,
	}
}
