// Here we implement each endpoint to its Gemini API configuration
package configs

type GeminiConfig struct {
	ModelId     string
	Temperature float64
	MaxTokens   int
	ClientName  string // e.g., "chat", "general"
	OtherParams map[string]interface{}
}

var GeminiEndpointConfigs = map[string]GeminiConfig{
	"/generate-lesson": {
		ModelId:     "gemini-pro",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "gemini-general",
	},
	"/generate-course": {
		ModelId:     "gemini-pro",
		Temperature: 0.6,
		MaxTokens:   2048,
		ClientName:  "gemini-general",
	},
	"/generate-notes": {
		ModelId:     "gemini-2.0-flash",
		Temperature: 1.0,
		MaxTokens:   1024,
		ClientName:  "gemini-general",
	},
	"/generate-exams": {
		ModelId:     "gemini-pro",
		Temperature: 0.8,
		MaxTokens:   1024,
		ClientName:  "gemini-general",
	},
	"/generate-resources": {
		ModelId:     "gemini-pro",
		Temperature: 0.4,
		MaxTokens:   1024,
		ClientName:  "gemini-general",
	},
	"/course-creation/generate": {
		ModelId:     "gemini-pro",
		Temperature: 0.5,
		MaxTokens:   3200,
		ClientName:  "gemini-general",
	},
	"/api/prompts/generate-project-prompt": {
		ModelId:     "gemini-pro",
		Temperature: 0.4,
		MaxTokens:   2048,
		ClientName:  "gemini-general",
	},
	"/api/prompts/generate": {
		ModelId:     "gemini-pro",
		Temperature: 0.4,
		MaxTokens:   2048,
		ClientName:  "gemini-general",
	},
	"/chat": {
		ModelId:     "gemini-pro",
		Temperature: 0.9,
		MaxTokens:   100,
		ClientName:  "gemini-chat",
	},
	"/professor-chat": {
		ModelId:     "gemini-pro",
		Temperature: 0.5,
		MaxTokens:   1024,
		ClientName:  "gemini-chat",
	},
	"/librarian-chat": {
		ModelId:     "gemini-pro",
		Temperature: 0.6,
		MaxTokens:   512,
		ClientName:  "gemini-chat",
	},
	"/examiner-chat": {
		ModelId:     "gemini-pro",
		Temperature: 0.6,
		MaxTokens:   512,
		ClientName:  "gemini-chat",
	},
	"/lessons/animated-knowledge-graph": {
		ModelId:     "gemini-2.0-flash",
		Temperature: 0.2,
		MaxTokens:   2200,
		ClientName:  "gemini-general",
	},
	"/lessons/animated-slide-content": {
		ModelId:     "gemini-2.0-flash",
		Temperature: 0.2,
		MaxTokens:   1400,
		ClientName:  "gemini-general",
	},
	"/lessons/animated-slide-scene-content": {
		ModelId:     "gemini-2.0-flash",
		Temperature: 0.2,
		MaxTokens:   1400,
		ClientName:  "gemini-general",
	},
	"/animated-video/tree-slide": {
		ModelId:     "gemini-2.0-flash",
		Temperature: 0.2,
		MaxTokens:   2200,
		ClientName:  "gemini-general",
	},
	"/animated-video/node-content": {
		ModelId:     "gemini-2.0-flash",
		Temperature: 0.2,
		MaxTokens:   1400,
		ClientName:  "gemini-general",
	},
}

// GetGeminiConfig returns the GeminiConfig for a given endpoint.
func GetGeminiConfig(endpoint string) (GeminiConfig, bool) {
	cfg, ok := GeminiEndpointConfigs[endpoint]
	return cfg, ok
}
