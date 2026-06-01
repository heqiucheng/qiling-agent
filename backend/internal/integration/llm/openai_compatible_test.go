package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpenAICompatibleClientSendsChatCompletionRequest(t *testing.T) {
	var captured chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("expected chat completions path, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("expected bearer key, got %s", r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model": "gpt-5.4",
			"choices": [{
				"message": {"role": "assistant", "content": "{\"customer_stage\":\"high_intent\",\"intent_level\":\"high\",\"main_concerns\":[\"budget\"],\"recommended_action\":\"confirm budget fit\",\"script\":\"I can map this to your budget.\",\"reasoning\":\"Customer mentioned budget.\",\"risk_flags\":[]}"},
				"finish_reason": "stop"
			}]
		}`))
	}))
	defer server.Close()

	client, err := NewOpenAICompatibleClient(OpenAICompatibleConfig{BaseURL: server.URL, APIKey: "test-key", Timeout: time.Second})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	response, err := client.Generate(t.Context(), GenerateRequest{
		Model: "gpt-5.4",
		Messages: []Message{
			{Role: "system", Content: "return json"},
			{Role: "user", Content: "hello"},
		},
		Metadata: map[string]any{"task_type": "test"},
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if captured.Model != "gpt-5.4" {
		t.Fatalf("expected model gpt-5.4, got %s", captured.Model)
	}
	if captured.ResponseFormat.Type != "json_object" {
		t.Fatalf("expected json_object response format, got %s", captured.ResponseFormat.Type)
	}
	if len(captured.Messages) != 2 {
		t.Fatalf("expected two messages, got %d", len(captured.Messages))
	}
	if response.Model != "gpt-5.4" {
		t.Fatalf("expected response model, got %s", response.Model)
	}
	if response.FinishReason != "stop" {
		t.Fatalf("expected stop finish reason, got %s", response.FinishReason)
	}
	if response.Content == "" {
		t.Fatal("expected response content")
	}
}

func TestOpenAICompatibleClientReturnsProviderErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	client, err := NewOpenAICompatibleClient(OpenAICompatibleConfig{BaseURL: server.URL, APIKey: "test-key", Timeout: time.Second})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, err = client.Generate(t.Context(), GenerateRequest{Model: "gpt-5.4"})
	if err == nil {
		t.Fatal("expected provider error")
	}
}

func TestOpenAICompatibleClientUsesCompleteChatCompletionsEndpoint(t *testing.T) {
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"role":"assistant","content":"{}"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	client, err := NewOpenAICompatibleClient(OpenAICompatibleConfig{BaseURL: server.URL + "/chat/completions", APIKey: "test-key", Timeout: time.Second})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if _, err := client.Generate(t.Context(), GenerateRequest{Model: "gpt-5.4"}); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if path != "/chat/completions" {
		t.Fatalf("expected complete endpoint to be used as-is, got %s", path)
	}
}

func TestOpenAICompatibleClientRequiresAPIKey(t *testing.T) {
	_, err := NewOpenAICompatibleClient(OpenAICompatibleConfig{BaseURL: "https://example.test", APIKey: ""})
	if err == nil {
		t.Fatal("expected missing api key error")
	}
}
