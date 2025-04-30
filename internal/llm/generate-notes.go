package llm

import (
	"encoding/json"
	"io"
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
	LessonTitle         string   `json:"lessonTitle"`
	StudentProfile StudentProfile  `json:"studentProfile"`
	CourseUnit     CourseUnit      `json:"courseUnit"`
}

type StudentProfile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	EducationLevel string `json:"educationLevel"`
	Interests []string `json:"interests"`
	CareerGoals string `json:"careerGoals"`
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

	// Inject system + user messages
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: string(userMessageBytes)},
	}

	// Generate embedding
	combinedBytes, _ := json.Marshal(messages)
	embeddingVec, err := embedding.GetEmbedding(embedding.JSONStringWrapper(string(combinedBytes)))
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}
	// key := string(combinedBytes)
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

	promptStr := systemPrompt + "\n" + string(userMessageBytes)
	payload := map[string]interface{}{
		"modelName": cfg.ModelName,
		"input": map[string]string{
			"prompt": promptStr,
		},
	}

	// Log the payload being sent to LLM
	rawResponse, err := llmClient.Call("/generate-notes", payload)
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate notes", http.StatusInternalServerError)
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

	// Cache (disabled)
	// _ = supabaseClient.CacheEmbedding(key, embeddingVec, string(finalBytes))

	// Return
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
