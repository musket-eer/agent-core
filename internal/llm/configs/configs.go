// Here we implement the each endpoint to its Bedrock API configuration

package configs

// BedrockConfig holds configuration for a Bedrock endpoint.
type BedrockConfig struct {
	ModelName    string
	Temperature  float64
	MaxTokens    int
	OtherParams  map[string]interface{}
}

// EndpointConfigs maps endpoint names to their BedrockConfig.
var EndpointConfigs = map[string]BedrockConfig{
	"/generate-lesson": {
		ModelName:   "bedrock-lesson-model",
		Temperature: 0.7,
		MaxTokens:   1024,
	},
	"/generate-course": {
		ModelName:   "bedrock-course-model",
		Temperature: 0.6,
		MaxTokens:   2048,
	},
	"/generate-notes": {
		ModelName:   "bedrock-notes-model",
		Temperature: 0.5,
		MaxTokens:   512,
	},
	"/generate-exams": {
		ModelName:   "bedrock-exams-model",
		Temperature: 0.8,
		MaxTokens:   1024,
	},
	"/generate-resources": {
		ModelName:   "bedrock-resources-model",
		Temperature: 0.4,
		MaxTokens:   512,
	},
	"/chat": {
		ModelName:   "bedrock-chat-model",
		Temperature: 0.9,
		MaxTokens:   2048,
	},
}

// GetConfig returns the BedrockConfig for a given endpoint.
func GetConfig(endpoint string) (BedrockConfig, bool) {
	cfg, ok := EndpointConfigs[endpoint]
	return cfg, ok
}