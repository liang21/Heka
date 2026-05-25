package testcase

import (
	"context"

	"github.com/liang21/heka/internal/domain/shared"
)

type ModuleRepository interface {
	Create(ctx context.Context, m *Module) error
	FindByID(ctx context.Context, id shared.ID) (*Module, error)
	Update(ctx context.Context, m *Module) error
	Delete(ctx context.Context, id shared.ID) error
	FindByProject(ctx context.Context, projectID shared.ID) ([]*Module, error)
}

type TagRepository interface {
	FindByProject(ctx context.Context, projectID shared.ID) ([]*Tag, error)
	Create(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, id shared.ID) error
}

type TestCaseRepository interface {
	Create(ctx context.Context, tc *TestCase) error
	FindByID(ctx context.Context, id shared.ID) (*TestCase, error)
	List(ctx context.Context, filter TestCaseFilter) ([]*TestCase, int64, error)
	Update(ctx context.Context, tc *TestCase) error
	SoftDelete(ctx context.Context, id shared.ID) error
	BatchUpdateStatus(ctx context.Context, ids []shared.ID, status CaseStatus) error
	BatchDelete(ctx context.Context, ids []shared.ID) error
	BatchMove(ctx context.Context, ids []shared.ID, moduleID *shared.ID) error
}

type CollectionRepository interface {
	Create(ctx context.Context, c *Collection) error
	AddCases(ctx context.Context, collectionID shared.ID, caseIDs []shared.ID) error
	RemoveCases(ctx context.Context, collectionID shared.ID, caseIDs []shared.ID) error
	ListCases(ctx context.Context, collectionID shared.ID, page, pageSize int) ([]*TestCase, int64, error)
}

type TestCaseFilter struct {
	ProjectID  shared.ID
	ModuleID   *shared.ID
	Status     *CaseStatus
	Priority   *Priority
	Keyword    string
	Tags       []string
	Page       int
	PageSize   int
	SortBy     string
	SortDesc   bool
}
