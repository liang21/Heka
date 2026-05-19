package execution

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/execution"
)

// tasks.md: T111 | spec.md: §4.9 执行记录 DTO

type SubmitResultReq struct {
	TestCaseID shared.ID            `json:"test_case_id" validate:"required,uuid"`
	Status     execution.ResultStatus `json:"status" validate:"required"`
	BugID      string               `json:"bug_id"`
	BugURL     string               `json:"bug_url"`
	Notes      string               `json:"notes"`
}

type BatchSubmitReq struct {
	Results []SubmitResultReq `json:"results" validate:"required,min=1,dive"`
}

type ExecutionResponse struct {
	ID          shared.ID          `json:"id"`
	PlanID      shared.ID          `json:"plan_id"`
	Name        string             `json:"name"`
	Status      execution.ExecStatus `json:"status"`
	ExecutorID  shared.ID          `json:"executor_id"`
	StartedAt   time.Time          `json:"started_at"`
	PausedAt    *time.Time         `json:"paused_at"`
	CompletedAt *time.Time         `json:"completed_at"`
	Notes       string             `json:"notes"`
}

type ExecutionSummaryResponse struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Blocked int `json:"blocked"`
	Skipped int `json:"skipped"`
}
