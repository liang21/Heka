package plan

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/plan"
)

// tasks.md: T108 | spec.md: §4.8 测试计划 DTO

type PlanCaseItem struct {
	TestCaseID shared.ID  `json:"test_case_id" validate:"required,uuid"`
	AssignedTo *shared.ID `json:"assigned_to"`
}

type CreatePlanReq struct {
	ProjectID   shared.ID      `json:"project_id" validate:"required,uuid"`
	Name        string         `json:"name" validate:"required"`
	Description string         `json:"description"`
	Cases       []PlanCaseItem `json:"cases" validate:"required,min=1"`
}

type PlanResponse struct {
	ID          shared.ID       `json:"id"`
	ProjectID   shared.ID       `json:"project_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      plan.PlanStatus `json:"status"`
	CreatedBy   shared.ID       `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type PlanDetailResponse struct {
	ID                 shared.ID       `json:"id"`
	ProjectID          shared.ID       `json:"project_id"`
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	Status             plan.PlanStatus `json:"status"`
	CurrentExecutionID *shared.ID      `json:"current_execution_id"`
	StartedAt          *time.Time      `json:"started_at"`
	PausedAt           *time.Time      `json:"paused_at"`
	EndedAt            *time.Time      `json:"ended_at"`
	CreatedBy          shared.ID       `json:"created_by"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}
