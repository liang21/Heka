package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/liang21/heka/internal/application/ai"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/interface/http/response"
)

// tasks.md: T139 | spec.md: §4.11 AI Handler Implementation

type AIService interface {
	GenerateTestCases(ctx context.Context, userID shared.ID, req ai.GenerateRequest) (*ai.TaskResponse, error)
	Analyze(ctx context.Context, userID shared.ID, req ai.AnalyzeRequest) (*ai.TaskResponse, error)
	GetTask(ctx context.Context, taskID shared.ID) (*ai.TaskResponse, error)
}

type AIHandler struct {
	svc AIService
}

func NewAIHandler(svc AIService) *AIHandler {
	return &AIHandler{svc: svc}
}

func (h *AIHandler) GenerateTestCases(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	var req ai.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	task, err := h.svc.GenerateTestCases(r.Context(), userID, req)
	if err != nil {
		response.Error(w, shared.NewAppError("AI-IE-001", "failed to start generation", http.StatusInternalServerError))
		return
	}

	response.Created(w, task)
}

func (h *AIHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid task id", http.StatusBadRequest))
		return
	}

	task, err := h.svc.GetTask(r.Context(), taskID)
	if err != nil {
		response.Error(w, shared.ErrTestCaseNotFound)
		return
	}

	response.Success(w, task)
}

func (h *AIHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	var req ai.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	task, err := h.svc.Analyze(r.Context(), userID, req)
	if err != nil {
		response.Error(w, shared.NewAppError("AI-IE-002", "failed to start analysis", http.StatusInternalServerError))
		return
	}

	response.Created(w, task)
}

func (h *AIHandler) TaskEvents(w http.ResponseWriter, r *http.Request) {
	taskID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid task id", http.StatusBadRequest))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, shared.NewAppError("SSE-001", "streaming not supported", http.StatusInternalServerError))
		return
	}

	for i := 0; i < 3; i++ {
		task, _ := h.svc.GetTask(r.Context(), taskID)

		data, _ := json.Marshal(task)
		w.Write([]byte("data: " + string(data) + "\n\n"))
		flusher.Flush()

		if task.Status == "completed" || task.Status == "failed" {
			break
		}
	}
}
