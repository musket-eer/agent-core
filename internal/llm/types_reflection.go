package llm

import (
	"agent-core/internal/llm/v2/types"
)

// ============================================================================
// Reflection Agent Types
// ============================================================================

type ReflectionGeneratePromptsRequest struct {
	// Legacy V1 fields
	Activity      Activity               `json:"activity"`
	UserContext   map[string]interface{} `json:"userContext"`
	CourseContext map[string]interface{} `json:"courseContext"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r ReflectionGeneratePromptsRequest) GetV2Context() *types.V2Context { return r.Context }
func (r ReflectionGeneratePromptsRequest) GetUserID() string              { return "" }
func (r ReflectionGeneratePromptsRequest) GetAgentID() string             { return r.AgentID }

type ReflectionGeneratePromptsResponse struct {
	Prompts []ReflectionPrompt `json:"prompts"`
}

type ReflectionPrompt struct {
	ID       string `json:"id"`
	Question string `json:"question"`
	Category string `json:"category"` // "understanding", "application", "metacognition", "emotion"
	FollowUp string `json:"followUp"`
}

type ReflectionAnalyzeRequest struct {
	// Legacy V1 fields
	Reflection    string      `json:"reflection"`
	Activity      Activity    `json:"activity"`
	UserProfile   UserProfile `json:"userProfile"`
	LearningGoals []string    `json:"learningGoals"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r ReflectionAnalyzeRequest) GetV2Context() *types.V2Context { return r.Context }
func (r ReflectionAnalyzeRequest) GetUserID() string              { return "" }
func (r ReflectionAnalyzeRequest) GetAgentID() string             { return r.AgentID }

type ReflectionAnalyzeResponse struct {
	Insights       string   `json:"insights"`
	Strengths      []string `json:"strengths"`
	AreasForGrowth []string `json:"areasForGrowth"`
	Suggestions    []string `json:"suggestions"`
	DepthScore     int      `json:"depthScore"`
}

type ReflectionLearningInsightsRequest struct {
	// Legacy V1 fields
	UserID        string       `json:"userId"`
	Timeframe     string       `json:"timeframe"` // "week", "month", "quarter", "year"
	Reflections   []Reflection `json:"reflections"`
	Activities    []Activity   `json:"activities"`
	LearningGoals []string     `json:"learningGoals"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r ReflectionLearningInsightsRequest) GetV2Context() *types.V2Context { return r.Context }
func (r ReflectionLearningInsightsRequest) GetUserID() string              { return r.UserID }
func (r ReflectionLearningInsightsRequest) GetAgentID() string             { return r.AgentID }

type Reflection struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	ActivityID string `json:"activityId"`
	Date       string `json:"date"`
}

type ReflectionLearningInsightsResponse struct {
	Patterns           []LearningPattern `json:"patterns"`
	GrowthAreas        []string          `json:"growthAreas"`
	Recommendations    []string          `json:"recommendations"`
	Milestones         []Milestone       `json:"milestones"`
	MetacognitiveScore int               `json:"metacognitiveScore"`
}

type LearningPattern struct {
	Pattern   string `json:"pattern"`
	Frequency string `json:"frequency"`
	Insight   string `json:"insight"`
}

type Milestone struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

type ReflectionPeriodSummaryRequest struct {
	// Legacy V1 fields
	UserID              string     `json:"userId"`
	StartDate           string     `json:"startDate"`
	EndDate             string     `json:"endDate"`
	CompletedActivities []Activity `json:"completedActivities"`
	Reflections         []string   `json:"reflections"`
	Achievements        []string   `json:"achievements"`
	Challenges          []string   `json:"challenges"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r ReflectionPeriodSummaryRequest) GetV2Context() *types.V2Context { return r.Context }
func (r ReflectionPeriodSummaryRequest) GetUserID() string              { return r.UserID }
func (r ReflectionPeriodSummaryRequest) GetAgentID() string             { return r.AgentID }

type ReflectionPeriodSummaryResponse struct {
	Summary string `json:"summary"`
}

type ReflectionChatRequest struct {
	PersonaRequest
	Message           string                 `json:"message"`
	UserID            string                 `json:"userId"`
	UserName          string                 `json:"userName"`
	UserProfile       UserProfile            `json:"userProfile"`
	ChatHistory       []ChatMessage          `json:"chatHistory"`
	ReflectionContext map[string]interface{} `json:"reflectionContext"`
}

func (r ReflectionChatRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
}

