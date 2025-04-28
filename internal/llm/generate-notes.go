package llm

import (
	"encoding/json"
	"net/http"
	"strings"

	"agent-core/internal/llm/configs"
	"agent-core/internal/embedding"
	"agent-core/internal/prompt"
	"agent-core/internal/llm/utils"

	"github.com/google/uuid"
)

// FrontendPayload expects model, messages, temperature

// TODO: - We want to maps this one to one with the frontend json
type NotesPayLoad struct {
	Messages []Message `json:"messages"`
}

// Response struct for notes
type NotesResponse struct {
	ID    string `json:"id"`
	Notes string `json:"notes"`
}

// Handler for /generate-notes
func GenerateNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the flat payload into a struct
	var notesReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&notesReq); err != nil {
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userMessageBytes, err := json.Marshal(notesReq)
	if err != nil {
		utils.WriteJSONError(w, "Failed to serialize user message", http.StatusInternalServerError)
		return
	}

	userMessage := Message{
		Role:    "user",
		Content: string(userMessageBytes),
	}

	cfg, ok := configs.GetConfig("/generate-notes")
	if !ok {
		utils.WriteJSONError(w, "No config for notes generation", http.StatusInternalServerError)
		return
	}

	// Load backend system prompt
	systemPrompt, err := prompt.GetPrompt("professorNotesPrompt")
	if err != nil {
		utils.WriteJSONError(w, "Failed to load notes system prompt", http.StatusInternalServerError)
		return
	}

	// Inject system + user messages
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		userMessage,
	}

	// Generate embedding
	combinedBytes, _ := json.Marshal(messages)
	embeddingVec, err := embedding.GetEmbedding(embedding.JSONStringWrapper(string(combinedBytes)))
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}
	key := string(combinedBytes)

	cachedResponse, _, hit, _ := supabaseClient.Check(embeddingVec)
	if hit {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResponse))
		return
	}

	// LLM call
	if llmClient == nil {
		utils.WriteJSONError(w, "LLM client not initialized", http.StatusInternalServerError)
		return
	}

	payload := map[string]interface{}{
		"model":       cfg.ModelName,
		"messages":    messages,
		"temperature": cfg.Temperature,
		"max_tokens":  cfg.MaxTokens,
	}

	rawResponse, err := llmClient.Call("/generate-lesson", payload)
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate lesson", http.StatusInternalServerError)
		return
	}

	// Clean response
	cleaned := strings.TrimSpace(rawResponse)
	if strings.HasPrefix(cleaned, "```markdown") {
		cleaned = strings.TrimPrefix(cleaned, "```markdown")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}
	if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}

	// Wrap final response
	result := NotesResponse{
		ID:    uuid.New().String(),
		Notes: cleaned,
	}

	finalBytes, err := json.Marshal(result)
	if err != nil {
		utils.WriteJSONError(w, "Failed to serialize notes", http.StatusInternalServerError)
		return
	}

	// Cache
	_ = supabaseClient.CacheEmbedding(key, embeddingVec, string(finalBytes))

	// Return
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
