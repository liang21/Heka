package rag

import (
	"context"

	"github.com/liang21/heka/internal/domain/rag"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/infrastructure/ai"
)

// tasks.md: T119 | spec.md: §4.11 RAGService Implementation

type Service struct {
	vectorRepo      rag.VectorRepository
	embeddingClient *ai.EmbeddingClient
}

func NewService(vectorRepo rag.VectorRepository, embeddingClient *ai.EmbeddingClient) *Service {
	return &Service{
		vectorRepo:      vectorRepo,
		embeddingClient: embeddingClient,
	}
}

func (s *Service) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	if req.Query == "" {
		return nil, shared.ErrSysValidation
	}

	if req.TopK <= 0 {
		req.TopK = 10 // default
	}

	embedding, err := s.embeddingClient.EmbedSingle(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	results, err := s.vectorRepo.Search(ctx, req.ProjectID, embedding, req.TopK)
	if err != nil {
		return nil, err
	}

	items := make([]SearchResultItem, len(results))
	for i, r := range results {
		items[i] = SearchResultItem{
			ChunkID:  r.ChunkID,
			Content:  r.Content,
			Score:    r.Score,
		}
	}

	return &SearchResponse{Results: items}, nil
}
