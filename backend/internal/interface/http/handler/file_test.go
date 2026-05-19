package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/liang21/heka/internal/application/file"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T136 | spec.md: §4.10 File Handler TDD RED

type mockFileService struct {
	mock.Mock
}

func (m *mockFileService) Upload(ctx context.Context, userID, projectID shared.ID, name, contentType string, size int64, reader io.Reader) (*file.FileResponse, error) {
	args := m.Called(ctx, userID, projectID, name, contentType, size, mock.Anything)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*file.FileResponse), args.Error(1)
}

func (m *mockFileService) FigmaUpload(ctx context.Context, userID, projectID shared.ID, req file.FigmaUploadReq) (*file.FileResponse, error) {
	args := m.Called(ctx, userID, projectID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*file.FileResponse), args.Error(1)
}

func (m *mockFileService) List(ctx context.Context, projectID shared.ID, page, pageSize int) ([]file.FileListResponse, int64, error) {
	args := m.Called(ctx, projectID, page, pageSize)
	if args.Get(0) == nil { return nil, args.Get(1).(int64), args.Error(2) }
	return args.Get(0).([]file.FileListResponse), args.Get(1).(int64), args.Error(2)
}

func (m *mockFileService) GetByID(ctx context.Context, fileID shared.ID) (*file.FileResponse, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*file.FileResponse), args.Error(1)
}

func (m *mockFileService) Reindex(ctx context.Context, fileID shared.ID) error {
	return m.Called(ctx, fileID).Error(0)
}

func (m *mockFileService) GetIndexStatus(ctx context.Context, fileID shared.ID) (*file.IndexStatusResponse, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*file.IndexStatusResponse), args.Error(1)
}

func (m *mockFileService) Delete(ctx context.Context, fileID shared.ID) error {
	return m.Called(ctx, fileID).Error(0)
}

func TestFileHandler_List(t *testing.T) {
	t.Parallel()
	svc := new(mockFileService)
	h := NewFileHandler(svc, 100*1024*1024) // 100MB default

	projectID := shared.NewID()
	svc.On("List", mock.Anything, projectID, 1, 20).Return([]file.FileListResponse{
		{ID: shared.NewID(), Name: "file1.pdf"},
	}, int64(1), nil)

	req := httptest.NewRequest("GET", "/api/v1/files?project_id="+projectID.String(), nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFileHandler_Delete(t *testing.T) {
	t.Parallel()
	svc := new(mockFileService)
	h := NewFileHandler(svc, 100*1024*1024) // 100MB default

	fileID := shared.NewID()
	svc.On("Delete", mock.Anything, fileID).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/files/"+fileID.String(), nil)
	req = setChiURLParam(req, "id", fileID.String())
	w := httptest.NewRecorder()

	h.Delete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFileHandler_GetIndexStatus(t *testing.T) {
	t.Parallel()
	svc := new(mockFileService)
	h := NewFileHandler(svc, 100*1024*1024) // 100MB default

	fileID := shared.NewID()
	svc.On("GetIndexStatus", mock.Anything, fileID).Return(&file.IndexStatusResponse{
		FileID: fileID, Status: "completed",
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/files/"+fileID.String()+"/index-status", nil)
	req = setChiURLParam(req, "id", fileID.String())
	w := httptest.NewRecorder()

	h.GetIndexStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFileHandler_Reindex(t *testing.T) {
	t.Parallel()
	svc := new(mockFileService)
	h := NewFileHandler(svc, 100*1024*1024) // 100MB default

	fileID := shared.NewID()
	svc.On("Reindex", mock.Anything, fileID).Return(nil)

	req := httptest.NewRequest("POST", "/api/v1/files/"+fileID.String()+"/reindex", nil)
	req = setChiURLParam(req, "id", fileID.String())
	w := httptest.NewRecorder()

	h.Reindex(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
