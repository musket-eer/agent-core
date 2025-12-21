package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"agent-core/internal/cache"
	"agent-core/internal/llm/utils"
	"agent-core/internal/llm/v2/middleware"
	"agent-core/internal/llm/v2/types"
	"agent-core/internal/prompt"
)

// TasksGenerateSuggestionsV2Handler - POST /tasks/v2/generate-suggestions
func TasksGenerateSuggestionsV2Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[TasksV2] Generate suggestions request received")

	var req types.V2TasksGenerateSuggestionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksV2] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check feature flag
	featureFlags := middleware.GetFeatureFlags()
	if !featureFlags.ShouldUseV2(req.UserID, req.AgentID) {
		log.Printf("[TasksV2] User %s not enabled for V2, falling back to V1", req.UserID)
		// TODO: Redirect to V1 handler
		utils.WriteJSONError(w, "V2 not enabled for this user", http.StatusForbidden)
		return
	}

	// Extract context
	contextData := middleware.ContextData{
		UserInstructions:      req.Context.UserInstructions,
		SessionMetadata:       req.Context.SessionMetadata,
		UserMemory:            req.Context.UserMemory,
		ConversationSummaries: req.Context.ConversationSummaries,
		CurrentSession:        req.Context.CurrentSession,
		ComponentContext:      req.Context.ComponentContext,
	}

	// Load base prompt
	basePrompt, err := prompt.GetPrompt("tasksSuggestionsPrompt")
	if err != nil {
		log.Printf("[TasksV2] Failed to load prompt: %v", err)
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}

	// Build context-enriched system prompt
	systemPrompt := middleware.BuildSystemPrompt(basePrompt, contextData)

	// Build user message from request data
	userMessage := map[string]interface{}{
		"learningGoals":  req.LearningGoals,
		"availableTime":  req.AvailableTime,
		"currentCourses": req.CurrentCourses,
		"preferences":    req.Preferences,
	}
	userMessageBytes, _ := json.Marshal(userMessage)

	// Check cache
	cacheKey := string(userMessageBytes) + systemPrompt
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(cacheKey); found {
		log.Printf("[TasksV2] Cache hit")
		writeV2Response(w, cached, contextData, startTime, true)
		return
	}

	// TODO: Call LLM with enriched prompt
	// payload := map[string]interface{}{
	// 	"model": "deepseek-chat",
	// 	"input": map[string]string{
	// 		"prompt": systemPrompt + "\n" + string(userMessageBytes),
	// 	},
	// }
	log.Printf("[TasksV2] LLM call would happen here with enriched prompt")

	// Placeholder response
	responseBytes := []byte(`{
		"suggestedTasks": [
			{
				"id": "task-1",
				"title": "Build Todo App with React Hooks",
				"description": "Practice useState and useEffect",
				"estimatedTime": 120,
				"priority": "high",
				"reasoning": "Builds on recent learning"
			}
		],
		"personalizationNotes": "Tasks tailored for visual learners"
	}`)

	// Cache response
	_ = cacheInstance.CacheByPayload(cacheKey, responseBytes)

	// Write V2 response
	writeV2Response(w, responseBytes, contextData, startTime, false)
}

// TasksPrioritizeV2Handler - POST /tasks/v2/prioritize
func TasksPrioritizeV2Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[TasksV2] Prioritize request received")

	var req types.V2TasksPrioritizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksV2] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Similar implementation as above...
	// Extract context, build prompt, call LLM, return response

	contextData := middleware.ContextData{
		UserInstructions:      req.Context.UserInstructions,
		SessionMetadata:       req.Context.SessionMetadata,
		UserMemory:            req.Context.UserMemory,
		ConversationSummaries: req.Context.ConversationSummaries,
		CurrentSession:        req.Context.CurrentSession,
		ComponentContext:      req.Context.ComponentContext,
	}

	// TODO: Load prompt and call LLM
	// For now, using placeholder
	response := []byte(`{"prioritizedTasks": [], "reasoning": "Implementation pending"}`)

	writeV2Response(w, response, contextData, startTime, false)
}

// TasksBreakDownV2Handler - POST /tasks/v2/break-down
func TasksBreakDownV2Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[TasksV2] Break down request received")

	var req types.V2TasksBreakDownRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksV2] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Implementation similar to above
	contextData := middleware.ContextData{
		UserInstructions:      req.Context.UserInstructions,
		SessionMetadata:       req.Context.SessionMetadata,
		UserMemory:            req.Context.UserMemory,
		ConversationSummaries: req.Context.ConversationSummaries,
		CurrentSession:        req.Context.CurrentSession,
		ComponentContext:      req.Context.ComponentContext,
	}

	response := []byte(`{"subtasks": [], "reasoning": "Implementation pending"}`)
	writeV2Response(w, response, contextData, startTime, false)
}

// TasksInsightsV2Handler - POST /tasks/v2/insights
func TasksInsightsV2Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[TasksV2] Insights request received")

	var req types.V2TasksInsightsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksV2] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	contextData := middleware.ContextData{
		UserInstructions:      req.Context.UserInstructions,
		SessionMetadata:       req.Context.SessionMetadata,
		UserMemory:            req.Context.UserMemory,
		ConversationSummaries: req.Context.ConversationSummaries,
		CurrentSession:        req.Context.CurrentSession,
		ComponentContext:      req.Context.ComponentContext,
	}

	response := []byte(`{"completionRate": 0.85, "patterns": [], "strengths": [], "improvements": [], "recommendations": []}`)
	writeV2Response(w, response, contextData, startTime, false)
}

// TasksChatV2Handler - POST /tasks/v2/chat
func TasksChatV2Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[TasksV2] Chat request received")

	var req types.V2TasksChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksV2] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	contextData := middleware.ContextData{
		UserInstructions:      req.Context.UserInstructions,
		SessionMetadata:       req.Context.SessionMetadata,
		UserMemory:            req.Context.UserMemory,
		ConversationSummaries: req.Context.ConversationSummaries,
		CurrentSession:        req.Context.CurrentSession,
		ComponentContext:      req.Context.ComponentContext,
	}

	response := []byte(`{"response": "Hello! How can I help with your tasks?", "suggestions": []}`)
	writeV2Response(w, response, contextData, startTime, false)
}

// writeV2Response writes a standardized V2 response
func writeV2Response(w http.ResponseWriter, dataBytes []byte, context middleware.ContextData, startTime time.Time, cacheHit bool) {
	processingTime := time.Since(startTime).Seconds()
	tokensUsed := middleware.EstimateTokens(string(dataBytes))
	contextSources := middleware.GetContextSources(context)

	// Parse the data
	var data interface{}
	json.Unmarshal(dataBytes, &data)

	response := types.V2BaseResponse{
		Success: true,
		Data:    data,
		Metadata: types.V2ResponseMetadata{
			TokensUsed:     tokensUsed,
			ProcessingTime: processingTime,
			CacheHit:       cacheHit,
			ContextSources: contextSources,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
