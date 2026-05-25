package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/liang21/heka/internal/application/testcase"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/interface/http/response"
)

// tasks.md: T131 | spec.md: §4.4-4.7 TestCase Handler Implementation

type TestCaseService interface {
	CreateModule(ctx context.Context, userID shared.ID, req testcase.CreateModuleReq) (*testcase.ModuleDTO, error)
	GetModuleTree(ctx context.Context, projectID shared.ID) ([]testcase.ModuleDTO, error)
	UpdateModule(ctx context.Context, userID, moduleID shared.ID, req testcase.UpdateModuleReq) (*testcase.ModuleDTO, error)
	DeleteModule(ctx context.Context, userID, moduleID shared.ID) error

	CreateTag(ctx context.Context, userID, projectID shared.ID, req testcase.CreateTagReq) (*testcase.TagDTO, error)
	ListTags(ctx context.Context, projectID shared.ID) ([]testcase.TagDTO, error)
	DeleteTag(ctx context.Context, tagID shared.ID) error

	CreateTestCase(ctx context.Context, userID, projectID shared.ID, req testcase.CreateTestCaseReq) (*testcase.TestCaseResponse, error)
	GetTestCase(ctx context.Context, caseID shared.ID) (*testcase.TestCaseResponse, error)
	UpdateTestCase(ctx context.Context, userID, caseID shared.ID, req testcase.UpdateTestCaseReq) (*testcase.TestCaseResponse, error)
	DeleteTestCase(ctx context.Context, userID, caseID shared.ID) error
	ListTestCases(ctx context.Context, f testcase.TestCaseFilter) ([]testcase.TestCaseListResponse, int64, error)

	CreateCollection(ctx context.Context, userID shared.ID, req testcase.CreateCollectionReq) (*testcase.CollectionDTO, error)
	ListCollections(ctx context.Context, projectID shared.ID) ([]testcase.CollectionDTO, error)
	AddToCollection(ctx context.Context, collectionID, caseID shared.ID) error
	RemoveFromCollection(ctx context.Context, collectionID, caseID shared.ID) error
}

type TestCaseHandler struct {
	svc TestCaseService
}

func NewTestCaseHandler(svc TestCaseService) *TestCaseHandler {
	return &TestCaseHandler{svc: svc}
}

// Module handlers
func (h *TestCaseHandler) CreateModule(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	var req testcase.CreateModuleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	module, err := h.svc.CreateModule(r.Context(), userID, req)
	if err != nil {
		response.Error(w, shared.NewAppError("TC-IE-001", "failed to create module", http.StatusInternalServerError))
		return
	}

	response.Created(w, module)
}

