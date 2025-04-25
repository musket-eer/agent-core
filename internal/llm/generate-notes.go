// Here we implement a Bedrock endpoint to generate notes
package llm

import (
	"encoding/json"
	"agent-core/internal/cache/supabaseclient"
	"log"
	"net/http"

	"agent-core/internal/prompt"
	"agent-core/internal/embedding"
)


// SupabaseClient is a global variable to hold the Supabase client instance
var SupabaseClient *supabaseclient.SupabasePgvectorClient

// Handler for /generate-notes endpoint
func GenerateNotesHandler(w http.ResponseWriter, r *http.Request) {
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
	notesPrompt, err := prompt.GetPrompt("professorNotesPrompt")
	if err != nil {
		log.Printf("Error getting notes generation prompt: %v", err)
		http.Error(w, "Failed to get notes generation prompt", http.StatusInternalServerError)
		return
	}
	combined := struct {
		Request CourseRequest
		Prompt  string
	}{
		Request: req,
		Prompt:  promptWithVars(notesPrompt, req.Title, req.Objective, req.EducationLevel, req.Difficulty),
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
	notesPrompt, err = prompt.GetPrompt("professorNotesPrompt")
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

	// --- Supabase Cache Store ---
	err = supabaseClient.CacheEmbedding(key, embeddingVec, notesResp)
	if err != nil {
		log.Printf("Failed to cache embedding in Supabase: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(notesResp))
}
