package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"agent-core/internal/cache"
	"agent-core/internal/llm"
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
	log.Printf("[Startup] Deepseek BaseURL: '%s'", baseURL)
	log.Printf("[Startup] Deepseek API Key set: %v", apiKey != "")
	deepseekClient := llm.NewDeepseekClient(apiKey, baseURL)
	llm.LlmClientSet(deepseekClient)

	// Set up global Supabase client
	supabaseClient = cache.NewSupabasePgvectorClient()
	log.Printf("[Startup] Supabase API URL: '%s'", supabaseClient.ApiUrl)
	log.Printf("[Startup] Supabase Table: '%s'", supabaseClient.TableName)

	port := 8080
	mux := llm.RegisterHandlers()
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
