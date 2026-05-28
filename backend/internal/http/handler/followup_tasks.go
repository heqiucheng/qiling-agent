package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
	"github.com/heqiucheng/qiling-agent/backend/internal/service"
)

type FollowupTasksHandler struct {
	Service *service.QilingService
}

func (h FollowupTasksHandler) List(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, h.Service.FollowupTasks(r, httpx.ActorFromRequest(r)))
}

func (h FollowupTasksHandler) Copy(w http.ResponseWriter, r *http.Request) {
	taskID := taskIDFromActionPath(r.URL.Path, "/copy")

	var req service.CopyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "请求 JSON 格式不正确", nil)
		return
	}

	result, err := h.Service.CopyTask(taskID, req)
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}

func (h FollowupTasksHandler) Skip(w http.ResponseWriter, r *http.Request) {
	taskID := taskIDFromActionPath(r.URL.Path, "/skip")

	var req service.SkipTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "请求 JSON 格式不正确", nil)
		return
	}

	result, err := h.Service.SkipTask(taskID, req)
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}

func (h FollowupTasksHandler) MarkWrong(w http.ResponseWriter, r *http.Request) {
	taskID := taskIDFromActionPath(r.URL.Path, "/mark-wrong")

	var req service.MarkWrongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "请求 JSON 格式不正确", nil)
		return
	}

	result, err := h.Service.MarkTaskWrong(taskID, req)
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}

func (h FollowupTasksHandler) Regenerate(w http.ResponseWriter, r *http.Request) {
	taskID := taskIDFromActionPath(r.URL.Path, "/regenerate")

	var req service.RegenerateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "请求 JSON 格式不正确", nil)
		return
	}

	result, err := h.Service.RegenerateTask(taskID, req)
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}

func taskIDFromActionPath(path string, action string) string {
	trimmed := strings.TrimSuffix(strings.TrimPrefix(path, "/api/followup-tasks/"), action)
	return strings.Trim(trimmed, "/")
}
