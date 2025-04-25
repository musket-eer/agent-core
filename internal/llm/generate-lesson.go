package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
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

	// Optional: Check cache here...
	// cachedResponse, found := checkCache(req)
	// if found {
	//     w.Header().Set("Content-Type", "application/json")
	//     w.Write(cachedResponse)
	//     return
	// }

	// Instantiate BedrockClient
	client := &BedrockClient{
		BaseURL:    "https://api.bedrock.aws.amazon.com/v1",
		APIKey:     "your-api-key", // Secure this
		Model:      "bedrock-lesson-model",
		HttpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Create prompt input
	prompt := map[string]interface{}{
		"model": client.Model,
		"input": map[string]string{
			"prompt": "Generate a lesson plan for a course titled '" + req.Title + 
				"' with objective: '" + req.Objective + 
				"', suitable for " + req.EducationLevel + 
				" level students and difficulty " + req.Difficulty + ".",
		},
	}

	// Call Bedrock
	lessonResp, err := client.CallBedrock("generate-lesson", prompt)
	if err != nil {
		log.Printf("Error calling Bedrock: %v", err)
		http.Error(w, "Failed to generate lesson", http.StatusInternalServerError)
		return
	}

	// Optional: Save to cache...
	// saveToCache(req, lessonResp)

	// Return the Bedrock response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(lessonResp))
}
