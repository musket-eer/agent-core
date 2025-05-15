package llm

import (
	"fmt"
	"agent-core/internal/llm/configs"
)	

// LLMClient defines the interface all clients must implement
type LLMClient interface {
	Call(endpoint string, params map[string]interface{}) (string, error)
}

// global client registry
var llmClients = make(map[string]LLMClient)

// RegisterLLMClient stores a client with a given name
func RegisterLLMClient(name string, client LLMClient) {
	llmClients[name] = client
}

// GetLLMClient retrieves a client by name
func GetLLMClient(name string) (LLMClient, error) {
	client, ok := llmClients[name]
	if !ok {
		return nil, fmt.Errorf("no LLM client registered with name: %s", name)
	}
	return client, nil
}

// Optional: List all registered clients
func ListLLMClients() []string {
	keys := make([]string, 0, len(llmClients))
	for k := range llmClients {
		keys = append(keys, k)
	}
	return keys
}
// Call routes to the appropriate client based on the endpoint
func Call(endpoint string, params map[string]interface{}) (string, error) {
	config, ok := configs.GetDeepseekConfig(endpoint)
	if !ok {
		return "", fmt.Errorf("no config found for endpoint: %s", endpoint)
	}

	clientName := config.ClientName
	if clientName == "" {
		clientName = "general" // fallback default
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		return "", fmt.Errorf("failed to route LLM call: %w", err)
	}

	return client.Call(endpoint, params)
}
