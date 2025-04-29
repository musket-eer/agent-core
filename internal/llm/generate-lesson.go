package llm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"agent-core/internal/embedding"
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
	ID     string `json:"id"`
	Leader string `json:"leader"`
	Title  string `json:"title"`
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
	fmt.Printf("[generate-lesson] Incoming request: %+v\n", lessonReq)
	if err := json.NewDecoder(r.Body).Decode(&lessonReq); err != nil {
		fmt.Printf("[generate-lesson] JSON decode error: %v\n", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		fmt.Printf("[generate-lesson] Invalid request body: %v\n", err)
		return
	}

	cfg, ok := configs.GetDeepseekConfig("/generate-lesson")
	if !ok {
		fmt.Printf("[generate-lesson] No config found for /generate-lesson\n")
		utils.WriteJSONError(w, "No config for generation", http.StatusInternalServerError)
		fmt.Printf("[generate-lesson] No config found for /generate-lesson\n")
		return
	}

	systemPrompt, err := prompt.GetPrompt("lessonSequencePrompt")
	if err != nil {
		fmt.Printf("[generate-lesson] Failed to load system prompt: %v\n", err)
		utils.WriteJSONError(w, "Failed to load system prompt", http.StatusInternalServerError)
		fmt.Printf("[generate-lesson] Failed to load system prompt: %v\n", err)
		return
	}

	lessonReqJson, err := json.Marshal(lessonReq)
	if err != nil {
		utils.WriteJSONError(w, "Failed to serialize lesson request", http.StatusInternalServerError)
		return
	}

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: string(lessonReqJson)},
	}

	// Cache Key
	combinedBytes, _ := json.Marshal(messages)
	embeddingVec, err := embedding.GetEmbedding(embedding.JSONStringWrapper(string(combinedBytes)))
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}

	cachedResponse, _, hit, _ := supabaseClient.Check(embeddingVec)
	if hit {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResponse))
		return
	}

	if llmClient == nil {
		utils.WriteJSONError(w, "LLM client not initialized", http.StatusInternalServerError)
		return
	}

	// Transform the incoming flat LessonRequest into a single prompt string for Deepseek
	lessonReqJson, err = json.Marshal(lessonReq)
	if err != nil {
		fmt.Printf("[generate-lesson] Failed to serialize lesson request: %v\n", err)
		utils.WriteJSONError(w, "Failed to serialize lesson request", http.StatusInternalServerError)
		return
	}
	promptStr := systemPrompt + "\n" + string(lessonReqJson)
	payload := map[string]interface{}{
		"model": cfg.ModelName,
		"input": map[string]string{"prompt": promptStr},
		"temperature": cfg.Temperature,
		"max_tokens":  cfg.MaxTokens,
	}
	fmt.Printf("[generate-lesson] Payload for LLM: %+v\n", payload)

	rawResponse, err := llmClient.Call("/generate-lesson", payload)
	if err != nil {
		fmt.Printf("[generate-lesson] LLM client error: %v\n", err)
		utils.WriteJSONError(w, "Failed to generate lesson", http.StatusInternalServerError)
		return
	}

	cleaned := utils.ExtractJsonFromCodeBlock(rawResponse)

	fmt.Printf("[generate-lesson] Raw LLM response before parsing: %s\n", cleaned)

	// Parse lesson data using the utility function for raw LLM output
	parsed := utils.ParseLessonData(cleaned)

	// Map into structured LessonPlan
	lessonMetadataRaw := parsed["lessonMetadata"].(map[string]interface{})
	activitiesRaw := parsed["activities"].([]map[string]string)

	metadata := LessonMetadata{
		Title:      lessonMetadataRaw["title"].(string),
		Objectives: toStringArray(lessonMetadataRaw["objectives"]),
		Duration:   "",
		Difficulty: "",
	}
	if v, ok := lessonMetadataRaw["duration"]; ok {
		metadata.Duration = v.(string)
	}
	if v, ok := lessonMetadataRaw["difficulty"]; ok {
		metadata.Difficulty = v.(string)
	}

	activities := []Activity{}
	for _, act := range activitiesRaw {
		id := act["id"]
		if id == "" {
			id = uuid.New().String()
		}
		activities = append(activities, Activity{
			ID:     id,
			Leader: act["leader"],
			Title:  act["title"],
		})
	}

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

	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}

// helper
func toStringArray(raw interface{}) []string {
	items, ok := raw.([]interface{})
	if !ok {
		return []string{}
	}
	strs := make([]string, 0, len(items))
	for _, i := range items {
		if s, ok := i.(string); ok {
			strs = append(strs, s)
		}
	}
	return strs
}
