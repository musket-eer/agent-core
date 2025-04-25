// Here we implement a Bedrock endpoint to generate exams
package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Handler for /generate-exams endpoint
func GenerateExamsHandler(w http.ResponseWriter, r *http.Request) {
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
		BaseURL:   "https://api.bedrock.aws.amazon.com/v1",
		APIKey:    "your-api-key", // Secure via env vars or secret management
		Model:     "bedrock-exams-model",
		HttpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Create prompt input
	prompt := map[string]interface{}{
		"model": client.Model,
		"input": map[string]string{
			"prompt": "Generate exam questions for a course titled '" + req.Title + "' with objective: '" + req.Objective + "', suitable for " + req.EducationLevel + " level students and difficulty " + req.Difficulty + ".",
		},
	}

	// Call Bedrock
	examResp, err := client.CallBedrock("generate-exams", prompt)
	if err != nil {
		log.Printf("Error calling Bedrock: %v", err)
		http.Error(w, "Failed to generate exams", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(examResp))
}
