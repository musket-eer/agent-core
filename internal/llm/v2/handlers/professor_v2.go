package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"agent-core/internal/cache"
	"agent-core/internal/llm/utils"
	"agent-core/internal/llm/v2/middleware"
	"agent-core/internal/llm/v2/types"
	"agent-core/internal/prompt"
)

// ============================================================================
// Professor Agent V2 Handlers
// ============================================================================

// ProfessorChatV2Handler handles student chat messages with memory context
func ProfessorChatV2Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[ProfessorV2] Chat request received")

	var req types.ProfessorChatV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProfessorV2] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check feature flag
	featureFlags := middleware.GetFeatureFlags()
	if !featureFlags.ShouldUseV2(req.UserID, req.AgentID) {
		log.Printf("[ProfessorV2] User %s not enabled for V2", req.UserID)
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
	basePrompt, err := prompt.GetPrompt("professorChatPrompt")
	if err != nil {
		log.Printf("[ProfessorV2] Failed to load prompt: %v", err)
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}

	// Build enriched system prompt
	systemPrompt := middleware.BuildSystemPrompt(basePrompt, contextData)
	systemPrompt += buildProfessorChatContext(req.LessonTitle, req.StudentMessage, req.CourseDetails, req.StudentProfile)

	// Check cache
	cacheKey := systemPrompt + req.StudentMessage
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(cacheKey); found {
		log.Printf("[ProfessorV2] Cache hit")
		writeV2Response(w, cached, contextData, startTime, true)
		return
	}

	// TODO: Call LLM with enriched prompt
	log.Printf("[ProfessorV2] LLM call would happen here")

	// Placeholder response
	responseBytes := []byte(fmt.Sprintf(`{"response": "[V2 Placeholder] Professor response to: %s"}`, req.StudentMessage))

	// Cache response
	_ = cacheInstance.CacheByPayload(cacheKey, responseBytes)

	writeV2Response(w, responseBytes, contextData, startTime, false)
}

// ProfessorGenerateNotesV2Handler generates lesson notes with memory context
func ProfessorGenerateNotesV2Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[ProfessorV2] Generate notes request received")

	var req types.ProfessorGenerateNotesV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProfessorV2] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	featureFlags := middleware.GetFeatureFlags()
	if !featureFlags.ShouldUseV2(req.UserID, req.AgentID) {
		log.Printf("[ProfessorV2] User %s not enabled for V2", req.UserID)
		utils.WriteJSONError(w, "V2 not enabled for this user", http.StatusForbidden)
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

	basePrompt, err := prompt.GetPrompt("generateNotesPrompt")
	if err != nil {
		log.Printf("[ProfessorV2] Failed to load prompt: %v", err)
		utils.WriteJSONError(w, "Failed to load prompt", http.StatusInternalServerError)
		return
	}

	systemPrompt := middleware.BuildSystemPrompt(basePrompt, contextData)
	systemPrompt += buildGenerateNotesContext(req)

	cacheKey := systemPrompt + req.LessonTitle
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(cacheKey); found {
		log.Printf("[ProfessorV2] Cache hit")
		writeV2Response(w, cached, contextData, startTime, true)
		return
	}

	log.Printf("[ProfessorV2] LLM call would happen here")

	responseBytes := []byte(fmt.Sprintf(`{"notes": "[V2 Placeholder] Lesson notes for: %s"}`, req.LessonTitle))
	_ = cacheInstance.CacheByPayload(cacheKey, responseBytes)

	writeV2Response(w, responseBytes, contextData, startTime, false)
}

// ProfessorInsightsV2Handler provides teaching insights
func ProfessorInsightsV2Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[ProfessorV2] Insights request received")

	var req types.ProfessorInsightsV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProfessorV2] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	featureFlags := middleware.GetFeatureFlags()
	if !featureFlags.ShouldUseV2(req.UserID, req.AgentID) {
		log.Printf("[ProfessorV2] User %s not enabled for V2", req.UserID)
		utils.WriteJSONError(w, "V2 not enabled for this user", http.StatusForbidden)
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

	basePrompt := "You are an expert educational analyst providing teaching insights."
	systemPrompt := middleware.BuildSystemPrompt(basePrompt, contextData)
	systemPrompt += fmt.Sprintf("\n\n## Analysis Timeframe\n%s\n", req.Timeframe)

	cacheKey := systemPrompt + req.Timeframe
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(cacheKey); found {
		log.Printf("[ProfessorV2] Cache hit")
		writeV2Response(w, cached, contextData, startTime, true)
		return
	}

	log.Printf("[ProfessorV2] LLM call would happen here")

	responseBytes := []byte(`{
		"insights": {
			"learningPace": "medium",
			"comprehensionLevel": "high",
			"strengths": ["Quick learner", "Asks good questions"],
			"areasForImprovement": ["Practice more examples"],
			"recommendedTopics": ["Advanced concepts", "Real-world projects"],
			"teachingStrategy": "Continue with hands-on approach"
		}
	}`)

	_ = cacheInstance.CacheByPayload(cacheKey, responseBytes)
	writeV2Response(w, responseBytes, contextData, startTime, false)
}

