package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"agent-core/internal/cache"
	"agent-core/internal/llm"
	"agent-core/internal/llm/clients"
	"agent-core/internal/llm/v2/middleware"
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

	// ============================================================================
	// Initialize V2 Feature Flags
	// ============================================================================
	log.Println("🚀 Initializing V2 feature flags...")
	featureFlags := middleware.InitFeatureFlags()

	// Read V2 settings from environment
	v2Enabled := os.Getenv("V2_ENABLED") == "true"
	v2RolloutPercentage := 0.0
	if rolloutStr := os.Getenv("V2_ROLLOUT_PERCENTAGE"); rolloutStr != "" {
		if val, err := strconv.ParseFloat(rolloutStr, 64); err == nil {
			v2RolloutPercentage = val
		}
	}

	// Configure feature flags
	featureFlags.SetV2Enabled(v2Enabled)
	featureFlags.SetV2RolloutPercentage(v2RolloutPercentage)

	// Enable specific agents if configured
	if enabledAgents := os.Getenv("V2_ENABLED_AGENTS"); enabledAgents != "" {
		// Parse comma-separated list
		// For now, start with V2 disabled by default
		log.Printf("✅ V2 Feature Flags: enabled=%v, rollout=%.1f%%", v2Enabled, v2RolloutPercentage)
	} else {
		log.Println("✅ V2 Feature Flags initialized (disabled by default)")
	}

	// Start HTTP server
	port := 8080
	mux := llm.RegisterHandlers()
	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("🚀 Server running at http://localhost%s\n", addr)
	fmt.Printf("📊 V2 Status: enabled=%v, rollout=%.1f%%\n", v2Enabled, v2RolloutPercentage)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("❌ Server failed: %v\n", err)
	}
}
