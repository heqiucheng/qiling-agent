package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heqiucheng/qiling-agent/backend/internal/config"
	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
)

func TestHealthEndpoint(t *testing.T) {
	router := NewRouter(config.Config{Addr: ":0", Env: "test"})
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header")
	}

	var body httpx.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error != nil {
		t.Fatalf("expected nil error, got %#v", body.Error)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", body.Data)
	}
	if data["status"] != "ok" {
		t.Fatalf("expected status ok, got %#v", data["status"])
	}
	if body.Meta.RequestID == "" {
		t.Fatal("expected response meta request id")
	}
}

func TestDashboardSummaryEndpoint(t *testing.T) {
	body := getJSON(t, "/api/dashboard/summary")
	data := responseData(t, body)

	metrics, ok := data["metrics"].([]any)
	if !ok || len(metrics) == 0 {
		t.Fatalf("expected metrics, got %#v", data["metrics"])
	}
	priorityTasks, ok := data["priority_tasks"].([]any)
	if !ok || len(priorityTasks) == 0 {
		t.Fatalf("expected priority tasks, got %#v", data["priority_tasks"])
	}
	if _, ok := data["daily_review"].(map[string]any); !ok {
		t.Fatalf("expected daily_review object, got %#v", data["daily_review"])
	}
}

func TestCustomersEndpointFiltersAndPaginates(t *testing.T) {
	body := getJSON(t, "/api/customers?intent=high&page=1&page_size=1")
	data := responseData(t, body)

	if data["page"] != float64(1) {
		t.Fatalf("expected page 1, got %#v", data["page"])
	}
	if data["page_size"] != float64(1) {
		t.Fatalf("expected page size 1, got %#v", data["page_size"])
	}
	if data["total"] != float64(2) {
		t.Fatalf("expected two high-intent customers, got %#v", data["total"])
	}
	items, ok := data["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected one paged item, got %#v", data["items"])
	}
}

func TestFollowupTasksEndpointFiltersByStatus(t *testing.T) {
	body := getJSON(t, "/api/followup-tasks?status=pending")
	data := responseData(t, body)

	if data["total"] != float64(3) {
		t.Fatalf("expected three pending tasks, got %#v", data["total"])
	}
	items, ok := data["items"].([]any)
	if !ok || len(items) != 3 {
		t.Fatalf("expected three task items, got %#v", data["items"])
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got %#v", items[0])
	}
	if _, ok := first["recommendation"].(map[string]any); !ok {
		t.Fatalf("expected recommendation object, got %#v", first["recommendation"])
	}
}

func TestUploadConversationFlow(t *testing.T) {
	router := NewRouter(config.Config{Addr: ":0", Env: "test"})

	uploadBody := postJSON(t, router, "/api/uploads/conversations", `{
		"source_type": "pasted_text",
		"content": "王女士 09:20 这个价格还能优惠吗？",
		"owner_id": "usr_001"
	}`, http.StatusOK)
	uploadData := responseData(t, uploadBody)
	uploadID, ok := uploadData["upload_id"].(string)
	if !ok || uploadID == "" {
		t.Fatalf("expected upload id, got %#v", uploadData["upload_id"])
	}
	if uploadData["status"] != "needs_confirmation" {
		t.Fatalf("expected needs_confirmation, got %#v", uploadData["status"])
	}
	parsedCustomer, ok := uploadData["parsed_customer"].(map[string]any)
	if !ok {
		t.Fatalf("expected parsed customer, got %#v", uploadData["parsed_customer"])
	}
	if parsedCustomer["name"] != "王女士" {
		t.Fatalf("expected parsed customer name 王女士, got %#v", parsedCustomer["name"])
	}

	getBody := requestJSON(t, router, http.MethodGet, "/api/uploads/"+uploadID, "", http.StatusOK)
	getData := responseData(t, getBody)
	if getData["id"] != uploadID {
		t.Fatalf("expected upload id %s, got %#v", uploadID, getData["id"])
	}

	confirmBody := postJSON(t, router, "/api/uploads/"+uploadID+"/confirm", `{
		"customer_name": "王女士",
		"owner_id": "usr_001"
	}`, http.StatusOK)
	confirmData := responseData(t, confirmBody)
	if confirmData["status"] != "confirmed" {
		t.Fatalf("expected confirmed, got %#v", confirmData["status"])
	}
	if confirmData["followup_task_id"] == "" {
		t.Fatalf("expected followup task id, got %#v", confirmData["followup_task_id"])
	}
}