// ProfessorExplainV2Handler explains concepts adaptively
func ProfessorExplainV2Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[ProfessorV2] Explain request received")

	var req types.ProfessorExplainV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProfessorV2] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	featureFlags := middleware.GetFeatureFlags()
	if !featureFlags.ShouldUseV2(req.UserID, req.AgentID) {
		log.Printf("[ProfessorV2] User %s not enabled for V2", req.UserID)
		utils.WriteJSONError(w, "V2 not enabled for this user", http.StatusForbidden)
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

	basePrompt := "You are an expert Professor explaining concepts clearly."
	systemPrompt := middleware.BuildSystemPrompt(basePrompt, contextData)
	systemPrompt += fmt.Sprintf("\n\n## Concept to Explain\n%s\n## Target Difficulty\n%s\n", req.Concept, req.Difficulty)

	cacheKey := systemPrompt + req.Concept
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(cacheKey); found {
		log.Printf("[ProfessorV2] Cache hit")
		writeV2Response(w, cached, contextData, startTime, true)
		return
	}

	log.Printf("[ProfessorV2] LLM call would happen here")

	responseBytes := []byte(fmt.Sprintf(`{"explanation": "[V2 Placeholder] Explanation of: %s at %s level"}`, req.Concept, req.Difficulty))
	_ = cacheInstance.CacheByPayload(cacheKey, responseBytes)

	writeV2Response(w, responseBytes, contextData, startTime, false)
}

// ProfessorPracticeQuestionsV2Handler generates practice questions
func ProfessorPracticeQuestionsV2Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	log.Printf("[ProfessorV2] Practice questions request received")

	var req types.ProfessorPracticeQuestionsV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProfessorV2] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	featureFlags := middleware.GetFeatureFlags()
	if !featureFlags.ShouldUseV2(req.UserID, req.AgentID) {
		log.Printf("[ProfessorV2] User %s not enabled for V2", req.UserID)
		utils.WriteJSONError(w, "V2 not enabled for this user", http.StatusForbidden)
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

	basePrompt := "You are an expert Professor creating practice questions."
	systemPrompt := middleware.BuildSystemPrompt(basePrompt, contextData)
	systemPrompt += fmt.Sprintf("\n\n## Task\nGenerate %d practice questions on: %s\nDifficulty: %s\nQuestion Type: %s\n", req.QuestionCount, req.Topic, req.Difficulty, req.QuestionType)

	cacheKey := systemPrompt + req.Topic
	cacheInstance := cache.Get()
	if cached, found, _ := cacheInstance.CheckByPayload(cacheKey); found {
		log.Printf("[ProfessorV2] Cache hit")
		writeV2Response(w, cached, contextData, startTime, true)
		return
	}

	log.Printf("[ProfessorV2] LLM call would happen here")

	responseBytes := []byte(fmt.Sprintf(`{
		"questions": [{
			"question": "Sample question about %s",
			"type": "%s",
			"options": ["Option A", "Option B", "Option C", "Option D"],
			"correctAnswer": "Option A",
			"explanation": "This is the correct answer because..."
		}]
	}`, req.Topic, req.QuestionType))

	_ = cacheInstance.CacheByPayload(cacheKey, responseBytes)
	writeV2Response(w, responseBytes, contextData, startTime, false)
}

// ============================================================================
// Helper Functions
// ============================================================================

func buildProfessorChatContext(
	lessonTitle string,
	studentMessage string,
	courseDetails map[string]interface{},
	studentProfile map[string]interface{},
) string {
	var context string

	context += fmt.Sprintf("\n\n## Current Lesson\n%s\n", lessonTitle)

	if courseDetails != nil {
		context += "\n## Course Context\n"
		if title, ok := courseDetails["title"].(string); ok {
			context += fmt.Sprintf("Course: %s\n", title)
		}
		if difficulty, ok := courseDetails["difficulty_level"].(string); ok {
			context += fmt.Sprintf("Difficulty: %s\n", difficulty)
		}
	}

	if studentProfile != nil {
		context += "\n## Student Profile\n"
		if name, ok := studentProfile["name"].(string); ok {
			context += fmt.Sprintf("Name: %s\n", name)
		}
		if level, ok := studentProfile["education_level"].(string); ok {
			context += fmt.Sprintf("Education Level: %s\n", level)
		}
		if style, ok := studentProfile["learning_style"].(string); ok {
			context += fmt.Sprintf("Learning Style: %s\n", style)
		}
	}

	context += fmt.Sprintf("\n## Student Question\n%s\n", studentMessage)

	return context
}

func buildGenerateNotesContext(req types.ProfessorGenerateNotesV2Request) string {
	var context string

	context += "\n\n## Course Information\n"
	context += fmt.Sprintf("Course: %s\n", req.CourseTitle)
	context += fmt.Sprintf("Description: %s\n", req.CourseDescription)
	context += fmt.Sprintf("Difficulty: %s\n", req.CourseDifficultyLevel)

	context += "\n## Lesson Information\n"
	context += fmt.Sprintf("Lesson: %s\n", req.LessonTitle)
	context += fmt.Sprintf("Objective: %s\n", req.LessonObjective)

	context += "\n## Student Context\n"
	context += fmt.Sprintf("Education Level: %s\n", req.EducationLevel)
	context += fmt.Sprintf("Learning Style: %s\n", req.LearningStyle)

	return context
}
