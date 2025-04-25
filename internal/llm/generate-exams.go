// Here we implement a Bedrock endpoint to generate exams
package llm

import (
	"encoding/json"
	"log"
	"net/http"

	"agent-core/internal/prompt"
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

	// Create prompt input
	promptStr, err := prompt.GetPrompt("examGenerationPrompt")
	if err != nil {
		http.Error(w, "Failed to load exam generation prompt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	prompt := map[string]interface{}{
		"input": map[string]string{
			"prompt": promptWithVars(promptStr, req.Title, req.Objective, req.EducationLevel, req.Difficulty),
		},
	}

	// Call LLM client
	examResp, err := llmClient.Call("/generate-exams", prompt)
	if err != nil {
		log.Printf("Error calling LLM client: %v", err)
		http.Error(w, "Failed to generate exams", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(examResp))
}