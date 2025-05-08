package utils

import (
	"fmt"
	"strings"
)

func ConvertToBedrockPayload(
	modelId string,
	canonical map[string]interface{},
) (map[string]interface{}, error) {
	input, ok := canonical["input"].(map[string]string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid input.prompt")
	}

	prompt, ok := input["prompt"]
	if !ok || prompt == "" {
		return nil, fmt.Errorf("prompt not found or empty in input")
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
		claudePrompt := fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", prompt)
		return map[string]interface{}{
			"prompt":                claudePrompt,
			"temperature":           temperature,
			"max_tokens_to_sample": maxTokens,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported model for Bedrock: %s", modelId)
	}
}
