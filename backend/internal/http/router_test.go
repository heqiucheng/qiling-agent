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
	body := requestJSONWithHeaders(t, NewRouter(config.Config{Addr: ":0", Env: "test"}), http.MethodGet, "/api/customers?intent=high&page=1&page_size=1", "", http.StatusOK, map[string]string{
		"X-Qiling-User-ID": "mgr_001",
		"X-Qiling-Role":    "manager",
	})
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

func TestSalesRoleOnlySeesOwnCustomers(t *testing.T) {
	body := requestJSONWithHeaders(t, NewRouter(config.Config{Addr: ":0", Env: "test"}), http.MethodGet, "/api/customers", "", http.StatusOK, map[string]string{
		"X-Qiling-User-ID": "usr_001",
		"X-Qiling-Role":    "sales",
	})
	data := responseData(t, body)

	if data["total"] != float64(2) {
		t.Fatalf("expected sales user to see two owned customers, got %#v", data["total"])
	}
}

func TestManagerRoleSeesAllCustomers(t *testing.T) {
	body := requestJSONWithHeaders(t, NewRouter(config.Config{Addr: ":0", Env: "test"}), http.MethodGet, "/api/customers", "", http.StatusOK, map[string]string{
		"X-Qiling-User-ID": "mgr_001",
		"X-Qiling-Role":    "manager",
	})
	data := responseData(t, body)

	if data["total"] != float64(3) {
		t.Fatalf("expected manager to see three customers, got %#v", data["total"])
	}
}

func TestSalesRoleCannotSeeOtherOwnersCustomerDetail(t *testing.T) {
	body := requestJSONWithHeaders(t, NewRouter(config.Config{Addr: ":0", Env: "test"}), http.MethodGet, "/api/customers/cus_003", "", http.StatusForbidden, map[string]string{
		"X-Qiling-User-ID": "usr_001",
		"X-Qiling-Role":    "sales",
	})
	if body.Error == nil || body.Error.Code != "FORBIDDEN" {
		t.Fatalf("expected FORBIDDEN, got %#v", body.Error)
	}
}

func TestCustomerDetailEndpoint(t *testing.T) {
	body := getJSON(t, "/api/customers/cus_001")
	data := responseData(t, body)

	if _, ok := data["customer"].(map[string]any); !ok {
		t.Fatalf("expected customer object, got %#v", data["customer"])
	}
	if _, ok := data["latest_recommendation"].(map[string]any); !ok {
		t.Fatalf("expected latest recommendation object, got %#v", data["latest_recommendation"])
	}
	evidence, ok := data["profile_evidence"].([]any)
	if !ok || len(evidence) == 0 {
		t.Fatalf("expected profile evidence, got %#v", data["profile_evidence"])
	}
}

func TestCustomerConversationsEndpoint(t *testing.T) {
	body := getJSON(t, "/api/customers/cus_001/conversations?page=1&page_size=10")
	data := responseData(t, body)

	if data["total"] != float64(2) {
		t.Fatalf("expected two messages, got %#v", data["total"])
	}
	items, ok := data["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected messages, got %#v", data["items"])
	}
}

func TestCustomerShortTermMemoryEndpoint(t *testing.T) {
	body := requestJSONWithHeaders(t, NewRouter(config.Config{Addr: ":0", Env: "test"}), http.MethodGet, "/api/customers/cus_001/short-term-memory", "", http.StatusOK, map[string]string{
		"X-Qiling-User-ID": "usr_001",
		"X-Qiling-Role":    "sales",
	})
	data := responseData(t, body)

	if _, ok := data["customer"].(map[string]any); !ok {
		t.Fatalf("expected customer object, got %#v", data["customer"])
	}
	if data["prompt_context"] == "" {
		t.Fatalf("expected prompt context, got %#v", data["prompt_context"])
	}
	highlights, ok := data["conversation_highlights"].([]any)
	if !ok || len(highlights) == 0 {
		t.Fatalf("expected conversation highlights, got %#v", data["conversation_highlights"])
	}
	tasks, ok := data["recent_tasks"].([]any)
	if !ok || len(tasks) == 0 {
		t.Fatalf("expected recent tasks, got %#v", data["recent_tasks"])
	}
}

