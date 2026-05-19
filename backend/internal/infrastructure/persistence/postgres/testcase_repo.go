package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
	"gorm.io/gorm"
)

// TestCaseModel is the GORM model for test_cases table
type TestCaseModel struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProjectID   string    `gorm:"type:uuid;not null;index"`
	ModuleID    *string   `gorm:"type:uuid;index"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	Status      string    `gorm:"type:varchar(50);not null;default:'draft';index"`
	Priority    int       `gorm:"type:int;not null;default:1;index"`
	Tags       []string  `gorm:"type:text[];default:'{}'"`
	Version     int       `gorm:"type:int;not null;default:1"`
	CreatedBy   string    `gorm:"type:uuid;not null"`
	UpdatedBy   *string   `gorm:"type:uuid"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
	DeletedAt   *time.Time `gorm:"index"`
	Steps       []StepModel `gorm:"foreignKey:TestCaseID;constraint:OnDelete:CASCADE"`
}

// StepModel is the GORM model for test_steps table
type StepModel struct {
	ID         string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TestCaseID string `gorm:"type:uuid;not null;index"`
	Number     int    `gorm:"type:int;not null"`
	Action     string `gorm:"type:text;not null"`
	Expected   string `gorm:"type:text;not null"`
}

// TableName specifies the table name for TestCaseModel
func (TestCaseModel) TableName() string {
	return "test_cases"
}

// TableName specifies the table name for StepModel
func (StepModel) TableName() string {
	return "test_steps"
}

// testCaseRepository implements testcase.TestCaseRepository using GORM
type testCaseRepository struct {
	db *gorm.DB
}

// NewTestCaseRepository creates a new TestCaseRepository instance
func NewTestCaseRepository(db *gorm.DB) testcase.TestCaseRepository {
	return &testCaseRepository{db: db}
}

// Create creates a new test case with its steps in a transaction
func (r *testCaseRepository) Create(ctx context.Context, tc *testcase.TestCase) error {
	model := r.domainToModel(tc)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the test case
		if err := tx.Create(model).Error; err != nil {
			return fmt.Errorf("failed to create test case: %w", err)
		}

		// Update domain entity with generated ID if needed
		tc.ID = shared.ID(model.ID)
		tc.Version = model.Version
		tc.CreatedAt = model.CreatedAt
		tc.UpdatedAt = model.UpdatedAt

		return nil
	})
}

// FindByID finds a test case by ID with steps preloaded
func (r *testCaseRepository) FindByID(ctx context.Context, id shared.ID) (*testcase.TestCase, error) {
	var model TestCaseModel

	err := r.db.WithContext(ctx).
		Preload("Steps").
		Where("id = ? AND deleted_at IS NULL", string(id)).
		First(&model).Error

	if err == gorm.ErrRecordNotFound {
		return nil, shared.ErrTestCaseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find test case: %w", err)
	}

	return r.modelToDomain(&model), nil
}

// List lists test cases with filtering and pagination
func (r *testCaseRepository) List(ctx context.Context, filter testcase.TestCaseFilter) ([]*testcase.TestCase, int64, error) {
	query := r.db.WithContext(ctx).Model(&TestCaseModel{}).
		Where("project_id = ? AND deleted_at IS NULL", string(filter.ProjectID))

	// Apply filters
	query = r.applyFilters(query, filter)

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count test cases: %w", err)
	}

	// Apply sorting
	query = r.applySorting(query, filter.SortBy, filter.SortDesc)

	// Apply pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	offset := (filter.Page - 1) * filter.PageSize

	var models []TestCaseModel
	if err := query.Preload("Steps").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list test cases: %w", err)
	}

	cases := make([]*testcase.TestCase, len(models))
	for i, model := range models {
		cases[i] = r.modelToDomain(&model)
	}

	return cases, total, nil
}

// Update updates a test case with optimistic locking
func (r *testCaseRepository) Update(ctx context.Context, tc *testcase.TestCase) error {
	model := r.domainToModel(tc)

	result := r.db.WithContext(ctx).
		Model(&TestCaseModel{}).
		Where("id = ? AND version = ? AND deleted_at IS NULL", string(model.ID), tc.Version).
		Updates(map[string]interface{}{
			"module_id":    model.ModuleID,
			"title":        model.Title,
			"description":  model.Description,
			"status":       model.Status,
			"priority":     model.Priority,
			"tags":         model.Tags,
			"updated_by":   model.UpdatedBy,
			"version":      model.Version + 1,
			"updated_at":   time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update test case: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.ErrTestCaseConflict
	}

	// Update version in domain entity
	tc.Version++

	return nil
}

// SoftDelete soft deletes a test case
func (r *testCaseRepository) SoftDelete(ctx context.Context, id shared.ID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&TestCaseModel{}).
		Where("id = ? AND deleted_at IS NULL", string(id)).
		Update("deleted_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to soft delete test case: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.ErrTestCaseNotFound
	}

	return nil
}

// BatchUpdateStatus batch updates test case statuses
func (r *testCaseRepository) BatchUpdateStatus(ctx context.Context, ids []shared.ID, status testcase.CaseStatus) error {
	if len(ids) == 0 {
		return nil
	}

	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = string(id)
	}

	result := r.db.WithContext(ctx).
		Model(&TestCaseModel{}).
		Where("id IN ? AND deleted_at IS NULL", idStrings).
		Update("status", string(status))

	if result.Error != nil {
		return fmt.Errorf("failed to batch update status: %w", result.Error)
	}

	return nil
}

