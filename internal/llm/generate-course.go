package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"agent-core/internal/llm/configs"
	"agent-core/internal/prompt"
	"agent-core/internal/embedding"
	"agent-core/internal/llm/utils"

	"github.com/google/uuid"
)

type CourseUnit struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Duration    int      `json:"duration"`
	Objectives  []string `json:"objectives"`
	Content     []string `json:"content"`
	Exercises   []string `json:"exercises"`
	Assessment  string   `json:"assessment"`
}

type CourseResponse struct {
	ID                  string       `json:"id"`
	Title               string       `json:"title"`
	Description         string       `json:"description"`
	Difficulty          string       `json:"difficulty"`
	TotalDuration       int          `json:"totalDuration"`
	WeeklyHours         int          `json:"weeklyHours"`
	Prerequisites       []string     `json:"prerequisites"`
	LearningObjectives  []string     `json:"learningObjectives"`
	Units               []CourseUnit `json:"units"`
	CreatedAt           string       `json:"createdAt"`
	UpdatedAt           string       `json:"updatedAt"`
	Progress            int          `json:"progress"`
	CurrentLessonIndex  int          `json:"currentLessonIndex"`
}

func GenerateCourseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse frontend payload (ignore model, temperature, etc.)
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// --- Load system prompt
	systemPrompt, err := prompt.GetPrompt("courseGenerationPrompt")
	if err != nil {
		utils.WriteJSONError(w, "Failed to load system prompt", http.StatusInternalServerError)
		return
	}

	// --- Combine user content and system prompt
	userMessage := map[string]interface{}{
		"course": req["course"],
		"profileData": req["profileData"],
	}
	userMessageJson, _ := json.Marshal(userMessage)
	fullMessages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: string(userMessageJson)},
	}

	cfg, ok := configs.GetDeepseekConfig("/generate-course")
	if !ok {
		utils.WriteJSONError(w, "No config for generation", http.StatusInternalServerError)
		return
	}

	// --- Embedding for cache key
	combinedBytes, _ := json.Marshal(fullMessages)
	embeddingVec, err := embedding.GetEmbedding(embedding.JSONStringWrapper(string(combinedBytes)))
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}

	key := string(combinedBytes)

	// --- Check Supabase cache
	cachedResponse, _, hit, _ := supabaseClient.Check(embeddingVec)
	if hit {
		log.Println("[generate-course] Cache hit.")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResponse))
		return
	}

	// --- Build LLM payload
	finalPrompt := systemPrompt + "\n" + string(combinedBytes)

	payload := map[string]interface{}{
		"model":       cfg.ModelName,
		"input":       map[string]string{"prompt": finalPrompt},
		"temperature": cfg.Temperature,
		"max_tokens":  cfg.MaxTokens,
	}

	response, err := llmClient.Call("/generate-course", payload)
	if err != nil {
		utils.WriteJSONError(w, "Failed to call LLM", http.StatusInternalServerError)
		return
	}

	// --- Clean response (strip ```json)
	cleaned := utils.ExtractJsonFromCodeBlock(response)

	// --- Parse into strongly typed struct
	var parsed CourseResponse
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		log.Printf("[generate-course] JSON parse error: %v", err)
		utils.WriteJSONError(w, "Invalid JSON format in LLM response", http.StatusBadRequest)
		return
	}

	// --- Enrich with backend fields
	parsed.ID = uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	parsed.CreatedAt = now
	parsed.UpdatedAt = now
	parsed.Progress = 0
	parsed.CurrentLessonIndex = 0

	// --- Marshal final response
	finalBytes, err := json.Marshal(parsed)
	if err != nil {
		utils.WriteJSONError(w, "Failed to serialize course", http.StatusInternalServerError)
		return
	}

	// --- Cache final response
	_ = supabaseClient.CacheEmbedding(key, embeddingVec, string(finalBytes))

	// --- Send response
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
