package llm

import (
	"agent-core/internal/llm/utils"
	"encoding/json"
	"log"
	"net/http"
)

// GenerateTasksHandler - POST /generate-tasks
// Breaks a milestone into actionable tasks
func GenerateTasksHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GenerateTasksHandler] Incoming request: %s %s", r.Method, r.URL.Path)

	var req GenerateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GenerateTasksHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.MilestoneTitle == "" {
		log.Printf("[GenerateTasksHandler] Missing required field: milestoneTitle")
		utils.WriteJSONError(w, "milestoneTitle is required", http.StatusBadRequest)
		return
	}

	// Ensure safe task count (3-10 range)
	safeTaskCount := req.TaskCount
	if safeTaskCount <= 0 || safeTaskCount > 10 {
		safeTaskCount = 5
	}
	req.TaskCount = safeTaskCount

	log.Printf("[GenerateTasksHandler] Generating %d tasks for milestone: %s", safeTaskCount, req.MilestoneTitle)

	// Call LLM with generate tasks prompt
	response := handleLLMRequest(w, req, "generateTasksPrompt", "/generate-tasks")
	if response == nil {
		return
	}

	// Parse LLM response
	var tasksResp GenerateTasksResponse
	if err := json.Unmarshal(response, &tasksResp); err != nil {
		log.Printf("[GenerateTasksHandler] Failed to parse LLM response: %v", err)
		utils.WriteJSONError(w, "Failed to parse LLM JSON response", http.StatusInternalServerError)
		return
	}

	// Normalize tasks for safety
	if tasksResp.Tasks == nil {
		tasksResp.Tasks = []GeneratedTask{}
	}

	normalizedTasks := make([]GeneratedTask, 0, len(tasksResp.Tasks))
	for i, task := range tasksResp.Tasks {
		// Ensure title is present
		title := task.Title
		if title == "" {
			title = "Task " + string(rune(i+1))
		}

		// Ensure description is string
		description := task.Description

		// Ensure estimated_minutes is positive
		estimatedMinutes := task.EstimatedMinutes
		if estimatedMinutes <= 0 {
			estimatedMinutes = 45
		}

		// Validate priority
		priority := task.Priority
		validPriorities := map[string]bool{
			"low":      true,
			"medium":   true,
			"high":     true,
			"critical": true,
		}
		if !validPriorities[priority] {
			priority = "medium"
		}

		// Ensure tags is array
		tags := task.Tags
		if tags == nil {
			tags = []string{}
		}

		normalizedTasks = append(normalizedTasks, GeneratedTask{
			Title:            title,
			Description:      description,
			EstimatedMinutes: estimatedMinutes,
			Priority:         priority,
			Tags:             tags,
		})
	}

	// Return normalized response
	finalResp := GenerateTasksResponse{
		Tasks: normalizedTasks,
	}

	finalBytes, _ := json.Marshal(finalResp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)

	log.Printf("[GenerateTasksHandler] Successfully generated %d tasks", len(normalizedTasks))
}
