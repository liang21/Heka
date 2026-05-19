package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/liang21/heka/internal/application/testcase"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T130 | spec.md: §4.4-4.7 TestCase Handler TDD RED

type mockTCService struct {
	mock.Mock
}

func (m *mockTCService) CreateModule(ctx context.Context, userID shared.ID, req testcase.CreateModuleReq) (*testcase.ModuleDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*testcase.ModuleDTO), args.Error(1)
}

func (m *mockTCService) ListTags(ctx context.Context, projectID shared.ID) ([]testcase.TagDTO, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]testcase.TagDTO), args.Error(1)
}

func (m *mockTCService) CreateTestCase(ctx context.Context, userID, projectID shared.ID, req testcase.CreateTestCaseReq) (*testcase.TestCaseResponse, error) {
	args := m.Called(ctx, userID, projectID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*testcase.TestCaseResponse), args.Error(1)
}

func (m *mockTCService) GetTestCase(ctx context.Context, caseID shared.ID) (*testcase.TestCaseResponse, error) {
	args := m.Called(ctx, caseID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*testcase.TestCaseResponse), args.Error(1)
}

func (m *mockTCService) ListTestCases(ctx context.Context, f testcase.TestCaseFilter) ([]testcase.TestCaseListResponse, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil { return nil, args.Get(1).(int64), args.Error(2) }
	return args.Get(0).([]testcase.TestCaseListResponse), args.Get(1).(int64), args.Error(2)
}

func (m *mockTCService) CreateCollection(ctx context.Context, userID shared.ID, req testcase.CreateCollectionReq) (*testcase.CollectionDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*testcase.CollectionDTO), args.Error(1)
}

func (m *mockTCService) ListCollections(ctx context.Context, projectID shared.ID) ([]testcase.CollectionDTO, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]testcase.CollectionDTO), args.Error(1)
}

func (m *mockTCService) AddToCollection(ctx context.Context, collectionID, caseID shared.ID) error {
	args := m.Called(ctx, collectionID, caseID)
	return args.Error(0)
}

func (m *mockTCService) RemoveFromCollection(ctx context.Context, collectionID, caseID shared.ID) error {
	args := m.Called(ctx, collectionID, caseID)
	return args.Error(0)
}

func (m *mockTCService) GetModuleTree(ctx context.Context, projectID shared.ID) ([]testcase.ModuleDTO, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]testcase.ModuleDTO), args.Error(1)
}

func (m *mockTCService) UpdateModule(ctx context.Context, moduleID shared.ID, req testcase.UpdateModuleReq) (*testcase.ModuleDTO, error) {
	args := m.Called(ctx, moduleID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*testcase.ModuleDTO), args.Error(1)
}

func (m *mockTCService) DeleteModule(ctx context.Context, moduleID shared.ID) error {
	args := m.Called(ctx, moduleID)
	return args.Error(0)
}

func (m *mockTCService) CreateTag(ctx context.Context, projectID shared.ID, req testcase.CreateTagReq) (*testcase.TagDTO, error) {
	args := m.Called(ctx, projectID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*testcase.TagDTO), args.Error(1)
}

func (m *mockTCService) DeleteTag(ctx context.Context, tagID shared.ID) error {
	args := m.Called(ctx, tagID)
	return args.Error(0)
}

func (m *mockTCService) UpdateTestCase(ctx context.Context, caseID shared.ID, req testcase.UpdateTestCaseReq) (*testcase.TestCaseResponse, error) {
	args := m.Called(ctx, caseID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*testcase.TestCaseResponse), args.Error(1)
}

func (m *mockTCService) DeleteTestCase(ctx context.Context, caseID shared.ID) error {
	args := m.Called(ctx, caseID)
	return args.Error(0)
}

func TestTestCaseHandler_CreateModule(t *testing.T) {
	t.Parallel()
	svc := new(mockTCService)
	h := NewTestCaseHandler(svc)

	userID := shared.NewID()
	projectID := shared.NewID()

	svc.On("CreateModule", mock.Anything, userID, mock.Anything).Return(&testcase.ModuleDTO{
		ID: shared.NewID(), ProjectID: projectID, Name: "Auth",
	}, nil)

	body := map[string]string{"project_id": projectID.String(), "name": "Auth"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/modules", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateModule(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTestCaseHandler_ListTestCases(t *testing.T) {
	t.Parallel()
	svc := new(mockTCService)
	h := NewTestCaseHandler(svc)

	projectID := shared.NewID()

	svc.On("ListTestCases", mock.Anything, mock.Anything).Return([]testcase.TestCaseListResponse{
		{ID: shared.NewID(), Title: "Case 1"},
	}, int64(1), nil)

	req := httptest.NewRequest("GET", "/api/v1/testcases?project_id="+projectID.String(), nil)
	w := httptest.NewRecorder()

	h.ListTestCases(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTestCaseHandler_CreateCollection(t *testing.T) {
	t.Parallel()
	svc := new(mockTCService)
	h := NewTestCaseHandler(svc)

	userID := shared.NewID()
	projectID := shared.NewID()

	svc.On("CreateCollection", mock.Anything, userID, mock.Anything).Return(&testcase.CollectionDTO{
		ID: shared.NewID(), ProjectID: projectID, Name: "Regression",
	}, nil)

	body := map[string]string{"project_id": projectID.String(), "name": "Regression"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/collections", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.CreateCollection(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