func TestSalesRoleCannotSeeOtherOwnersShortTermMemory(t *testing.T) {
	body := requestJSONWithHeaders(t, NewRouter(config.Config{Addr: ":0", Env: "test"}), http.MethodGet, "/api/customers/cus_003/short-term-memory", "", http.StatusForbidden, map[string]string{
		"X-Qiling-User-ID": "usr_001",
		"X-Qiling-Role":    "sales",
	})
	if body.Error == nil || body.Error.Code != "FORBIDDEN" {
		t.Fatalf("expected FORBIDDEN, got %#v", body.Error)
	}
}

func TestFollowupTasksEndpointFiltersByStatus(t *testing.T) {
	body := requestJSONWithHeaders(t, NewRouter(config.Config{Addr: ":0", Env: "test"}), http.MethodGet, "/api/followup-tasks?status=pending", "", http.StatusOK, map[string]string{
		"X-Qiling-User-ID": "mgr_001",
		"X-Qiling-Role":    "manager",
	})
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

func TestReviewSummaryEndpoint(t *testing.T) {
	body := getJSON(t, "/api/review-reports/summary")
	data := responseData(t, body)

	metrics, ok := data["metrics"].([]any)
	if !ok || len(metrics) == 0 {
		t.Fatalf("expected metrics, got %#v", data["metrics"])
	}
	insights, ok := data["insights"].([]any)
	if !ok || len(insights) == 0 {
		t.Fatalf("expected insights, got %#v", data["insights"])
	}
	if data["sample_warning"] == nil {
		t.Fatal("expected sample warning for small mock dataset")
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

func TestAuditEventsEndpointRecordsWriteActions(t *testing.T) {
	router := NewRouter(config.Config{Addr: ":0", Env: "test"})

	requestJSONWithHeaders(t, router, http.MethodPost, "/api/followup-tasks/task_001/copy", `{
		"copied_script": "script",
		"client_copied_at": "2026-05-28T10:05:00Z"
	}`, http.StatusOK, map[string]string{
		"X-Qiling-User-ID": "usr_001",
		"X-Qiling-Role":    "sales",
	})

	body := requestJSONWithHeaders(t, router, http.MethodGet, "/api/audit-events?action=followup_task.copied", "", http.StatusOK, map[string]string{
		"X-Qiling-User-ID": "usr_001",
		"X-Qiling-Role":    "sales",
	})
	data := responseData(t, body)
	if data["total"] != float64(1) {
		t.Fatalf("expected one audit event, got %#v", data["total"])
	}
	items, ok := data["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected one audit item, got %#v", data["items"])
	}
	event, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected audit event object, got %#v", items[0])
	}
	if event["entity_id"] != "task_001" {
		t.Fatalf("expected task_001 entity, got %#v", event["entity_id"])
	}
	if _, ok := event["metadata"].(map[string]any); !ok {
		t.Fatalf("expected metadata object, got %#v", event["metadata"])
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

func TestAgentRunEndpointReturnsPromptAndValidationTrace(t *testing.T) {
	router := NewRouter(config.Config{Addr: ":0", Env: "test"})

	uploadBody := postJSON(t, router, "/api/uploads/conversations", `{
		"source_type": "pasted_text",
		"content": "王女士 09:20 价格和效果需要再看看",
		"owner_id": "usr_001"
	}`, http.StatusOK)
	uploadID := responseData(t, uploadBody)["upload_id"].(string)

	confirmBody := postJSON(t, router, "/api/uploads/"+uploadID+"/confirm", `{
		"customer_name": "王女士",
		"owner_id": "usr_001"
	}`, http.StatusOK)
	agentRunID := responseData(t, confirmBody)["agent_run_id"].(string)

	runBody := requestJSON(t, router, http.MethodGet, "/api/agent-runs/"+agentRunID, "", http.StatusOK)
	run := responseData(t, runBody)
	if run["model"] != "mock-local-v1" {
		t.Fatalf("expected mock model, got %#v", run["model"])
	}
	if run["prompt_version"] != "followup_v1" {
		t.Fatalf("expected followup prompt, got %#v", run["prompt_version"])
	}
	if _, ok := run["output"].(map[string]any); !ok {
		t.Fatalf("expected structured output, got %#v", run["output"])
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
	return requestJSONWithHeaders(t, router, method, path, payload, expectedStatus, nil)
}

func requestJSONWithHeaders(t *testing.T, router http.Handler, method string, path string, payload string, expectedStatus int, headers map[string]string) httpx.Response {
	t.Helper()

	var body *bytes.Reader
	if payload == "" {
		body = bytes.NewReader(nil)
	} else {
		body = bytes.NewReader([]byte(payload))
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
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