func (h *TestCaseHandler) GetModuleTree(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(r.URL.Query().Get("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	tree, err := h.svc.GetModuleTree(r.Context(), projectID)
	if err != nil {
		response.Error(w, shared.NewAppError("TC-IE-002", "failed to get module tree", http.StatusInternalServerError))
		return
	}

	response.Success(w, tree)
}

func (h *TestCaseHandler) UpdateModule(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	moduleID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-006", "invalid module id", http.StatusBadRequest))
		return
	}

	var req testcase.UpdateModuleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	module, err := h.svc.UpdateModule(r.Context(), userID, moduleID, req)
	if err != nil {
		response.Error(w, shared.NewAppError("TC-IE-003", "failed to update module", http.StatusInternalServerError))
		return
	}

	response.Success(w, module)
}

func (h *TestCaseHandler) DeleteModule(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	moduleID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-006", "invalid module id", http.StatusBadRequest))
		return
	}

	if err := h.svc.DeleteModule(r.Context(), userID, moduleID); err != nil {
		response.Error(w, shared.NewAppError("TC-IE-004", "failed to delete module", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "module deleted"})
}

// Tag handlers
func (h *TestCaseHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	var req testcase.CreateTagReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	tag, err := h.svc.CreateTag(r.Context(), userID, req.ProjectID, req)
	if err != nil {
		response.Error(w, shared.NewAppError("TC-IE-005", "failed to create tag", http.StatusInternalServerError))
		return
	}

	response.Created(w, tag)
}

func (h *TestCaseHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(r.URL.Query().Get("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	tags, err := h.svc.ListTags(r.Context(), projectID)
	if err != nil {
		response.Error(w, shared.NewAppError("TC-IE-006", "failed to list tags", http.StatusInternalServerError))
		return
	}

	response.Success(w, tags)
}

func (h *TestCaseHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	tagID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-007", "invalid tag id", http.StatusBadRequest))
		return
	}

	if err := h.svc.DeleteTag(r.Context(), tagID); err != nil {
		response.Error(w, shared.NewAppError("TC-IE-007", "failed to delete tag", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "tag deleted"})
}

// TestCase handlers
func (h *TestCaseHandler) CreateTestCase(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	projectID, err := shared.ParseID(r.URL.Query().Get("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	var req testcase.CreateTestCaseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	tc, err := h.svc.CreateTestCase(r.Context(), userID, projectID, req)
	if err != nil {
		response.Error(w, shared.NewAppError("TC-IE-008", "failed to create test case", http.StatusInternalServerError))
		return
	}

	response.Created(w, tc)
}

func (h *TestCaseHandler) GetTestCase(w http.ResponseWriter, r *http.Request) {
	caseID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-008", "invalid test case id", http.StatusBadRequest))
		return
	}

	tc, err := h.svc.GetTestCase(r.Context(), caseID)
	if err != nil {
		response.Error(w, shared.ErrTestCaseNotFound)
		return
	}

	response.Success(w, tc)
}

func (h *TestCaseHandler) UpdateTestCase(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	caseID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-008", "invalid test case id", http.StatusBadRequest))
		return
	}

	var req testcase.UpdateTestCaseReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	tc, err := h.svc.UpdateTestCase(r.Context(), userID, caseID, req)
	if err != nil {
		response.Error(w, shared.NewAppError("TC-IE-009", "failed to update test case", http.StatusInternalServerError))
		return
	}

	response.Success(w, tc)
}

func (h *TestCaseHandler) DeleteTestCase(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	caseID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-008", "invalid test case id", http.StatusBadRequest))
		return
	}

	if err := h.svc.DeleteTestCase(r.Context(), userID, caseID); err != nil {
		response.Error(w, shared.NewAppError("TC-IE-010", "failed to delete test case", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "test case deleted"})
}

func (h *TestCaseHandler) ListTestCases(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	projectID, err := shared.ParseID(query.Get("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	cases, total, err := h.svc.ListTestCases(r.Context(), testcase.TestCaseFilter{
		ProjectID: projectID,
		ModuleID:  parseOptionalID(query.Get("module_id")),
		TagID:     parseOptionalID(query.Get("tag_id")),
		Status:    query.Get("status"),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		response.Error(w, shared.NewAppError("TC-IE-011", "failed to list test cases", http.StatusInternalServerError))
		return
	}

	response.PageResult(w, cases, total, page, pageSize)
}

// Collection handlers
func (h *TestCaseHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	var req testcase.CreateCollectionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	coll, err := h.svc.CreateCollection(r.Context(), userID, req)
	if err != nil {
		response.Error(w, shared.NewAppError("TC-IE-012", "failed to create collection", http.StatusInternalServerError))
		return
	}

	response.Created(w, coll)
}

func (h *TestCaseHandler) ListCollections(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(r.URL.Query().Get("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	colls, err := h.svc.ListCollections(r.Context(), projectID)
	if err != nil {
		response.Error(w, shared.NewAppError("TC-IE-013", "failed to list collections", http.StatusInternalServerError))
		return
	}

	response.Success(w, colls)
}

func (h *TestCaseHandler) AddToCollection(w http.ResponseWriter, r *http.Request) {
	collectionID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-010", "invalid collection id", http.StatusBadRequest))
		return
	}

	var req struct {
		CaseID shared.ID `json:"case_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, shared.NewAppError("VL-001", "invalid request", http.StatusBadRequest))
		return
	}

	if err := h.svc.AddToCollection(r.Context(), collectionID, req.CaseID); err != nil {
		response.Error(w, shared.NewAppError("TC-IE-014", "failed to add to collection", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "added to collection"})
}

func (h *TestCaseHandler) RemoveFromCollection(w http.ResponseWriter, r *http.Request) {
	collectionID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-010", "invalid collection id", http.StatusBadRequest))
		return
	}

	caseID, err := shared.ParseID(chi.URLParam(r, "case_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-009", "invalid case id", http.StatusBadRequest))
		return
	}

	if err := h.svc.RemoveFromCollection(r.Context(), collectionID, caseID); err != nil {
		response.Error(w, shared.NewAppError("TC-IE-015", "failed to remove from collection", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "removed from collection"})
}

func parseOptionalID(s string) *shared.ID {
	if s == "" {
		return nil
	}
	id, err := shared.ParseID(s)
	if err != nil {
		return nil
	}
	return &id
}
