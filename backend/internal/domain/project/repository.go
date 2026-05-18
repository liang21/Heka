package project

import (
	"context"

	"github.com/liang21/heka/internal/domain/shared"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	FindByID(ctx context.Context, id shared.ID) (*Project, error)
	FindByUserID(ctx context.Context, userID shared.ID) ([]*Project, error)
	IsMember(ctx context.Context, projectID, userID shared.ID) (bool, error)
	AddMember(ctx context.Context, member *ProjectMember) error
	CountMembers(ctx context.Context, projectID shared.ID) (int64, error)
}
