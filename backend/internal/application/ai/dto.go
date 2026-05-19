package ai

import (
	"encoding/json"

	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T120 | spec.md: §4.11 AI 功能 DTO

type GenerateRequest struct {
	ProjectID      shared.ID `json:"project_id" validate:"required,uuid"`
	FileID         shared.ID `json:"file_id" validate:"required,uuid"`
	Count          int       `json:"count"`
	IncludeNegative bool     `json:"include_negative"`
}

type AnalyzeRequest struct {
	ProjectID   shared.ID `json:"project_id" validate:"required,uuid"`
	Description string    `json:"description" validate:"required"`
}

type TaskResponse struct {
	TaskID    shared.ID       `json:"task_id"`
	Status    string          `json:"status"`
	Progress  int             `json:"progress"`
	Total     int             `json:"total"`
	Result    json.RawMessage `json:"result"`
	Error     string          `json:"error"`
}
