package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/liang21/heka/internal/application/project"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/interface/http/response"
)

// tasks.md: T129 | spec.md: §4.3 Project Handler Implementation

type ProjectService interface {
	Create(ctx context.Context, userID shared.ID, req project.CreateProjectReq) (*project.ProjectResponse, error)
	GetByID(ctx context.Context, projectID shared.ID) (*project.ProjectResponse, error)
	ListByUser(ctx context.Context, userID shared.ID) ([]*project.ProjectResponse, error)
	AddMember(ctx context.Context, projectID, userID shared.ID, req project.AddMemberReq) error
}

type ProjectHandler struct {
	svc ProjectService
}

func NewProjectHandler(svc ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	var req project.CreateProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	proj, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.Error(w, shared.NewAppError("PJ-IE-001", "failed to create project", http.StatusInternalServerError))
		return
	}

	response.Created(w, proj)
}

func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	proj, err := h.svc.GetByID(r.Context(), projectID)
	if err != nil {
		response.Error(w, shared.ErrTestCaseNotFound)
		return
	}

	response.Success(w, proj)
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	projs, err := h.svc.ListByUser(r.Context(), userID)
	if err != nil {
		response.Error(w, shared.NewAppError("PJ-IE-002", "failed to list projects", http.StatusInternalServerError))
		return
	}

	response.Success(w, projs)
}

func (h *ProjectHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	var req project.AddMemberReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	if err := h.svc.AddMember(r.Context(), projectID, userID, req); err != nil {
		response.Error(w, shared.NewAppError("PJ-IE-003", "failed to add member", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "member added"})
}

func (h *ProjectHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	proj, err := h.svc.GetByID(r.Context(), projectID)
	if err != nil {
		response.Error(w, shared.ErrTestCaseNotFound)
		return
	}

	response.Success(w, proj)
}
