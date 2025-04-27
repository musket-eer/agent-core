package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"agent-core/internal/prompt"
	"agent-core/internal/cache/supabaseclient"
	"agent-core/internal/embedding"
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
	resourcePrompt, err := prompt.GetPrompt("librarianResourcesPrompt")
	if err != nil {
		log.Printf("Error getting course generation prompt: %v", err)
		http.Error(w, "Failed to get course generation prompt", http.StatusInternalServerError)
		return
	}
	combined := struct {
		Request CourseRequest
		Prompt  string
	}{
		Request: req,
		Prompt:  promptWithVars(resourcePrompt, req.Title, req.Objective, req.EducationLevel, req.Difficulty),
	}
	combinedBytes, _ := json.Marshal(combined)
	embeddingVec, err := embedding.GetEmbedding(embedding.JSONStringWrapper(string(combinedBytes)))
	if err != nil {
		log.Printf("Embedding error: %v", err)
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}
	key := string(combinedBytes)
	supabaseClient := &supabaseclient.SupabasePgvectorClient{
		ApiUrl:    os.Getenv("SUPABASE_API_URL"),
		ApiKey:    os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		TableName: os.Getenv("SUPABASE_TABLE_NAME"),
	}
	var cachedResponse string
	var hit bool
	_, cachedResponse, hit, err = supabaseClient.Check(key, embeddingVec)
	if err != nil {
		log.Printf("Supabase cache check error: %v", err)
	}
	if hit {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResponse))
		return
	}

	// Create prompt input
	resourcePrompt, err = prompt.GetPrompt("librarianResourcesPrompt")
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
	// --- Supabase Cache Store ---
	err = supabaseClient.CacheEmbedding(key, embeddingVec, resourcesResp)
	if err != nil {
		log.Printf("Failed to cache embedding in Supabase: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resourcesResp))
}
