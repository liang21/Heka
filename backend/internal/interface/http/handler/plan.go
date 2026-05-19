package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/liang21/heka/internal/application/plan"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/interface/http/response"
)

// tasks.md: T133 | spec.md: §4.8 Plan Handler Implementation

type PlanService interface {
	Create(ctx context.Context, userID shared.ID, req plan.CreatePlanReq) (*plan.PlanResponse, error)
	GetByID(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error)
	List(ctx context.Context, projectID shared.ID) ([]plan.PlanResponse, error)
	AddCase(ctx context.Context, planID, caseID shared.ID) error
	RemoveCase(ctx context.Context, planID, caseID shared.ID) error
	Start(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error)
	Pause(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error)
	Resume(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error)
	Complete(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error)
	Cancel(ctx context.Context, planID shared.ID) (*plan.PlanDetailResponse, error)
}

type PlanHandler struct {
	svc PlanService
}

func NewPlanHandler(svc PlanService) *PlanHandler {
	return &PlanHandler{svc: svc}
}

func (h *PlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	var req plan.CreatePlanReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	p, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.Error(w, shared.NewAppError("TP-IE-001", "failed to create plan", http.StatusInternalServerError))
		return
	}

	response.Created(w, p)
}

func (h *PlanHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	planID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-003", "invalid plan id", http.StatusBadRequest))
		return
	}

	p, err := h.svc.GetByID(r.Context(), planID)
	if err != nil {
		response.Error(w, shared.ErrPlanNotFound)
		return
	}

	response.Success(w, p)
}

func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(r.URL.Query().Get("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	plans, err := h.svc.List(r.Context(), projectID)
	if err != nil {
		response.Error(w, shared.NewAppError("TP-IE-002", "failed to list plans", http.StatusInternalServerError))
		return
	}

	response.Success(w, plans)
}

func (h *PlanHandler) AddCase(w http.ResponseWriter, r *http.Request) {
	planID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-003", "invalid plan id", http.StatusBadRequest))
		return
	}

	var req struct {
		CaseID shared.ID `json:"case_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	if err := h.svc.AddCase(r.Context(), planID, req.CaseID); err != nil {
		response.Error(w, shared.NewAppError("TP-IE-003", "failed to add case", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "case added"})
}

func (h *PlanHandler) RemoveCase(w http.ResponseWriter, r *http.Request) {
	planID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-003", "invalid plan id", http.StatusBadRequest))
		return
	}

	caseID, err := shared.ParseID(chi.URLParam(r, "case_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-009", "invalid case id", http.StatusBadRequest))
		return
	}

	if err := h.svc.RemoveCase(r.Context(), planID, caseID); err != nil {
		response.Error(w, shared.NewAppError("TP-IE-004", "failed to remove case", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "case removed"})
}

func (h *PlanHandler) Start(w http.ResponseWriter, r *http.Request) {
	planID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-003", "invalid plan id", http.StatusBadRequest))
		return
	}

	p, err := h.svc.Start(r.Context(), planID)
	if err != nil {
		response.Error(w, shared.NewAppError("TP-IE-005", "failed to start plan", http.StatusInternalServerError))
		return
	}

	response.Success(w, p)
}

func (h *PlanHandler) Pause(w http.ResponseWriter, r *http.Request) {
	planID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-003", "invalid plan id", http.StatusBadRequest))
		return
	}

	p, err := h.svc.Pause(r.Context(), planID)
	if err != nil {
		response.Error(w, shared.NewAppError("TP-IE-006", "failed to pause plan", http.StatusInternalServerError))
		return
	}

	response.Success(w, p)
}

func (h *PlanHandler) Resume(w http.ResponseWriter, r *http.Request) {
	planID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-003", "invalid plan id", http.StatusBadRequest))
		return
	}

	p, err := h.svc.Resume(r.Context(), planID)
	if err != nil {
		response.Error(w, shared.NewAppError("TP-IE-007", "failed to resume plan", http.StatusInternalServerError))
		return
	}

	response.Success(w, p)
}

func (h *PlanHandler) Complete(w http.ResponseWriter, r *http.Request) {
	planID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-003", "invalid plan id", http.StatusBadRequest))
		return
	}

	p, err := h.svc.Complete(r.Context(), planID)
	if err != nil {
		response.Error(w, shared.NewAppError("TP-IE-008", "failed to complete plan", http.StatusInternalServerError))
		return
	}

	response.Success(w, p)
}

func (h *PlanHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	planID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-003", "invalid plan id", http.StatusBadRequest))
		return
	}

	p, err := h.svc.Cancel(r.Context(), planID)
	if err != nil {
		response.Error(w, shared.NewAppError("TP-IE-009", "failed to cancel plan", http.StatusInternalServerError))
		return
	}

	response.Success(w, p)
}
