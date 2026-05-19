package file

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/liang21/heka/internal/domain/file"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T115 | spec.md: §4.10 FileService TDD RED

type mockFileRepo struct {
	mock.Mock
}

func (m *mockFileRepo) Create(ctx context.Context, f *file.File) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}

func (m *mockFileRepo) FindByID(ctx context.Context, id shared.ID) (*file.File, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*file.File), args.Error(1)
}

func (m *mockFileRepo) FindByProject(ctx context.Context, projectID shared.ID, page, pageSize int) ([]*file.File, int64, error) {
	args := m.Called(ctx, projectID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*file.File), args.Get(1).(int64), args.Error(2)
}

func (m *mockFileRepo) UpdateIndexStatus(ctx context.Context, id shared.ID, status file.IndexStatus, errMsg string) error {
	args := m.Called(ctx, id, status, errMsg)
	return args.Error(0)
}

func (m *mockFileRepo) SoftDelete(ctx context.Context, id shared.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockFileRepo) CreateVersion(ctx context.Context, v *file.FileVersion) error {
	args := m.Called(ctx, v)
	return args.Error(0)
}

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Save(ctx context.Context, projectID shared.ID, name string, reader io.Reader) (string, error) {
	args := m.Called(ctx, projectID, name, reader)
	return args.String(0), args.Error(1)
}

func (m *mockStorage) GetPath(name string) string {
	args := m.Called(name)
	return args.String(0)
}

func (m *mockStorage) Delete(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

type mockFileEventBus struct {
	mock.Mock
}

func (m *mockFileEventBus) Publish(ctx context.Context, events ...shared.Event) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func (m *mockFileEventBus) Subscribe(eventName string, handler shared.EventHandler) {
	m.Called(eventName, handler)
}

func TestFileService_Upload(t *testing.T) {
	t.Parallel()
	repo := new(mockFileRepo)
	storage := new(mockStorage)
	bus := new(mockFileEventBus)
	svc := NewService(repo, storage, nil, bus, 100*1024*1024)

	projectID := shared.NewID()
	userID := shared.NewID()

	storage.On("Save", mock.Anything, projectID, "test.pdf", mock.Anything).Return("2024/01/test.pdf", nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*file.File")).Return(nil)
	bus.On("Publish", mock.Anything, mock.Anything).Return(nil)

	content := strings.NewReader("test pdf content")
	resp, err := svc.Upload(context.Background(), userID, projectID, "test.pdf", "application/pdf", 1024, content)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "test.pdf", resp.Name)
	assert.Equal(t, file.FilePDF, resp.Type)
}

func TestFileService_Delete(t *testing.T) {
	t.Parallel()
	repo := new(mockFileRepo)
	svc := NewService(repo, new(mockStorage), nil, new(mockFileEventBus), 100*1024*1024)

	fileID := shared.NewID()
	repo.On("SoftDelete", mock.Anything, fileID).Return(nil)

	err := svc.Delete(context.Background(), fileID)

	require.NoError(t, err)
}

func TestFileService_GetByID(t *testing.T) {
	t.Parallel()
	repo := new(mockFileRepo)
	svc := NewService(repo, new(mockStorage), nil, new(mockFileEventBus), 100*1024*1024)

	fileID := shared.NewID()
	projectID := shared.NewID()
	userID := shared.NewID()
	now := time.Now()

	f := &file.File{
		ID:         fileID,
		ProjectID:  projectID,
		Name:       "test.pdf",
		Type:       file.FilePDF,
		Size:       1024,
		Path:       "2024/01/test.pdf",
		SourceType: file.SourceUpload,
		UploadedBy: userID,
		UploadedAt: now,
	}

	repo.On("FindByID", mock.Anything, fileID).Return(f, nil)

	resp, err := svc.GetByID(context.Background(), fileID)

	require.NoError(t, err)
	assert.Equal(t, fileID, resp.ID)
	assert.Equal(t, "test.pdf", resp.Name)
}

func TestFileService_Reindex(t *testing.T) {
	t.Parallel()
	repo := new(mockFileRepo)
	svc := NewService(repo, new(mockStorage), nil, new(mockFileEventBus), 100*1024*1024)

	fileID := shared.NewID()
	repo.On("UpdateIndexStatus", mock.Anything, fileID, file.IndexPending, "").Return(nil)

	err := svc.Reindex(context.Background(), fileID)

	require.NoError(t, err)
}

func TestFileService_GetIndexStatus(t *testing.T) {
	t.Parallel()
	repo := new(mockFileRepo)
	svc := NewService(repo, new(mockStorage), nil, new(mockFileEventBus), 100*1024*1024)

	fileID := shared.NewID()
	now := time.Now()
	f := &file.File{
		ID:          fileID,
		IndexStatus: file.IndexCompleted,
		IndexedAt:   &now,
	}

	repo.On("FindByID", mock.Anything, fileID).Return(f, nil)

	resp, err := svc.GetIndexStatus(context.Background(), fileID)

	require.NoError(t, err)
	assert.Equal(t, fileID, resp.FileID)
	assert.Equal(t, file.IndexCompleted, resp.Status)
}
