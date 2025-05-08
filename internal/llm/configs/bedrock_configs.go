package configs

type BedRockConfig struct {
	ModelId            string
	InferenceProfileArn string
	Temperature        float64
	MaxTokensToSample  int
	OtherParams        map[string]interface{}
}

// BedrockEndpointConfigs maps endpoint names to Bedrock configuration.
var BedrockEndpointConfigs = map[string]BedRockConfig{
	"/generate-lesson": {
		ModelId:             "anthropic.claude-v2",
		InferenceProfileArn: "",
		Temperature:         0.6,
		MaxTokensToSample:   1024,
	},
	"/generate-course": {
		ModelId:             "anthropic.claude-v2",
		InferenceProfileArn: "",
		Temperature:         0.5,
		MaxTokensToSample:   4096,
	},
	"/generate-notes": {
		ModelId:             "anthropic.claude-v2",
		InferenceProfileArn: "",
		Temperature:         0.4,
		MaxTokensToSample:   1024,
	},
	"/generate-exams": {
		ModelId:             "anthropic.claude-v2",
		InferenceProfileArn: "",
		Temperature:         0.7,
		MaxTokensToSample:   1024,
	},
	"/generate-resources": {
		ModelId:             "anthropic.claude-v2",
		InferenceProfileArn: "",
		Temperature:         0.5,
		MaxTokensToSample:   1024,
	},
	"/chat": {
		ModelId:             "anthropic.claude-v2",
		InferenceProfileArn: "",
		Temperature:         0.8,
		MaxTokensToSample:   300,
	},
}

// GetBedrockConfig returns the BedrockConfig for a given endpoint.
func GetBedrockConfig(endpoint string) (BedRockConfig, bool) {
	cfg, ok := BedrockEndpointConfigs[endpoint]
	return cfg, ok
}