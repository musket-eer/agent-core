package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// CourseRequest defines the input structure used across all generation endpoints
type LibraryRequestExpectation struct {
	Title          string `json:"title"`
	Objective      string `json:"objective"`
	Difficulty     string `json:"difficulty"`
	EducationLevel string `json:"education_level"`
}

// Handler for /generate-resources endpoint
func GenerateResourcesHandler(w http.ResponseWriter, r *http.Request) {
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
		APIKey:     "your-api-key", // Load from env/secret manager
		Model:      "bedrock-resources-model",
		HttpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Create prompt input
	prompt := map[string]interface{}{
		"model": client.Model,
		"input": map[string]string{
			"prompt": "Generate supplementary resources for a course titled '" + req.Title +
				"' with objective: '" + req.Objective +
				"', suitable for " + req.EducationLevel +
				" level students and difficulty " + req.Difficulty + ".",
		},
	}

	// Call Bedrock
	resourcesResp, err := client.CallBedrock("generate-resources", prompt)
	if err != nil {
		log.Printf("Error calling Bedrock: %v", err)
		http.Error(w, "Failed to generate resources", http.StatusInternalServerError)
		return
	}

	// Optional: Save to cache...
	// saveToCache(req, resourcesResp)

	// Return Bedrock response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resourcesResp))
}
