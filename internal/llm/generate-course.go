package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Define the expected structure of input for course generation
type CourseRequest struct {
	Title       string `json:"title"`
	Objective   string `json:"objective"`
	Difficulty  string `json:"difficulty"`
	EducationLevel string `json:"education_level"`
}

// Handler for /generate-course endpoint
func GenerateCourseHandler(w http.ResponseWriter, r *http.Request) {
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

	// TODO: Check Rate Limiting
	// TODO: RETRIEVE prompt BUFFER by accessing the RAM persisted promtps in configs/promts
	
	// TODO: Check cache here...
	// cachedResponse, found := checkCache(req)
	// if found {
	//     w.Header().Set("Content-Type", "application/json")
	//     w.Write(cachedResponse)
	//     return
	// }

	// Instantiate BedrockClient
	client := &BedrockClient{
		BaseURL:   "https://api.bedrock.aws.amazon.com/v1", // Use actual Bedrock endpoint
		APIKey:    "your-api-key", // Secure via env vars or secret management
		Model:     "claude-v2",
		HttpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Create prompt input
	prompt := map[string]interface{}{
		"model": client.Model,
		"input": map[string]string{
			"prompt": "Generate a course outline for a course titled '" + req.Title + "' with objective: '" + req.Objective + "', suitable for " + req.EducationLevel + " level students and difficulty " + req.Difficulty + ".",
		},
	}

	// Call Bedrock
	bedrockResp, err := client.CallBedrock("generate-course", prompt)
	if err != nil {
		log.Printf("Error calling Bedrock: %v", err)
		http.Error(w, "Failed to generate course", http.StatusInternalServerError)
		return
	}

	// Optional: Save to cache...
	// saveToCache(req, bedrockResp)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(bedrockResp))
}
