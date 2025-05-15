package llm

import (
	"encoding/json"
	"net/http"
	"strings"

	"agent-core/internal/llm/configs"
	"agent-core/internal/prompt"
	"agent-core/internal/llm/utils"
	"agent-core/internal/cache"

	"github.com/google/uuid"
)

type ResourcesPayload struct {
	Messages []Message `json:"messages"`
}

// Resource object
type Resource struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	ID          string `json:"id"`
}

// ResourcesResponse wraps the final return
type ResourcesResponse struct {
	Resources []Resource `json:"resources"`
}

// Handler for /generate-study-resources
func GenerateResourcesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ResourcesPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Messages) < 2 {
		utils.WriteJSONError(w, "Messages array must include resource context.", http.StatusBadRequest)
		return
	}

	userMessage := req.Messages[1]
	// Prepare cache key
	cachePayload, _ := json.Marshal(req.Messages)
	payloadStr := string(cachePayload)
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}
	cfg, ok := configs.GetDeepseekConfig("/generate-resources")
	if !ok {
		utils.WriteJSONError(w, "No config for resource generation", http.StatusInternalServerError)
		return
	}

	systemPrompt, err := prompt.GetPrompt("librarianResourcesPrompt")
	if err != nil {
		utils.WriteJSONError(w, "Failed to load resource system prompt", http.StatusInternalServerError)
		return
	}

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		userMessage,
	}

	payload := map[string]interface{}{
		"model":       cfg.ModelId,
		"messages":    messages,
		"temperature": cfg.Temperature,
		"max_tokens":  cfg.MaxTokens,
	}

	// TODO:- Hash Payload, check cache which wraps supabase, if hit return
	endpoint := "/chat"
	config, ok := configs.GetDeepseekConfig(endpoint)
	if !ok {
		utils.WriteJSONError(w, "no config found for endpoint: "+endpoint, http.StatusInternalServerError)
		return
	}

	clientName := config.ClientName
	if clientName == "" {
		clientName = "general" // fallback
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		utils.WriteJSONError(w, "failed to get client "+clientName+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	rawResponse, err := client.Call(endpoint, payload)
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate study resources", http.StatusInternalServerError)
		return
	}

	cleaned := strings.TrimSpace(rawResponse)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}
	if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}

	// Parse LLM output
	var parsed []map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		utils.WriteJSONError(w, "Failed to parse study resources response", http.StatusBadRequest)
		return
	}

	// Map parsed into Resource struct
	resources := make([]Resource, 0, len(parsed))
	for _, item := range parsed {
		res := Resource{
			Title:       item["title"].(string),
			Description: item["description"].(string),
			Link:        item["link"].(string),
			ID:          uuid.New().String(),
		}
		resources = append(resources, res)
	}

	final := ResourcesResponse{
		Resources: resources,
	}

	finalBytes, err := json.Marshal(final)
	if err != nil {
		utils.WriteJSONError(w, "Failed to serialize resources", http.StatusInternalServerError)
		return
	}
	// Store response in cache
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
