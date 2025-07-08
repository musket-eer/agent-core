package llm

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"log"

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
	// DEBUG LOG: Handler entry
	log.Printf("[GenerateNotesHandler] Incoming request: %s %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		log.Printf("[GenerateNotesHandler] Invalid method: %s", r.Method)
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the payload into the NotesPayLoad struct
	var notesReq NotesPayLoad
	var rawBodyBytes []byte
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // limit to 1MB
	rawBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[GenerateNotesHandler] Failed to read request body: %v", err)
		utils.WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	log.Printf("[GenerateNotesHandler] Raw body: %s", string(rawBodyBytes))
	// Need to decode again since we've read the body
	err = json.Unmarshal(rawBodyBytes, &notesReq)
	if err != nil {
		log.Printf("[GenerateNotesHandler] Invalid request payload: %v", err)
		utils.WriteJSONError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("[GenerateNotesHandler] Parsed NotesPayLoad: %+v", notesReq)
	userMessageBytes, err := json.Marshal(notesReq)
	if err != nil {
		log.Printf("[GenerateNotesHandler] Failed to serialize user message: %v", err)
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
		log.Printf("[GenerateNotesHandler] Failed to load notes system prompt: %v", err)
		utils.WriteJSONError(w, "Failed to load notes system prompt", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateNotesHandler] Loaded system prompt")
	promptStr := systemPrompt + "\n" + string(userMessageBytes)
	// Use Gemini config for payload
	// geminiCfg, ok := configs.GetGeminiConfig("/generate-notes")
	// if !ok {
	// 	log.Printf("[GenerateNotesHandler] No Gemini config for notes generation")
	// 	utils.WriteJSONError(w, "No Gemini config for notes generation", http.StatusInternalServerError)
	// 	return
	// }
	// log.Printf("[GenerateNotesHandler] Loaded Gemini config: %+v", cfg)
	payload := map[string]interface{}{
		"model": cfg.ModelId,
		"input": map[string]string{
			"prompt": promptStr,
		},
	}
	log.Printf("[GenerateNotesHandler] Built payload for LLM: %+v", payload)

	// Stringify the payload for hashing
	payloadStr := string(userMessageBytes)
	resp, hit, err := cache.Get().CheckByPayload(payloadStr)
	if err != nil {
		log.Printf("[GenerateNotesHandler] Cache error: %v", err)
		utils.WriteJSONError(w, "Cache error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if hit {
		log.Printf("[GenerateNotesHandler] Cache hit for payload")
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
		return
	}
	log.Printf("[GenerateNotesHandler] Cache miss for payload")
	endpoint := "/generate-notes"
	// Commented out Deepseek client selection for Gemini test
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

	// Use Gemini client directly for testing
	// client, err = GetLLMClient("deepseek")
	if err != nil {
		log.Printf("[GenerateNotesHandler] Failed to get Gemini client: %v", err)
		utils.WriteJSONError(w, "Failed to get Gemini client", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateNotesHandler] Calling Gemini client...")
	rawResponse, err := client.Call(endpoint, payload)
	if err != nil {
		log.Printf("[GenerateNotesHandler] Failed to generate notes: %v", err)
		utils.WriteJSONError(w, "Failed to generate notes", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateNotesHandler] Raw LLM response: %s", rawResponse)
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
		log.Printf("[GenerateNotesHandler] Failed to serialize notes: %v", err)
		utils.WriteJSONError(w, "Failed to serialize notes", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateNotesHandler] Final notes response: %s", string(finalBytes))
	// Store response in cache with payload hash as key (stringified payload, stringified response)
	cacheErr := cache.Get().CacheByPayload(payloadStr, finalBytes)
	if cacheErr != nil {
		log.Printf("[GenerateNotesHandler] Failed to cache response: %v", cacheErr)
	}
	// Return
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(finalBytes)
	if writeErr != nil {
		log.Printf("[GenerateNotesHandler] Failed to write response: %v", writeErr)
	}
}
