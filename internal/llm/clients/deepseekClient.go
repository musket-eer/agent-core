package clients

import (
	"agent-core/internal/llm/configs"
	"context"
	"fmt"
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

	if !ok {
		return "", fmt.Errorf("no Deepseek config for endpoint: %s", endpoint)
	}
	model, _ := params["model"].(string)
	if model == "" {
		model = config.ModelId
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

	if overrideTemp, ok := params["temperature"].(float64); ok {
		req.Temperature = float32(overrideTemp)
	}
	if overrideMaxTokens, ok := params["max_tokens"].(int); ok && overrideMaxTokens > 0 {
		req.MaxTokens = overrideMaxTokens
	}
	if seed, ok := params["seed"].(int); ok {
		req.Seed = &seed
	}
	if responseFormat, ok := params["response_format"].(map[string]interface{}); ok {
		if formatType, ok := responseFormat["type"].(string); ok && formatType == "json_object" {
			req.ResponseFormat = &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			}
		}
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
