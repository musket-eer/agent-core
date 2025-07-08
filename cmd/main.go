package main

import (
	"fmt"
	"net/http"
	"log"
	"os"

	"agent-core/internal/cache"
	"agent-core/internal/llm"
	"agent-core/internal/llm/clients"
	"agent-core/internal/videos"

	"github.com/joho/godotenv"
)

var upstashClient *cache.UpstashClient

func main() {
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Println("⚠️ .env file not loaded")
	}

	// Initialize Upstash Redis client from environment variables
	upstashRestUrl := os.Getenv("UPSTASH_REDIS_REST_URL")
	upstashRestToken := os.Getenv("UPSTASH_REDIS_REST_TOKEN")
	upstashClient = cache.NewUpstashClient(upstashRestUrl, upstashRestToken)
	cache.Upstash = upstashClient

	upstashCache := cache.NewCache(upstashClient)

	// ✅ Initialize global cache after clearing
	cache.InitGlobal(upstashCache)

	// Set up Deepseek clients
	baseURL := os.Getenv("DEEPSEEK_GENERAL_URL")
	generalKey := os.Getenv("DEEPSEEK_GENERAL_API_KEY")
	chatKey := os.Getenv("DEEPSEEK_CHAT_API_KEY")
	youtubeAPIKey := os.Getenv("YOUTUBE_API_KEY")
	youtubeBaseURL := os.Getenv("YOUTUBE_URL")

	videos.GlobalClient = videos.NewYouTubeClient(youtubeAPIKey, youtubeBaseURL)

	log.Printf("🔐 UPSTASH REST TOKEN: %s", upstashRestToken)

	// Create Deepseek clients
	generalClient := clients.NewDeepseekClient(generalKey, baseURL)
	chatClient := clients.NewDeepseekClient(chatKey, baseURL)

	// Create Gemini clients
	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	geminiBaseURL := os.Getenv("GEMINI_BASE_URL")
	log.Printf("🔐 GEMINI API KEY: %s", geminiApiKey)
	geminiGeneralClient := clients.NewGeminiClient(geminiApiKey, geminiBaseURL)
	geminiChatClient := clients.NewGeminiClient(geminiApiKey, geminiBaseURL)

	// Register all LLM clients globally
	llm.RegisterLLMClient("general", generalClient)
	llm.RegisterLLMClient("chat", chatClient)
	llm.RegisterLLMClient("gemini-general", geminiGeneralClient)
	llm.RegisterLLMClient("gemini-chat", geminiChatClient)

	// Start HTTP server
	port := 8080
	mux := llm.RegisterHandlers()
	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("🚀 Server running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("❌ Server failed: %v\n", err)
	}
}
