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

	"github.com/liang21/heka/internal/application/ai"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T138 | spec.md: §4.11 AI Handler TDD RED

type mockAIService struct {
	mock.Mock
}

func (m *mockAIService) GenerateTestCases(ctx context.Context, userID shared.ID, req ai.GenerateRequest) (*ai.TaskResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*ai.TaskResponse), args.Error(1)
}

func (m *mockAIService) Analyze(ctx context.Context, userID shared.ID, req ai.AnalyzeRequest) (*ai.TaskResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*ai.TaskResponse), args.Error(1)
}

func (m *mockAIService) GetTask(ctx context.Context, taskID shared.ID) (*ai.TaskResponse, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*ai.TaskResponse), args.Error(1)
}

func TestAIHandler_GenerateTestCases(t *testing.T) {
	t.Parallel()
	svc := new(mockAIService)
	h := NewAIHandler(svc)

	userID := shared.NewID()
	projectID := shared.NewID()
	fileID := shared.NewID()

	svc.On("GenerateTestCases", mock.Anything, userID, mock.Anything).Return(&ai.TaskResponse{
		TaskID: shared.NewID(), Status: "pending",
	}, nil)

	body := map[string]interface{}{"project_id": projectID.String(), "file_id": fileID.String(), "count": 5}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/ai/generate-testcases", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GenerateTestCases(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["data"])
}

func TestAIHandler_GetTask(t *testing.T) {
	t.Parallel()
	svc := new(mockAIService)
	h := NewAIHandler(svc)

	taskID := shared.NewID()
	svc.On("GetTask", mock.Anything, taskID).Return(&ai.TaskResponse{
		TaskID: taskID, Status: "completed",
	}, nil)

	req := httptest.NewRequest("GET", "/api/v1/ai/tasks/"+taskID.String(), nil)
	req = setChiURLParam(req, "id", taskID.String())
	w := httptest.NewRecorder()

	h.GetTask(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAIHandler_Analyze(t *testing.T) {
	t.Parallel()
	svc := new(mockAIService)
	h := NewAIHandler(svc)

	userID := shared.NewID()
	projectID := shared.NewID()

	svc.On("Analyze", mock.Anything, userID, mock.Anything).Return(&ai.TaskResponse{
		TaskID: shared.NewID(), Status: "pending",
	}, nil)

	body := map[string]interface{}{"project_id": projectID.String(), "description": "Changed login flow"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/v1/ai/analyze", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.Analyze(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
