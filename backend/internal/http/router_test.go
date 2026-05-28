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
