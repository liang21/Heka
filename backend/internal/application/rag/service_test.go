package rag

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liang21/heka/internal/domain/rag"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T118 | spec.md: §4.11 RAG Service TDD RED

type mockVectorRepo struct {
	mock.Mock
}

func (m *mockVectorRepo) Upsert(ctx context.Context, chunks []*rag.DocumentChunk, embeddings [][]float32) error {
	args := m.Called(ctx, chunks, embeddings)
	return args.Error(0)
}

func (m *mockVectorRepo) DeleteByFile(ctx context.Context, fileID shared.ID) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *mockVectorRepo) Search(ctx context.Context, projectID shared.ID, query []float32, topK int) ([]*rag.SearchResult, error) {
	args := m.Called(ctx, projectID, query, topK)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*rag.SearchResult), args.Error(1)
}

func TestRAGService_Search(t *testing.T) {
	t.Parallel()
	vecRepo := new(mockVectorRepo)
	svc := NewService(vecRepo, nil) // nil embedding client for now

	projectID := shared.NewID()

	results := []*rag.SearchResult{
		{ChunkID: shared.NewID(), Content: "relevant text", Score: 0.95},
		{ChunkID: shared.NewID(), Content: "more text", Score: 0.85},
	}

	vecRepo.On("Search", mock.Anything, projectID, mock.Anything, 5).Return(results, nil)

	// We pass pre-computed embedding via a helper or the service computes it
	resp, err := svc.Search(context.Background(), SearchRequest{
		ProjectID: projectID,
		Query:     "test query",
		TopK:      5,
	})

	require.NoError(t, err)
	assert.Len(t, resp.Results, 2)
}

func TestRAGService_SearchEmptyQuery(t *testing.T) {
	t.Parallel()
	vecRepo := new(mockVectorRepo)
	svc := NewService(vecRepo, nil)

	resp, err := svc.Search(context.Background(), SearchRequest{
		ProjectID: shared.NewID(),
		Query:     "",
		TopK:      5,
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
}
