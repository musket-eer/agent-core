// Here we implement a Bedrock endpoint to generate exams
package llm

import (
	"encoding/json"
	"log"
	"net/http"

	"agent-core/internal/prompt"
	"agent-core/internal/embedding"
)


// Handler for /generate-exams endpoint
func GenerateExamsHandler(w http.ResponseWriter, r *http.Request) {
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

	// --- Supabase Cache Check ---
	promptStr, err := prompt.GetPrompt("examGenerationPrompt")
	if err != nil {
		http.Error(w, "Failed to load exam generation prompt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	combined := struct {
		Request CourseRequest
		Prompt  string
	}{
		Request: req,
		Prompt:  promptWithVars(promptStr, req.Title, req.Objective, req.EducationLevel, req.Difficulty),
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
	_, cachedResponse, hit, err = SupabaseClient.Check(key, embeddingVec) // use the global SupabaseClient
	if err != nil {
		log.Printf("Supabase cache check error: %v", err)
	}
	if hit {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResponse))
		return
	}

	// Create prompt input
	promptStr, err = prompt.GetPrompt("examGenerationPrompt")
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

	// --- Supabase Cache Store ---
	err = supabaseClient.CacheEmbedding(key, embeddingVec, examResp)
	if err != nil {
		log.Printf("Failed to cache embedding in Supabase: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(examResp))
}