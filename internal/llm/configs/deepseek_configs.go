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
	// Orchestrator Agent endpoint
	"/ai/orchestrator": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   2048,
		ClientName:  "general",
	},
	// Project Generation endpoint (stateless, aligned with frontend)
	"/generate-project": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   4096,
		ClientName:  "general",
	},
	// Generate Tasks endpoint (for milestone task generation)
	"/generate-tasks": {
		ModelId:     "deepseek-chat",
		Temperature: 0.6,
		MaxTokens:   1536,
		ClientName:  "general",
	},
	// Course Creation Chat endpoint (dialogue-based course creation)
	"/course-creation/chat": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   2048,
		ClientName:  "general",
	},
	"/ai/projects/milestone-to-tasks": {
		ModelId:     "deepseek-chat",
		Temperature: 0.6,
		MaxTokens:   2048,
		ClientName:  "general",
	},
	// Dashboard Agent endpoints
	"/dashboard/insights": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/dashboard/recommendations": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/dashboard/analyze-progress": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/dashboard/chat": {
		ModelId:     "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   512,
		ClientName:  "chat",
	},
	// Course Browse Agent endpoints
	"/courses/intelligent-search": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/courses/explain-relevance": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   512,
		ClientName:  "general",
	},
	"/courses/compare": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/courses/suggest-path": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/courses/chat": {
		ModelId:     "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   512,
		ClientName:  "chat",
	},
	// Reflection Agent endpoints
	"/reflection/generate-prompts": {
		ModelId:     "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   512,
		ClientName:  "general",
	},
	"/reflection/analyze": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/reflection/learning-insights": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/reflection/period-summary": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/reflection/chat": {
		ModelId:     "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   512,
		ClientName:  "chat",
	},
	"/reflection/suggest-exercises": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/reflection/generate-task-reflection": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   2048,
		ClientName:  "general",
	},
	// Tasks Agent endpoints
	"/tasks/generate-suggestions": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/tasks/prioritize": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/tasks/break-down": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/tasks/productivity-insights": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/tasks/generate-schedule": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/tasks/chat": {
		ModelId:     "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   512,
		ClientName:  "chat",
	},
	"/tasks/motivation": {
		ModelId:     "deepseek-chat",
		Temperature: 0.9,
		MaxTokens:   512,
		ClientName:  "chat",
	},
	// Librarian Agent endpoints
	"/chat-video": {
		ModelId:     "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   512,
		ClientName:  "chat",
	},
	"/recommend-videos": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/summarize-video": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/extract-concepts": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/generate-quiz": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/find-topic": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   512,
		ClientName:  "general",
	},
	// Project Assistant Agent endpoints
	"/projects/chat": {
		ModelId:     "deepseek-chat",
		Temperature: 0.8,
		MaxTokens:   512,
		ClientName:  "chat",
	},
	"/projects/suggestions": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/projects/help": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/projects/break-down": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	"/projects/resources": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   1024,
		ClientName:  "general",
	},
	// Project AI Agent endpoints (stateless)
	"/projects/create": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   3072,
		ClientName:  "general",
	},
	"/projects/assist": {
		ModelId:     "deepseek-chat",
		Temperature: 0.7,
		MaxTokens:   2048,
		ClientName:  "general",
	},
	"/projects/translate-milestones": {
		ModelId:     "deepseek-chat",
		Temperature: 0.6,
		MaxTokens:   2048,
		ClientName:  "general",
	},
}

// GetDeepseekConfig returns the DeepseekConfig for a given endpoint.
func GetDeepseekConfig(endpoint string) (DeepseekConfig, bool) {
	cfg, ok := DeepseekEndpointConfigs[endpoint]
	return cfg, ok
}
