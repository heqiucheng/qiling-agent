package http

import (
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

func responseData(t *testing.T, body httpx.Response) map[string]any {
	t.Helper()

	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object data, got %T", body.Data)
	}
	return data
}
