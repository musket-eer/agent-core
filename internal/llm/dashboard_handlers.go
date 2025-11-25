package llm

import (
	"encoding/json"
	"log"
	"net/http"

	"agent-core/internal/cache"
	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"
)

// DashboardInsightsHandler - POST /dashboard/insights
func DashboardInsightsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DashboardInsightsHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		log.Printf("[DashboardInsightsHandler] Invalid method: %s", r.Method)
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DashboardInsightsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[DashboardInsightsHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("[DashboardInsightsHandler] Parsed request")

	// Prepare cache key
	cachePayload, _ := json.Marshal(req)
	payloadStr := string(cachePayload)
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
		log.Printf("[DashboardInsightsHandler] Cache hit")
		w.Header().Set("Content-Type", "application/json")
		w.Write(cached)
		return
	}
	log.Printf("[DashboardInsightsHandler] Cache miss, calling LLM")

	systemPrompt, err := prompt.GetPrompt("dashboardInsightsPrompt")
	if err != nil {
		log.Printf("[DashboardInsightsHandler] Failed to load prompt: %v", err)
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}
	log.Printf("[DashboardInsightsHandler] Loaded system prompt")

	userMessage, _ := json.Marshal(req)
	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"input": map[string]string{
			"prompt": systemPrompt + "\n" + string(userMessage),
		},
	}

	endpoint := "/dashboard/insights"
	config, ok := configs.GetDeepseekConfig(endpoint)
	if !ok {
		config, _ = configs.GetDeepseekConfig("/chat")
	}

	clientName := config.ClientName
	if clientName == "" {
		clientName = "general"
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		log.Printf("[DashboardInsightsHandler] Failed to get client %s: %v", clientName, err)
		utils.WriteJSONError(w, "failed to get client "+clientName+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[DashboardInsightsHandler] Calling LLM client: %s", clientName)

	rawResponse, err := client.Call(endpoint, payload)
	if err != nil {
		log.Printf("[DashboardInsightsHandler] LLM call failed: %v", err)
		utils.WriteJSONError(w, "Failed to generate insights", http.StatusInternalServerError)
		return
	}
	log.Printf("[DashboardInsightsHandler] Received LLM response")

	// Clean up JSON response
	cleaned := utils.CleanJSONResponse(rawResponse)

	var response DashboardInsightsResponse
	if err := json.Unmarshal([]byte(cleaned), &response); err != nil {
		// If JSON parsing fails, create a structured response
		response = DashboardInsightsResponse{
			Insights: InsightsSummary{
				Summary:         rawResponse,
				Strengths:       []string{},
				Recommendations: []string{},
				NextSteps:       []string{},
			},
			MotivationalMessage: "Keep up the great work!",
			ProgressHighlights:  []ProgressHighlight{},
		}
	}

	finalBytes, _ := json.Marshal(response)
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)
	log.Printf("[DashboardInsightsHandler] Successfully generated insights, cached response")
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}

// DashboardRecommendationsHandler - POST /dashboard/recommendations
func DashboardRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DashboardRecommendationsHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		log.Printf("[DashboardRecommendationsHandler] Invalid method: %s", r.Method)
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DashboardRecommendationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cachePayload, _ := json.Marshal(req)
	payloadStr := string(cachePayload)
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cached)
		return
	}

	systemPrompt, err := prompt.GetPrompt("dashboardRecommendationsPrompt")
	if err != nil {
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}

	userMessage, _ := json.Marshal(req)
	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"input": map[string]string{
			"prompt": systemPrompt + "\n" + string(userMessage),
		},
	}

	endpoint := "/dashboard/recommendations"
	config, ok := configs.GetDeepseekConfig(endpoint)
	if !ok {
		config, _ = configs.GetDeepseekConfig("/chat")
	}

	clientName := config.ClientName
	if clientName == "" {
		clientName = "general"
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		utils.WriteJSONError(w, "failed to get client", http.StatusInternalServerError)
		return
	}

	rawResponse, err := client.Call(endpoint, payload)
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate recommendations", http.StatusInternalServerError)
		return
	}

	var response DashboardRecommendationsResponse
	if err := json.Unmarshal([]byte(rawResponse), &response); err != nil {
		response = DashboardRecommendationsResponse{
			Recommendations: []CourseRecommendation{},
		}
	}

	finalBytes, _ := json.Marshal(response)
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}

// DashboardAnalyzeProgressHandler - POST /dashboard/analyze-progress
func DashboardAnalyzeProgressHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DashboardAnalyzeProgressHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		log.Printf("[DashboardAnalyzeProgressHandler] Invalid method: %s", r.Method)
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DashboardAnalyzeProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	cachePayload, _ := json.Marshal(req)
	payloadStr := string(cachePayload)
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cached)
		return
	}

	systemPrompt, err := prompt.GetPrompt("dashboardAnalyzeProgressPrompt")
	if err != nil {
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}

	userMessage, _ := json.Marshal(req)
	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"input": map[string]string{
			"prompt": systemPrompt + "\n" + string(userMessage),
		},
	}

	endpoint := "/dashboard/analyze-progress"
	config, ok := configs.GetDeepseekConfig(endpoint)
	if !ok {
		config, _ = configs.GetDeepseekConfig("/chat")
	}

	clientName := config.ClientName
	if clientName == "" {
		clientName = "general"
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		utils.WriteJSONError(w, "failed to get client", http.StatusInternalServerError)
		return
	}

	rawResponse, err := client.Call(endpoint, payload)
	if err != nil {
		utils.WriteJSONError(w, "Failed to analyze progress", http.StatusInternalServerError)
		return
	}

	var response DashboardAnalyzeProgressResponse
	if err := json.Unmarshal([]byte(rawResponse), &response); err != nil {
		response = DashboardAnalyzeProgressResponse{
			Analysis: ProgressAnalysis{
				Summary:             rawResponse,
				Trends:              []Trend{},
				Achievements:        []string{},
				AreasForImprovement: []string{},
				Recommendations:     []string{},
			},
		}
	}

	finalBytes, _ := json.Marshal(response)
	_ = cache.Get().CacheByPayload(payloadStr, finalBytes)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}

// DashboardChatHandler - POST /dashboard/chat
func DashboardChatHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DashboardChatHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		log.Printf("[DashboardChatHandler] Invalid method: %s", r.Method)
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DashboardChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	systemPrompt, err := prompt.GetPrompt("dashboardChatPrompt")
	if err != nil {
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}

	userMessage, _ := json.Marshal(req)
	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"input": map[string]string{
			"prompt": systemPrompt + "\n" + string(userMessage),
		},
	}

	endpoint := "/dashboard/chat"
	config, ok := configs.GetDeepseekConfig(endpoint)
	if !ok {
		config, _ = configs.GetDeepseekConfig("/chat")
	}

	clientName := config.ClientName
	if clientName == "" {
		clientName = "chat"
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		utils.WriteJSONError(w, "failed to get client", http.StatusInternalServerError)
		return
	}

	rawResponse, err := client.Call(endpoint, payload)
	if err != nil {
		utils.WriteJSONError(w, "Failed to generate chat response", http.StatusInternalServerError)
		return
	}

	var response DashboardChatResponse
	if err := json.Unmarshal([]byte(rawResponse), &response); err != nil {
		response = DashboardChatResponse{
			Response:    rawResponse,
			Suggestions: []string{},
		}
	}

	finalBytes, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}
