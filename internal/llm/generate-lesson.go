package llm

import (
	"encoding/json"
	"log"
	"net/http"

	"agent-core/internal/prompt"
)

// LessonRequestExpectation defines the expected input structure for lesson generation
type LessonRequestExpectation struct {
	Title          string `json:"title"`
	Objective      string `json:"objective"`
	Difficulty     string `json:"difficulty"`
	EducationLevel string `json:"education_level"`
}

// CourseRequest is defined in generate-course.go and reused here.

// Handler for /generate-lesson endpoint
func GenerateLessonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body
	var req CourseRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Check cache here...
	// cachedResponse, found := checkCache(req)
	// if found {
	//     w.Header().Set("Content-Type", "application/json")
	//     w.Write(cachedResponse)
	//     return
	// }

	// Create prompt input
	promptStr, err := prompt.GetPrompt("lessonSequencePrompt")
	if err != nil {
		http.Error(w, "Failed to get prompt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	promptMap := map[string]interface{}{
		"input": map[string]string{
			"prompt": promptWithVars(promptStr, req.Title, req.Objective, req.EducationLevel, req.Difficulty),
		},
	}

	// Call LLM client
	lessonResp, err := llmClient.Call("/generate-lesson", promptMap)
	if err != nil {
		log.Printf("Error calling LLM client: %v", err)
		http.Error(w, "Failed to generate lesson", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(lessonResp))
}
