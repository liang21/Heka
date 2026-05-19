package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/liang21/heka/internal/application/report"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/interface/http/response"
)

// tasks.md: T141 | spec.md: §4.12 Report Handler Implementation

type ReportService interface {
	PlanReport(ctx context.Context, planID shared.ID) (*report.PlanReportResponse, error)
	Coverage(ctx context.Context, projectID shared.ID) (*report.CoverageResponse, error)
	Trend(ctx context.Context, projectID shared.ID, days int) (*report.TrendResponse, error)
	BugDistribution(ctx context.Context, projectID shared.ID) (*report.BugDistributionResponse, error)
	Workload(ctx context.Context, userID shared.ID) (*report.WorkloadResponse, error)
}

type ReportHandler struct {
	svc ReportService
}

func NewReportHandler(svc ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) PlanReport(w http.ResponseWriter, r *http.Request) {
	planID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-003", "invalid plan id", http.StatusBadRequest))
		return
	}

	rep, err := h.svc.PlanReport(r.Context(), planID)
	if err != nil {
		response.Error(w, shared.NewAppError("RP-IE-001", "failed to get plan report", http.StatusInternalServerError))
		return
	}

	response.Success(w, rep)
}

func (h *ReportHandler) Coverage(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(r.URL.Query().Get("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	rep, err := h.svc.Coverage(r.Context(), projectID)
	if err != nil {
		response.Error(w, shared.NewAppError("RP-IE-002", "failed to get coverage", http.StatusInternalServerError))
		return
	}

	response.Success(w, rep)
}

func (h *ReportHandler) Trend(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(r.URL.Query().Get("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}

	rep, err := h.svc.Trend(r.Context(), projectID, days)
	if err != nil {
		response.Error(w, shared.NewAppError("RP-IE-003", "failed to get trend", http.StatusInternalServerError))
		return
	}

	response.Success(w, rep)
}

func (h *ReportHandler) BugDistribution(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(r.URL.Query().Get("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	rep, err := h.svc.BugDistribution(r.Context(), projectID)
	if err != nil {
		response.Error(w, shared.NewAppError("RP-IE-004", "failed to get bug distribution", http.StatusInternalServerError))
		return
	}

	response.Success(w, rep)
}

func (h *ReportHandler) Workload(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	rep, err := h.svc.Workload(r.Context(), userID)
	if err != nil {
		response.Error(w, shared.NewAppError("RP-IE-005", "failed to get workload", http.StatusInternalServerError))
		return
	}

	response.Success(w, rep)
}
