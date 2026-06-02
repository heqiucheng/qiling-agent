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
	report, err := h.Service.Report(reportID, httpx.ActorFromRequest(r))
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "" {
		format = "markdown"
	}
	if format != "markdown" {
		httpx.WriteError(w, r, http.StatusBadRequest, "UNSUPPORTED_EXPORT_FORMAT", "暂不支持该报告导出格式", map[string]any{"format": format})
		return
	}

	filename := strings.ReplaceAll(report.ID, "/", "_") + ".md"
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(report.Markdown))
}
