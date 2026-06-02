package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
	"github.com/heqiucheng/qiling-agent/backend/internal/service"
)

type ReportsHandler struct {
	Service *service.QilingService
}

func (h ReportsHandler) CustomerIntent(w http.ResponseWriter, r *http.Request) {
	var req service.CustomerIntentReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "请求 JSON 格式不正确", nil)
		return
	}
	report, err := h.Service.CustomerIntentReport(req, httpx.ActorFromRequest(r))
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, report)
}

func (h ReportsHandler) List(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, h.Service.Reports(r, httpx.ActorFromRequest(r)))
}

func (h ReportsHandler) Get(w http.ResponseWriter, r *http.Request) {
	reportID := strings.TrimPrefix(r.URL.Path, "/api/reports/")
	report, err := h.Service.Report(reportID, httpx.ActorFromRequest(r))
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, report)
}

func (h ReportsHandler) Export(w http.ResponseWriter, r *http.Request) {
	reportID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/reports/"), "/export")
	export, err := h.Service.ExportReport(reportID, r.URL.Query().Get("format"), httpx.ActorFromRequest(r))
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", export.ContentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+export.Filename+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(export.Body)
}
