package file

import (
	"context"

	"github.com/liang21/heka/internal/domain/shared"
)

type FileRepository interface {
	Create(ctx context.Context, file *File) error
	FindByID(ctx context.Context, id shared.ID) (*File, error)
	FindByProject(ctx context.Context, projectID shared.ID, page, pageSize int) ([]*File, int64, error)
	UpdateIndexStatus(ctx context.Context, id shared.ID, status IndexStatus, errMsg string) error
	SoftDelete(ctx context.Context, id shared.ID) error
	CreateVersion(ctx context.Context, version *FileVersion) error
}
