package main

import (
	"fmt"
	"net/http"
	"os"
	"agent-core/internal/cache"
	"agent-core/internal/llm"
	"agent-core/internal/llm/clients"

	"github.com/joho/godotenv"
	// "agent-core/internal/cache/supabaseclient"

)
var upstashClient *cache.UpstashClient

func main() {
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		// .env file not loaded, but continue
	}

	// Initialize Upstash Redis client from environment variables
	upstashRestUrl := os.Getenv("UPSTASH_REDIS_REST_URL")
	upstashRestToken := os.Getenv("UPSTASH_REDIS_REST_TOKEN")
	upstashClient = cache.NewUpstashClient(upstashRestUrl, upstashRestToken)
	cache.Upstash = upstashClient
	upstashCache := cache.NewCache(upstashClient)
	cache.InitGlobal(upstashCache)

	// Set up Deepseek client from environment variables
	baseURL := os.Getenv("DEEPSEEK_GENERAL_URL")
	generalKey := os.Getenv("DEEPSEEK_GENERAL_API_KEY")
	chatKey := os.Getenv("DEEPSEEK_CHAT_API_KEY")




	// Register HTTP handlers

// Create clients

	generalClient := clients.NewDeepseekClient(generalKey, baseURL)
	chatClient := clients.NewDeepseekClient(chatKey, baseURL)


	// Register both clients globally
	llm.RegisterLLMClient("general", generalClient)
	llm.RegisterLLMClient("chat", chatClient)
	
	port := 8080
	mux := llm.RegisterHandlers()
	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("Server running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
