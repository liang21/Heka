package shared

import (
	"encoding/json"
	"time"
)

type AITaskCompletedEvent struct {
	TaskID ID
	Result json.RawMessage
	time   time.Time
}

func (e *AITaskCompletedEvent) EventName() string {
	return "ai.task.completed"
}

func (e *AITaskCompletedEvent) OccurredAt() time.Time {
	return e.time
}

type AITaskFailedEvent struct {
	TaskID ID
	Error  string
	time   time.Time
}

func (e *AITaskFailedEvent) EventName() string {
	return "ai.task.failed"
}

func (e *AITaskFailedEvent) OccurredAt() time.Time {
	return e.time
}
