package clients

import (
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"agent-core/internal/llm/configs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type BedrockClient struct {
	AwsClient *bedrockruntime.Client
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
		log.Printf("[BedrockClient.Call] No Bedrock config found for endpoint: %s", endpoint)
		return "", fmt.Errorf("no Bedrock config found for endpoint: %s", endpoint)
	}
	log.Printf("[BedrockClient.Call] Loaded config for endpoint: %s", endpoint)

	// Extract prompt from params["input"]["prompt"]
	input, ok := params["prompt"].(string)
	if !ok {
		log.Printf("[BedrockClient.Call] Missing 'prompt' field in params")
		return "", fmt.Errorf("missing 'prompt' field")
	}
	promptStr := input
	log.Printf("[BedrockClient.Call] Extracted prompt from params")

	// Format Claude-style prompt
	fullPrompt := fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", promptStr)
	log.Printf("[BedrockClient.Call] Formatted Claude-style prompt")

	// Build payload
	payload := map[string]interface{}{
		"prompt":                fullPrompt,
		"temperature":           cfg.Temperature,
		"max_tokens_to_sample":  cfg.MaxTokensToSample,
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
			ModelId: aws.String(cfg.ModelId),
			Body:    body,
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
		if msg, ok := decoded["completion"].(string); ok {
			log.Printf("[BedrockClient.Call] Found 'completion' in response, returning it")
			return msg, nil
		}
	} else {
		log.Printf("[BedrockClient.Call] Failed to decode JSON from response: %v", err)
	}

	return string(result), nil
}