package llm

// ============================================================================
// Course Creation Chat Types - Dialogue-based course creation
// POST /course-creation/chat
// ============================================================================

type CourseCreationChatRequest struct {
	AgentID     string            `json:"agentId"`
	UserID      string            `json:"userId"`
	CurrentStep string            `json:"currentStep"`
	CourseDraft *CourseDraft      `json:"courseDraft"`
	UserMessage string            `json:"userMessage"`
	ChatHistory []ChatHistoryTurn `json:"chatHistory"`
}

type CourseDraft struct {
	ID                 *string     `json:"id"`
	Title              string      `json:"title"`
	Description        string      `json:"description"`
	Objectives         []string    `json:"objectives"`
	PriorKnowledge     string      `json:"priorKnowledge"`
	DifficultyLevel    string      `json:"difficultyLevel"`
	WeeklyHours        int         `json:"weeklyHours"`
	TotalDurationWeeks int         `json:"totalDurationWeeks"`
	ReferenceBooks     []string    `json:"referenceBooks"`
	Units              []DraftUnit `json:"units"`
}

type DraftUnit struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Objective  string `json:"objective"`
	Assessment string `json:"assessment"`
	Position   int    `json:"position"`
}

type ChatHistoryTurn struct {
	Sender string `json:"sender"` // "assistant" or "user"
	Text   string `json:"text"`
}

type CourseCreationChatResponse struct {
	Reply           string                 `json:"reply"`
	UpdatedDraft    map[string]interface{} `json:"updatedDraft"`
	NextStep        string                 `json:"nextStep"`
	Suggestions     []string               `json:"suggestions,omitempty"`
	CreatedCourseID *string                `json:"createdCourseId,omitempty"`
}

// Valid step values for validation
var ValidCourseCreationSteps = map[string]bool{
	"greeting":            true,
	"ask_topic":           true,
	"ask_level":           true,
	"ask_prior_knowledge": true,
	"ask_time_commitment": true,
	"ask_goals":           true,
	"draft_outline":       true,
	"refine_outline":      true,
	"final_review":        true,
	"ready_to_create":     true,
}
