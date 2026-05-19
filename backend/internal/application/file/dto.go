package file

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/file"
)

// tasks.md: T114 | spec.md: §4.10 文件管理 DTO

type FileResponse struct {
	ID             shared.ID          `json:"id"`
	ProjectID      shared.ID          `json:"project_id"`
	Name           string             `json:"name"`
	Type           file.FileType      `json:"type"`
	Size           int64              `json:"size"`
	Path           string             `json:"path"`
	SourceType     file.SourceType    `json:"source_type"`
	SourceURL      string             `json:"source_url"`
	ContentPreview string             `json:"content_preview"`
	IsIndexed      bool               `json:"is_indexed"`
	IndexStatus    file.IndexStatus   `json:"index_status"`
	IndexError     string             `json:"index_error"`
	IndexedAt      *time.Time         `json:"indexed_at"`
	UploadedBy     shared.ID          `json:"uploaded_by"`
	Version        int                `json:"version"`
	UploadedAt     time.Time          `json:"uploaded_at"`
}

type FileListResponse struct {
	ID           shared.ID        `json:"id"`
	ProjectID    shared.ID        `json:"project_id"`
	Name         string           `json:"name"`
	Type         file.FileType    `json:"type"`
	Size         int64            `json:"size"`
	IndexStatus  file.IndexStatus `json:"index_status"`
	UploadedBy   shared.ID        `json:"uploaded_by"`
	UploadedAt   time.Time        `json:"uploaded_at"`
}

type IndexStatusResponse struct {
	FileID shared.ID        `json:"file_id"`
	Status file.IndexStatus `json:"status"`
	Error  string           `json:"error"`
}

type FigmaUploadReq struct {
	URL string `json:"url" validate:"required,url"`
}
