package handler

import (
	"errors"
	"net/http"

	"github.com/heqiucheng/qiling-agent/backend/internal/apperror"
	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
)

func WriteServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		httpx.WriteError(w, r, statusForCode(appErr.Code), appErr.Code, appErr.Message, appErr.Details)
		return
	}
	httpx.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误", nil)
}

func statusForCode(code string) int {
	switch code {
	case "VALIDATION_ERROR", "EMPTY_CONTENT", "UNSUPPORTED_UPLOAD_TYPE":
		return http.StatusBadRequest
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT", "TASK_ALREADY_FINALIZED":
		return http.StatusConflict
	case "RATE_LIMITED":
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
