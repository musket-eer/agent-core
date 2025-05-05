package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"agent-core/internal/cache"
	"agent-core/internal/llm"
	"agent-core/internal/embedding"
)

var supabaseClient *cache.SupabasePgvectorClient

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not loaded: %v", err)
	}
	// Set up Deepseek client from environment variables
	baseURL := os.Getenv("DEEPSEEK_URL")
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	deepseekClient := llm.NewDeepseekClient(apiKey, baseURL)
	llm.LlmClientSet(deepseekClient)

	// Set up global Supabase client
	supabaseClient = cache.NewSupabasePgvectorClient()

	// setup embedding dimension directly from environment
	dimStr := os.Getenv("EMBEDDING_DIMENSION")
	if dimStr != "" {
		var d int
		_, err := fmt.Sscanf(dimStr, "%d", &d)
		if err == nil && d > 0 {
			embedding.GlobalEmbeddingConfig.Dimension = d
			log.Printf("[Startup] Set embedding dimension: %d", d)
		} else {
			log.Printf("[Startup] Invalid EMBEDDING_DIMENSION value: %s", dimStr)
		}
	} else {
		log.Printf("[Startup] EMBEDDING_DIMENSION not set, using default: %d", embedding.GlobalEmbeddingConfig.Dimension)
	}

	port := 8080
	mux := llm.RegisterHandlers()
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
