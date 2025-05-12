package llm

import (
	"encoding/json"
	"net/http"
	"agent-core/internal/llm/configs"
	"agent-core/internal/prompt"
	"agent-core/internal/llm/utils"

	"github.com/google/uuid"

)

// FrontendPayload represents the incoming structure
type LessonRequest struct {
	LessonTitle       string   `json:"lesson_title"`
	LessonDescription string   `json:"lesson_description"`
	LessonObjectives  []string `json:"lesson_objectives"`
	LessonSize        int      `json:"lesson_size"`
	EducationLevel    string   `json:"education_level"`
	Interests         []string `json:"interests"`
	CareerGoals       string   `json:"career_goals"`
}



type LessonMetadata struct {
	Title      string   `json:"title"`
	Objectives []string `json:"objectives"`
	Duration   string   `json:"duration"`
	Difficulty string   `json:"difficulty"`
}

type Activity struct {
	ID       string    `json:"id"`
	Leader   string    `json:"leader"`
	Title    string    `json:"title"`
	Messages []Message `json:"messages"`
}

type LessonPlan struct {
	LessonMetadata LessonMetadata `json:"lessonMetadata"`
	Activities     []Activity      `json:"activities"`
}

// POST /generate-lesson
func GenerateLessonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var lessonReq LessonRequest
	if err := json.NewDecoder(r.Body).Decode(&lessonReq); err != nil {
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cfg, ok := configs.GetDeepseekConfig("/generate-lesson")
	if !ok {
		utils.WriteJSONError(w, "No config for generation", http.StatusInternalServerError)
		return
	}

	systemPrompt, err := prompt.GetPrompt("lessonSequencePrompt")
	if err != nil {
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}

	// Transform input into prompt
	userMessage, _ := json.Marshal(lessonReq)
	fullPrompt := systemPrompt + "\n" + string(userMessage)

	// Use hash-based cache: payload is fullPrompt
	// cached, hit, err := supabaseClient.CheckByPayload(fullPrompt)
	// if err != nil {
	// 	utils.WriteJSONError(w, "Cache error", http.StatusInternalServerError)
	// 	return
	// }
	// if hit {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write([]byte(cached))
	// 	return
	// }

	// Make LLM call
	payload := map[string]interface{}{
		"model":       cfg.ModelId,
		"input":       map[string]string{"prompt": fullPrompt},
		"temperature": cfg.Temperature,
		"max_tokens":  cfg.MaxTokens,
	}
	rawResponse, err := llmClient.Call("/generate-lesson", payload)
	if err != nil {
		utils.WriteJSONError(w, "Failed to call LLM", http.StatusInternalServerError)
		return
	}

	// Parse plain-text lines
	lines := utils.ExtractNonEmptyLines(rawResponse)
	if len(lines) != 10 {
		utils.WriteJSONError(w, "Expected exactly 10 activities", http.StatusBadRequest)
		return
	}

	activities := make([]Activity, 0, 10)
	for _, line := range lines {
		parts := utils.SplitByPipe(line)
		if len(parts) != 2 {
			continue
		}
		leader := utils.Trim(parts[0])
		title := utils.Trim(parts[1])
		activities = append(activities, Activity{
			ID:     uuid.New().String(),
			Leader: leader,
			Title:  title,
			
		})
	}

	// Synthesize lesson metadata from request
	metadata := LessonMetadata{
		Title:      lessonReq.LessonTitle,
		Objectives: lessonReq.LessonObjectives,
		Duration:   "1 hour", // or based on LessonSize
		Difficulty: lessonReq.EducationLevel,
	}

	// Compose final lesson
	lesson := LessonPlan{
		LessonMetadata: metadata,
		Activities:     activities,
	}

	final := map[string]interface{}{
		"lessonPlan": lesson,
	}
	finalBytes, err := json.Marshal(final)
	if err != nil {
		utils.WriteJSONError(w, "Failed to serialize lesson", http.StatusInternalServerError)
		return
	}

	// Cache final result
	// _ = supabaseClient.CacheByPayload(fullPrompt, string(finalBytes))

	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}