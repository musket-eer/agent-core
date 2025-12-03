package llm

import (
	"agent-core/internal/llm/utils"
	"encoding/json"
	"log"
	"net/http"
)

// ============================================================================
// Milestone Generation Handlers - Stateless JSON in/out
// ============================================================================

// GenerateProjectHandler - POST /generate-project
// Generates complete project with milestones and string array tasks
func GenerateProjectHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GenerateProjectHandler] Incoming request: %s %s", r.Method, r.URL.Path)

	var req GenerateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GenerateProjectHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.User.ID == "" || req.Project.Title == "" {
		log.Printf("[GenerateProjectHandler] Missing required fields")
		utils.WriteJSONError(w, "user.id and project.title are required", http.StatusBadRequest)
		return
	}

	log.Printf("[GenerateProjectHandler] Generating project: %s", req.Project.Title)

	// Call LLM with project generation prompt
	response := handleLLMRequest(w, req, "generateProjectPrompt", "/generate-project")
	if response == nil {
		return
	}

	// Parse LLM response
	var projectResp GenerateProjectResponse
	if err := json.Unmarshal(response, &projectResp); err != nil {
		log.Printf("[GenerateProjectHandler] Failed to parse LLM response: %v", err)
		utils.WriteJSONError(w, "LLM JSON parse error", http.StatusInternalServerError)
		return
	}

	// Light normalization - ensure arrays are never nil
	if projectResp.LearningObjectives == nil {
		projectResp.LearningObjectives = []string{}
	}
	if projectResp.Prerequisites == nil {
		projectResp.Prerequisites = []string{}
	}
	if projectResp.Technologies == nil {
		projectResp.Technologies = []string{}
	}
	if projectResp.Deliverables == nil {
		projectResp.Deliverables = []string{}
	}
	if projectResp.Milestones == nil {
		projectResp.Milestones = []GeneratedMilestone{}
	}

	// Ensure each milestone has non-nil tasks array
	for i := range projectResp.Milestones {
		if projectResp.Milestones[i].Tasks == nil {
			projectResp.Milestones[i].Tasks = []string{}
		}
	}

	finalBytes, _ := json.Marshal(projectResp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)

	log.Printf("[GenerateProjectHandler] Successfully generated project with %d milestones", len(projectResp.Milestones))
}

// MilestoneToTasksHandler - POST /ai/projects/milestone-to-tasks
// Refines tasks for a single milestone
func MilestoneToTasksHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[MilestoneToTasksHandler] Incoming request: %s %s", r.Method, r.URL.Path)

	var req MilestoneToTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[MilestoneToTasksHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Milestone.Title == "" || req.Project.Title == "" {
		log.Printf("[MilestoneToTasksHandler] Missing required fields")
		utils.WriteJSONError(w, "milestone.title and project.title are required", http.StatusBadRequest)
		return
	}

	log.Printf("[MilestoneToTasksHandler] Generating tasks for milestone: %s", req.Milestone.Title)

	// Call LLM with milestone-to-tasks prompt
	response := handleLLMRequest(w, req, "milestoneToTasksPrompt", "/ai/projects/milestone-to-tasks")
	if response == nil {
		return
	}

	// Parse LLM response
	var tasksResp MilestoneToTasksResponse
	if err := json.Unmarshal(response, &tasksResp); err != nil {
		log.Printf("[MilestoneToTasksHandler] Failed to parse LLM response: %v", err)
		// Return fallback response with empty tasks
		tasksResp = MilestoneToTasksResponse{
			Tasks: []string{},
		}
	}

	// Ensure tasks array is never nil
	if tasksResp.Tasks == nil {
		tasksResp.Tasks = []string{}
	}

	finalBytes, _ := json.Marshal(tasksResp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)

	log.Printf("[MilestoneToTasksHandler] Successfully generated %d tasks", len(tasksResp.Tasks))
}
