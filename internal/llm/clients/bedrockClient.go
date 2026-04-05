package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"agent-core/internal/llm/configs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type BedrockClient struct {
	AwsClient *bedrockruntime.Client
}

type bedrockMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type extractedBedrockInput struct {
	System   string
	Messages []bedrockMessage
}

// NewBedrockClient initializes a BedrockClient for the specified AWS region
func NewBedrockClient(region string) (*BedrockClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := bedrockruntime.NewFromConfig(cfg)
	return &BedrockClient{AwsClient: client}, nil
}

// Call implements LLMClient for Bedrock models (Claude 3.x etc.)
func (c *BedrockClient) Call(endpoint string, params map[string]interface{}) (string, error) {
	for k := range params {
		log.Println("param key:", k)
	}

	// Get model configuration for this endpoint
	cfg, ok := configs.GetBedrockConfig(endpoint)
	if !ok {
		cfg = configs.DefaultBedrockConfig()
		log.Printf("[BedrockClient.Call] No Bedrock config found for endpoint: %s, using default config", endpoint)
	}
	log.Printf("[BedrockClient.Call] Loaded config for endpoint: %s", endpoint)

	input, err := extractBedrockInput(params)
	if err != nil {
		log.Printf("[BedrockClient.Call] Missing prompt in params: %v", err)
		return "", err
	}
	log.Printf("[BedrockClient.Call] Extracted Bedrock input from params")

	// Build Claude 3 messages payload
	payload := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"temperature":       cfg.Temperature,
		"max_tokens":        cfg.MaxTokensToSample,
		"messages":          input.Messages,
	}
	if strings.TrimSpace(input.System) != "" {
		payload["system"] = input.System
	}
	for k, v := range cfg.OtherParams {
		payload[k] = v
	}
	log.Printf("[BedrockClient.Call] Built payload: %v", payload)

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[BedrockClient.Call] Failed to marshal payload: %v", err)
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	log.Printf("[BedrockClient.Call] Marshaled payload to JSON")

	var respBody io.ReadCloser
	if cfg.InferenceProfileArn != "" {
		log.Printf("[BedrockClient.Call] Using InvokeModelWithConfig for Claude 3.5+")
		resp, err := c.AwsClient.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(cfg.ModelId),
			Body:        body,
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
		})
		if err != nil {
			log.Printf("[BedrockClient.Call] InvokeModelWithConfig error: %v", err)
			return "", fmt.Errorf("invoke model with config error: %w", err)
		}
		log.Printf("[BedrockClient.Call] Model invoked successfully (with config)")
		respBody = io.NopCloser(bytes.NewReader(resp.Body)) // now it's io.ReadCloser

	} else {
		log.Printf("[BedrockClient.Call] Using InvokeModel for Claude v2")
		resp, err := c.AwsClient.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(cfg.ModelId),
			Body:        body,
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
		})
		if err != nil {
			log.Printf("[BedrockClient.Call] InvokeModel error: %v", err)
			return "", fmt.Errorf("invoke model error: %w", err)
		}
		log.Printf("[BedrockClient.Call] Model invoked successfully (Claude v2)")
		respBody = io.NopCloser(bytes.NewReader(resp.Body)) // now it's io.ReadCloser
	}

	defer respBody.Close()
	result, err := io.ReadAll(respBody)
	if err != nil {
		log.Printf("[BedrockClient.Call] Failed to read response: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	log.Printf("[BedrockClient.Call] Successfully read response body")

	// Optionally decode JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(result, &decoded); err == nil {
		log.Printf("[BedrockClient.Call] Decoded JSON from response body")
		if msg, ok := decoded["completion"].(string); ok && strings.TrimSpace(msg) != "" {
			log.Printf("[BedrockClient.Call] Found legacy 'completion' in response, returning it")
			return msg, nil
		}
		if content, ok := decoded["content"].([]interface{}); ok {
			var parts []string
			for _, item := range content {
				part, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				if text, ok := part["text"].(string); ok && strings.TrimSpace(text) != "" {
					parts = append(parts, text)
				}
			}
			if len(parts) > 0 {
				return strings.Join(parts, "\n"), nil
			}
		}
	} else {
		log.Printf("[BedrockClient.Call] Failed to decode JSON from response: %v", err)
	}

	return string(result), nil
}

func extractBedrockInput(params map[string]interface{}) (extractedBedrockInput, error) {
	if params == nil {
		return extractedBedrockInput{}, fmt.Errorf("missing request params")
	}
	if input, ok := params["prompt"].(string); ok && strings.TrimSpace(input) != "" {
		return extractedBedrockInput{
			Messages: []bedrockMessage{{Role: "user", Content: strings.TrimSpace(input)}},
		}, nil
	}
	if inputMap, ok := params["input"].(map[string]string); ok {
		if prompt := strings.TrimSpace(inputMap["prompt"]); prompt != "" {
			return extractedBedrockInput{
				Messages: []bedrockMessage{{Role: "user", Content: prompt}},
			}, nil
		}
	}
	if inputMap, ok := params["input"].(map[string]interface{}); ok {
		if prompt, ok := inputMap["prompt"].(string); ok && strings.TrimSpace(prompt) != "" {
			return extractedBedrockInput{
				Messages: []bedrockMessage{{Role: "user", Content: strings.TrimSpace(prompt)}},
			}, nil
		}
	}
	if messages, ok := params["messages"].([]interface{}); ok {
		input := normalizeMessages(messages)
		if len(input.Messages) > 0 {
			return input, nil
		}
	}
	if messages, ok := params["messages"].([]map[string]interface{}); ok {
		items := make([]interface{}, 0, len(messages))
		for _, message := range messages {
			items = append(items, message)
		}
		input := normalizeMessages(items)
		if len(input.Messages) > 0 {
			return input, nil
		}
	}
	return extractedBedrockInput{}, fmt.Errorf("missing prompt field")
}

func normalizeMessages(messages []interface{}) extractedBedrockInput {
	out := extractedBedrockInput{
		Messages: make([]bedrockMessage, 0, len(messages)),
	}
	systemParts := make([]string, 0, 2)
	for _, raw := range messages {
		msg, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		content, _ := msg["content"].(string)
		role = strings.TrimSpace(role)
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		if role == "" || role == "user" || role == "assistant" {
			normalizedRole := role
			if normalizedRole == "" {
				normalizedRole = "user"
			}
			out.Messages = append(out.Messages, bedrockMessage{
				Role:    normalizedRole,
				Content: content,
			})
			continue
		}
		if role == "system" {
			systemParts = append(systemParts, content)
			continue
		}
		out.Messages = append(out.Messages, bedrockMessage{
			Role:    "user",
			Content: role + ": " + content,
		})
	}
	out.System = strings.Join(systemParts, "\n\n")
	return out
}
