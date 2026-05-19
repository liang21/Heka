package response

import (
	"encoding/json"
	"net/http"

	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T098 | spec.md: §4.1 统一响应封装

// Success writes a success JSON response.
func Success(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"data":    data,
		"message": "success",
	})
}

// Created writes a 201 Created JSON response.
func Created(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"data":    data,
		"message": "created",
	})
}

// Error writes an error JSON response from AppError.
func Error(w http.ResponseWriter, appErr *shared.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    appErr.Code,
		"message": appErr.Message,
	})
}

// PageResult writes a paginated success response.
func PageResult(w http.ResponseWriter, data interface{}, total int64, page, pageSize int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"data":    data,
		"message": "success",
		"total":   total,
		"page":    page,
		"page_size": pageSize,
	})
}
