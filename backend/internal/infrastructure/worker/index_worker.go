// tasks.md: T095 | spec.md: Index worker for async file indexing
package worker

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/liang21/heka/internal/domain/file"
	"github.com/liang21/heka/internal/domain/rag"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/infrastructure/ai"
	"github.com/liang21/heka/internal/infrastructure/parser"
)

// IndexWorker handles async file indexing tasks
type IndexWorker struct {
	fileRepo      file.FileRepository
	chunkRepo     rag.ChunkRepository
	vectorRepo    rag.VectorRepository
	indexTaskRepo shared.IndexTaskRepository
	parser        *parser.Registry
	chunker       *parser.Chunker
	embedding     *ai.EmbeddingClient
	eventBus      shared.EventBus
	ticker        *time.Ticker
}

// NewIndexWorker creates a new index worker
func NewIndexWorker(
	fileRepo file.FileRepository,
	chunkRepo rag.ChunkRepository,
	vectorRepo rag.VectorRepository,
	indexTaskRepo shared.IndexTaskRepository,
	parser *parser.Registry,
	chunker *parser.Chunker,
	embedding *ai.EmbeddingClient,
	eventBus shared.EventBus,
) *IndexWorker {
	return &IndexWorker{
		fileRepo:      fileRepo,
		chunkRepo:     chunkRepo,
		vectorRepo:    vectorRepo,
		indexTaskRepo: indexTaskRepo,
		parser:        parser,
		chunker:       chunker,
		embedding:     embedding,
		eventBus:      eventBus,
	}
}

// Start begins the worker loop
func (w *IndexWorker) Start(ctx context.Context) {
	w.ticker = time.NewTicker(5 * time.Second)
	defer w.ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.ticker.C:
			w.processTasks(ctx)
		}
	}
}

// processTasks processes pending index tasks
func (w *IndexWorker) processTasks(ctx context.Context) {
	// Fetch pending tasks
	tasks, err := w.indexTaskRepo.FindPending(ctx, 10)
	if err != nil {
		fmt.Printf("Failed to fetch pending tasks: %v\n", err)
		return
	}

	// Fetch stale tasks (compensation for missed tasks)
	staleTasks, err := w.indexTaskRepo.FindStale(ctx, "-10m", 5)
	if err != nil {
		fmt.Printf("Failed to fetch stale tasks: %v\n", err)
	} else {
		tasks = append(tasks, staleTasks...)
	}

	for _, task := range tasks {
		w.processTask(ctx, task)
	}
}

// processTask processes a single index task
func (w *IndexWorker) processTask(ctx context.Context, task *shared.IndexTask) {
	// Update status to processing
	task.Status = "processing"
	if err := w.indexTaskRepo.Update(ctx, task); err != nil {
		fmt.Printf("Failed to update task status: %v\n", err)
		return
	}

	// Fetch file
	fileEntity, err := w.fileRepo.FindByID(ctx, task.FileID)
	if err != nil {
		w.failTask(ctx, task, fmt.Sprintf("failed to fetch file: %v", err))
		return
	}

	// Parse file
	text, err := w.parseFile(ctx, fileEntity)
	if err != nil {
		w.failTask(ctx, task, fmt.Sprintf("failed to parse file: %v", err))
		return
	}

	// Chunk content
	chunks := w.chunker.Chunk(text, rag.DefaultChunkConfig())
	if len(chunks) == 0 {
		w.failTask(ctx, task, "no chunks generated")
		return
	}

	// Assign IDs and FileID to chunks
	for i, chunk := range chunks {
		chunk.ID = shared.NewID()
		chunk.FileID = fileEntity.ID
		chunk.Index = i
	}

	// Generate embeddings
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	embeddings, err := w.embedding.BatchEmbed(ctx, texts, 100)
	if err != nil {
		w.failTask(ctx, task, fmt.Sprintf("failed to generate embeddings: %v", err))
		return
	}

	// Save chunks to database
	if err := w.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		w.failTask(ctx, task, fmt.Sprintf("failed to save chunks: %v", err))
		return
	}

	// Upsert vectors to Milvus
	if err := w.vectorRepo.Upsert(ctx, chunks, embeddings); err != nil {
		w.failTask(ctx, task, fmt.Sprintf("failed to upsert vectors: %v", err))
		return
	}

	// Update file index status
	if err := w.fileRepo.UpdateIndexStatus(ctx, fileEntity.ID, file.IndexCompleted, ""); err != nil {
		fmt.Printf("Failed to update file index status: %v\n", err)
	}

	// Mark task as completed
	task.Status = "completed"
	now := time.Now()
	task.CompletedAt = &now
	if err := w.indexTaskRepo.Update(ctx, task); err != nil {
		fmt.Printf("Failed to mark task as completed: %v\n", err)
	}

	// Publish event
	if w.eventBus != nil {
		event := &shared.IndexCompletedEvent{
			FileID: fileEntity.ID,
			At:     time.Now(),
		}
		if err := w.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish index completed event: %v\n", err)
		}
	}
}

// parseFile parses a file based on its type
func (w *IndexWorker) parseFile(ctx context.Context, fileEntity *file.File) (string, error) {
	switch fileEntity.SourceType {
	case file.SourceFigma:
		// For Figma files, the content should already be in ContentPreview
		return fileEntity.ContentPreview, nil

	case file.SourceUpload:
		// For uploaded files, read from disk and parse
		return w.parseUploadedFile(ctx, fileEntity)

	default:
		return "", fmt.Errorf("unsupported source type: %s", fileEntity.SourceType)
	}
}

// parseUploadedFile parses an uploaded file
func (w *IndexWorker) parseUploadedFile(ctx context.Context, fileEntity *file.File) (string, error) {
	// Open file
	file, err := os.Open(fileEntity.Path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get parser for file type
	p, err := w.parser.GetParser(string(fileEntity.Type))
	if err != nil {
		return "", fmt.Errorf("failed to get parser: %w", err)
	}

	// Parse file
	text, err := p.Parse(ctx, file)
	if err != nil {
		return "", fmt.Errorf("failed to parse: %w", err)
	}

	return text, nil
}

// failTask marks a task as failed and increments retry count
func (w *IndexWorker) failTask(ctx context.Context, task *shared.IndexTask, errMsg string) {
	task.RetryCount++
	task.Error = errMsg

	if task.RetryCount >= task.MaxRetries {
		task.Status = "failed"
	} else {
		task.Status = "pending"
	}

	if err := w.indexTaskRepo.Update(ctx, task); err != nil {
		fmt.Printf("Failed to update failed task: %v\n", err)
	}

	// Publish failed event
	if w.eventBus != nil {
		event := &shared.IndexFailedEvent{
			FileID: task.FileID,
			Error:  errMsg,
			At:     time.Now(),
		}
		if err := w.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish index failed event: %v\n", err)
		}
	}
}

// Stop stops the worker
func (w *IndexWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
}
