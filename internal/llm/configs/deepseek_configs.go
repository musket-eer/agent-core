// Here we implement each endpoint to its Deepseek API configuration
package configs

// DeepseekConfig holds configuration for a Deepseek endpoint.
type DeepseekConfig struct {
	ModelId     string
	Temperature float64
	MaxTokens   int
	ClientName  string // NEW: "chat", "general", etc.
	OtherParams map[string]interface{}
}


// DeepseekEndpointConfigs maps endpoint names to their DeepseekConfig.
var DeepseekEndpointConfigs = map[string]DeepseekConfig{
	"/generate-lesson": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/generate-course": {
		ModelId:     "deepseek-chat",
		Temperature: 0.6,
		MaxTokens:   2048,
		ClientName:  "general",
	},
	"/generate-notes": {
		ModelId:     "deepseek-chat",
		Temperature: 1.0,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/generate-exams": {
		ModelId:     "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/generate-resources": {
		ModelId:     "deepseek-chat",
		Temperature: 0.4,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/chat": {
		ModelId:     "deepseek-chat",
		Temperature: 0.9,
		MaxTokens:   100,
		ClientName:  "chat",
	},
	"/professor-chat": {
		ModelId:     "deepseek-chat",
		Temperature: 0.5,
		MaxTokens:   512,
		ClientName:  "chat",
	},
	"/librarian-chat": {
		ModelId:     "deepseek-chat",
		Temperature: 0.6,
		MaxTokens:   512,
		ClientName:  "chat",
	},
	"/examiner-chat": {
		ModelId:     "deepseek-chat",
		Temperature: 0.6,
		MaxTokens:   512,
		ClientName:  "chat",
	},
}

// GetDeepseekConfig returns the DeepseekConfig for a given endpoint.
func GetDeepseekConfig(endpoint string) (DeepseekConfig, bool) {
	cfg, ok := DeepseekEndpointConfigs[endpoint]
	return cfg, ok
}
