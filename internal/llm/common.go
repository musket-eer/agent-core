package llm

// Message represents a single chat message
// Shared across LLM endpoints
//
// Message struct is now defined here for all handlers to import and use.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
