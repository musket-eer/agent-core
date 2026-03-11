package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

	youtubeAPIKey := os.Getenv("YOUTUBE_API_KEY")
	youtubeBaseURL := os.Getenv("YOUTUBE_URL")

	videos.GlobalClient = videos.NewYouTubeClient(youtubeAPIKey, youtubeBaseURL)

	// Register all available LLM clients from environment.
	activeLLMAliases := registerLLMClients()
	configureLLMQueueWorkers(activeLLMAliases)

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

func registerLLMClients() []string {
	registered := 0
	aliases := make([]string, 0, 8)
	register := func(alias string, client llm.LLMClient) {
		llm.RegisterLLMClient(alias, client)
		registered++
		aliases = append(aliases, alias)
		log.Printf("✅ Registered LLM client alias: %s", alias)
	}

	deepseekBaseURL := firstNonEmpty(
		os.Getenv("DEEPSEEK_BASE_URL"),
		os.Getenv("DEEPSEEK_GENERAL_URL"),
	)
	deepseekGeneralKey := os.Getenv("DEEPSEEK_GENERAL_API_KEY")
	deepseekChatKey := os.Getenv("DEEPSEEK_CHAT_API_KEY")
	deepseekChatURL := firstNonEmpty(os.Getenv("DEEPSEEK_CHAT_URL"), deepseekBaseURL)

	// Backward-compatible aliases used by existing endpoint configs.
	if strings.TrimSpace(deepseekGeneralKey) != "" && strings.TrimSpace(deepseekBaseURL) != "" {
		client := clients.NewDeepseekClient(deepseekGeneralKey, deepseekBaseURL)
		register("general", client)
		register("deepseek-general", client)
	}
	if strings.TrimSpace(deepseekChatKey) != "" && strings.TrimSpace(deepseekChatURL) != "" {
		client := clients.NewDeepseekClient(deepseekChatKey, deepseekChatURL)
		register("chat", client)
		register("deepseek-chat", client)
	}

	// Optional extra Deepseek/OpenAI-compatible aliases.
	registerDeepseekAlias := func(alias, keyEnv, urlEnv string) {
		key := os.Getenv(keyEnv)
		url := firstNonEmpty(os.Getenv(urlEnv), deepseekBaseURL)
		if strings.TrimSpace(key) == "" || strings.TrimSpace(url) == "" {
			return
		}
		register(alias, clients.NewDeepseekClient(key, url))
	}
	registerDeepseekAlias("deepseek-reasoner", "DEEPSEEK_REASONER_API_KEY", "DEEPSEEK_REASONER_URL")
	registerDeepseekAlias("deepseek-fast", "DEEPSEEK_FAST_API_KEY", "DEEPSEEK_FAST_URL")
	registerDeepseekAlias("openai-general", "OPENAI_API_KEY", "OPENAI_BASE_URL")
	registerDeepseekAlias("openai-chat", "OPENAI_CHAT_API_KEY", "OPENAI_CHAT_BASE_URL")

	// Gemini aliases are intentionally disabled for now.

	// Bedrock aliases.
	awsRegion := firstNonEmpty(os.Getenv("AWS_REGION"), os.Getenv("AWS_DEFAULT_REGION"))
	if strings.TrimSpace(awsRegion) != "" {
		bedrockClient, err := clients.NewBedrockClient(awsRegion)
		if err != nil {
			log.Printf("⚠️ Skipping bedrock clients: %v", err)
		} else {
			register("bedrock-general", bedrockClient)
			register("bedrock-chat", bedrockClient)
		}
	}

	log.Printf("✅ LLM client registry ready: total aliases=%d", registered)
	log.Printf("ℹ️ Available aliases: %v", llm.ListLLMClients())
	return aliases
}

func configureLLMQueueWorkers(activeAliases []string) {
	// Respect explicit override.
	if strings.TrimSpace(os.Getenv("LLM_QUEUE_WORKERS")) != "" {
		log.Printf("✅ LLM queue workers configured from env: %s", os.Getenv("LLM_QUEUE_WORKERS"))
		return
	}

	// We treat these as distinct worker lanes by key/provider for your setup.
	workerLanes := 0
	if contains(activeAliases, "deepseek-general") || contains(activeAliases, "general") {
		workerLanes++
	}
	if contains(activeAliases, "deepseek-chat") || contains(activeAliases, "chat") {
		workerLanes++
	}
	if workerLanes < 1 {
		workerLanes = 1
	}
	_ = os.Setenv("LLM_QUEUE_WORKERS", strconv.Itoa(workerLanes))
	log.Printf("✅ LLM queue workers auto-configured: %d", workerLanes)
	log.Printf("ℹ️ Worker lanes active: deepseek-general=%t deepseek-chat=%t",
		contains(activeAliases, "deepseek-general") || contains(activeAliases, "general"),
		contains(activeAliases, "deepseek-chat") || contains(activeAliases, "chat"),
	)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func contains(values []string, needle string) bool {
	for _, v := range values {
		if v == needle {
			return true
		}
	}
	return false
}
