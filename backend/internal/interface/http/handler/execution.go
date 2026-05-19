package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/liang21/heka/internal/application/execution"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/interface/http/response"
)

// tasks.md: T135 | spec.md: §4.9 Execution Handler Implementation

type ExecutionService interface {
	Create(ctx context.Context, executorID, planID shared.ID, name string) (*execution.ExecutionResponse, error)
	GetByID(ctx context.Context, executionID shared.ID) (*execution.ExecutionResponse, error)
	GetSummary(ctx context.Context, executionID shared.ID) (*execution.ExecutionSummaryResponse, error)
	SubmitResult(ctx context.Context, executorID, executionID shared.ID, req execution.SubmitResultReq) error
	BatchSubmit(ctx context.Context, executorID, executionID shared.ID, req execution.BatchSubmitReq) error
}

type ExecutionHandler struct {
	svc ExecutionService
}

func NewExecutionHandler(svc ExecutionService) *ExecutionHandler {
	return &ExecutionHandler{svc: svc}
}

func (h *ExecutionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	var req struct {
		PlanID shared.ID `json:"plan_id"`
		Name   string    `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	exec, err := h.svc.Create(r.Context(), userID, req.PlanID, req.Name)
	if err != nil {
		response.Error(w, shared.NewAppError("EX-IE-001", "failed to create execution", http.StatusInternalServerError))
		return
	}

	response.Created(w, exec)
}

func (h *ExecutionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	executionID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-004", "invalid execution id", http.StatusBadRequest))
		return
	}

	exec, err := h.svc.GetByID(r.Context(), executionID)
	if err != nil {
		response.Error(w, shared.ErrExecutionNotFound)
		return
	}

	response.Success(w, exec)
}

func (h *ExecutionHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	executionID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-004", "invalid execution id", http.StatusBadRequest))
		return
	}

	summary, err := h.svc.GetSummary(r.Context(), executionID)
	if err != nil {
		response.Error(w, shared.NewAppError("EX-IE-002", "failed to get summary", http.StatusInternalServerError))
		return
	}

	response.Success(w, summary)
}

func (h *ExecutionHandler) SubmitResult(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	executionID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-004", "invalid execution id", http.StatusBadRequest))
		return
	}

	var req execution.SubmitResultReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	if err := h.svc.SubmitResult(r.Context(), userID, executionID, req); err != nil {
		response.Error(w, shared.NewAppError("EX-IE-003", "failed to submit result", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "result submitted"})
}

func (h *ExecutionHandler) BatchSubmit(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	executionID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-004", "invalid execution id", http.StatusBadRequest))
		return
	}

	var req execution.BatchSubmitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	if err := h.svc.BatchSubmit(r.Context(), userID, executionID, req); err != nil {
		response.Error(w, shared.NewAppError("EX-IE-004", "failed to batch submit", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "batch submitted"})
}
