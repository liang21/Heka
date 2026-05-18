package file

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

type FileType string

const (
	FilePDF   FileType = "pdf"
	FileDOCX  FileType = "docx"
	FileXLSX  FileType = "xlsx"
	FileImage FileType = "image"
	FileFigma FileType = "figma"
)

type SourceType string

const (
	SourceUpload SourceType = "upload"
	SourceFigma  SourceType = "figma"
)

type IndexStatus string

const (
	IndexPending    IndexStatus = "pending"
	IndexProcessing IndexStatus = "processing"
	IndexCompleted  IndexStatus = "completed"
	IndexFailed     IndexStatus = "failed"
)

type File struct {
	ID            shared.ID
	ProjectID     shared.ID
	Name          string
	Type          FileType
	Size          int64
	Path          string
	SourceType    SourceType
	SourceURL     string
	ContentPreview string
	IsIndexed     bool
	IndexStatus   IndexStatus
	IndexError    string
	IndexedAt     *time.Time
	UploadedBy    shared.ID
	Version       int
	UploadedAt    time.Time
	DeletedAt     *time.Time
}

type FileVersion struct {
	ID         shared.ID
	FileID     shared.ID
	Version    int
	Path       string
	Size       int64
	UploadedBy shared.ID
	UploadedAt time.Time
}
