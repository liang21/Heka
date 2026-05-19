package plan

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

type PlanTestCase struct {
	PlanID     shared.ID
	TestCaseID shared.ID
	AssignedTo *shared.ID
	OrderIndex int
}

type TestPlan struct {
	ID                shared.ID
	ProjectID         shared.ID
	Name              string
	Description       string
	Status            PlanStatus
	CurrentExecutionID *shared.ID
	StartedAt         *time.Time
	PausedAt          *time.Time
	EndedAt           *time.Time
	CreatedBy         shared.ID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	Cases             []PlanTestCase
}
