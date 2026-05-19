package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
	"gorm.io/gorm"
)

// tagGorm represents the GORM model for the tags table
type tagGorm struct {
	ID        string `gorm:"column:id;primaryKey"`
	ProjectID string `gorm:"column:project_id;not null;index:idx_tag_project_name"`
	Name      string `gorm:"column:name;not null;index:idx_tag_project_name"`
	Color     string `gorm:"column:color;not null"`
	CreatedBy string `gorm:"column:created_by;not null"`
}

// TableName specifies the table name for GORM
func (tagGorm) TableName() string {
	return "tags"
}

// toDomain converts a GORM model to a domain entity
func (t *tagGorm) toDomain() *testcase.Tag {
	return &testcase.Tag{
		ID:        shared.ID(t.ID),
		ProjectID: shared.ID(t.ProjectID),
		Name:      t.Name,
		Color:     t.Color,
		CreatedBy: shared.ID(t.CreatedBy),
	}
}

// toGorm converts a domain entity to a GORM model
func toTagGorm(tag *testcase.Tag) *tagGorm {
	return &tagGorm{
		ID:        string(tag.ID),
		ProjectID: string(tag.ProjectID),
		Name:      tag.Name,
		Color:     tag.Color,
		CreatedBy: string(tag.CreatedBy),
	}
}

// tagRepository implements testcase.TagRepository using GORM
type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository creates a new TagRepository instance
func NewTagRepository(db *gorm.DB) testcase.TagRepository {
	return &tagRepository{db: db}
}

// Create creates a new tag in the database
// Enforces UNIQUE constraint on (project_id, name)
func (r *tagRepository) Create(ctx context.Context, tag *testcase.Tag) error {
	if tag == nil {
		return fmt.Errorf("tag cannot be nil")
	}

	gormModel := toTagGorm(tag)

	result := DBOrTx(ctx, r.db).Create(gormModel)
	if result.Error != nil {
		// Check for unique constraint violation (duplicate name in same project)
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return shared.ErrTagDuplicateName
		}
		return fmt.Errorf("failed to create tag: %w", result.Error)
	}

	return nil
}

// FindByProject finds all tags for a given project ID
// Returns tags ordered by created_at DESC (most recent first)
func (r *tagRepository) FindByProject(ctx context.Context, projectID shared.ID) ([]*testcase.Tag, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID cannot be empty")
	}

	var gormModels []*tagGorm
	result := DBOrTx(ctx, r.db).
		Where("project_id = ?", string(projectID)).
		Order("created_at DESC").
		Find(&gormModels)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to find tags: %w", result.Error)
	}

	// Convert to domain entities
	tags := make([]*testcase.Tag, len(gormModels))
	for i, gormModel := range gormModels {
		tags[i] = gormModel.toDomain()
	}

	return tags, nil
}
