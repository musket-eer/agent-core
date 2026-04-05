package configs

import (
	"os"
	"strings"
)

const defaultBedrockModelID = "anthropic.claude-3-sonnet-20240229-v1:0"

type BedRockConfig struct {
	ModelId             string
	InferenceProfileArn string
	Temperature         float64
	MaxTokensToSample   int
	OtherParams         map[string]interface{}
}

// BedrockEndpointConfigs maps endpoint names to Bedrock configuration.
var BedrockEndpointConfigs = map[string]BedRockConfig{
	"/generate-lesson": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.6,
		MaxTokensToSample:   1024,
	},
	"/generate-course": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.5,
		MaxTokensToSample:   4096,
	},
	"/generate-notes": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.4,
		MaxTokensToSample:   1024,
	},
	"/generate-exams": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.7,
		MaxTokensToSample:   1024,
	},
	"/generate-resources": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.5,
		MaxTokensToSample:   1024,
	},
	"/lessons/animated-knowledge-graph": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.2,
		MaxTokensToSample:   2200,
	},
	"/lessons/animated-slide-content": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.2,
		MaxTokensToSample:   1400,
	},
	"/lessons/animated-slide-scene-content": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.2,
		MaxTokensToSample:   1400,
	},
	"/animated-video/tree-slide": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.2,
		MaxTokensToSample:   2200,
	},
	"/animated-video/node-content": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.2,
		MaxTokensToSample:   1400,
	},
	"/chat": {
		ModelId:             defaultBedrockModelID,
		InferenceProfileArn: "",
		Temperature:         0.8,
		MaxTokensToSample:   300,
	},
}

func DefaultBedrockConfig() BedRockConfig {
	return BedRockConfig{
		ModelId:             activeBedrockModelID(),
		InferenceProfileArn: "",
		Temperature:         0.4,
		MaxTokensToSample:   2048,
	}
}

// GetBedrockConfig returns the BedrockConfig for a given endpoint.
func GetBedrockConfig(endpoint string) (BedRockConfig, bool) {
	cfg, ok := BedrockEndpointConfigs[endpoint]
	if ok {
		cfg.ModelId = activeBedrockModelID()
	}
	return cfg, ok
}

func activeBedrockModelID() string {
	if v := strings.TrimSpace(os.Getenv("BEDROCK_MODEL_ID")); v != "" {
		return v
	}
	return defaultBedrockModelID
}
