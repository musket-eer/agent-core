package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"agent-core/internal/llm/configs"
	"github.com/google/uuid"
	"agent-core/internal/llm/utils"
	"agent-core/internal/cache"
)

// Add another type or data for examiner instructions

// GenerateExamRequest expects messages
type GenerateExamRequest struct {
	Messages []Message `json:"messages"`
}

// ExamQuestionOld represents a single exam question (legacy format)
type ExamQuestionOld struct {
	QuestionText string   `json:"question_text"`
	Options      []string `json:"options"`
	QuestionID   string   `json:"question_id"`
}

// GenerateExamsHandler - POST /generate-exams
func GenerateExamsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GenerateExamsHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		log.Printf("[GenerateExamsHandler] Invalid method: %s", r.Method)
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateExamRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || len(req.Messages) == 0 {
		log.Printf("[GenerateExamsHandler] Invalid request: %v", err)
		utils.WriteJSONError(w, "Invalid or missing messages array", http.StatusBadRequest)
		return
	}
	log.Printf("[GenerateExamsHandler] Parsed request with %d messages", len(req.Messages))

	// --- Prepare model payload ---
	cfg, ok := configs.GetDeepseekConfig("/generate-exams")
	if !ok {
		log.Printf("[GenerateExamsHandler] No config found for /generate-exams")
		utils.WriteJSONError(w, "No config for exam generation", http.StatusInternalServerError)
		return
	}

	// Prepare cache key
	cachePayload, _ := json.Marshal(req.Messages)
	payloadStr := string(cachePayload)
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
		log.Printf("[GenerateExamsHandler] Cache hit for exam generation")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}
	log.Printf("[GenerateExamsHandler] Cache miss, calling LLM")

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
		log.Printf("[GenerateExamsHandler] No config found for endpoint: %s", endpoint)
		utils.WriteJSONError(w, "no config found for endpoint: "+endpoint, http.StatusInternalServerError)
		return
	}

		clientName := config.ClientName
		if clientName == "" {
			clientName = "general" // fallback
		}

		client, err := GetLLMClient(clientName)
		if err != nil {
			log.Printf("[GenerateExamsHandler] Failed to get client %s: %v", clientName, err)
			utils.WriteJSONError(w, "failed to get client "+clientName+": "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("[GenerateExamsHandler] Calling LLM client: %s", clientName)

		rawResponse, err := client.Call(endpoint, payload)
	if err != nil {
		log.Printf("[GenerateExamsHandler] LLM call failed: %v", err)
		utils.WriteJSONError(w, "Failed to generate exams", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateExamsHandler] Received LLM response")

	// Clean up JSON response
	cleaned := utils.CleanJSONResponse(rawResponse)

	// --- Parse response ---
	var rawQuestions []struct {
		QuestionText string   `json:"question_text"`
		Options      []string `json:"options"`
	}
	err = json.Unmarshal([]byte(cleaned), &rawQuestions)
	if err != nil {
		log.Printf("[GenerateExamsHandler] Failed to parse exam JSON: %v", err)
		utils.WriteJSONError(w, "Failed to parse exam JSON", http.StatusBadRequest)
		return
	}
	log.Printf("[GenerateExamsHandler] Successfully parsed %d exam questions", len(rawQuestions))

	// --- Post-process ---
	var questions []ExamQuestionOld
	for _, q := range rawQuestions {
		questions = append(questions, ExamQuestionOld{
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
	log.Printf("[GenerateExamsHandler] Successfully generated exam with %d questions, cached response", len(questions))
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
