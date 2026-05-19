package file

import (
	"context"
	"io"

	"github.com/liang21/heka/internal/domain/file"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T116 | spec.md: §4.10 FileService Implementation

type Service struct {
	repo         file.FileRepository
	storage      Storage
	figma        FigmaClient
	eventBus     shared.EventBus
	maxUploadSize int64
}

type Storage interface {
	Save(ctx context.Context, projectID shared.ID, filename string, reader io.Reader) (string, error)
	GetPath(filename string) string
	Delete(ctx context.Context, filename string) error
}

type FigmaClient interface {
	GetFile(ctx context.Context, fileURL string) (interface{}, error)
}

func NewService(repo file.FileRepository, storage Storage, figma FigmaClient, eventBus shared.EventBus, maxUploadSize int64) *Service {
	return &Service{
		repo:         repo,
		storage:      storage,
		figma:        figma,
		eventBus:     eventBus,
		maxUploadSize: maxUploadSize,
	}
}

func (s *Service) Upload(ctx context.Context, userID, projectID shared.ID, name, contentType string, size int64, reader io.Reader) (*FileResponse, error) {
	path, err := s.storage.Save(ctx, projectID, name, reader)
	if err != nil {
		return nil, err
	}

	f := &file.File{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        name,
		Type:        s.fileTypeFromMime(contentType),
		Size:        size,
		Path:        path,
		SourceType:  file.SourceUpload,
		IndexStatus: file.IndexPending,
		UploadedBy:  userID,
	}

	if err := s.repo.Create(ctx, f); err != nil {
		return nil, err
	}

	s.eventBus.Publish(ctx, &file.FileUploadedEvent{
		ProjectID: projectID,
		FileID:     f.ID,
	})

	return s.toResponse(f), nil
}

func (s *Service) Delete(ctx context.Context, fileID shared.ID) error {
	return s.repo.SoftDelete(ctx, fileID)
}

func (s *Service) GetByID(ctx context.Context, fileID shared.ID) (*FileResponse, error) {
	f, err := s.repo.FindByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return s.toResponse(f), nil
}

func (s *Service) List(ctx context.Context, projectID shared.ID, page, pageSize int) ([]FileListResponse, int64, error) {
	files, total, err := s.repo.FindByProject(ctx, projectID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]FileListResponse, len(files))
	for i, f := range files {
		resp[i] = FileListResponse{
			ID:          f.ID,
			ProjectID:   f.ProjectID,
			Name:        f.Name,
			Type:        f.Type,
			Size:        f.Size,
			IndexStatus: f.IndexStatus,
			UploadedBy:  f.UploadedBy,
			UploadedAt:  f.UploadedAt,
		}
	}

	return resp, total, nil
}

func (s *Service) Reindex(ctx context.Context, fileID shared.ID) error {
	return s.repo.UpdateIndexStatus(ctx, fileID, file.IndexPending, "")
}

func (s *Service) GetIndexStatus(ctx context.Context, fileID shared.ID) (*IndexStatusResponse, error) {
	f, err := s.repo.FindByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return &IndexStatusResponse{
		FileID: f.ID,
		Status: f.IndexStatus,
		Error:  f.IndexError,
	}, nil
}

func (s *Service) fileTypeFromMime(mime string) file.FileType {
	switch mime {
	case "application/pdf":
		return file.FilePDF
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return file.FileDOCX
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return file.FileXLSX
	case "image/png", "image/jpeg", "image/gif":
		return file.FileImage
	default:
		return file.FilePDF
	}
}

func (s *Service) toResponse(f *file.File) *FileResponse {
	return &FileResponse{
		ID:             f.ID,
		ProjectID:      f.ProjectID,
		Name:           f.Name,
		Type:           f.Type,
		Size:           f.Size,
		Path:           f.Path,
		SourceType:     f.SourceType,
		SourceURL:      f.SourceURL,
		ContentPreview: f.ContentPreview,
		IsIndexed:      f.IsIndexed,
		IndexStatus:    f.IndexStatus,
		IndexError:     f.IndexError,
		IndexedAt:      f.IndexedAt,
		UploadedBy:     f.UploadedBy,
		Version:        f.Version,
		UploadedAt:     f.UploadedAt,
	}
}
