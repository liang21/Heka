package testcase

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

type TestCaseCreatedEvent struct {
	ProjectID  shared.ID
	TestCaseID shared.ID
	time       time.Time
}

func (e *TestCaseCreatedEvent) EventName() string {
	return "testcase.created"
}

func (e *TestCaseCreatedEvent) OccurredAt() time.Time {
	return e.time
}

type TestCaseUpdatedEvent struct {
	ProjectID  shared.ID
	TestCaseID shared.ID
	time       time.Time
}

func (e *TestCaseUpdatedEvent) EventName() string {
	return "testcase.updated"
}

func (e *TestCaseUpdatedEvent) OccurredAt() time.Time {
	return e.time
}

type TestCaseDeletedEvent struct {
	ProjectID  shared.ID
	TestCaseID shared.ID
	time       time.Time
}

func (e *TestCaseDeletedEvent) EventName() string {
	return "testcase.deleted"
}

func (e *TestCaseDeletedEvent) OccurredAt() time.Time {
	return e.time
}
