package postgres

import (
	"context"
	"time"

	"github.com/liang21/heka/internal/domain/rag"
	"github.com/liang21/heka/internal/domain/shared"
	"gorm.io/gorm"
)

type chunkRepo struct {
	db *gorm.DB
}

// NewChunkRepository creates a new ChunkRepository instance
func NewChunkRepository(db *gorm.DB) rag.ChunkRepository {
	return &chunkRepo{db: db}
}

type chunkModel struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	FileID    string    `gorm:"index;type:varchar(36);not null"`
	Content   string    `gorm:"type:text;not null"`
	Index     int       `gorm:"not null"`
	Tokens    int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

func (m chunkModel) TableName() string {
	return "chunks"
}

func toChunkModel(chunk *rag.DocumentChunk) chunkModel {
	return chunkModel{
		ID:        chunk.ID.String(),
		FileID:    chunk.FileID.String(),
		Content:   chunk.Content,
		Index:     chunk.Index,
		Tokens:    chunk.Tokens,
		CreatedAt: chunk.CreatedAt,
	}
}

func toChunkEntity(model chunkModel) *rag.DocumentChunk {
	return &rag.DocumentChunk{
		ID:        shared.ID(model.ID),
		FileID:    shared.ID(model.FileID),
		Content:   model.Content,
		Index:     model.Index,
		Tokens:    model.Tokens,
		CreatedAt: model.CreatedAt,
	}
}

func toChunkEntities(models []chunkModel) []*rag.DocumentChunk {
	chunks := make([]*rag.DocumentChunk, len(models))
	for i, model := range models {
		chunks[i] = toChunkEntity(model)
	}
	return chunks
}

// CreateBatch implements efficient batch insertion using GORM's CreateInBatches
func (r *chunkRepo) CreateBatch(ctx context.Context, chunks []*rag.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil // Handle empty batches gracefully
	}

	models := make([]chunkModel, len(chunks))
	for i, chunk := range chunks {
		models[i] = toChunkModel(chunk)
	}

	db := DBOrTx(ctx, r.db)

	// Use CreateInBatches for efficient bulk insert
	// Batch size of 100 is optimal for PostgreSQL
	batchSize := 100
	if len(chunks) < batchSize {
		batchSize = len(chunks)
	}

	if err := db.CreateInBatches(models, batchSize).Error; err != nil {
		return err
	}

	return nil
}

// FindByFile retrieves all chunks for a given file ID, ordered by index ASC
func (r *chunkRepo) FindByFile(ctx context.Context, fileID shared.ID) ([]*rag.DocumentChunk, error) {
	db := DBOrTx(ctx, r.db)

	var models []chunkModel
	if err := db.Where("file_id = ?", fileID.String()).
		Order("index ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	// Return empty slice instead of nil when no chunks found
	if len(models) == 0 {
		return []*rag.DocumentChunk{}, nil
	}

	return toChunkEntities(models), nil
}

// DeleteByFile deletes all chunks associated with a given file ID
func (r *chunkRepo) DeleteByFile(ctx context.Context, fileID shared.ID) error {
	db := DBOrTx(ctx, r.db)

	// Delete all chunks for the file - this is idempotent (no error if not found)
	return db.Where("file_id = ?", fileID.String()).Delete(&chunkModel{}).Error
}
