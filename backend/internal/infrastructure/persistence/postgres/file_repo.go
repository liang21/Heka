package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/liang21/heka/internal/domain/file"
	"github.com/liang21/heka/internal/domain/shared"
	"gorm.io/gorm"
)

// FileModel is the GORM model for files table
type FileModel struct {
	ID             string             `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProjectID      string             `gorm:"type:uuid;not null;index"`
	Name           string             `gorm:"type:varchar(255);not null"`
	Type           string             `gorm:"type:varchar(50);not null;index"`
	Size           int64              `gorm:"type:bigint;not null"`
	Path           string             `gorm:"type:varchar(500);not null"`
	SourceType     string             `gorm:"type:varchar(50);not null"`
	SourceURL      string             `gorm:"type:varchar(500)"`
	ContentPreview string             `gorm:"type:text"`
	IsIndexed      bool               `gorm:"type:boolean;not null;default:false;index"`
	IndexStatus    string             `gorm:"type:varchar(50);not null;default:'pending';index"`
	IndexError     string             `gorm:"type:text"`
	IndexedAt      *time.Time         `gorm:"type:timestamptz"`
	UploadedBy     string             `gorm:"type:uuid;not null"`
	Version        int                `gorm:"type:int;not null;default:1"`
	UploadedAt     time.Time          `gorm:"type:timestamptz;not null"`
	DeletedAt      *time.Time         `gorm:"index"`
	Versions       []FileVersionModel `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE"`
}

// FileVersionModel is the GORM model for file_versions table
type FileVersionModel struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FileID     string    `gorm:"type:uuid;not null;index"`
	Version    int       `gorm:"type:int;not null"`
	Path       string    `gorm:"type:varchar(500);not null"`
	Size       int64     `gorm:"type:bigint;not null"`
	UploadedBy string    `gorm:"type:uuid;not null"`
	UploadedAt time.Time `gorm:"type:timestamptz;not null"`
}

// TableName specifies the table name for FileModel
func (FileModel) TableName() string {
	return "files"
}

// TableName specifies the table name for FileVersionModel
func (FileVersionModel) TableName() string {
	return "file_versions"
}

// fileRepository implements file.FileRepository using GORM
type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new FileRepository instance
func NewFileRepository(db *gorm.DB) file.FileRepository {
	return &fileRepository{db: db}
}

// Create creates a new file record
func (r *fileRepository) Create(ctx context.Context, f *file.File) error {
	model := r.domainToModel(f)

	db := DBOrTx(ctx, r.db)

	if err := db.Create(model).Error; err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	// Update domain entity with generated values
	f.ID = shared.ID(model.ID)
	f.Version = model.Version
	f.UploadedAt = model.UploadedAt

	return nil
}

// FindByID finds a file by ID with versions preloaded
func (r *fileRepository) FindByID(ctx context.Context, id shared.ID) (*file.File, error) {
	var model FileModel

	db := DBOrTx(ctx, r.db)

	err := db.
		Preload("Versions").
		Where("id = ? AND deleted_at IS NULL", string(id)).
		First(&model).Error

	if err == gorm.ErrRecordNotFound {
		return nil, shared.ErrFileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find file: %w", err)
	}

	return r.modelToDomain(&model), nil
}

// FindByProject finds files by project ID with pagination
func (r *fileRepository) FindByProject(ctx context.Context, projectID shared.ID, page, pageSize int) ([]*file.File, int64, error) {
	db := DBOrTx(ctx, r.db)

	query := db.Model(&FileModel{}).
		Where("project_id = ? AND deleted_at IS NULL", string(projectID))

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	// Validate pagination parameters
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var models []FileModel
	if err := query.
		Preload("Versions").
		Order("uploaded_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find files by project: %w", err)
	}

	files := make([]*file.File, len(models))
	for i, model := range models {
		files[i] = r.modelToDomain(&model)
	}

	return files, total, nil
}