// BatchDelete batch deletes test cases
func (r *testCaseRepository) BatchDelete(ctx context.Context, ids []shared.ID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = string(id)
	}

	result := r.db.WithContext(ctx).
		Model(&TestCaseModel{}).
		Where("id IN ? AND deleted_at IS NULL", idStrings).
		Update("deleted_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to batch delete: %w", result.Error)
	}

	return nil
}

// BatchMove batch moves test cases to a different module
func (r *testCaseRepository) BatchMove(ctx context.Context, ids []shared.ID, moduleID *shared.ID) error {
	if len(ids) == 0 {
		return nil
	}

	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = string(id)
	}

	updateData := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if moduleID != nil {
		updateData["module_id"] = string(*moduleID)
	} else {
		updateData["module_id"] = nil
	}

	result := r.db.WithContext(ctx).
		Model(&TestCaseModel{}).
		Where("id IN ? AND deleted_at IS NULL", idStrings).
		Updates(updateData)

	if result.Error != nil {
		return fmt.Errorf("failed to batch move: %w", result.Error)
	}

	return nil
}

// applyFilters applies filter conditions to the query
func (r *testCaseRepository) applyFilters(query *gorm.DB, filter testcase.TestCaseFilter) *gorm.DB {
	// Filter by module
	if filter.ModuleID != nil {
		query = query.Where("module_id = ?", string(*filter.ModuleID))
	}

	// Filter by status
	if filter.Status != nil {
		query = query.Where("status = ?", string(*filter.Status))
	}

	// Filter by priority
	if filter.Priority != nil {
		query = query.Where("priority = ?", int(*filter.Priority))
	}

	// Filter by tags (GIN index optimized)
	if len(filter.Tags) > 0 {
		query = query.Where("tags && ?", filter.Tags)
	}

	// Filter by keyword (full-text search on title and description)
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", keyword, keyword)
	}

	return query
}

// applySorting applies sorting to the query
func (r *testCaseRepository) applySorting(query *gorm.DB, sortBy string, sortDesc bool) *gorm.DB {
	// Default sorting
	if sortBy == "" {
		sortBy = "created_at"
		sortDesc = true
	}

	orderClause := sortBy
	if sortDesc {
		orderClause += " DESC"
	}
	return query.Order(orderClause)
}

// domainToModel converts domain entity to GORM model
func (r *testCaseRepository) domainToModel(tc *testcase.TestCase) *TestCaseModel {
	model := &TestCaseModel{
		ID:          string(tc.ID),
		ProjectID:   string(tc.ProjectID),
		Title:       tc.Title,
		Description: tc.Description,
		Status:      string(tc.Status),
		Priority:    int(tc.Priority),
		Tags:        tc.Tags,
		Version:     tc.Version,
		CreatedBy:   string(tc.CreatedBy),
		CreatedAt:   tc.CreatedAt,
		UpdatedAt:   tc.UpdatedAt,
	}

	if tc.ModuleID != nil {
		moduleIDStr := string(*tc.ModuleID)
		model.ModuleID = &moduleIDStr
	}

	if tc.UpdatedBy != nil {
		updatedByStr := string(*tc.UpdatedBy)
		model.UpdatedBy = &updatedByStr
	}

	if tc.DeletedAt != nil {
		model.DeletedAt = tc.DeletedAt
	}

	// Convert steps
	if len(tc.Steps) > 0 {
		model.Steps = make([]StepModel, len(tc.Steps))
		for i, step := range tc.Steps {
			model.Steps[i] = StepModel{
				ID:         string(step.ID),
				TestCaseID: string(step.TestCaseID),
				Number:     step.Number,
				Action:     step.Action,
				Expected:   step.Expected,
			}
		}
	}

	return model
}

// modelToDomain converts GORM model to domain entity
func (r *testCaseRepository) modelToDomain(model *TestCaseModel) *testcase.TestCase {
	tc := &testcase.TestCase{
		ID:          shared.ID(model.ID),
		ProjectID:   shared.ID(model.ProjectID),
		Title:       model.Title,
		Description: model.Description,
		Status:      testcase.CaseStatus(model.Status),
		Priority:    testcase.Priority(model.Priority),
		Tags:        model.Tags,
		Version:     model.Version,
		CreatedBy:   shared.ID(model.CreatedBy),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	if model.ModuleID != nil {
		moduleID := shared.ID(*model.ModuleID)
		tc.ModuleID = &moduleID
	}

	if model.UpdatedBy != nil {
		updatedBy := shared.ID(*model.UpdatedBy)
		tc.UpdatedBy = &updatedBy
	}

	if model.DeletedAt != nil {
		tc.DeletedAt = model.DeletedAt
	}

	// Convert steps
	if len(model.Steps) > 0 {
		tc.Steps = make([]testcase.Step, len(model.Steps))
		for i, stepModel := range model.Steps {
			tc.Steps[i] = testcase.Step{
				ID:         shared.ID(stepModel.ID),
				TestCaseID: shared.ID(stepModel.TestCaseID),
				Number:     stepModel.Number,
				Action:     stepModel.Action,
				Expected:   stepModel.Expected,
			}
		}
	}

	return tc
}
