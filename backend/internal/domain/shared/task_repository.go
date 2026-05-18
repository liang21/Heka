package shared

import "context"

type AsyncTaskRepository interface {
	Create(ctx context.Context, task *AsyncTask) error
	FindByID(ctx context.Context, id ID) (*AsyncTask, error)
	FindPendingByType(ctx context.Context, projectID ID, taskType string, limit int) ([]*AsyncTask, error)
	Update(ctx context.Context, task *AsyncTask) error
}

type IndexTaskRepository interface {
	Create(ctx context.Context, task *IndexTask) error
	FindPending(ctx context.Context, limit int) ([]*IndexTask, error)
	FindStale(ctx context.Context, olderThan string, limit int) ([]*IndexTask, error)
	Update(ctx context.Context, task *IndexTask) error
}
