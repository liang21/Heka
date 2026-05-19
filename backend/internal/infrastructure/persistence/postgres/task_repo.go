package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"gorm.io/gorm"
)

// AsyncTaskModel is the GORM model for async_tasks table
type AsyncTaskModel struct {
	ID              string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProjectID       string     `gorm:"type:uuid;not null;index"`
	Type            string     `gorm:"type:varchar(50);not null;index"`
	Status          string     `gorm:"type:varchar(50);not null;index"`
	ProgressCurrent int        `gorm:"type:int;not null;default:0"`
	ProgressTotal   int        `gorm:"type:int;not null;default:0"`
	Input           []byte     `gorm:"type:jsonb"`
	Result          []byte     `gorm:"type:jsonb"`
	Error           string     `gorm:"type:text"`
	CreatedBy       string     `gorm:"type:uuid;not null"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;not null"`
	StartedAt       *time.Time `gorm:"type:timestamptz"`
	CompletedAt     *time.Time `gorm:"type:timestamptz"`
}

// IndexTaskModel is the GORM model for index_tasks table
type IndexTaskModel struct {
	ID          string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FileID      string     `gorm:"type:uuid;not null;index"`
	Status      string     `gorm:"type:varchar(50);not null;index"`
	RetryCount  int        `gorm:"type:int;not null;default:0"`
	MaxRetries  int        `gorm:"type:int;not null;default:3"`
	Error       string     `gorm:"type:text"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;not null"`
	CompletedAt *time.Time `gorm:"type:timestamptz"`
}

// TableName specifies the table name for AsyncTaskModel
func (AsyncTaskModel) TableName() string {
	return "async_tasks"
}

// TableName specifies the table name for IndexTaskModel
func (IndexTaskModel) TableName() string {
	return "index_tasks"
}

// AsyncTaskRepository implements shared.AsyncTaskRepository using GORM
type AsyncTaskRepository struct {
	db *gorm.DB
}

// NewAsyncTaskRepository creates a new AsyncTaskRepository instance
func NewAsyncTaskRepository(db *gorm.DB) shared.AsyncTaskRepository {
	return &AsyncTaskRepository{db: db}
}

// Create creates a new async task record
func (r *AsyncTaskRepository) Create(ctx context.Context, task *shared.AsyncTask) error {
	if task.ID == "" {
		return shared.ErrSysValidation
	}

	model := r.domainToModel(task)

	db := DBOrTx(ctx, r.db)

	if err := db.Create(model).Error; err != nil {
		return fmt.Errorf("failed to create async task: %w", err)
	}

	return nil
}

// FindByID finds an async task by ID
func (r *AsyncTaskRepository) FindByID(ctx context.Context, id shared.ID) (*shared.AsyncTask, error) {
	var model AsyncTaskModel

	db := DBOrTx(ctx, r.db)

	err := db.Where("id = ?", string(id)).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, shared.NewAppError("TS-NF-001", "async task not found", 404)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find async task: %w", err)
	}

	return r.modelToDomain(&model), nil
}

// FindPendingByType finds pending tasks by type with limit
func (r *AsyncTaskRepository) FindPendingByType(ctx context.Context, projectID shared.ID, taskType string, limit int) ([]*shared.AsyncTask, error) {
	db := DBOrTx(ctx, r.db)

	var models []AsyncTaskModel
	err := db.
		Where("project_id = ? AND status = ? AND type = ?", string(projectID), "pending", taskType).
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find pending tasks by type: %w", err)
	}

	tasks := make([]*shared.AsyncTask, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}

	return tasks, nil
}

// Update updates an async task record
func (r *AsyncTaskRepository) Update(ctx context.Context, task *shared.AsyncTask) error {
	db := DBOrTx(ctx, r.db)

	model := r.domainToModel(task)

	result := db.
		Model(&AsyncTaskModel{}).
		Where("id = ?", string(task.ID)).
		Updates(map[string]interface{}{
			"status":           model.Status,
			"progress_current": model.ProgressCurrent,
			"progress_total":   model.ProgressTotal,
			"result":           model.Result,
			"error":            model.Error,
			"started_at":       model.StartedAt,
			"completed_at":     model.CompletedAt,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update async task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.NewAppError("TS-NF-001", "async task not found", 404)
	}

	return nil
}

// IndexTaskRepository implements shared.IndexTaskRepository using GORM
type IndexTaskRepository struct {
	db *gorm.DB
}

// NewIndexTaskRepository creates a new IndexTaskRepository instance
func NewIndexTaskRepository(db *gorm.DB) shared.IndexTaskRepository {
	return &IndexTaskRepository{db: db}
}

// Create creates a new index task record
func (r *IndexTaskRepository) Create(ctx context.Context, task *shared.IndexTask) error {
	if task.ID == "" {
		return shared.ErrSysValidation
	}

	model := r.indexDomainToModel(task)

	db := DBOrTx(ctx, r.db)

	if err := db.Create(model).Error; err != nil {
		return fmt.Errorf("failed to create index task: %w", err)
	}

	return nil
}

// FindPending finds pending index tasks with limit
func (r *IndexTaskRepository) FindPending(ctx context.Context, limit int) ([]*shared.IndexTask, error) {
	db := DBOrTx(ctx, r.db)

	var models []IndexTaskModel
	err := db.
		Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find pending index tasks: %w", err)
	}

	tasks := make([]*shared.IndexTask, len(models))
	for i, model := range models {
		tasks[i] = r.indexModelToDomain(&model)
	}

	return tasks, nil
}

