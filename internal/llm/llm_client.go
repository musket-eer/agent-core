package llm

var llmClient LLMClient

// LlmClientSet sets the global LLM client to be used by all handlers
func LlmClientSet(client LLMClient) {
	llmClient = client
}

type LLMClient interface {
	Call(endpoint string, params map[string]interface{}) (string, error)
}
