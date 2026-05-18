package project

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

type Project struct {
	ID          shared.ID
	Name        string
	Description string
	CreatedBy   shared.ID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type ProjectMember struct {
	ProjectID shared.ID
	UserID    shared.ID
	JoinedAt  time.Time
}
