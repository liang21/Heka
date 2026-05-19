package shared

import (
	"encoding/json"
	"time"
)

type AITaskCompletedEvent struct {
	TaskID ID
	Result json.RawMessage
	At     time.Time
}

func (e *AITaskCompletedEvent) EventName() string {
	return "ai.task.completed"
}

func (e *AITaskCompletedEvent) OccurredAt() time.Time {
	return e.At
}

type AITaskFailedEvent struct {
	TaskID ID
	Error  string
	At     time.Time
}

func (e *AITaskFailedEvent) EventName() string {
	return "ai.task.failed"
}

func (e *AITaskFailedEvent) OccurredAt() time.Time {
	return e.At
}

type IndexCompletedEvent struct {
	FileID ID
	At     time.Time
}

func (e *IndexCompletedEvent) EventName() string {
	return "file.index.completed"
}

func (e *IndexCompletedEvent) OccurredAt() time.Time {
	return e.At
}

type IndexFailedEvent struct {
	FileID ID
	Error  string
	At     time.Time
}

func (e *IndexFailedEvent) EventName() string {
	return "file.index.failed"
}

func (e *IndexFailedEvent) OccurredAt() time.Time {
	return e.At
}
