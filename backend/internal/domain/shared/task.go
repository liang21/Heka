package shared

import (
	"encoding/json"
	"time"
)

type AsyncTask struct {
	ID              ID
	ProjectID       ID
	Type            string
	Status          string
	ProgressCurrent int
	ProgressTotal   int
	Input           json.RawMessage
	Result          json.RawMessage
	Error           string
	CreatedBy       ID
	CreatedAt       time.Time
	StartedAt       *time.Time
	CompletedAt     *time.Time
}

type IndexTask struct {
	ID         ID
	FileID     ID
	Status     string
	RetryCount int
	MaxRetries int
	Error      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	CompletedAt *time.Time
}
