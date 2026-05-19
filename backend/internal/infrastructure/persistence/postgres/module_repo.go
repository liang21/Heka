package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
	"gorm.io/gorm"
)

// moduleGorm represents the GORM model for the modules table
type moduleGorm struct {
	ID          string  `gorm:"column:id;primaryKey"`
	ProjectID   string  `gorm:"column:project_id;not null"`
	Name        string  `gorm:"column:name;not null"`
	Description string  `gorm:"column:description"`
	ParentID    *string `gorm:"column:parent_id"`
	OrderIndex  int     `gorm:"column:order_index;not null;default:0"`
	CreatedBy   string  `gorm:"column:created_by;not null"`
}

// TableName specifies the table name for GORM
func (moduleGorm) TableName() string {
	return "modules"
}

// toDomain converts a GORM model to a domain entity
func (m *moduleGorm) toDomain() *testcase.Module {
	module := &testcase.Module{
		ID:          shared.ID(m.ID),
		ProjectID:   shared.ID(m.ProjectID),
		Name:        m.Name,
		Description: m.Description,
		OrderIndex:  m.OrderIndex,
		CreatedBy:   shared.ID(m.CreatedBy),
	}

	if m.ParentID != nil {
		parentID := shared.ID(*m.ParentID)
		module.ParentID = &parentID
	}

	return module
}

// toGorm converts a domain entity to a GORM model
func toModuleGorm(m *testcase.Module) *moduleGorm {
	gorm := &moduleGorm{
		ID:          string(m.ID),
		ProjectID:   string(m.ProjectID),
		Name:        m.Name,
		Description: m.Description,
		OrderIndex:  m.OrderIndex,
		CreatedBy:   string(m.CreatedBy),
	}

	if m.ParentID != nil {
		parentID := string(*m.ParentID)
		gorm.ParentID = &parentID
	}

	return gorm
}

// moduleRepository implements testcase.ModuleRepository using GORM
type moduleRepository struct {
	db *gorm.DB
}

// NewModuleRepository creates a new ModuleRepository instance
func NewModuleRepository(db *gorm.DB) testcase.ModuleRepository {
	return &moduleRepository{db: db}
}

// Create creates a new module in the database
func (r *moduleRepository) Create(ctx context.Context, m *testcase.Module) error {
	if m == nil {
		return fmt.Errorf("module cannot be nil")
	}

	gormModel := toModuleGorm(m)

	// Auto-generate path based on parent
	if err := r.generatePath(ctx, gormModel); err != nil {
		return fmt.Errorf("failed to generate path: %w", err)
	}

	result := DBOrTx(ctx, r.db).Create(gormModel)
	if result.Error != nil {
		// Check for unique constraint violation (duplicate name in same level)
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return shared.ErrModuleConflict
		}
		return fmt.Errorf("failed to create module: %w", result.Error)
	}

	return nil
}

// Update updates an existing module in the database
func (r *moduleRepository) Update(ctx context.Context, m *testcase.Module) error {
	if m == nil {
		return fmt.Errorf("module cannot be nil")
	}

	gormModel := toModuleGorm(m)

	// Regenerate path if parent changed
	if err := r.generatePath(ctx, gormModel); err != nil {
		return fmt.Errorf("failed to regenerate path: %w", err)
	}

	result := DBOrTx(ctx, r.db).Save(gormModel)
	if result.Error != nil {
		// Check for unique constraint violation
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return shared.ErrModuleConflict
		}
		return fmt.Errorf("failed to update module: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.ErrModuleNotFound
	}

	return nil
}

// Delete deletes a module from the database
// Note: Foreign key constraint with ON DELETE CASCADE will handle children deletion
func (r *moduleRepository) Delete(ctx context.Context, id shared.ID) error {
	if id == "" {
		return fmt.Errorf("module ID cannot be empty")
	}

	result := DBOrTx(ctx, r.db).Delete(&moduleGorm{}, "id = ?", string(id))
	if result.Error != nil {
		return fmt.Errorf("failed to delete module: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.ErrModuleNotFound
	}

	return nil
}

// FindByProject finds all modules for a given project ID
// Returns flat list with all modules, which can be assembled into a tree structure in memory
func (r *moduleRepository) FindByProject(ctx context.Context, projectID shared.ID) ([]*testcase.Module, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	var gormModels []*moduleGorm
	result := DBOrTx(ctx, r.db).
		Where("project_id = ?", string(projectID)).
		Order("order_index ASC, name ASC").
		Find(&gormModels)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to find modules: %w", result.Error)
	}

	// Convert to domain entities
	modules := make([]*testcase.Module, len(gormModels))
	for i, gormModel := range gormModels {
		modules[i] = gormModel.toDomain()
	}

	return modules, nil
}

// generatePath generates and sets the path for a module based on its parent
// Path format: "/Parent/Child" or "/" for root modules
func (r *moduleRepository) generatePath(ctx context.Context, gormModel *moduleGorm) error {
	// For now, we'll skip path generation as it's not in the current schema
	// This can be implemented later if needed for hierarchical queries
	return nil
}
