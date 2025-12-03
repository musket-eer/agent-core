package llm

// ============================================================================
// Project Generation Types - Stateless JSON in/out
// Aligned with frontend projectService.createProject contract
// ============================================================================

// ============================================================================
// Generate Complete Project (with milestones and tasks)
// POST /generate-project
// ============================================================================

type GenerateProjectRequest struct {
	User          ProjectUser        `json:"user"`
	Project       ProjectDefinition  `json:"project"`
	CourseContext CourseContext      `json:"course_context,omitempty"`
	Constraints   ProjectConstraints `json:"constraints"`
}

type ProjectUser struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type ProjectDefinition struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	ProjectType        string   `json:"project_type"` // "application", "research", "learning", etc.
	Technologies       []string `json:"technologies"`
	LearningObjectives []string `json:"learning_objectives"`
	Prerequisites      []string `json:"prerequisites"`
	Deliverables       []string `json:"deliverables"`
}

type CourseContext struct {
	CourseID          string   `json:"courseId,omitempty"`
	UnitID            string   `json:"unitId,omitempty"`
	CourseTitle       string   `json:"courseTitle,omitempty"`
	CourseDescription string   `json:"courseDescription,omitempty"`
	UnitTitle         string   `json:"unitTitle,omitempty"`
	UnitObjective     string   `json:"unitObjective,omitempty"`
	TopicsCovered     []string `json:"topicsCovered,omitempty"`
}

type ProjectConstraints struct {
	MaxMilestones        int    `json:"maxMilestones"`
	MaxTasksPerMilestone int    `json:"maxTasksPerMilestone"`
	EstimatedHours       int    `json:"estimatedHours"`
	TargetDifficulty     string `json:"targetDifficulty"` // "beginner", "intermediate", "advanced"
}

type GenerateProjectResponse struct {
	Title              string               `json:"title"`
	Description        string               `json:"description"`
	ProjectType        string               `json:"project_type"`
	EstimatedHours     *int                 `json:"estimated_hours"`
	LearningObjectives []string             `json:"learning_objectives"`
	Prerequisites      []string             `json:"prerequisites"`
	Technologies       []string             `json:"technologies"`
	Deliverables       []string             `json:"deliverables"`
	Milestones         []GeneratedMilestone `json:"milestones"`
	GenerationPrompt   *string              `json:"generation_prompt"`
}

type GeneratedMilestone struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	EstimatedHours *int     `json:"estimated_hours"`
	Status         string   `json:"status"` // "not_started", "in_progress", "completed"
	Tasks          []string `json:"tasks"`
}

// ============================================================================
// Refine Tasks for Single Milestone (optional)
// POST /ai/projects/milestone-to-tasks
// ============================================================================

type MilestoneToTasksRequest struct {
	Milestone   MilestoneSpec        `json:"milestone"`
	Project     ProjectContext       `json:"project"`
	Constraints TaskGenerationLimits `json:"constraints"`
}

type MilestoneSpec struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ProjectContext struct {
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
}

type TaskGenerationLimits struct {
	MaxTasks int `json:"maxTasks"`
}

type MilestoneToTasksResponse struct {
	Tasks []string `json:"tasks"`
}
