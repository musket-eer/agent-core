package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"agent-core/internal/cache"
	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"
)

// CourseIntelligentSearchHandler - POST /courses/intelligent-search
func CourseIntelligentSearchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CourseIntelligentSearchHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req CourseIntelligentSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CourseIntelligentSearchHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("[CourseIntelligentSearchHandler] Parsed request for query: %s", req.Query)

	response := handleLLMRequest(w, req, "courseSearchPrompt", "/courses/intelligent-search")
	if response != nil {
		var searchResp CourseIntelligentSearchResponse
		if err := json.Unmarshal(response, &searchResp); err != nil {
			searchResp = CourseIntelligentSearchResponse{
				Results:        []SearchResult{},
				SearchInsights: SearchInsights{InterpretedQuery: req.Query, SuggestedRefinements: []string{}},
			}
		}
		finalBytes, _ := json.Marshal(searchResp)
		w.Write(finalBytes)
	}
}

// CourseExplainRelevanceHandler - POST /courses/explain-relevance
func CourseExplainRelevanceHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CourseExplainRelevanceHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req CourseExplainRelevanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CourseExplainRelevanceHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := handleLLMRequest(w, req, "courseRelevancePrompt", "/courses/explain-relevance")
	if response != nil {
		var explainResp CourseExplainRelevanceResponse
		if err := json.Unmarshal(response, &explainResp); err != nil {
			explainResp = CourseExplainRelevanceResponse{Explanation: string(response)}
		}
		finalBytes, _ := json.Marshal(explainResp)
		w.Write(finalBytes)
	}
}

// CourseCompareHandler - POST /courses/compare
func CourseCompareHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CourseCompareHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req CourseCompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CourseCompareHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := handleLLMRequest(w, req, "courseComparePrompt", "/courses/compare")
	if response != nil {
		var compareResp CourseCompareResponse
		if err := json.Unmarshal(response, &compareResp); err != nil {
			compareResp = CourseCompareResponse{
				Comparison: CourseComparison{
					Summary:         string(response),
					Recommendations: "",
					ComparisonTable: []ComparisonAspect{},
					BestFor:         BestForCategories{},
				},
			}
		}
		finalBytes, _ := json.Marshal(compareResp)
		w.Write(finalBytes)
	}
}

// CourseSuggestPathHandler - POST /courses/suggest-path
func CourseSuggestPathHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CourseSuggestPathHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req CourseSuggestPathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CourseSuggestPathHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := handleLLMRequest(w, req, "courseSuggestPathPrompt", "/courses/suggest-path")
	if response != nil {
		var pathResp CourseSuggestPathResponse
		if err := json.Unmarshal(response, &pathResp); err != nil {
			pathResp = CourseSuggestPathResponse{
				LearningPath: LearningPath{
					Title:             "Learning Path",
					Description:       string(response),
					EstimatedDuration: 0,
					Phases:            []LearningPhase{},
					Milestones:        []string{},
					Reasoning:         "",
				},
			}
		}
		finalBytes, _ := json.Marshal(pathResp)
		w.Write(finalBytes)
	}
}

// CourseChatHandler - POST /courses/chat
func CourseChatHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CourseChatHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req CourseChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CourseChatHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := handleLLMRequest(w, req, "courseChatPrompt", "/courses/chat")
	if response != nil {
		var chatResp CourseChatResponse
		if err := json.Unmarshal(response, &chatResp); err != nil {
			chatResp = CourseChatResponse{
				Response:         string(response),
				SuggestedCourses: []string{},
				SuggestedFilters: []FilterSuggestion{},
			}
		}
		finalBytes, _ := json.Marshal(chatResp)
		w.Write(finalBytes)
	}
}

// handleLLMRequest is a helper function to reduce code duplication
func handleLLMRequest(w http.ResponseWriter, req interface{}, promptKey string, endpoint string) []byte {
	log.Printf("[handleLLMRequest] Processing request for endpoint: %s", endpoint)
	cachePayload, _ := json.Marshal(req)
	payloadStr := string(cachePayload)
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
		log.Printf("[handleLLMRequest] Cache hit for endpoint: %s", endpoint)
		w.Header().Set("Content-Type", "application/json")
		return cached
	}
	log.Printf("[handleLLMRequest] Cache miss for endpoint: %s, calling LLM", endpoint)

	basePrompt, err := prompt.GetPrompt(promptKey)
	if err != nil {
		log.Printf("[handleLLMRequest] Failed to load prompt %s: %v", promptKey, err)
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return nil
	}

	// Check if this is a V2 request with memory context
	logRequestType(req, endpoint)
	var systemPrompt string

	if isV2Request(req) {
		// V2: Use memory-based context extraction
		log.Printf("[handleLLMRequest] Using V2 memory context for endpoint: %s", endpoint)
		contextData := extractV2ContextCompat(req)
		systemPrompt = buildV2SystemPrompt(basePrompt, contextData)
	} else {
		// V1: Use legacy persona-based system
		log.Printf("[handleLLMRequest] Using V1 persona system for endpoint: %s", endpoint)
		agentID, persona := ExtractPersona(req)
		systemPrompt = BuildPersonaSystemPrompt(agentID, persona, basePrompt)
	}

	userMessage, _ := json.Marshal(req)

	// Build the final enriched prompt
	enrichedPrompt := systemPrompt + "\n" + string(userMessage)

	// Print the entire enriched prompt before sending to AI
	log.Printf("\n" + strings.Repeat("=", 80))
	log.Printf("[ENRICHED PROMPT] Endpoint: %s", endpoint)
	log.Printf(strings.Repeat("=", 80))
	log.Printf("%s", enrichedPrompt)
	log.Printf(strings.Repeat("=", 80) + "\n")

	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"input": map[string]string{
			"prompt": enrichedPrompt,
		},
	}

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
		log.Printf("[handleLLMRequest] Failed to get client %s: %v", clientName, err)
		utils.WriteJSONError(w, "failed to get client", http.StatusInternalServerError)
		return nil
	}
	log.Printf("[handleLLMRequest] Calling LLM client: %s", clientName)

	rawResponse, err := client.Call(endpoint, payload)
	if err != nil {
		log.Printf("[handleLLMRequest] LLM call failed for endpoint %s: %v", endpoint, err)
		utils.WriteJSONError(w, "Failed to call LLM", http.StatusInternalServerError)
		return nil
	}
	log.Printf("[handleLLMRequest] Received LLM response for endpoint: %s", endpoint)

	// Clean the JSON response to remove markdown code blocks and extra text
	cleanedResponse := utils.CleanJSONResponse(rawResponse)
	responseBytes := []byte(cleanedResponse)

	_ = cache.Get().CacheByPayload(payloadStr, responseBytes)
	log.Printf("[handleLLMRequest] Successfully processed request for endpoint: %s, cached response", endpoint)
	w.Header().Set("Content-Type", "application/json")
	return responseBytes
}
