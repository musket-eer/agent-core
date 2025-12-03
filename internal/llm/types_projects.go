package llm

// ============================================================================
// Project AI Agent Types - Stateless JSON in/out
// ============================================================================

// ============================================================================
// Project Creator Agent (agentId: "project_creator")
// ============================================================================

type ProjectCreatorRequest struct {
	PersonaRequest
	UserID         string                 `json:"userId"`
	Goal           string                 `json:"goal"`
	TimeframeWeeks int                    `json:"timeframeWeeks"`
	Difficulty     string                 `json:"difficulty"` // "beginner", "intermediate", "advanced"
	Constraints    []string               `json:"constraints"`
	UserProfile    map[string]interface{} `json:"userProfile,omitempty"`
	ActiveCourses  []interface{}          `json:"activeCourses,omitempty"`
	RecentTasks    []interface{}          `json:"recentTasks,omitempty"`
}

func (r ProjectCreatorRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
}

type ProjectCreatorResponse struct {
	Project    ProjectPlan     `json:"project"`
	Milestones []MilestonePlan `json:"milestones"`
	Courses    []CourseLink    `json:"courses"`
}

type ProjectPlan struct {
	Title                  string `json:"title"`
	Description            string `json:"description"`
	Status                 string `json:"status"`   // "planned"
	Priority               string `json:"priority"` // "low", "medium", "high"
	Domain                 string `json:"domain,omitempty"`
	EstimatedDurationWeeks int    `json:"estimatedDurationWeeks"`
}

type MilestonePlan struct {
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	MilestoneType  string          `json:"milestone_type"` // "milestone", "subproject", "task_like"
	EstimatedHours float64         `json:"estimated_hours"`
	DueInDays      int             `json:"due_in_days,omitempty"`
	Children       []MilestonePlan `json:"children,omitempty"`
}

type CourseLink struct {
	CourseID string `json:"courseId"`
	Role     string `json:"role"` // "primary", "supporting", "prerequisite"
}

// ============================================================================
// Project Assistant Agent (agentId: "project_assistant")
// ============================================================================

type ProjectAssistantRequest struct {
	PersonaRequest
	UserID      string                 `json:"userId"`
	Message     string                 `json:"message"`
	RequestType string                 `json:"requestType"` // "assist", "replan", "prioritize", "summarize", "risks"
	Project     map[string]interface{} `json:"project"`
	Milestones  []interface{}          `json:"milestones"`
	Courses     []interface{}          `json:"courses,omitempty"`
	UserProfile map[string]interface{} `json:"userProfile,omitempty"`
	RecentTasks []interface{}          `json:"recentTasks,omitempty"`
	TimeLogs    []interface{}          `json:"timeLogs,omitempty"`
}

func (r ProjectAssistantRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
}

type ProjectAssistantResponse struct {
	AssistantMessage string           `json:"assistantMessage"`
	SuggestedChanges SuggestedChanges `json:"suggestedChanges,omitempty"`
}

type SuggestedChanges struct {
	MilestoneUpdates []MilestoneUpdate `json:"milestoneUpdates,omitempty"`
	NewMilestones    []NewMilestone    `json:"newMilestones,omitempty"`
}

type MilestoneUpdate struct {
	MilestoneID    string  `json:"milestoneId"`
	Status         string  `json:"status,omitempty"` // "planned", "in_progress", "completed", "blocked", "dropped"
	Title          string  `json:"title,omitempty"`
	Description    string  `json:"description,omitempty"`
	EstimatedHours float64 `json:"estimated_hours,omitempty"`
	Notes          string  `json:"notes,omitempty"`
}

type NewMilestone struct {
	ParentMilestoneID string  `json:"parentMilestoneId,omitempty"`
	Title             string  `json:"title"`
	Description       string  `json:"description"`
	MilestoneType     string  `json:"milestone_type"` // "milestone", "subproject", "task_like"
	EstimatedHours    float64 `json:"estimated_hours"`
	DueInDays         int     `json:"due_in_days,omitempty"`
}

// ============================================================================
// Milestone Translator Agent (agentId: "milestone_translator")
// ============================================================================

type MilestoneTranslatorRequest struct {
	PersonaRequest
	UserID                string                 `json:"userId"`
	ProjectID             string                 `json:"projectId"`
	Milestones            []MilestoneInput       `json:"milestones"`
	TaskSystemConstraints TaskConstraints        `json:"taskSystemConstraints"`
	UserProfile           map[string]interface{} `json:"userProfile,omitempty"`
	ActiveCourses         []interface{}          `json:"activeCourses,omitempty"`
}

func (r MilestoneTranslatorRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
}

type MilestoneInput struct {
	ID             string                 `json:"id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	EstimatedHours float64                `json:"estimated_hours"`
	CourseContext  map[string]interface{} `json:"courseContext,omitempty"`
}

type TaskConstraints struct {
	MaxTasksPerMilestone    int     `json:"maxTasksPerMilestone"`
	MaxDurationPerTaskHours float64 `json:"maxDurationPerTaskHours"`
}

type MilestoneTranslatorResponse struct {
	Tasks []TranslatedTask `json:"tasks"`
}

type TranslatedTask struct {
	MilestoneID      string `json:"milestoneId"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	Difficulty       string `json:"difficulty"` // "easy", "medium", "hard"
	CourseID         string `json:"courseId,omitempty"`
	LessonID         string `json:"lessonId,omitempty"`
}
