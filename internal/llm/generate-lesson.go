package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"agent-core/internal/llm/configs"
	"agent-core/internal/prompt"
	"agent-core/internal/llm/utils"
	"agent-core/internal/cache"

	"github.com/google/uuid"
)

// Add another data called lesson planner instructions. To add more control to number of lessons
type LessonRequest struct {
	// course parameters
	CourseTitle       string   `json:"courseTitle"`
	CourseDescription string   `json:"courseDescription"`
	CourseDifficultyLevel        string    `json:"difficulty"`

	// lesson parameters
	LessonTitle       string   `json:"lessonTitle"`
	LessonObjective   string   `json:"lessonObjective"`
	LessonAssessment string    `json:"lessonAssessment"`

	// profile parameters
	EducationLevel    string   `json:"educationLevel"`
	Interests         []string `json:"interests"`
	CareerGoals       string   `json:"careerGoals"`
}

// LessonMetadata is now defined in types_agents.go

type LessonActivity struct {
	ID       string    `json:"id"`
	Leader   string    `json:"leader"`
	Title    string    `json:"title"`
	Objective string   `json:"objective"`
	Messages []Message `json:"messages"`
}

type LessonPlan struct {
	LessonMetadata LessonMetadata `json:"lessonMetadata"`
	Activities     []LessonActivity      `json:"activities"`
}

// POST /generate-lesson
func GenerateLessonHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GenerateLessonHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		log.Printf("[GenerateLessonHandler] Invalid method: %s", r.Method)
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var lessonReq LessonRequest
	if err := json.NewDecoder(r.Body).Decode(&lessonReq); err != nil {
		log.Printf("[GenerateLessonHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("[GenerateLessonHandler] Parsed request for lesson: %s", lessonReq.LessonTitle)

	cfg, ok := configs.GetDeepseekConfig("/generate-lesson")
	if !ok {
		log.Printf("[GenerateLessonHandler] No config found for /generate-lesson")
		utils.WriteJSONError(w, "No config for generation", http.StatusInternalServerError)
		return
	}
	// Prepare cache key
	cachePayload, _ := json.Marshal(lessonReq)
	payloadStr := string(cachePayload)
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
		log.Printf("[GenerateLessonHandler] Cache hit for lesson generation")
		w.Header().Set("Content-Type", "application/json")
		w.Write(cached)
		return
	}
	log.Printf("[GenerateLessonHandler] Cache miss, calling LLM")
	systemPrompt, err := prompt.GetPrompt("lessonSequencePrompt")
	if err != nil {
		log.Printf("[GenerateLessonHandler] Failed to load prompt: %v", err)
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateLessonHandler] Loaded system prompt")

	// Transform input into prompt
	userMessage, _ := json.Marshal(lessonReq)

	// Make LLM call
	payload := map[string]interface{}{
		"model": cfg.ModelId,
		"input": map[string]string{
			"prompt": systemPrompt + "\n" + string(userMessage),
		},
	}
	endpoint := "/generate-lesson"
	config, ok := configs.GetDeepseekConfig(endpoint)
	if !ok {
		log.Printf("[GenerateLessonHandler] No config found for endpoint: %s", endpoint)
		utils.WriteJSONError(w, "no config found for endpoint: "+endpoint, http.StatusInternalServerError)
		return
	}
	clientName := config.ClientName
	if clientName == "" {
		clientName = "general" // fallback
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		log.Printf("[GenerateLessonHandler] Failed to get client %s: %v", clientName, err)
		utils.WriteJSONError(w, "failed to get client "+clientName+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateLessonHandler] Calling LLM client: %s", clientName)

	rawResponse, err := client.Call(endpoint, payload)
	if err != nil {
		log.Printf("[GenerateLessonHandler] LLM call failed: %v", err)
		utils.WriteJSONError(w, "Failed to generate lesson", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateLessonHandler] Received LLM response")

	// Clean up JSON response
	cleaned := utils.CleanJSONResponse(rawResponse)

	// Parse plain-text lines
	lines := utils.ExtractNonEmptyLines(cleaned)
	if len(lines) != 10 {
		log.Printf("[GenerateLessonHandler] Expected 10 activities, got %d", len(lines))
		utils.WriteJSONError(w, "Expected exactly 10 activities", http.StatusBadRequest)
		return
	}
	log.Printf("[GenerateLessonHandler] Successfully parsed %d activities", len(lines))

	activities := make([]LessonActivity, 0, 10)
	for _, line := range lines {

		parts := utils.SplitByPipe(line)
		if len(parts) < 2 {
			continue
		}
		leader := utils.Trim(parts[0])
		title := utils.Trim(parts[1])
		objective := utils.Trim(parts[2])
		activities = append(activities, LessonActivity{
			ID: uuid.New().String(),
			Leader: leader,
			Title:  title,
			Objective: objective,
		})
	}

	// Synthesize lesson metadata from request
	metadata := LessonMetadata{
		Title:      lessonReq.LessonTitle,
		Objective: lessonReq.LessonObjective,
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
		log.Printf("[GenerateLessonHandler] Failed to serialize lesson: %v", err)
		utils.WriteJSONError(w, "Failed to serialize lesson", http.StatusInternalServerError)
		return
	}
	// Store response in cache
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)
	log.Printf("[GenerateLessonHandler] Successfully generated lesson, cached response")
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(finalBytes)
	if writeErr != nil {
		log.Printf("[GenerateLessonHandler] Failed to write response: %v", writeErr)
		utils.WriteJSONError(w, "Failed to write response", http.StatusInternalServerError)
	}
}