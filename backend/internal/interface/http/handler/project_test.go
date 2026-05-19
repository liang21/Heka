package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/liang21/heka/internal/application/project"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T128 | spec.md: §4.3 Project Handler TDD RED

type mockProjService struct {
	mock.Mock
}

func (m *mockProjService) Create(ctx context.Context, userID shared.ID, req project.CreateProjectReq) (*project.ProjectResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*project.ProjectResponse), args.Error(1)
}

func (m *mockProjService) GetByID(ctx context.Context, projectID shared.ID) (*project.ProjectResponse, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*project.ProjectResponse), args.Error(1)
}

func (m *mockProjService) ListByUser(ctx context.Context, userID shared.ID) ([]*project.ProjectResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*project.ProjectResponse), args.Error(1)
}

func (m *mockProjService) AddMember(ctx context.Context, projectID, userID shared.ID, req project.AddMemberReq) error {
	args := m.Called(ctx, projectID, userID, req)
	return args.Error(0)
}

func TestProjectHandler_GetProject(t *testing.T) {
	t.Parallel()
	svc := new(mockProjService)
	h := NewProjectHandler(svc)

	projectID := shared.NewID()
	svc.On("GetByID", mock.Anything, projectID).Return(&project.ProjectResponse{
		ID: projectID, Name: "Test Project",
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/projects/"+projectID.String(), nil)
	req = setChiURLParam(req, "id", projectID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProjectHandler_ListProjects(t *testing.T) {
	t.Parallel()
	svc := new(mockProjService)
	h := NewProjectHandler(svc)

	userID := shared.NewID()
	svc.On("ListByUser", mock.Anything, userID).Return([]*project.ProjectResponse{
		{ID: shared.NewID(), Name: "Project 1"},
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/projects", nil)
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
