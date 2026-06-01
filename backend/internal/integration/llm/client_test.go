package llm

import (
	"context"
	"errors"
	"testing"
)

func TestMockClientGenerateUsesRequestModel(t *testing.T) {
	response, err := MockClient{}.Generate(context.Background(), GenerateRequest{
		Model:         "mock-model",
		PromptVersion: "followup_v1",
		Messages:      []Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if response.Model != "mock-model" {
		t.Fatalf("expected request model, got %s", response.Model)
	}
	if response.Content != "{}" {
		t.Fatalf("expected empty JSON object, got %s", response.Content)
	}
}

func TestMockClientCanReturnError(t *testing.T) {
	_, err := MockClient{Err: errors.New("timeout")}.Generate(context.Background(), GenerateRequest{
		Model: "mock-model",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
