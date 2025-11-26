package llm

import (
	"fmt"
	"log"
)

// Persona represents the agent's personality and behavior settings
type Persona struct {
	AgentID       string                 `json:"agentId"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Icon          string                 `json:"icon"`
	DefaultTraits map[string]interface{} `json:"defaultTraits"`
	UserTraits    map[string]interface{} `json:"userTraits"`
	MergedTraits  map[string]interface{} `json:"mergedTraits"`
	Instructions  string                 `json:"instructions"`
}

// PersonaRequest is embedded in all agent requests
type PersonaRequest struct {
	AgentID string   `json:"agentId,omitempty"`
	Persona *Persona `json:"persona,omitempty"`
}

// ExtractPersona extracts agentId and persona from a request
func ExtractPersona(req interface{}) (string, *Persona) {
	// Try to extract persona using type assertion
	if pr, ok := req.(interface{ GetPersona() (string, *Persona) }); ok {
		return pr.GetPersona()
	}

	// Default fallback
	return "unknown", nil
}

// BuildPersonaSystemPrompt creates a persona-aware system prompt
func BuildPersonaSystemPrompt(agentID string, persona *Persona, basePrompt string) string {
	name := agentID
	instructions := ""

	if persona != nil {
		if persona.Name != "" {
			name = persona.Name
		}
		if persona.Instructions != "" {
			instructions = persona.Instructions
		}
	}

	// Build the persona-aware system prompt
	personaPrefix := fmt.Sprintf(`You are the %s agent in the Most Learned platform.

Follow these persona instructions exactly when you write responses:

%s

If traits indicate specific tone, verbosity, strictness, or style,
reflect that consistently in your responses.

When relevant, explain your reasoning clearly but stay aligned
with the persona settings.

---

`, name, instructions)

	log.Printf("[Persona] Using persona: %s with custom instructions: %v", name, instructions != "")

	return personaPrefix + basePrompt
}

// GetPersonaName returns the persona name or falls back to agentId
func GetPersonaName(agentID string, persona *Persona) string {
	if persona != nil && persona.Name != "" {
		return persona.Name
	}
	return agentID
}
