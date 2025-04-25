package llm

import (
	"encoding/json"
	"log"
	"net/http"	
	"agent-core/internal/prompt"
	"agent-core/internal/llm/configs"
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

	// Use the global LLM client
	client := llmClient
	if client == nil {
		http.Error(w, "LLM client not initialized", http.StatusInternalServerError)
		return
	}

	// Get the course generation prompt, handle error if any
	coursePrompt, err := prompt.GetPrompt("courseGenerationPrompt")
	if err != nil {
		log.Printf("Error getting course generation prompt: %v", err)
		http.Error(w, "Failed to get course generation prompt", http.StatusInternalServerError)
		return
	}
	// Fetch model name from config
	cfg, ok := configs.GetConfig("/generate-course")
	if !ok {
		log.Printf("Error: no config found for /generate-course")
		http.Error(w, "No config for course generation", http.StatusInternalServerError)
		return
	}
	prompt := map[string]interface{}{
		"model": cfg.ModelName,
		"input": map[string]string{
			"prompt": promptWithVars(coursePrompt, req.Title, req.Objective, req.EducationLevel, req.Difficulty),
		},
	}
	// call cache for embeddings match.
	// cachedResponse, found := checkCache(prompt)
	// if found {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Write(cachedResponse)
	// 	return
	// }

	// Call LLM client
	response, err := client.Call("/generate-course", prompt)
	if err != nil {
		log.Printf("Error calling LLM client: %v", err)
		http.Error(w, "Failed to generate course", http.StatusInternalServerError)
		return
	}

	// Optional: Save to cache...
	// saveToCache(req, bedrockResp)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}
