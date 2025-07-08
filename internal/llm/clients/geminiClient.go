package clients

import (
	"context"
	"fmt"
	"agent-core/internal/llm/configs"
	genai "google.golang.org/genai"
)

// GeminiClient implements the LLMClient interface for Google Gemini API using the official SDK
type GeminiClient struct {
	ApiKey string
	Client *genai.Client
}

// NewGeminiClient constructs a GeminiClient using the official SDK
func NewGeminiClient(apiKey, _ string) *GeminiClient {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create Gemini client: %v", err))
	}
	return &GeminiClient{
		ApiKey: apiKey,
		Client: client,
	}
}

// Call implements the LLMClient interface for Gemini using the SDK
func (c *GeminiClient) Call(endpoint string, params map[string]interface{}) (string, error) {
	config, ok := configs.GetGeminiConfig(endpoint)
	if !ok {
		return "", fmt.Errorf("no Gemini config for endpoint: %s", endpoint)
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

	temp := float32(config.Temperature)
	maxOutputTokens := int32(config.MaxTokens)

	genaiConfig := &genai.GenerateContentConfig{
		Temperature:     &temp,
		MaxOutputTokens: maxOutputTokens,
		ResponseMIMEType: "application/json",
	}

	ctx := context.Background()
	result, err := c.Client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(prompt),
		genaiConfig,
	)
	if err != nil {
		return "", err
	}
	return result.Text(), nil
}
