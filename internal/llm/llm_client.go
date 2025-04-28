package llm

import "agent-core/internal/cache"

var llmClient LLMClient
var supabaseClient = cache.NewSupabasePgvectorClient()

// LlmClientSet sets the global LLM client to be used by all handlers
func LlmClientSet(client LLMClient) {
	llmClient = client
}

type LLMClient interface {
	Call(endpoint string, params map[string]interface{}) (string, error)
}
