package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
	"github.com/heqiucheng/qiling-agent/backend/internal/service"
)

type UploadsHandler struct {
	Service *service.QilingService
}

func (h UploadsHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	var req service.UploadConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "请求 JSON 格式不正确", nil)
		return
	}

	result, err := h.Service.UploadConversation(req)
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}

func (h UploadsHandler) Get(w http.ResponseWriter, r *http.Request) {
	uploadID := strings.TrimPrefix(r.URL.Path, "/api/uploads/")
	record, err := h.Service.Upload(uploadID)
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, record)
}

func (h UploadsHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	uploadID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/uploads/"), "/confirm")

	var req service.ConfirmUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "请求 JSON 格式不正确", nil)
		return
	}

	result, err := h.Service.ConfirmUpload(uploadID, req)
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}