type ReflectionChatResponse struct {
	Response          string     `json:"response"`
	FollowUpQuestions []string   `json:"followUpQuestions"`
	Resources         []Resource `json:"resources"`
	Insights          []string   `json:"insights,omitempty"`
	Patterns          []string   `json:"patterns,omitempty"`
	Recommendations   []string   `json:"recommendations,omitempty"`
	EmotionalTone     string     `json:"emotionalTone,omitempty"`
}

type ReflectionSuggestExercisesRequest struct {
	// Legacy V1 fields
	UserID            string   `json:"userId"`
	LearningStyle     string   `json:"learningStyle"`
	CurrentChallenges []string `json:"currentChallenges"`
	LearningGoals     []string `json:"learningGoals"`
	RecentActivities  []string `json:"recentActivities"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r ReflectionSuggestExercisesRequest) GetV2Context() *types.V2Context { return r.Context }
func (r ReflectionSuggestExercisesRequest) GetUserID() string              { return r.UserID }
func (r ReflectionSuggestExercisesRequest) GetAgentID() string             { return r.AgentID }

type ReflectionSuggestExercisesResponse struct {
	Exercises []ReflectionExercise `json:"exercises"`
}

type ReflectionExercise struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Duration     int      `json:"duration"`
	Type         string   `json:"type"` // "journaling", "visualization", "discussion", "analysis"
	Instructions []string `json:"instructions"`
}

type TaskInfo struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Priority     string `json:"priority"`
	UrgencyLevel string `json:"urgencyLevel"`
	LearningGoal string `json:"learningGoal"`
	Why          string `json:"why"`
	How          string `json:"how"`
}

type TimeData struct {
	EstimatedMinutes int    `json:"estimatedMinutes"`
	TimeSpent        int    `json:"timeSpent"`
	ScheduledTime    string `json:"scheduledTime"`
	StartedAt        string `json:"startedAt"`
	CompletedAt      string `json:"completedAt"`
}

type CompletionData struct {
	Completed            bool   `json:"completed"`
	TaskOutcome          string `json:"taskOutcome"`
	CompletionPercentage int    `json:"completionPercentage"`
	ProductivityRating   int    `json:"productivityRating"`
	FocusQuality         int    `json:"focusQuality"`
}

type TaskReflectionRequest struct {
	// Legacy V1 fields
	UserID         string                 `json:"userId"`
	TaskID         string                 `json:"taskId"`
	TaskInfo       TaskInfo               `json:"taskInfo"`
	TimeData       TimeData               `json:"timeData"`
	CompletionData CompletionData         `json:"completionData"`
	UserResponses  map[string]string      `json:"userResponses"`
	LegacyContext  map[string]interface{} `json:"context,omitempty"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	V2Context     *types.V2Context  `json:"v2Context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r TaskReflectionRequest) GetV2Context() *types.V2Context { return r.V2Context }
func (r TaskReflectionRequest) GetUserID() string              { return r.UserID }
func (r TaskReflectionRequest) GetAgentID() string             { return r.AgentID }

type TaskReflectionResponse struct {
	Reflection       string   `json:"reflection"`
	Insights         []string `json:"insights"`
	Patterns         []string `json:"patterns"`
	Recommendations  []string `json:"recommendations"`
	DepthScore       float64  `json:"depthScore"`
	EmotionalTone    string   `json:"emotionalTone"`
	KeyLearnings     []string `json:"keyLearnings"`
	ChallengesFaced  []string `json:"challengesFaced"`
	SolutionsApplied []string `json:"solutionsApplied"`
	SkillsPracticed  []string `json:"skillsPracticed"`
}

type TaskPromptsRequest struct {
	// Legacy V1 fields
	TaskType string `json:"taskType"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r TaskPromptsRequest) GetV2Context() *types.V2Context { return r.Context }
func (r TaskPromptsRequest) GetUserID() string              { return "" }
func (r TaskPromptsRequest) GetAgentID() string             { return r.AgentID }

type TaskPromptsResponse struct {
	Prompts []string `json:"prompts"`
}

type AnalyzeTaskRequest struct {
	// Legacy V1 fields
	UserResponses map[string]string `json:"userResponses"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r AnalyzeTaskRequest) GetV2Context() *types.V2Context { return r.Context }
func (r AnalyzeTaskRequest) GetUserID() string              { return "" }
func (r AnalyzeTaskRequest) GetAgentID() string             { return r.AgentID }

type AnalyzeTaskResponse struct {
	Insights       []string `json:"insights"`
	Strengths      []string `json:"strengths"`
	AreasForGrowth []string `json:"areasForGrowth"`
	Suggestions    []string `json:"suggestions"`
	DepthScore     int      `json:"depthScore"`
}
