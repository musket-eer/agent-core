package llm

import (
	"encoding/json"
	"log"
	"net/http"

	"agent-core/internal/prompt"
	"agent-core/internal/llm/configs"
	"agent-core/internal/embedding"
	"agent-core/internal/cache"
)


// SupabaseClient is a global variable to hold the Supabase client instance
var supabaseClient = cache.NewSupabasePgvectorClient()

// Define the expected structure of input for course generation
type CourseRequest struct {
	Title       string `json:"title"`
	Objective   string `json:"objective"`
	Difficulty  string `json:"difficulty"`
	EducationLevel string `json:"education_level"`
}

// Handler for /generate-course endpoint
func GenerateCourseHandler(w http.ResponseWriter, r *http.Request) {
	embedding.PrintEmbeddingConfig()
	

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
	
	// --- Supabase Cache Check ---
	// Combine request and prompt for embedding
	cfg, ok := configs.GetConfig("/generate-course")
	if !ok {
		log.Printf("Error: no config found for /generate-course")
		http.Error(w, "No config for course generation", http.StatusInternalServerError)
		return
	}
	coursePrompt, err := prompt.GetPrompt("courseGenerationPrompt")
	if err != nil {
		log.Printf("Error getting course generation prompt: %v", err)
		http.Error(w, "Failed to get course generation prompt", http.StatusInternalServerError)
		return
	}
	promptStr := promptWithVars(coursePrompt, req.Title, req.Objective, req.EducationLevel, req.Difficulty)
	combined := struct {
		Request CourseRequest
		Prompt  string
	}{
		Request: req,
		Prompt:  promptStr,
	}
	combinedBytes, _ := json.Marshal(combined)
	embeddingVec, err := embedding.GetEmbedding(embedding.JSONStringWrapper(string(combinedBytes)))
	if err != nil {
		log.Printf("Embedding error: %v", err)
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}
	key := string(combinedBytes)
	var cachedResponse string
	var hit bool
	cachedResponse, _, hit, err = supabaseClient.Check(embeddingVec) // use the global supabaseClient
	if err != nil {
		log.Printf("Supabase cache check error: %v", err)
	}
	if hit {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResponse))
		return
	}

	// Use the global LLM client
	client := llmClient
	if client == nil {
		http.Error(w, "LLM client not initialized", http.StatusInternalServerError)
		return
	}

	// Get the course generation prompt, handle error if any
	coursePrompt, err = prompt.GetPrompt("courseGenerationPrompt")
	if err != nil {
		log.Printf("Error getting course generation prompt: %v", err)
		http.Error(w, "Failed to get course generation prompt", http.StatusInternalServerError)
		return
	}
	// Fetch model name from config
	cfg, ok = configs.GetConfig("/generate-course")
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

	// --- Supabase Cache Store ---
	err = supabaseClient.CacheEmbedding(key, embeddingVec, response)
	if err != nil {
		log.Printf("Failed to cache embedding in Supabase: %v", err)
	}


	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}
