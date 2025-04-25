package llm

import (
	"context"
	"fmt"
	"agent-core/internal/llm/configs"
	openai "github.com/sashabaranov/go-openai"
)

// DeepseekClient wraps the OpenAI client for Deepseek API compatibility
// Implements LLMClient interface
// Assumes Deepseek API is OpenAI-compatible

type DeepseekClient struct {
	Client *openai.Client
}

// NewDeepseekClient constructs a DeepseekClient using OpenAI-compatible API
func NewDeepseekClient(apiKey, baseURL string) *DeepseekClient {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL // e.g., "https://api.deepseek.com/v1"
	client := openai.NewClientWithConfig(cfg)
	return &DeepseekClient{Client: client}
}

// Call implements the LLMClient interface for chat completions
func (c *DeepseekClient) Call(endpoint string, params map[string]interface{}) (string, error) {
	config, ok := configs.GetDeepseekConfig(endpoint)
	fmt.Println("Deepseek config:", config)
	if !ok {
		return "", fmt.Errorf("no Deepseek config for endpoint: %s", endpoint)
	}
	model, _ := params["modelName"].(string)
	if model == "" {
		model = config.ModelName
	}
	inputMap, _ := params["input"].(map[string]string)
	prompt := inputMap["prompt"]
	if prompt == "" {
		return "", fmt.Errorf("no prompt provided in params")
	}
	// Build chat completion request
	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: float32(config.Temperature),
		MaxTokens:   config.MaxTokens,
	}
	resp, err := c.Client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from Deepseek")
	}
	return resp.Choices[0].Message.Content, nil
}