// UpdateIndexStatus updates the index status of a file
func (r *fileRepository) UpdateIndexStatus(ctx context.Context, id shared.ID, status file.IndexStatus, errMsg string) error {
	db := DBOrTx(ctx, r.db)

	updateData := map[string]interface{}{
		"index_status": string(status),
	}

	if status == file.IndexCompleted {
		now := time.Now()
		updateData["indexed_at"] = &now
		updateData["is_indexed"] = true
		updateData["index_error"] = ""
	} else if status == file.IndexFailed {
		updateData["index_error"] = errMsg
		updateData["is_indexed"] = false
	}

	result := db.
		Model(&FileModel{}).
		Where("id = ? AND deleted_at IS NULL", string(id)).
		Updates(updateData)

	if result.Error != nil {
		return fmt.Errorf("failed to update index status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.ErrFileNotFound
	}

	return nil
}

// SoftDelete soft deletes a file
func (r *fileRepository) SoftDelete(ctx context.Context, id shared.ID) error {
	db := DBOrTx(ctx, r.db)

	now := time.Now()
	result := db.
		Model(&FileModel{}).
		Where("id = ? AND deleted_at IS NULL", string(id)).
		Update("deleted_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to soft delete file: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return shared.ErrFileNotFound
	}

	return nil
}

// CreateVersion creates a new file version
func (r *fileRepository) CreateVersion(ctx context.Context, version *file.FileVersion) error {
	db := DBOrTx(ctx, r.db)

	model := r.versionDomainToModel(version)

	// Verify that the file exists
	var fileModel FileModel
	if err := db.Where("id = ? AND deleted_at IS NULL", string(model.FileID)).First(&fileModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return shared.ErrFileNotFound
		}
		return fmt.Errorf("failed to verify file exists: %w", err)
	}

	if err := db.Create(model).Error; err != nil {
		return fmt.Errorf("failed to create file version: %w", err)
	}

	// Update domain entity with generated ID
	version.ID = shared.ID(model.ID)

	return nil
}

// domainToModel converts domain File entity to GORM model
func (r *fileRepository) domainToModel(f *file.File) *FileModel {
	return &FileModel{
		ID:             string(f.ID),
		ProjectID:      string(f.ProjectID),
		Name:           f.Name,
		Type:           string(f.Type),
		Size:           f.Size,
		Path:           f.Path,
		SourceType:     string(f.SourceType),
		SourceURL:      f.SourceURL,
		ContentPreview: f.ContentPreview,
		IsIndexed:      f.IsIndexed,
		IndexStatus:    string(f.IndexStatus),
		IndexError:     f.IndexError,
		IndexedAt:      f.IndexedAt,
		UploadedBy:     string(f.UploadedBy),
		Version:        f.Version,
		UploadedAt:     f.UploadedAt,
		DeletedAt:      f.DeletedAt,
	}
}

// modelToDomain converts GORM model to domain File entity
func (r *fileRepository) modelToDomain(model *FileModel) *file.File {
	return &file.File{
		ID:             shared.ID(model.ID),
		ProjectID:      shared.ID(model.ProjectID),
		Name:           model.Name,
		Type:           file.FileType(model.Type),
		Size:           model.Size,
		Path:           model.Path,
		SourceType:     file.SourceType(model.SourceType),
		SourceURL:      model.SourceURL,
		ContentPreview: model.ContentPreview,
		IsIndexed:      model.IsIndexed,
		IndexStatus:    file.IndexStatus(model.IndexStatus),
		IndexError:     model.IndexError,
		IndexedAt:      model.IndexedAt,
		UploadedBy:     shared.ID(model.UploadedBy),
		Version:        model.Version,
		UploadedAt:     model.UploadedAt,
		DeletedAt:      model.DeletedAt,
	}
}

// versionDomainToModel converts domain FileVersion entity to GORM model
func (r *fileRepository) versionDomainToModel(v *file.FileVersion) *FileVersionModel {
	return &FileVersionModel{
		ID:         string(v.ID),
		FileID:     string(v.FileID),
		Version:    v.Version,
		Path:       v.Path,
		Size:       v.Size,
		UploadedBy: string(v.UploadedBy),
		UploadedAt: v.UploadedAt,
	}
}
