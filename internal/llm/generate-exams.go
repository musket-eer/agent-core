package llm

import (
	"encoding/json"
	"net/http"
	"time"

	"agent-core/internal/llm/configs"
	"github.com/google/uuid"
	"agent-core/internal/llm/utils"
	"agent-core/internal/cache"
)

// GenerateExamRequest expects messages
type GenerateExamRequest struct {
	Messages []Message `json:"messages"`
}

// ExamQuestion represents a single exam question
type ExamQuestion struct {
	QuestionText string   `json:"question_text"`
	Options      []string `json:"options"`
	QuestionID   string   `json:"question_id"`
}

// GenerateExamsHandler - POST /generate-exams
func GenerateExamsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateExamRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || len(req.Messages) == 0 {
		utils.WriteJSONError(w, "Invalid or missing messages array", http.StatusBadRequest)
		return
	}

	// --- Prepare model payload ---
	cfg, ok := configs.GetDeepseekConfig("/generate-exams")
	if !ok {
		utils.WriteJSONError(w, "No config for exam generation", http.StatusInternalServerError)
		return
	}

	// Prepare cache key
	cachePayload, _ := json.Marshal(req.Messages)
	payloadStr := string(cachePayload)
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	// Inject **exam generation formatting rules** (system message)
	systemMsg := Message{
		Role: "system",
		Content: `Generate exam questions in JSON format:
[
  {
    "question_text": "What is the main concept of...",
    "options": ["Option A", "Option B", "Option C", "Option D"]
  }
]
Respond with pure JSON — no extra text.`,
	}

	messages := append(req.Messages, systemMsg)

	payload := map[string]interface{}{
		"model":       cfg.ModelId,
		"messages":    messages,
		"temperature": cfg.Temperature,
		"max_tokens":  cfg.MaxTokens,
	}

	endpoint := "/generate-exams"
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
		utils.WriteJSONError(w, "Failed to generate exams", http.StatusInternalServerError)
		return
	}

	// --- Parse response ---
	var rawQuestions []struct {
		QuestionText string   `json:"question_text"`
		Options      []string `json:"options"`
	}
	err = json.Unmarshal([]byte(rawResponse), &rawQuestions)
	if err != nil {
		utils.WriteJSONError(w, "Failed to parse exam JSON", http.StatusBadRequest)
		return
	}

	// --- Post-process ---
	var questions []ExamQuestion
	for _, q := range rawQuestions {
		questions = append(questions, ExamQuestion{
			QuestionText: q.QuestionText,
			Options:      q.Options,
			QuestionID:   uuid.New().String(),
		})
	}

	final := map[string]interface{}{
		"id":         uuid.New().String(),
		"questions":  questions,
		"createdAt":  time.Now().Format(time.RFC3339),
		"updatedAt":  time.Now().Format(time.RFC3339),
	}

	finalBytes, _ := json.Marshal(final)
	// Store response in cache
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
