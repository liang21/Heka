// tasks.md: T046 | TDD GREEN Phase for ProjectRepository
package postgres

import (
	"context"

	"github.com/liang21/heka/internal/domain/project"
	"github.com/liang21/heka/internal/domain/shared"
	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new ProjectRepository instance
func NewProjectRepository(db *gorm.DB) project.ProjectRepository {
	return &projectRepository{db: db}
}

// Create creates a new project in the database
func (r *projectRepository) Create(ctx context.Context, proj *project.Project) error {
	db := DBOrTx(ctx, r.db)
	if err := db.Create(proj).Error; err != nil {
		return err
	}
	return nil
}

// FindByID retrieves a project by its ID
func (r *projectRepository) FindByID(ctx context.Context, id shared.ID) (*project.Project, error) {
	db := DBOrTx(ctx, r.db)
	var proj project.Project
	err := db.Preload("Members").Where("id = ?", id).First(&proj).Error
	if err == gorm.ErrRecordNotFound {
		return nil, shared.ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}
	return &proj, nil
}

// FindByUserID retrieves all projects created by a specific user
func (r *projectRepository) FindByUserID(ctx context.Context, userID shared.ID) ([]*project.Project, error) {
	db := DBOrTx(ctx, r.db)
	var projects []*project.Project
	err := db.Where("created_by = ?", userID).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// IsMember checks if a user is a member of a project
func (r *projectRepository) IsMember(ctx context.Context, projectID, userID shared.ID) (bool, error) {
	db := DBOrTx(ctx, r.db)
	var count int64
	err := db.Table("project_members").
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddMember adds a member to a project
func (r *projectRepository) AddMember(ctx context.Context, member *project.ProjectMember) error {
	db := DBOrTx(ctx, r.db)
	if err := db.Create(member).Error; err != nil {
		return err
	}
	return nil
}

// CountMembers counts the number of members in a project
func (r *projectRepository) CountMembers(ctx context.Context, projectID shared.ID) (int64, error) {
	db := DBOrTx(ctx, r.db)
	var count int64
	err := db.Table("project_members").
		Where("project_id = ?", projectID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
