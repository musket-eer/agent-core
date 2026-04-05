package main

import (
	"crypto/sha256"
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
	bedrockOnly := strings.EqualFold(strings.TrimSpace(os.Getenv("BEDROCK_ONLY")), "true")
	register := func(alias string, client llm.LLMClient) {
		llm.RegisterLLMClient(alias, client)
		registered++
		aliases = append(aliases, alias)
		log.Printf("✅ Registered LLM client alias: %s", alias)
	}
	logClientSource := func(alias, url, key string) {
		log.Printf("ℹ️ LLM client source alias=%s url=%s key=%s", alias, maskURL(url), maskSecret(key))
	}

	// Bedrock aliases.
	awsRegion := firstNonEmpty(os.Getenv("AWS_REGION"), os.Getenv("AWS_DEFAULT_REGION"))
	if strings.TrimSpace(awsRegion) != "" {
		bedrockClient, err := clients.NewBedrockClient(awsRegion)
		if err != nil {
			log.Printf("⚠️ Skipping bedrock clients: %v", err)
		} else {
			if bedrockOnly {
				log.Printf("⚠️ BEDROCK_ONLY=true: registering Bedrock for all default aliases")
				register("general", bedrockClient)
				logClientSource("general", awsRegion, "aws-config")
				register("chat", bedrockClient)
				logClientSource("chat", awsRegion, "aws-config")
				register("deepseek-general", bedrockClient)
				logClientSource("deepseek-general", awsRegion, "aws-config")
				register("deepseek-chat", bedrockClient)
				logClientSource("deepseek-chat", awsRegion, "aws-config")
			}
			register("bedrock-general", bedrockClient)
			logClientSource("bedrock-general", awsRegion, "aws-config")
			register("bedrock-chat", bedrockClient)
			logClientSource("bedrock-chat", awsRegion, "aws-config")
		}
	}

	if !bedrockOnly {
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
			logClientSource("general", deepseekBaseURL, deepseekGeneralKey)
			register("deepseek-general", client)
			logClientSource("deepseek-general", deepseekBaseURL, deepseekGeneralKey)
		}
		if strings.TrimSpace(deepseekChatKey) != "" && strings.TrimSpace(deepseekChatURL) != "" {
			client := clients.NewDeepseekClient(deepseekChatKey, deepseekChatURL)
			register("chat", client)
			logClientSource("chat", deepseekChatURL, deepseekChatKey)
			register("deepseek-chat", client)
			logClientSource("deepseek-chat", deepseekChatURL, deepseekChatKey)
		}

		// Optional extra Deepseek/OpenAI-compatible aliases.
		registerDeepseekAlias := func(alias, key, url string) {
			if strings.TrimSpace(key) == "" || strings.TrimSpace(url) == "" {
				return
			}
			register(alias, clients.NewDeepseekClient(key, url))
			logClientSource(alias, url, key)
		}
		registerDeepseekAlias("deepseek-reasoner", os.Getenv("DEEPSEEK_REASONER_API_KEY"), firstNonEmpty(os.Getenv("DEEPSEEK_REASONER_URL"), deepseekBaseURL))
		registerDeepseekAlias("deepseek-fast", os.Getenv("DEEPSEEK_FAST_API_KEY"), firstNonEmpty(os.Getenv("DEEPSEEK_FAST_URL"), deepseekBaseURL))
		registerDeepseekAlias("openai-general", firstNonEmptyEnv("OPENAI_GENERAL_KEY", "OPENAI_API_KEY"), firstNonEmptyEnv("OPENAI_GENERAL_URL", "OPENAI_BASE_URL"))
		registerDeepseekAlias("openai-chat", firstNonEmptyEnv("OPENAI_CHAT_KEY", "OPENAI_CHAT_API_KEY"), firstNonEmptyEnv("OPENAI_CHAT_URL", "OPENAI_CHAT_BASE_URL"))
		registerDeepseekAlias("openai-exams", firstNonEmptyEnv("OPENAI_EXAMS_KEY", "OPENAI_GENERAL_KEY", "OPENAI_API_KEY"), firstNonEmptyEnv("OPENAI_EXAMS_URL", "OPENAI_GENERAL_URL", "OPENAI_BASE_URL"))
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

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
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

func maskSecret(secret string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "<empty>"
	}
	sum := sha256.Sum256([]byte(secret))
	short := fmt.Sprintf("%x", sum[:6])
	if len(secret) <= 8 {
		return fmt.Sprintf("<len=%d sha256=%s>", len(secret), short)
	}
	return fmt.Sprintf("%s...%s <len=%d sha256=%s>", secret[:4], secret[len(secret)-4:], len(secret), short)
}

func maskURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "<empty>"
	}
	return raw
}
