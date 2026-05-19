package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/liang21/heka/internal/application/file"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/interface/http/response"
)

// tasks.md: T137 | spec.md: §4.10 File Handler Implementation

type FileService interface {
	Upload(ctx context.Context, userID, projectID shared.ID, name, contentType string, size int64, reader io.Reader) (*file.FileResponse, error)
	GetByID(ctx context.Context, fileID shared.ID) (*file.FileResponse, error)
	List(ctx context.Context, projectID shared.ID, page, pageSize int) ([]file.FileListResponse, int64, error)
	Delete(ctx context.Context, fileID shared.ID) error
	Reindex(ctx context.Context, fileID shared.ID) error
	GetIndexStatus(ctx context.Context, fileID shared.ID) (*file.IndexStatusResponse, error)
}

type FileHandler struct {
	svc           FileService
	maxUploadSize int64
}

func NewFileHandler(svc FileService, maxUploadSize int64) *FileHandler {
	return &FileHandler{
		svc:           svc,
		maxUploadSize: maxUploadSize,
	}
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(shared.ID)
	if !ok {
		response.Error(w, shared.ErrAuthForbidden)
		return
	}

	projectID, err := shared.ParseID(r.FormValue("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	uploadedFile, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, shared.NewAppError("FILE-VL-003", "no file uploaded", http.StatusBadRequest))
		return
	}
	defer uploadedFile.Close()

	contentType := header.Header.Get("Content-Type")
	size := header.Size

	if size > h.maxUploadSize {
		response.Error(w, shared.ErrFileTooLarge)
		return
	}

	f, err := h.svc.Upload(r.Context(), userID, projectID, header.Filename, contentType, size, uploadedFile)
	if err != nil {
		response.Error(w, shared.NewAppError("FILE-IE-001", "failed to upload file", http.StatusInternalServerError))
		return
	}

	response.Created(w, f)
}

func (h *FileHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	fileID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-005", "invalid file id", http.StatusBadRequest))
		return
	}

	f, err := h.svc.GetByID(r.Context(), fileID)
	if err != nil {
		response.Error(w, shared.ErrFileNotFound)
		return
	}

	response.Success(w, f)
}

func (h *FileHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := shared.ParseID(r.URL.Query().Get("project_id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-002", "invalid project id", http.StatusBadRequest))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	files, total, err := h.svc.List(r.Context(), projectID, page, pageSize)
	if err != nil {
		response.Error(w, shared.NewAppError("FILE-IE-002", "failed to list files", http.StatusInternalServerError))
		return
	}

	response.PageResult(w, files, total, page, pageSize)
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	fileID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-005", "invalid file id", http.StatusBadRequest))
		return
	}

	if err := h.svc.Delete(r.Context(), fileID); err != nil {
		response.Error(w, shared.NewAppError("FILE-IE-003", "failed to delete file", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "file deleted"})
}

func (h *FileHandler) Reindex(w http.ResponseWriter, r *http.Request) {
	fileID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-005", "invalid file id", http.StatusBadRequest))
		return
	}

	if err := h.svc.Reindex(r.Context(), fileID); err != nil {
		response.Error(w, shared.NewAppError("FILE-IE-004", "failed to reindex file", http.StatusInternalServerError))
		return
	}

	response.Success(w, map[string]string{"message": "reindex started"})
}

func (h *FileHandler) GetIndexStatus(w http.ResponseWriter, r *http.Request) {
	fileID, err := shared.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, shared.NewAppError("VL-005", "invalid file id", http.StatusBadRequest))
		return
	}

	status, err := h.svc.GetIndexStatus(r.Context(), fileID)
	if err != nil {
		response.Error(w, shared.NewAppError("FILE-IE-005", "failed to get index status", http.StatusInternalServerError))
		return
	}

	response.Success(w, status)
}
