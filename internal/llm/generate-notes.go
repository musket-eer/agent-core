package llm

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"agent-core/internal/cache"
	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"

	"github.com/google/uuid"
)

// FrontendPayload expects model, messages, temperature

// TODO: - We want to maps this one to one with the frontend json
type NotesPayLoad struct {
	// course parameters
	CourseTitle       string   `json:"courseTitle"`
	CourseDescription string   `json:"courseDescription"`
	Difficulty        string    `json:"courseDifficultyLevel"`

	// lesson parameters
	LessonTitle       string   `json:"lessonTitle"`
	LessonObjective   string   `json:"lessonObjective"`
	LessonAssessment string    `json:"lessonAssessment"`

	// profile parameters
	EducationLevel    string   `json:"educationLevel"`
	Interests         []string `json:"interests"`
	CareerGoals       string   `json:"careerGoals"`
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

	// Decode the payload into the NotesPayLoad struct
	var notesReq NotesPayLoad
	var rawBodyBytes []byte
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // limit to 1MB
	rawBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	// Need to decode again since we've read the body
	err = json.Unmarshal(rawBodyBytes, &notesReq)
	if err != nil {
		utils.WriteJSONError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	userMessageBytes, err := json.Marshal(notesReq)
	if err != nil {
		utils.WriteJSONError(w, "Failed to serialize user message", http.StatusInternalServerError)
		return
	}
	cfg, ok := configs.GetDeepseekConfig("/generate-notes")
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
	promptStr := systemPrompt + "\n" + string(userMessageBytes)
	payload := map[string]interface{}{
		"model": cfg.ModelId,
		"input": map[string]string{
			"prompt": promptStr,
		},
	}

	// Stringify the payload for hashing
	payloadStr := string(userMessageBytes)
	resp, hit, err := cache.Get().CheckByPayload(payloadStr)
	if err != nil {
		utils.WriteJSONError(w, "Cache error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if hit {
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
		return
	}
	endpoint := "/generate-notes"
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
		utils.WriteJSONError(w, "Failed to generate notes", http.StatusInternalServerError)
		return
	}
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
	// Store response in cache with payload hash as key (stringified payload, stringified response)
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)

	// Return
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(finalBytes)
	if writeErr != nil {
	}
}
