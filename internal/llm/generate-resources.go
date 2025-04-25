package llm

import (
	"encoding/json"
	"log"
	"net/http"

	"agent-core/internal/prompt"
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

	// Create prompt input
	resourcePrompt, err := prompt.GetPrompt("resourceGenerationPrompt")
	if err != nil {
		log.Printf("Error getting course generation prompt: %v", err)
		http.Error(w, "Failed to get course generation prompt", http.StatusInternalServerError)
		return
	}
	promptMap := map[string]interface{}{
		"input": map[string]string{
			"prompt": promptWithVars(resourcePrompt, req.Title, req.Objective, req.EducationLevel, req.Difficulty),
		},
	}

	// Call LLM client
	resourcesResp, err := llmClient.Call("/generate-resources", promptMap)
	if err != nil {
		log.Printf("Error calling LLM client: %v", err)
		http.Error(w, "Failed to generate resources", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resourcesResp))
	prompt := map[string]interface{}{
		"prompt": promptWithVars(resourcePrompt, req.Title, req.Objective, req.EducationLevel, req.Difficulty),
	}

	// Call LLM client
	resourcesResp, err = llmClient.Call("generate-resources", prompt)
	if err != nil {
		log.Printf("Error calling LLM client: %v", err)
		http.Error(w, "Failed to generate resources", http.StatusInternalServerError)
		return
	}

	// Optional: Save to cache...
	// saveToCache(req, resourcesResp)

	// Return Bedrock response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resourcesResp))
}
