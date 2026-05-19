// tasks.md: T096 | spec.md: AI task worker for async AI operations
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/infrastructure/ai"
)

// AIWorker handles async AI task processing
type AIWorker struct {
	manager  *ai.Manager
	taskRepo shared.AsyncTaskRepository
	eventBus shared.EventBus
	ticker   *time.Ticker
}

// NewAIWorker creates a new AI task worker
func NewAIWorker(
	manager *ai.Manager,
	taskRepo shared.AsyncTaskRepository,
	eventBus shared.EventBus,
) *AIWorker {
	return &AIWorker{
		manager:  manager,
		taskRepo: taskRepo,
		eventBus: eventBus,
	}
}

// Start begins the worker loop
func (w *AIWorker) Start(ctx context.Context) {
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

// processTasks processes pending AI tasks
func (w *AIWorker) processTasks(ctx context.Context) {
	// Fetch pending tasks (limit 10 per batch)
	tasks, err := w.taskRepo.FindPendingByType(ctx, "", "ai", 10)
	if err != nil {
		fmt.Printf("Failed to fetch pending AI tasks: %v\n", err)
		return
	}

	for _, task := range tasks {
		w.processTask(ctx, task)
	}
}

// processTask processes a single AI task
func (w *AIWorker) processTask(ctx context.Context, task *shared.AsyncTask) {
	// Update status to in_progress
	task.Status = "in_progress"
	now := time.Now()
	task.StartedAt = &now

	if err := w.taskRepo.Update(ctx, task); err != nil {
		fmt.Printf("Failed to update task status: %v\n", err)
		return
	}

	// Parse input to get chat request
	var chatReq ai.ChatRequest
	if err := json.Unmarshal(task.Input, &chatReq); err != nil {
		w.failTask(ctx, task, fmt.Sprintf("invalid input format: %v", err))
		return
	}

	// Call LLM via Manager
	resp, err := w.manager.Chat(ctx, chatReq)
	if err != nil {
		w.failTask(ctx, task, fmt.Sprintf("AI call failed: %v", err))
		return
	}

	// Serialize result
	resultJSON, err := json.Marshal(resp)
	if err != nil {
		w.failTask(ctx, task, fmt.Sprintf("failed to serialize result: %v", err))
		return
	}

	// Update task with result
	task.Status = "completed"
	task.Result = resultJSON
	task.ProgressCurrent = task.ProgressTotal
	completedAt := time.Now()
	task.CompletedAt = &completedAt

	if err := w.taskRepo.Update(ctx, task); err != nil {
		fmt.Printf("Failed to update task with result: %v\n", err)
		return
	}

	// Publish completion event
	if w.eventBus != nil {
		event := &shared.AITaskCompletedEvent{
			TaskID: task.ID,
			Result: resultJSON,
			At:     time.Now(),
		}
		if err := w.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish AI task completed event: %v\n", err)
		}
	}
}

// failTask marks a task as failed
func (w *AIWorker) failTask(ctx context.Context, task *shared.AsyncTask, errMsg string) {
	task.Status = "failed"
	task.Error = errMsg
	completedAt := time.Now()
	task.CompletedAt = &completedAt

	if err := w.taskRepo.Update(ctx, task); err != nil {
		fmt.Printf("Failed to update failed task: %v\n", err)
	}

	// Publish failed event
	if w.eventBus != nil {
		event := &shared.AITaskFailedEvent{
			TaskID: task.ID,
			Error:  errMsg,
			At:     time.Now(),
		}
		if err := w.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish AI task failed event: %v\n", err)
		}
	}
}

// Stop stops the worker
func (w *AIWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
}
