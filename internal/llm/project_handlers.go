package llm

import (
	"agent-core/internal/llm/utils"
	"encoding/json"
	"log"
	"net/http"
)

// ============================================================================
// Project AI Agent Handlers - Stateless JSON in/out
// ============================================================================

// ProjectCreatorHandler - POST /projects/create
// Turns a goal + context into a project plan (project + milestones + courses)
func ProjectCreatorHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ProjectCreatorHandler] Incoming request: %s %s", r.Method, r.URL.Path)

	var req ProjectCreatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProjectCreatorHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Goal == "" {
		log.Printf("[ProjectCreatorHandler] Missing required fields")
		utils.WriteJSONError(w, "userId and goal are required", http.StatusBadRequest)
		return
	}

	log.Printf("[ProjectCreatorHandler] Creating project plan for goal: %s", req.Goal)

	// Call LLM with project creator prompt
	response := handleLLMRequest(w, req, "projectCreatorPrompt", "/projects/create")
	if response == nil {
		return
	}

	// Parse LLM response
	var creatorResp ProjectCreatorResponse
	if err := json.Unmarshal(response, &creatorResp); err != nil {
		log.Printf("[ProjectCreatorHandler] Failed to parse LLM response: %v", err)
		// Return fallback response
		creatorResp = ProjectCreatorResponse{
			Project: ProjectPlan{
				Title:                  "Project Plan",
				Description:            "Unable to generate detailed plan",
				Status:                 "planned",
				Priority:               "medium",
				EstimatedDurationWeeks: req.TimeframeWeeks,
			},
			Milestones: []MilestonePlan{},
			Courses:    []CourseLink{},
		}
	}

	finalBytes, _ := json.Marshal(creatorResp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)

	log.Printf("[ProjectCreatorHandler] Successfully created project plan with %d milestones", len(creatorResp.Milestones))
}

// ProjectAssistantHandler - POST /projects/assist
// Acts as assistant on existing project: answers questions and proposes changes
func ProjectAssistantHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ProjectAssistantHandler] Incoming request: %s %s", r.Method, r.URL.Path)

	var req ProjectAssistantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProjectAssistantHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Message == "" {
		log.Printf("[ProjectAssistantHandler] Missing required fields")
		utils.WriteJSONError(w, "userId and message are required", http.StatusBadRequest)
		return
	}

	log.Printf("[ProjectAssistantHandler] Processing assistant request: %s", req.RequestType)

	// Call LLM with project assistant prompt
	response := handleLLMRequest(w, req, "projectAssistantPrompt", "/projects/assist")
	if response == nil {
		return
	}

	// Parse LLM response
	var assistantResp ProjectAssistantResponse
	if err := json.Unmarshal(response, &assistantResp); err != nil {
		log.Printf("[ProjectAssistantHandler] Failed to parse LLM response: %v", err)
		// Return fallback response with just the raw text
		assistantResp = ProjectAssistantResponse{
			AssistantMessage: string(response),
			SuggestedChanges: SuggestedChanges{},
		}
	}

	finalBytes, _ := json.Marshal(assistantResp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)

	log.Printf("[ProjectAssistantHandler] Successfully processed assistant request")
}

// MilestoneTranslatorHandler - POST /projects/translate-milestones
// Turns milestones into concrete tasks
func MilestoneTranslatorHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[MilestoneTranslatorHandler] Incoming request: %s %s", r.Method, r.URL.Path)

	var req MilestoneTranslatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[MilestoneTranslatorHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.ProjectID == "" || len(req.Milestones) == 0 {
		log.Printf("[MilestoneTranslatorHandler] Missing required fields")
		utils.WriteJSONError(w, "userId, projectId, and milestones are required", http.StatusBadRequest)
		return
	}

	log.Printf("[MilestoneTranslatorHandler] Translating %d milestones to tasks", len(req.Milestones))

	// Call LLM with milestone translator prompt
	response := handleLLMRequest(w, req, "milestoneTranslatorPrompt", "/projects/translate-milestones")
	if response == nil {
		return
	}

	// Parse LLM response
	var translatorResp MilestoneTranslatorResponse
	if err := json.Unmarshal(response, &translatorResp); err != nil {
		log.Printf("[MilestoneTranslatorHandler] Failed to parse LLM response: %v", err)
		// Return empty tasks array as fallback
		translatorResp = MilestoneTranslatorResponse{
			Tasks: []TranslatedTask{},
		}
	}

	finalBytes, _ := json.Marshal(translatorResp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)

	log.Printf("[MilestoneTranslatorHandler] Successfully translated to %d tasks", len(translatorResp.Tasks))
}
