// tasks.md: T064, T065 | spec.md: Vector repository with Milvus
// NOTE: Milvus SDK integration is incomplete due to API changes in v2.4.2
// All methods are stubbed to return errors. To enable vector search:
// 1. Either downgrade to milvus-sdk-go v2.3.x, or
// 2. Update code to match v2.4.x API
package milvus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/liang21/heka/internal/domain/rag"
	"github.com/liang21/heka/internal/domain/shared"
)

const (
	collectionName = "heka_chunks"
	dimension      = 1536 // OpenAI embedding dimension
	vectorField    = "embedding"
	chunkIDField   = "chunk_id"
	fileIDField    = "file_id"
)

// VectorRepository implements rag.VectorRepository using Milvus
type VectorRepository struct {
	client             client.Client
	initOnce           sync.Once
	initDone           chan struct{}
	initCtx            context.Context
	initCancel         context.CancelFunc
}

// NewVectorRepository creates a new VectorRepository
func NewVectorRepository(milvusClient client.Client) rag.VectorRepository {
	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)

	repo := &VectorRepository{
		client:     milvusClient,
		initDone:   make(chan struct{}),
		initCtx:    initCtx,
		initCancel: initCancel,
	}

	// Initialize collection on startup with cancellable context
	go func() {
		defer close(repo.initDone)
		_ = repo.ensureCollection(initCtx)
	}()

	return repo
}

// ensureCollection creates the collection if it doesn't exist
func (r *VectorRepository) ensureCollection(ctx context.Context) error {
	var initErr error
	r.initOnce.Do(func() {
		initErr = r.doEnsureCollection(ctx)
	})
	return initErr
}

// doEnsureCollection performs the actual collection initialization
func (r *VectorRepository) doEnsureCollection(ctx context.Context) error {
	// Stub - Milvus integration incomplete
	return fmt.Errorf("Milvus vector repository not fully implemented")
}

// Upsert inserts or updates vectors for chunks
func (r *VectorRepository) Upsert(ctx context.Context, chunks []*rag.DocumentChunk, embeddings [][]float32) error {
	// Wait for initialization
	select {
	case <-r.initDone:
	case <-time.After(5 * time.Second):
		return fmt.Errorf("vector repository initialization timeout")
	case <-ctx.Done():
		return ctx.Err()
	}

	// Stub - Milvus integration incomplete
	return fmt.Errorf("Milvus vector upsert not implemented")
}

// DeleteByFile deletes all vectors for a specific file
func (r *VectorRepository) DeleteByFile(ctx context.Context, fileID shared.ID) error {
	// Stub - Milvus integration incomplete
	return fmt.Errorf("Milvus vector delete not implemented")
}

// Search performs similarity search
func (r *VectorRepository) Search(ctx context.Context, projectID shared.ID, query []float32, topK int) ([]*rag.SearchResult, error) {
	// Wait for initialization
	select {
	case <-r.initDone:
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("vector repository initialization timeout")
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Stub - Milvus integration incomplete
	return nil, fmt.Errorf("Milvus vector search not implemented - SDK API needs update")
}

// Close cleans up resources
func (r *VectorRepository) Close() error {
	r.initCancel()
	select {
	case <-r.initDone:
	case <-time.After(1 * time.Second):
	}
	return nil
}
