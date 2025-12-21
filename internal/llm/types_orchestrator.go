package llm

// ============================================================================
// Orchestrator Agent Types
// ============================================================================

// OrchestratorRequest represents the incoming request to the orchestrator
type OrchestratorRequest struct {
	PersonaRequest
	UserID     string                 `json:"userId"`
	SessionID  string                 `json:"sessionId,omitempty"`
	Message    string                 `json:"message"`
	AppContext map[string]interface{} `json:"appContext,omitempty"`
}

func (r OrchestratorRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
}

// OrchestratorResponse represents the orchestrator's response
type OrchestratorResponse struct {
	Reply    string                `json:"reply"`
	Commands []OrchestratorCommand `json:"commands"`
	Meta     OrchestratorMeta      `json:"meta"`
}

// OrchestratorMeta contains metadata about the orchestrator response
type OrchestratorMeta struct {
	SessionID  string   `json:"sessionId,omitempty"`
	UsedAgents []string `json:"usedAgents,omitempty"`
}

// OrchestratorCommand represents a high-level action the frontend can execute
type OrchestratorCommand struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// Valid command types
const (
	// Task commands
	CommandCreateTask             = "create_task"
	CommandCompleteTask           = "complete_task"
	CommandUpdateTaskNotes        = "update_task_notes"
	CommandGenerateTaskReflection = "generate_task_reflection"

	// Navigation commands
	CommandNavigate  = "navigate"
	CommandFocusTask = "focus_task"
)

// CreateTaskPayload represents the payload for create_task command
type CreateTaskPayload struct {
	Title        string  `json:"title"`
	When         string  `json:"when,omitempty"`
	Priority     string  `json:"priority,omitempty"`
	ProjectTitle *string `json:"projectTitle,omitempty"`
	CourseTitle  *string `json:"courseTitle,omitempty"`
	Description  string  `json:"description,omitempty"`
}

// CompleteTaskPayload represents the payload for complete_task command
type CompleteTaskPayload struct {
	TaskID string `json:"taskId"`
}

// UpdateTaskNotesPayload represents the payload for update_task_notes command
type UpdateTaskNotesPayload struct {
	TaskID string `json:"taskId"`
	Notes  string `json:"notes"`
}

// GenerateTaskReflectionPayload represents the payload for generate_task_reflection command
type GenerateTaskReflectionPayload struct {
	TaskID         string  `json:"taskId"`
	UserReflection *string `json:"userReflection,omitempty"`
}

// NavigatePayload represents the payload for navigate command
type NavigatePayload struct {
	View string `json:"view"` // "tasks", "completed_tasks", "courses", "dashboard"
}

// FocusTaskPayload represents the payload for focus_task command
type FocusTaskPayload struct {
	TaskID string `json:"taskId"`
}