func TestUploadConversationRejectsEmptyContent(t *testing.T) {
	router := NewRouter(config.Config{Addr: ":0", Env: "test"})
	body := postJSON(t, router, "/api/uploads/conversations", `{
		"source_type": "pasted_text",
		"content": "",
		"owner_id": "usr_001"
	}`, http.StatusBadRequest)

	if body.Error == nil || body.Error.Code != "EMPTY_CONTENT" {
		t.Fatalf("expected EMPTY_CONTENT, got %#v", body.Error)
	}
}

func TestFollowupTaskActions(t *testing.T) {
	router := NewRouter(config.Config{Addr: ":0", Env: "test"})

	copyBody := postJSON(t, router, "/api/followup-tasks/task_001/copy", `{
		"copied_script": "您好，刚才您提到比较关注价格...",
		"client_copied_at": "2026-05-28T10:05:00Z"
	}`, http.StatusOK)
	copyData := responseData(t, copyBody)
	if copyData["status"] != "copied" {
		t.Fatalf("expected copied, got %#v", copyData["status"])
	}

	repeatBody := postJSON(t, router, "/api/followup-tasks/task_001/skip", `{
		"reason": "重复处理"
	}`, http.StatusConflict)
	if repeatBody.Error == nil || repeatBody.Error.Code != "TASK_ALREADY_FINALIZED" {
		t.Fatalf("expected TASK_ALREADY_FINALIZED, got %#v", repeatBody.Error)
	}

	skipBody := postJSON(t, router, "/api/followup-tasks/task_002/skip", `{
		"reason": "客户刚刚回复"
	}`, http.StatusOK)
	skipData := responseData(t, skipBody)
	if skipData["status"] != "skipped" {
		t.Fatalf("expected skipped, got %#v", skipData["status"])
	}

	markWrongBody := postJSON(t, router, "/api/followup-tasks/task_003/mark-wrong", `{
		"reason": "客户不是价格异议",
		"wrong_fields": ["customer_stage"]
	}`, http.StatusOK)
	markWrongData := responseData(t, markWrongBody)
	if markWrongData["status"] != "marked_wrong" {
		t.Fatalf("expected marked_wrong, got %#v", markWrongData["status"])
	}
}

func TestRegenerateTaskKeepsTaskPending(t *testing.T) {
	router := NewRouter(config.Config{Addr: ":0", Env: "test"})

	body := postJSON(t, router, "/api/followup-tasks/task_001/regenerate", `{
		"instruction": "语气更自然一点"
	}`, http.StatusOK)
	data := responseData(t, body)
	if data["agent_run_id"] == "" {
		t.Fatalf("expected agent run id, got %#v", data["agent_run_id"])
	}
	recommendation, ok := data["recommendation"].(map[string]any)
	if !ok {
		t.Fatalf("expected recommendation object, got %#v", data["recommendation"])
	}
	if recommendation["script"] == "" {
		t.Fatalf("expected regenerated script, got %#v", recommendation["script"])
	}
}

func getJSON(t *testing.T, path string) httpx.Response {
	t.Helper()

	router := NewRouter(config.Config{Addr: ":0", Env: "test"})
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body httpx.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error != nil {
		t.Fatalf("expected nil error, got %#v", body.Error)
	}
	return body
}

func postJSON(t *testing.T, router http.Handler, path string, payload string, expectedStatus int) httpx.Response {
	t.Helper()
	return requestJSON(t, router, http.MethodPost, path, payload, expectedStatus)
}

func requestJSON(t *testing.T, router http.Handler, method string, path string, payload string, expectedStatus int) httpx.Response {
	t.Helper()

	var body *bytes.Reader
	if payload == "" {
		body = bytes.NewReader(nil)
	} else {
		body = bytes.NewReader([]byte(payload))
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d with body %s", expectedStatus, rec.Code, rec.Body.String())
	}

	var response httpx.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response
}

func responseData(t *testing.T, body httpx.Response) map[string]any {
	t.Helper()

	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", body.Data)
	}
	return data
}
