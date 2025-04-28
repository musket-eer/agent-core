// Here we implement each endpoint to its Deepseek API configuration
package configs

// DeepseekConfig holds configuration for a Deepseek endpoint.
type DeepseekConfig struct {
	ModelName    string
	Temperature  float64
	MaxTokens    int
	OtherParams  map[string]interface{}
}

// DeepseekEndpointConfigs maps endpoint names to their DeepseekConfig.
var DeepseekEndpointConfigs = map[string]DeepseekConfig{
	"/generate-lesson": {
		ModelName:   "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
	},
	"/generate-course": {
		ModelName:   "deepseek-chat",
		Temperature: 0.6,
		MaxTokens:   2048,
	},
	"/generate-notes": {
		ModelName:   "deepseek-chat",
		Temperature: 0.5,
		MaxTokens:   512,
	},
	"/generate-exams": {
		ModelName:   "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   1024,
	},
	"/generate-resources": {
		ModelName:   "deepseek-chat",
		Temperature: 0.4,
		MaxTokens:   512,
	},
	"/chat": {
		ModelName:   "deepseek-chat",
		Temperature: 0.9,
		MaxTokens:   200,
	},
}

// GetDeepseekConfig returns the DeepseekConfig for a given endpoint.
func GetDeepseekConfig(endpoint string) (DeepseekConfig, bool) {
	cfg, ok := DeepseekEndpointConfigs[endpoint]
	return cfg, ok
}
