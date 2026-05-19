// tasks.md: T064, T065 | spec.md: Vector repository with Milvus
package milvus

import (
	"context"
	"fmt"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
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
	client *Client
}

// NewVectorRepository creates a new VectorRepository
func NewVectorRepository(client *Client) rag.VectorRepository {
	repo := &VectorRepository{
		client: client,
	}

	// Initialize collection on startup
	go repo.ensureCollection(context.Background())

	return repo
}

// ensureCollection creates the collection if it doesn't exist
func (r *VectorRepository) ensureCollection(ctx context.Context) error {
	// Check if collection exists
	has, err := r.client.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if has {
		// Collection exists, verify it's loaded
		return r.client.LoadCollection(ctx, collectionName, false)
	}

	// Create schema
	schema := &entity.Schema{
		CollectionName: collectionName,
		Fields: []*entity.Field{
			{
				Name:       chunkIDField,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "36"},
			},
			{
				Name:       fileIDField,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "36"},
			},
			{
				Name:     vectorField,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dimension),
				},
			},
		},
	}

	// Create collection
	if err := r.client.CreateCollection(ctx, collectionName, schema, 1); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Wait for collection to be ready
	for i := 0; i < 10; i++ {
		has, err := r.client.HasCollection(ctx, collectionName)
		if err == nil && has {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Create IVF_FLAT index
	idx, err := entity.NewIndexIvfFlat(entity.L2, 64)
	if err != nil {
		return fmt.Errorf("failed to create index config: %w", err)
	}

	if err := r.client.CreateIndex(ctx, collectionName, vectorField, idx, false); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Load collection into memory
	return r.client.LoadCollection(ctx, collectionName, false)
}

// Upsert inserts or updates vectors for chunks
func (r *VectorRepository) Upsert(ctx context.Context, chunks []*rag.DocumentChunk, embeddings [][]float32) error {
	if len(chunks) != len(embeddings) {
		return fmt.Errorf("chunks and embeddings count mismatch")
	}

	if len(chunks) == 0 {
		return nil
	}

	// Prepare column data
	chunkIDs := make([]string, len(chunks))
	fileIDs := make([]string, len(chunks))
	vectors := make([][]float32, len(embeddings))

	for i, chunk := range chunks {
		chunkIDs[i] = chunk.ID.String()
		fileIDs[i] = chunk.FileID.String()
		vectors[i] = embeddings[i]
	}

	// Create columns
	chunkIDColumn := entity.NewColumnVarChar(chunkIDField, chunkIDs)
	fileIDColumn := entity.NewColumnVarChar(fileIDField, fileIDs)
	vectorColumn := entity.NewColumnFloatVector(vectorField, dimension, vectors)

	// Insert data
	if err := r.client.Insert(ctx, collectionName, "", chunkIDColumn, fileIDColumn, vectorColumn); err != nil {
		return fmt.Errorf("failed to insert vectors: %w", err)
	}

	// Flush to ensure data is persisted
	return r.client.Flush(ctx, collectionName)
}

// DeleteByFile deletes all vectors for a specific file
func (r *VectorRepository) DeleteByFile(ctx context.Context, fileID shared.ID) error {
	expr := fmt.Sprintf("%s == '%s'", fileIDField, fileID.String())

	if err := r.client.Delete(ctx, collectionName, expr); err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	// Flush to ensure deletion is persisted
	return r.client.Flush(ctx, collectionName)
}

// Search performs similarity search
func (r *VectorRepository) Search(ctx context.Context, projectID shared.ID, query []float32, topK int) ([]*rag.SearchResult, error) {
	// Note: projectID is passed but we don't filter by it in the current schema
	// In a production system, you'd need to either:
	// 1. Add project_id to the schema and filter by it
	// 2. Maintain separate collections per project
	// 3. Implement a two-phase search (filter then search)

	vectors := []entity.Vector{entity.FloatVector(query)}

	// Search parameters for IVF_FLAT
	searchParams := entity.NewIndexIvfFlatSearchParam(16) // nprobe

	results, err := r.client.Search(
		ctx,
		collectionName,
		nil, // partitions
		"",  // expr (no filtering for now)
		[]string{chunkIDField, fileIDField},
		vectors,
		vectorField,
		entity.L2,
		topK,
		searchParams,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	if len(results) == 0 {
		return nil, shared.ErrRAGNoResults
	}

	// Convert results
	searchResults := make([]*rag.SearchResult, 0, len(results[0]))
	for _, result := range results[0] {
		searchResults = append(searchResults, &rag.SearchResult{
			ChunkID: shared.ID(result.ID.(string)),
			Score:   result.Score,
		})
	}

	return searchResults, nil
}
