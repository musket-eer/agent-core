package llm

import (
	"agent-core/internal/cache"
	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// add another type called course desinger instruction
type RawCourseUnit struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Objective   string `json:"objective"`
	Assessment  string `json:"assessment"`
}

type RawCourse struct {
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	Units          []RawCourseUnit `json:"units"`
	ReferenceBooks []string        `json:"referenceBooks"`
}

type CourseUnit struct {
	ID          string `json:"id"`
	Title       string `json:"lessonTitle"`
	Description string `json:"description"`
	Objective   string `json:"objective"`
	Assessment  string `json:"assessment"`
}

type CourseResponse struct {
	ID                 string       `json:"id"`
	Title              string       `json:"title"`
	Description        string       `json:"description"`
	Difficulty         string       `json:"difficultyLevel"`
	WeeklyHours        int          `json:"weeklyHours"`
	PriorKnowledge     []string     `json:"priorKnowledge"`
	Units              []CourseUnit `json:"units"`
	CreatedAt          string       `json:"createdAt"`
	UpdatedAt          string       `json:"updatedAt"`
	Progress           int          `json:"progress"`
	CurrentLessonIndex int          `json:"currentLessonIndex"`
	Lessons            int          `json:"lessons"`
	ReferenceBooks     []string     `json:"referenceBooks"`
}

func GenerateCourseHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GenerateCourseHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		log.Printf("[GenerateCourseHandler] Invalid method: %s", r.Method)
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract payload from frontend (goals, profile, etc.)
	type Course struct {
		Title              string   `json:"title"`
		Description        string   `json:"description"`
		Difficulty         string   `json:"difficultyLevel"`
		WeeklyHours        int      `json:"weeklyHours"`
		PriorKnowledge     []string `json:"priorKnowledge"`
		LearningObjectives []string `json:"objectives"`
		ReferenceBooks     []string `json:"referenceBooks"`
		Lessons            int      `json:"lessons"`
	}
	type ProfileData struct {
		Name           string   `json:"name"`
		EducationLevel string   `json:"educationLevel"`
		Interests      []string `json:"interests"`
		CareerGoals    []string `json:"careerGoals"`
	}
	type CoursePayLoad struct {
		Course      Course      `json:"course"`
		ProfileData ProfileData `json:"profileData"`
	}

	var req CoursePayLoad
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GenerateCourseHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("[GenerateCourseHandler] Parsed request for course: %s", req.Course.Title)

	systemPrompt, err := prompt.GetPrompt("courseGenerationPrompt")
	if err != nil {
		log.Printf("[GenerateCourseHandler] Failed to load prompt: %v", err)
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateCourseHandler] Loaded system prompt")

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
		log.Printf("[GenerateCourseHandler] No config found for /generate-course")
		utils.WriteJSONError(w, "No config found", http.StatusInternalServerError)
		return
	}

	// Embedding for cache key
	combinedBytes, _ := json.Marshal(fullMessages)

	// Check cache
	payloadStr := string(combinedBytes)
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
		log.Printf("[GenerateCourseHandler] Cache hit for course generation")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}
	log.Printf("[GenerateCourseHandler] Cache miss, calling LLM")
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
		log.Printf("[GenerateCourseHandler] No config found for endpoint: %s", endpoint)
		utils.WriteJSONError(w, "no config found for endpoint: "+endpoint, http.StatusInternalServerError)
		return
	}

	clientName := config.ClientName
	if clientName == "" {
		clientName = "general" // fallback
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		log.Printf("[GenerateCourseHandler] Failed to get client %s: %v", clientName, err)
		utils.WriteJSONError(w, "failed to get client "+clientName, http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateCourseHandler] Calling LLM client: %s", clientName)

	response, err := client.Call(endpoint, payload)
	if err != nil {
		log.Printf("[GenerateCourseHandler] LLM call failed: %v", err)
		utils.WriteJSONError(w, "LLM call failed", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateCourseHandler] Received LLM response")

	// Clean up JSON response
	cleaned := utils.CleanJSONResponse(response)
	// Parse the raw structure from LLM
	var raw RawCourse
	if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		log.Printf("[GenerateCourseHandler] Failed to unmarshal LLM response: %v", err)
		utils.WriteJSONError(w, "LLM response unmarshal failed", http.StatusBadRequest)
		return
	}
	log.Printf("[GenerateCourseHandler] Successfully parsed course: %s with %d units", raw.Title, len(raw.Units))

	// Enrich into full structure
	now := time.Now().Format(time.RFC3339)
	courseID := uuid.New().String()

	var enrichedUnits []CourseUnit
	for _, u := range raw.Units {
		log.Printf("Enriched Unit Title: %s", u.Title)
		enrichedUnits = append(enrichedUnits, CourseUnit{
			ID:          uuid.New().String(),
			Title:       u.Title,
			Description: u.Description,
			Objective:   u.Objective,
			Assessment:  u.Assessment,
		})
	}

	parsed := CourseResponse{
		ID:                 courseID,
		Title:              raw.Title,
		Description:        raw.Description,
		Difficulty:         req.Course.Difficulty,
		WeeklyHours:        req.Course.WeeklyHours,
		PriorKnowledge:     req.Course.PriorKnowledge,
		Units:              enrichedUnits,
		CreatedAt:          now,
		UpdatedAt:          now,
		Progress:           0,
		CurrentLessonIndex: 0,
		Lessons:            len(enrichedUnits),
		ReferenceBooks:     raw.ReferenceBooks,
	}

	finalBytes, err := json.Marshal(parsed)
	if err != nil {
		log.Printf("[GenerateCourseHandler] Failed to serialize response: %v", err)
		utils.WriteJSONError(w, "Failed to serialize response", http.StatusInternalServerError)
		return
	}
	// Store response in cache
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)
	log.Printf("[GenerateCourseHandler] Successfully generated course, cached response")
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
