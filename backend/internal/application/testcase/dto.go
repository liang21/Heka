package testcase

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
)

// tasks.md: T105 | spec.md: §4.4-4.7 测试用例 DTO

// --- Module DTOs ---

type CreateModuleReq struct {
	ProjectID   shared.ID  `json:"project_id" validate:"required,uuid"`
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description"`
	ParentID    *shared.ID `json:"parent_id"`
}

type UpdateModuleReq struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type ModuleDTO struct {
	ID          shared.ID           `json:"id"`
	ProjectID   shared.ID           `json:"project_id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	ParentID    *shared.ID          `json:"parent_id"`
	OrderIndex  int                 `json:"order_index"`
	Children    []ModuleDTO         `json:"children"`
}

// --- Tag DTOs ---

type CreateTagReq struct {
	ProjectID shared.ID `json:"project_id" validate:"required,uuid"`
	Name      string    `json:"name" validate:"required"`
	Color     string    `json:"color"`
}

type TagDTO struct {
	ID        shared.ID `json:"id"`
	ProjectID shared.ID `json:"project_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
}

// --- TestCase DTOs ---

type StepInput struct {
	Action   string `json:"action" validate:"required"`
	Expected string `json:"expected" validate:"required"`
}

type CreateTestCaseReq struct {
	ProjectID   shared.ID             `json:"project_id" validate:"required,uuid"`
	ModuleID    *shared.ID            `json:"module_id"`
	Title       string                `json:"title" validate:"required"`
	Description string                `json:"description"`
	Priority    testcase.Priority     `json:"priority"`
	Tags        []string              `json:"tags"`
	Steps       []StepInput           `json:"steps" validate:"required,dive"`
}

type UpdateTestCaseReq struct {
	ModuleID    *shared.ID            `json:"module_id"`
	Title       string                `json:"title" validate:"required"`
	Description string                `json:"description"`
	Priority    testcase.Priority     `json:"priority"`
	Tags        []string              `json:"tags"`
	Steps       []StepInput           `json:"steps" validate:"required,dive"`
	Version     int                   `json:"version"`
}

type TestCaseResponse struct {
	ID          shared.ID             `json:"id"`
	ProjectID   shared.ID             `json:"project_id"`
	ModuleID    *shared.ID            `json:"module_id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Status      testcase.CaseStatus   `json:"status"`
	Priority    testcase.Priority     `json:"priority"`
	Tags        []string              `json:"tags"`
	Steps       []StepResponse        `json:"steps"`
	CreatedBy   shared.ID             `json:"created_by"`
	UpdatedBy   *shared.ID            `json:"updated_by"`
	Version     int                   `json:"version"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

type StepResponse struct {
	ID         shared.ID `json:"id"`
	Number     int       `json:"number"`
	Action     string    `json:"action"`
	Expected   string    `json:"expected"`
}

type TestCaseListResponse struct {
	ID        shared.ID           `json:"id"`
	ProjectID shared.ID           `json:"project_id"`
	ModuleID  *shared.ID          `json:"module_id"`
	Title     string              `json:"title"`
	Status    testcase.CaseStatus `json:"status"`
	Priority  testcase.Priority   `json:"priority"`
	Tags      []string            `json:"tags"`
	CreatedBy shared.ID           `json:"created_by"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type TestCaseFilter struct {
	ProjectID shared.ID  `json:"project_id"`
	ModuleID  *shared.ID `json:"module_id"`
	TagID     *shared.ID `json:"tag_id"`
	Status    string     `json:"status"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
}

type BatchStatusReq struct {
	IDs    []shared.ID         `json:"ids" validate:"required,min=1"`
	Status testcase.CaseStatus `json:"status" validate:"required"`
}

type BatchDeleteReq struct {
	IDs []shared.ID `json:"ids" validate:"required,min=1"`
}

type BatchMoveReq struct {
	IDs      []shared.ID  `json:"ids" validate:"required,min=1"`
	ModuleID *shared.ID   `json:"module_id"`
}

// --- Collection DTOs ---

type CreateCollectionReq struct {
	ProjectID   shared.ID `json:"project_id" validate:"required,uuid"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
}

type CollectionDTO struct {
	ID          shared.ID `json:"id"`
	ProjectID   shared.ID `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type AddCollectionCasesReq struct {
	CaseIDs []shared.ID `json:"case_ids" validate:"required,min=1"`
}

type RemoveCollectionCasesReq struct {
	CaseIDs []shared.ID `json:"case_ids" validate:"required,min=1"`
}
