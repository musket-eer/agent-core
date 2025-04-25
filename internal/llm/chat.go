// Here we implement a Bedrock endpoint for chat
package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)


// Handler for /chat endpoint
func ChatHandler(w http.ResponseWriter, r *http.Request) {
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
			"prompt": "Chat with the assistant about the course titled '" + req.Title + "' with objective: '" + req.Objective + "', suitable for " + req.EducationLevel + " level students and difficulty " + req.Difficulty + ".",
		},
	}

	// Call Bedrock
	chatResp, err := client.CallBedrock("chat", prompt)
	if err != nil {
		log.Printf("Error calling Bedrock: %v", err)
		http.Error(w, "Failed to generate chat response", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(chatResp))
}
