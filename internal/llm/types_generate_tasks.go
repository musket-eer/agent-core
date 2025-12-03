package llm

// ============================================================================
// Generate Tasks Types - For milestone task generation
// POST /generate-tasks
// ============================================================================

type GenerateTasksRequest struct {
	MilestoneID          string `json:"milestoneId"`
	MilestoneTitle       string `json:"milestoneTitle"`
	MilestoneDescription string `json:"milestoneDescription"`
	ProjectContext       string `json:"projectContext"`
	TaskCount            int    `json:"taskCount"`
}

type GenerateTasksResponse struct {
	Tasks []GeneratedTask `json:"tasks"`
}

type GeneratedTask struct {
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	EstimatedMinutes int      `json:"estimated_minutes"`
	Priority         string   `json:"priority,omitempty"` // "low", "medium", "high", "critical"
	Tags             []string `json:"tags,omitempty"`
}
