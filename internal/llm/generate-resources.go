package llm

import (
	"agent-core/internal/cache"
	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"
	"agent-core/internal/videos"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type ResourcesPayLoad struct {
	// course parameters
	CourseTitle       string `json:"courseTitle"`
	CourseDescription string `json:"courseDescription"`
	Difficulty        string `json:"difficulty"`

	// lesson parameters
	LessonTitle      string `json:"lessonTitle"`
	LessonObjective  string `json:"lessonObjective"`
	LessonAssessment string `json:"lessonAssessment"`

	// profile parameters
	EducationLevel string   `json:"educationLevel"`
	Interests      []string `json:"interests"`
	CareerGoals    string   `json:"careerGoals"`
}

// ResourcesResponse wraps the final return
type ResourcesResponse struct {
	Resources []Resource `json:"resources"`
}

func GenerateResourcesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GenerateResourcesHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		log.Printf("[GenerateResourcesHandler] Invalid method: %s", r.Method)
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resourcesReq ResourcesPayLoad
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // limit to 1MB
	rawBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[GenerateResourcesHandler] Failed to read request body: %v", err)
		utils.WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(rawBodyBytes, &resourcesReq); err != nil {
		log.Printf("[GenerateResourcesHandler] Invalid request payload: %v", err)
		utils.WriteJSONError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("[GenerateResourcesHandler] Parsed request for lesson: %s", resourcesReq.LessonTitle)

	userMessageBytes, err := json.Marshal(resourcesReq)
	if err != nil {
		log.Printf("[GenerateResourcesHandler] Failed to serialize user message: %v", err)
		utils.WriteJSONError(w, "Failed to serialize user message", http.StatusInternalServerError)
		return
	}
	payloadStr := string(userMessageBytes)
	if resp, hit, _ := cache.Get().CheckByPayload(payloadStr); hit {
		log.Printf("[GenerateResourcesHandler] Cache hit for resources generation")
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
		return
	}
	log.Printf("[GenerateResourcesHandler] Cache miss, calling LLM")

	cfg, ok := configs.GetDeepseekConfig("/generate-resources")
	if !ok {
		log.Printf("[GenerateResourcesHandler] No config found for /generate-resources")
		utils.WriteJSONError(w, "No config for resources generation", http.StatusInternalServerError)
		return
	}

	systemPrompt, err := prompt.GetPrompt("librarianResourcesPrompt")
	if err != nil {
		log.Printf("[GenerateResourcesHandler] Failed to load prompt: %v", err)
		utils.WriteJSONError(w, "Failed to load resources system prompt", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateResourcesHandler] Loaded system prompt")

	promptStr := systemPrompt + "\n" + string(userMessageBytes)
	payload := map[string]interface{}{
		"model": cfg.ModelId,
		"input": map[string]string{
			"prompt": promptStr,
		},
	}

	clientName := cfg.ClientName

	if clientName == "" {
		clientName = "general"
	}
	client, err := GetLLMClient(clientName)
	if err != nil {
		log.Printf("[GenerateResourcesHandler] Failed to get client %s: %v", clientName, err)
		utils.WriteJSONError(w, "failed to get client "+clientName+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateResourcesHandler] Calling LLM client: %s", clientName)

	rawResponse, err := client.Call("/generate-resources", payload)
	if err != nil {
		log.Printf("[GenerateResourcesHandler] LLM call failed: %v", err)
		utils.WriteJSONError(w, "Failed to generate search query", http.StatusInternalServerError)
		return
	}
	log.Printf("[GenerateResourcesHandler] Received LLM response")

	// Clean up JSON response
	cleaned := utils.CleanJSONResponse(rawResponse)

	// Parse plain text list
	lines := strings.Split(cleaned, "\n")
	resources := []Resource{}
	var current Resource
	fieldCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "Title:"):
			if fieldCount == 3 {
				current.ID = uuid.New().String()
				resources = append(resources, current)
				current = Resource{}
			}
			current.Title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
			fieldCount = 1
		case strings.HasPrefix(line, "Description:"):
			current.Description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
			fieldCount = 2
		case strings.HasPrefix(line, "Link:"):
			current.Link = strings.TrimSpace(strings.TrimPrefix(line, "Link:"))
			fieldCount = 3
		}
	}
	if fieldCount == 3 {
		current.ID = uuid.New().String()
		resources = append(resources, current)
	}

	// 🔍 YouTube search
	searchQuery := rawResponse
	youtubeVideos, err := videos.GlobalClient.SearchVideos(searchQuery, 3)
	if err == nil {
		log.Printf("[GenerateResourcesHandler] Found %d YouTube videos", len(youtubeVideos))
		for _, v := range youtubeVideos {
			resources = append(resources, Resource{
				Title:       v.Title,
				Description: "YouTube Video",
				Link:        v.EmbedURL,
				ID:          uuid.New().String(),
			})
		}
	} else {
		log.Printf("[GenerateResourcesHandler] YouTube search failed: %v", err)
	}

	final := ResourcesResponse{Resources: resources}
	finalBytes, err := json.Marshal(final)
	if err != nil {
		log.Printf("[GenerateResourcesHandler] Failed to serialize resources: %v", err)
		utils.WriteJSONError(w, "Failed to serialize resources", http.StatusInternalServerError)
		return
	}
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)
	log.Printf("[GenerateResourcesHandler] Successfully generated %d resources, cached response", len(resources))

	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
