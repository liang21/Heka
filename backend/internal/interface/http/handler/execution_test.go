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

	"github.com/liang21/heka/internal/application/execution"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T134 | spec.md: §4.9 Execution Handler TDD RED

type mockExecutionService struct {
	mock.Mock
}

func (m *mockExecutionService) Create(ctx context.Context, executorID, planID shared.ID, name string) (*execution.ExecutionResponse, error) {
	args := m.Called(ctx, executorID, planID, name)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*execution.ExecutionResponse), args.Error(1)
}

func (m *mockExecutionService) GetByID(ctx context.Context, executionID shared.ID) (*execution.ExecutionResponse, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*execution.ExecutionResponse), args.Error(1)
}

func (m *mockExecutionService) GetSummary(ctx context.Context, executionID shared.ID) (*execution.ExecutionSummaryResponse, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*execution.ExecutionSummaryResponse), args.Error(1)
}

func (m *mockExecutionService) SubmitResult(ctx context.Context, executorID, executionID shared.ID, req execution.SubmitResultReq) error {
	return m.Called(ctx, executorID, executionID, req).Error(0)
}

func (m *mockExecutionService) BatchSubmit(ctx context.Context, executorID, executionID shared.ID, req execution.BatchSubmitReq) error {
	return m.Called(ctx, executorID, executionID, req).Error(0)
}

func TestExecutionHandler_SubmitResult(t *testing.T) {
	t.Parallel()
	svc := new(mockExecutionService)
	h := NewExecutionHandler(svc)

	execID := shared.NewID()
	caseID := shared.NewID()
	userID := shared.NewID()

	svc.On("SubmitResult", mock.Anything, userID, execID, mock.Anything).Return(nil)

	body := map[string]interface{}{"test_case_id": caseID.String(), "status": "passed"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/executions/"+execID.String()+"/results", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)
	req = setChiURLParam(req, "id", execID.String())
	w := httptest.NewRecorder()

	h.SubmitResult(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExecutionHandler_BatchSubmit(t *testing.T) {
	t.Parallel()
	svc := new(mockExecutionService)
	h := NewExecutionHandler(svc)

	execID := shared.NewID()
	userID := shared.NewID()

	svc.On("BatchSubmit", mock.Anything, userID, execID, mock.Anything).Return(nil)

	body := map[string]interface{}{"results": []map[string]interface{}{
		{"test_case_id": shared.NewID().String(), "status": "passed"},
		{"test_case_id": shared.NewID().String(), "status": "failed"},
	}}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/executions/"+execID.String()+"/results/batch", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)
	req = setChiURLParam(req, "id", execID.String())
	w := httptest.NewRecorder()

	h.BatchSubmit(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExecutionHandler_GetByID(t *testing.T) {
	t.Parallel()
	svc := new(mockExecutionService)
	h := NewExecutionHandler(svc)

	execID := shared.NewID()
	svc.On("GetByID", mock.Anything, execID).Return(&execution.ExecutionResponse{
		ID: execID, Status: "in_progress",
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/executions/"+execID.String(), nil)
	req = setChiURLParam(req, "id", execID.String())
	w := httptest.NewRecorder()

	h.GetByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
