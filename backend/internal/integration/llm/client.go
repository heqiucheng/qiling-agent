package llm

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GenerateRequest struct {
	Model          string         `json:"model"`
	PromptVersion  string         `json:"prompt_version"`
	ResponseSchema string         `json:"response_schema"`
	Messages       []Message      `json:"messages"`
	Metadata       map[string]any `json:"metadata"`
}

type GenerateResponse struct {
	Model        string `json:"model"`
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
}

type Client interface {
	Generate(ctx context.Context, request GenerateRequest) (GenerateResponse, error)
}

type MockClient struct {
	Response GenerateResponse
	Err      error
}

func (c MockClient) Generate(ctx context.Context, request GenerateRequest) (GenerateResponse, error) {
	if c.Err != nil {
		return GenerateResponse{}, c.Err
	}
	if c.Response.Content != "" {
		return c.Response, nil
	}
	return GenerateResponse{
		Model:        request.Model,
		Content:      "{}",
		FinishReason: "stop",
	}, nil
}
