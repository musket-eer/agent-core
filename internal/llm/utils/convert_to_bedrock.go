package utils

import (
	"errors"
	"strings"
)

func ConvertToBedrockPayload(
	modelId string,
	canonical map[string]interface{},
) (map[string]interface{}, error) {
	input, ok := canonical["input"].(map[string]string)
	if !ok {
		return nil, errors.New("missing or invalid input.prompt")
	}

	prompt, ok := input["prompt"]
	if !ok || prompt == "" {
		return nil, errors.New("prompt not found or empty in input")
	}

	// Default fallbacks in case type assertion fails
	temperature := 0.7
	if temp, ok := canonical["temperature"].(float64); ok {
		temperature = temp
	}

	maxTokens := 1024
	if mt, ok := canonical["max_tokens"].(int); ok {
		maxTokens = mt
	}

	// Handle Claude prompt format
	switch {
	case strings.HasPrefix(modelId, "anthropic.claude"):
		return map[string]interface{}{
			"anthropic_version": "bedrock-2023-05-31",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": prompt,
				},
			},
			"temperature": temperature,
			"max_tokens":  maxTokens,
		}, nil

	default:
		return nil, errors.New("unsupported model for Bedrock: " + modelId)
	}
}
