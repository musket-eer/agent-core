// Here we implement a Bedrock endpoint to generate notes
package llm

import (
	"encoding/json"
	"log"
	"net/http"

	"agent-core/internal/prompt"
)


// Handler for /generate-notes endpoint
func GenerateNotesHandler(w http.ResponseWriter, r *http.Request) {
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
	notesPrompt, err := prompt.GetPrompt("professorNotesPrompt")
	if err != nil {
		log.Printf("Error getting notes generation prompt: %v", err)
		http.Error(w, "Failed to get notes generation prompt", http.StatusInternalServerError)
		return
	}
	promptMap := map[string]interface{}{
		"input": map[string]string{
			"prompt": promptWithVars(notesPrompt, req.Title, req.Objective, req.EducationLevel, req.Difficulty),
		},
	}

	// Call LLM client
	notesResp, err := llmClient.Call("/generate-notes", promptMap)
	if err != nil {
		log.Printf("Error calling LLM client: %v", err)
		http.Error(w, "Failed to generate notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(notesResp))
}
