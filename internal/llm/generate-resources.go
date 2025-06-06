package llm

import (
	"encoding/json"
	"net/http"
	"strings"
	"io"
	"agent-core/internal/llm/configs"
	"agent-core/internal/prompt"
	"agent-core/internal/llm/utils"
	"agent-core/internal/cache"
	"agent-core/internal/videos"

	"github.com/google/uuid"
)


type ResourcesPayLoad struct {
	// course parameters
	courseTitle       string   `json:"courseTitle"`
	CourseDescription string   `json:"courseDescription"`
	Difficulty        string    `json:"difficulty"`

	// lesson parameters
	LessonTitle       string   `json:"lessonTitle"`
	LessonObjective   string   `json:"lessonObjective"`
	LessonAssessment string    `json:"lessonAssessment"`

	// profile parameters
	EducationLevel    string   `json:"educationLevel"`
	Interests         []string `json:"interests"`
	CareerGoals       string   `json:"careerGoals"`
}
// Resource object
type Resource struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	ID          string `json:"id"`
}

// ResourcesResponse wraps the final return
type ResourcesResponse struct {
	Resources []Resource `json:"resources"`
}

func GenerateResourcesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resourcesReq ResourcesPayLoad
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // limit to 1MB
	rawBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(rawBodyBytes, &resourcesReq); err != nil {
		utils.WriteJSONError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	userMessageBytes, err := json.Marshal(resourcesReq)
	if err != nil {
		utils.WriteJSONError(w, "Failed to serialize user message", http.StatusInternalServerError)
		return
	}
	payloadStr := string(userMessageBytes)
	if resp, hit, _ := cache.Get().CheckByPayload(payloadStr); hit {
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
		return
	}

	cfg, ok := configs.GetDeepseekConfig("/generate-resources")
	if !ok {
		utils.WriteJSONError(w, "No config for resources generation", http.StatusInternalServerError)
		return
	}

	systemPrompt, err := prompt.GetPrompt("librarianResourcesPrompt")
	if err != nil {
		utils.WriteJSONError(w, "Failed to load resources system prompt", http.StatusInternalServerError)
		return
	}

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
		utils.WriteJSONError(w, "failed to get client "+clientName+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	rawResponse, err := client.Call("/generate-resources", payload)
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate search query", http.StatusInternalServerError)
		return
	}

	// Clean up Markdown fences
	cleaned := strings.TrimSpace(rawResponse)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

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
		for _, v := range youtubeVideos {
			resources = append(resources, Resource{
				Title:       v.Title,
				Description: "YouTube Video",
				Link:        v.EmbedURL,
				ID:          uuid.New().String(),
			})
		}
	}

	final := ResourcesResponse{Resources: resources}
	finalBytes, err := json.Marshal(final)
	if err != nil {
		utils.WriteJSONError(w, "Failed to serialize resources", http.StatusInternalServerError)
		return
	}
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)

	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}