package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAICompatibleClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type OpenAICompatibleConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

func NewOpenAICompatibleClient(cfg OpenAICompatibleConfig) (*OpenAICompatibleClient, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("llm base url is required")
	}
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("llm api key is required")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 45 * time.Second
	}
	return &OpenAICompatibleClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *OpenAICompatibleClient) Generate(ctx context.Context, request GenerateRequest) (GenerateResponse, error) {
	payload := chatCompletionRequest{
		Model:          request.Model,
		Messages:       toChatMessages(request.Messages),
		ResponseFormat: chatResponseFormat{Type: "json_object"},
		Metadata:       request.Metadata,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("marshal llm request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointURL(), bytes.NewReader(raw))
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("build llm request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("call llm provider: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("read llm response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return GenerateResponse{}, fmt.Errorf("llm provider returned %d: %s", response.StatusCode, compactBody(body))
	}

	var decoded chatCompletionResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return GenerateResponse{}, fmt.Errorf("decode llm response from %s: %w: %s", c.endpointURL(), err, compactBody(body))
	}
	if len(decoded.Choices) == 0 {
		return GenerateResponse{}, fmt.Errorf("llm response has no choices")
	}

	return GenerateResponse{
		Model:        firstNonEmpty(decoded.Model, request.Model),
		Content:      decoded.Choices[0].Message.Content,
		FinishReason: decoded.Choices[0].FinishReason,
	}, nil
}

type chatCompletionRequest struct {
	Model          string             `json:"model"`
	Messages       []chatMessage      `json:"messages"`
	ResponseFormat chatResponseFormat `json:"response_format"`
	Metadata       map[string]any     `json:"metadata,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponseFormat struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
}

func toChatMessages(messages []Message) []chatMessage {
	converted := make([]chatMessage, 0, len(messages))
	for _, message := range messages {
		converted = append(converted, chatMessage{Role: message.Role, Content: message.Content})
	}
	return converted
}

func (c *OpenAICompatibleClient) endpointURL() string {
	switch {
	case strings.HasSuffix(c.baseURL, "/chat/completions"):
		return c.baseURL
	case strings.HasSuffix(c.baseURL, "/responses"):
		return c.baseURL
	default:
		return c.baseURL + "/v1/chat/completions"
	}
}

func compactBody(body []byte) string {
	value := strings.Join(strings.Fields(string(body)), " ")
	if len(value) > 500 {
		return value[:500] + "..."
	}
	return value
}

func firstNonEmpty(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
