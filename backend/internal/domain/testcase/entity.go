package testcase

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

type Module struct {
	ID          shared.ID
	ProjectID   shared.ID
	Name        string
	Description string
	ParentID    *shared.ID
	OrderIndex  int
	CreatedBy   shared.ID
}

type Tag struct {
	ID        shared.ID
	ProjectID shared.ID
	Name      string
	Color     string
	CreatedBy shared.ID
}

type Step struct {
	ID         shared.ID
	TestCaseID shared.ID
	Number     int
	Action     string
	Expected   string
}

type TestCase struct {
	ID         shared.ID
	ProjectID  shared.ID
	ModuleID   *shared.ID
	Title      string
	Description string
	Status     CaseStatus
	Priority   Priority
	Tags       []string
	Steps      []Step
	CreatedBy  shared.ID
	UpdatedBy  *shared.ID
	Version    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

type Collection struct {
	ID          shared.ID
	ProjectID   shared.ID
	Name        string
	Description string
	CreatedBy   shared.ID
	CreatedAt   time.Time
}

type CollectionCase struct {
	CollectionID shared.ID
	TestCaseID   shared.ID
	AddedAt      time.Time
}
