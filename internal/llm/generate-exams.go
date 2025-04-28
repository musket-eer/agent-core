package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"agent-core/internal/llm/configs"
	"agent-core/internal/embedding"
	"github.com/google/uuid"
	"agent-core/internal/llm/utils"
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

	// --- Embed input for caching ---
	combinedBytes, _ := json.Marshal(req.Messages)
	embeddingVec, err := embedding.GetEmbedding(embedding.JSONStringWrapper(string(combinedBytes)))
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}
	key := string(combinedBytes)

	cachedResponse, _, hit, err := supabaseClient.Check(embeddingVec)
	if err != nil {
	}
	if hit {
		log.Printf("[generate-exams] Cache hit. Responding with cached data: %s", cachedResponse)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResponse))
		return
	}

	// --- Prepare model payload ---
	cfg, ok := configs.GetConfig("/generate-exams")
	if !ok {
		utils.WriteJSONError(w, "No config for exam generation", http.StatusInternalServerError)
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
		"model":       cfg.ModelName,
		"messages":    messages,
		"temperature": cfg.Temperature,
		"max_tokens":  cfg.MaxTokens,
	}

	// --- LLM Call ---
	response, err := llmClient.Call("/generate-exams", payload)
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate exams", http.StatusInternalServerError)
		return
	}

	// --- Parse response ---
	var rawQuestions []struct {
		QuestionText string   `json:"question_text"`
		Options      []string `json:"options"`
	}
	err = json.Unmarshal([]byte(response), &rawQuestions)
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

	// --- Supabase Cache Store ---
	err = supabaseClient.CacheEmbedding(key, embeddingVec, string(finalBytes))
	if err != nil {
	}

	// --- Respond ---
	log.Printf("[generate-exams] LLM response sent to frontend: %s", string(finalBytes))
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