// FindStale finds stale index tasks (older than the specified duration)
func (r *IndexTaskRepository) FindStale(ctx context.Context, olderThan string, limit int) ([]*shared.IndexTask, error) {
	db := DBOrTx(ctx, r.db)

	// Parse the duration string (e.g., "1h", "30m")
	duration, err := time.ParseDuration(olderThan)
	if err != nil {
		return nil, fmt.Errorf("invalid duration format: %w", err)
	}

	staleTime := time.Now().Add(-duration)

	var models []IndexTaskModel
	err = db.
		Where("status = ? AND updated_at < ?", "pending", staleTime).
		Order("updated_at ASC").
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find stale index tasks: %w", err)
	}

	tasks := make([]*shared.IndexTask, len(models))
	for i, model := range models {
		tasks[i] = r.indexModelToDomain(&model)
	}

	return tasks, nil
}

// Update updates an index task record
func (r *IndexTaskRepository) Update(ctx context.Context, task *shared.IndexTask) error {
	db := DBOrTx(ctx, r.db)

	model := r.indexDomainToModel(task)

	result := db.
		Model(&IndexTaskModel{}).
		Where("id = ?", string(task.ID)).
		Updates(map[string]interface{}{
			"status":       model.Status,
			"retry_count":  model.RetryCount,
			"error":        model.Error,
			"updated_at":   model.UpdatedAt,
			"completed_at": model.CompletedAt,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update index task: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.NewAppError("TS-NF-001", "index task not found", 404)
	}

	return nil
}

// domainToModel converts domain AsyncTask entity to GORM model
func (r *AsyncTaskRepository) domainToModel(task *shared.AsyncTask) *AsyncTaskModel {
	model := &AsyncTaskModel{
		ID:              string(task.ID),
		ProjectID:       string(task.ProjectID),
		Type:            task.Type,
		Status:          task.Status,
		ProgressCurrent: task.ProgressCurrent,
		ProgressTotal:   task.ProgressTotal,
		Input:           task.Input,
		Result:          task.Result,
		Error:           task.Error,
		CreatedBy:       string(task.CreatedBy),
		CreatedAt:       task.CreatedAt,
		StartedAt:       task.StartedAt,
		CompletedAt:     task.CompletedAt,
	}

	return model
}

// modelToDomain converts GORM model to domain AsyncTask entity
func (r *AsyncTaskRepository) modelToDomain(model *AsyncTaskModel) *shared.AsyncTask {
	task := &shared.AsyncTask{
		ID:              shared.ID(model.ID),
		ProjectID:       shared.ID(model.ProjectID),
		Type:            model.Type,
		Status:          model.Status,
		ProgressCurrent: model.ProgressCurrent,
		ProgressTotal:   model.ProgressTotal,
		Input:           model.Input,
		Result:          model.Result,
		Error:           model.Error,
		CreatedBy:       shared.ID(model.CreatedBy),
		CreatedAt:       model.CreatedAt,
		StartedAt:       model.StartedAt,
		CompletedAt:     model.CompletedAt,
	}

	return task
}

// indexDomainToModel converts domain IndexTask entity to GORM model
func (r *IndexTaskRepository) indexDomainToModel(task *shared.IndexTask) *IndexTaskModel {
	model := &IndexTaskModel{
		ID:          string(task.ID),
		FileID:      string(task.FileID),
		Status:      task.Status,
		RetryCount:  task.RetryCount,
		MaxRetries:  task.MaxRetries,
		Error:       task.Error,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		CompletedAt: task.CompletedAt,
	}

	return model
}

// indexModelToDomain converts GORM model to domain IndexTask entity
func (r *IndexTaskRepository) indexModelToDomain(model *IndexTaskModel) *shared.IndexTask {
	task := &shared.IndexTask{
		ID:          shared.ID(model.ID),
		FileID:      shared.ID(model.FileID),
		Status:      model.Status,
		RetryCount:  model.RetryCount,
		MaxRetries:  model.MaxRetries,
		Error:       model.Error,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		CompletedAt: model.CompletedAt,
	}

	return task
}
