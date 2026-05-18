package shared

import "fmt"

// AppError is the universal error type used across all domains.
type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAppError constructs a new AppError.
func NewAppError(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

// ---------------------------------------------------------------------------
// Auth errors (AUTH)
// ---------------------------------------------------------------------------

var (
	ErrAuthInvalidCredentials = NewAppError("AUTH-AU-001", "invalid credentials", 401)
	ErrAuthTokenExpired       = NewAppError("AUTH-AU-002", "token expired", 401)
	ErrAuthForbidden          = NewAppError("AUTH-AU-003", "forbidden", 403)
)

// ---------------------------------------------------------------------------
// Test-case errors (TC)
// ---------------------------------------------------------------------------

var (
	ErrTestCaseNotFound      = NewAppError("TC-NF-001", "test case not found", 404)
	ErrTestCaseInvalidStatus = NewAppError("TC-VL-001", "invalid test case status transition", 400)
	ErrTestCaseConflict      = NewAppError("TC-CF-001", "test case version conflict", 409)
)

// ---------------------------------------------------------------------------
// Test-plan errors (TP)
// ---------------------------------------------------------------------------

var (
	ErrPlanNotFound        = NewAppError("TP-NF-001", "test plan not found", 404)
	ErrPlanInvalidStatus   = NewAppError("TP-VL-001", "invalid plan status transition", 400)
	ErrPlanNeedsCases      = NewAppError("TP-VL-002", "plan requires at least one test case to start", 400)
)

// ---------------------------------------------------------------------------
// Execution errors (EX)
// ---------------------------------------------------------------------------

var (
	ErrExecutionNotFound   = NewAppError("EX-NF-001", "execution not found", 404)
	ErrExecutionConflict   = NewAppError("EX-CF-001", "duplicate execution result", 409)
)

// ---------------------------------------------------------------------------
// File errors (FILE)
// ---------------------------------------------------------------------------

var (
	ErrFileNotFound     = NewAppError("FILE-NF-001", "file not found", 404)
	ErrFileInvalidType  = NewAppError("FILE-VL-001", "invalid file type", 400)
	ErrFileTooLarge     = NewAppError("FILE-VL-002", "file too large", 400)
)

// ---------------------------------------------------------------------------
// AI errors (AI)
// ---------------------------------------------------------------------------

var (
	ErrAIServiceUnavailable = NewAppError("AI-IE-001", "AI service unavailable", 500)
	ErrAIRateLimited        = NewAppError("AI-RT-001", "AI rate limited", 429)
	ErrAIInvalidInput       = NewAppError("AI-VL-001", "invalid AI input", 400)
)

// ---------------------------------------------------------------------------
// Project errors (PROJ)
// ---------------------------------------------------------------------------

var (
	ErrProjectNotFound  = NewAppError("PROJ-NF-001", "project not found", 404)
	ErrProjectNotMember = NewAppError("PROJ-AU-001", "not a project member", 403)
)

// ---------------------------------------------------------------------------
// User errors (USER)
// ---------------------------------------------------------------------------

var (
	ErrUserNotFound    = NewAppError("USER-NF-001", "user not found", 404)
	ErrUserEmailExists = NewAppError("USER-CF-001", "email already exists", 409)
)

// ---------------------------------------------------------------------------
// RAG errors (RAG)
// ---------------------------------------------------------------------------

var (
	ErrRAGIndexFailed  = NewAppError("RAG-IE-001", "indexing failed", 500)
	ErrRAGNoResults    = NewAppError("RAG-NF-001", "no results found", 404)
)

// ---------------------------------------------------------------------------
// System errors (SYS)
// ---------------------------------------------------------------------------

var (
	ErrSysInternal    = NewAppError("SYS-IE-001", "internal error", 500)
	ErrSysRateLimited = NewAppError("SYS-RT-001", "rate limited", 429)
	ErrSysValidation  = NewAppError("SYS-VL-001", "validation error", 400)
)
