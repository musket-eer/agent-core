package clients

import (
	"agent-core/internal/llm/configs"
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"reflect"
	"strings"
)

// DeepseekClient wraps the OpenAI client for Deepseek API compatibility
// Implements LLMClient interface
// Assumes Deepseek API is OpenAI-compatible

type DeepseekClient struct {
	Client  *openai.Client
	BaseURL string
}

// NewDeepseekClient constructs a DeepseekClient using OpenAI-compatible API
func NewDeepseekClient(apiKey, baseURL string) *DeepseekClient {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = normalizeOpenAICompatibleBaseURL(baseURL)
	client := openai.NewClientWithConfig(cfg)
	return &DeepseekClient{Client: client, BaseURL: cfg.BaseURL}
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
	messages, err := buildChatMessages(params)
	if err != nil {
		return "", err
	}
	// Build chat completion request
	req := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: float32(config.Temperature),
		MaxTokens:   config.MaxTokens,
	}

	if overrideTemp, ok := params["temperature"].(float64); ok {
		req.Temperature = float32(overrideTemp)
	}
	if overrideMaxTokens, ok := params["max_tokens"].(int); ok && overrideMaxTokens > 0 {
		req.MaxTokens = overrideMaxTokens
	}
	restrictedModel := shouldUseMaxCompletionTokens(model, c.BaseURL)
	if restrictedModel {
		req.MaxCompletionTokens = req.MaxTokens
		req.MaxTokens = 0
		if req.Temperature != 0 && req.Temperature != 1 {
			req.Temperature = 0
		}
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

func buildChatMessages(params map[string]interface{}) ([]openai.ChatCompletionMessage, error) {
	if messages, ok := buildTypedMessages(params["messages"]); ok {
		return messages, nil
	}

	if rawMessages, ok := params["messages"].([]interface{}); ok && len(rawMessages) > 0 {
		messages := make([]openai.ChatCompletionMessage, 0, len(rawMessages))
		for _, raw := range rawMessages {
			msg, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			role, _ := msg["role"].(string)
			content, _ := msg["content"].(string)
			if role == "" || content == "" {
				continue
			}
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    role,
				Content: content,
			})
		}
		if len(messages) > 0 {
			return messages, nil
		}
	}

	inputMap, _ := params["input"].(map[string]string)
	prompt := inputMap["prompt"]
	if prompt == "" {
		return nil, fmt.Errorf("no prompt or messages provided in params")
	}

	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}, nil
}

func buildTypedMessages(raw any) ([]openai.ChatCompletionMessage, bool) {
	v := reflect.ValueOf(raw)
	if !v.IsValid() || v.Kind() != reflect.Slice || v.Len() == 0 {
		return nil, false
	}

	messages := make([]openai.ChatCompletionMessage, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		skip := false
		for item.Kind() == reflect.Pointer {
			if item.IsNil() {
				skip = true
				break
			}
			item = item.Elem()
		}
		if skip {
			continue
		}

		var role, content string
		switch item.Kind() {
		case reflect.Map:
			role = stringValue(item.MapIndex(reflect.ValueOf("role")))
			content = stringValue(item.MapIndex(reflect.ValueOf("content")))
		case reflect.Struct:
			role = stringValue(item.FieldByName("Role"))
			content = stringValue(item.FieldByName("Content"))
		}

		if role == "" || content == "" {
			continue
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: content,
		})
	}

	if len(messages) == 0 {
		return nil, false
	}
	return messages, true
}

func stringValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	return ""
}

func normalizeOpenAICompatibleBaseURL(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	switch trimmed {
	case "https://api.openai.com", "https://api.deepseek.com":
		return trimmed + "/v1"
	default:
		return trimmed
	}
}

func shouldUseMaxCompletionTokens(model string, baseURL string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	baseURL = strings.ToLower(strings.TrimSpace(baseURL))
	if strings.Contains(baseURL, "api.openai.com") {
		return strings.HasPrefix(model, "gpt-5") || strings.HasPrefix(model, "o1") || strings.HasPrefix(model, "o3")
	}
	return strings.HasPrefix(model, "o1") || strings.HasPrefix(model, "o3")
}
