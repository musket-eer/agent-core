package llm

import (
	"encoding/json"
	"net/http"
	"time"
	"agent-core/internal/cache"
	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"

	"github.com/google/uuid"
)

type RawCourseUnit struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Objective string `json:"objective"`
	Assessment string `json:"assessment"`

}

type RawCourse struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Units       []RawCourseUnit `json:"units"`
}

type CourseUnit struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Objective  string `json:"objective"`
	Assessment string `json:"assessment"`
}

type CourseResponse struct {
	ID                 string       `json:"id"`
	Title              string       `json:"title"`
	Description        string       `json:"description"`
	Difficulty         string       `json:"difficulty"`
	TotalDuration      int          `json:"totalDuration"`
	WeeklyHours        int          `json:"weeklyHours"`
	PriorKnowledge     []string     `json:"priorKnowledge"`
	Units              []CourseUnit `json:"units"`
	CreatedAt          string       `json:"createdAt"`
	UpdatedAt          string       `json:"updatedAt"`
	Progress           int          `json:"progress"`
	CurrentLessonIndex int          `json:"currentLessonIndex"`
}

func GenerateCourseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract payload from frontend (goals, profile, etc.)
	type Course struct {
		Title              string   `json:"title"`
		Description        string   `json:"description"`
		Difficulty         string   `json:"difficulty"`
		TotalDuration      int      `json:"total_duration"`
		WeeklyHours        int      `json:"weekly_hours"`
		PriorKnowledge     []string `json:"prior_knowledge"`
		LearningObjectives []string `json:"learning_objectives"`
	}
	type ProfileData struct {
		Name  string `json:"name"`
		EducationLevel string `json:"education_level"`
		Interests []string `json:"interests"`
		CareerGoals []string `json:"career_goals"`

	}
	type CoursePayLoad struct {
		Course      Course      `json:"course"`
		ProfileData ProfileData `json:"profile_data"`
	}

	var req CoursePayLoad
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	systemPrompt, err := prompt.GetPrompt("courseGenerationPrompt")
	if err != nil {
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}

	// --- Combine input
	userMessage := map[string]interface{}{
		"course":      req.Course,
		"profileData": req.ProfileData,
	}
	userMessageJson, _ := json.Marshal(userMessage)
	fullMessages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: string(userMessageJson)},
	}

	cfg, ok := configs.GetDeepseekConfig("/generate-course")
	if !ok {
		utils.WriteJSONError(w, "No config found", http.StatusInternalServerError)
		return
	}

	// Embedding for cache key
	combinedBytes, _ := json.Marshal(fullMessages)

	// Check cache
	payloadStr := string(combinedBytes)
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}
	// Final prompt
	finalPrompt := systemPrompt + "\n" + string(combinedBytes)
	payload := map[string]interface{}{
		"model":       cfg.ModelId,
		"input":       map[string]string{"prompt": finalPrompt},
		"temperature": cfg.Temperature,
		"max_tokens":  cfg.MaxTokens,
	}
    
	endpoint := "/generate-course"
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
		utils.WriteJSONError(w, "failed to get client "+clientName, http.StatusInternalServerError)
		return
	}

	response, err := client.Call(endpoint, payload)
	if err != nil {
		utils.WriteJSONError(w, "LLM call failed", http.StatusInternalServerError)
		return
	}

	cleaned := utils.ExtractJsonFromCodeBlock(response)

	// Parse the raw structure from LLM
	var raw RawCourse
	if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		utils.WriteJSONError(w, "LLM response unmarshal failed", http.StatusBadRequest)
		return
	}

	// Enrich into full structure
	now := time.Now().Format(time.RFC3339)
	courseID := uuid.New().String()

	var enrichedUnits []CourseUnit
	for _, u := range raw.Units {
		enrichedUnits = append(enrichedUnits, CourseUnit{
			ID:          uuid.New().String(),
			Title:       u.Title,
			Description: u.Description,
			Objective:  u.Objective,
			Assessment: u.Assessment,
		})
	}

	parsed := CourseResponse{
		ID:                 courseID,
		Title:              raw.Title,
		Description:        raw.Description,
		Difficulty:         req.Course.Difficulty,
		TotalDuration:      req.Course.TotalDuration,
		WeeklyHours:        req.Course.WeeklyHours,
		PriorKnowledge:     req.Course.PriorKnowledge,
		Units:              enrichedUnits,
		CreatedAt:          now,
		UpdatedAt:          now,
		Progress:           0,
		CurrentLessonIndex: 0,
	}

	finalBytes, err := json.Marshal(parsed)
	if err != nil {
		utils.WriteJSONError(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}
	// Store response in cache
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
