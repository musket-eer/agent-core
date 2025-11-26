package llm

// ============================================================================
// Reflection Agent Types
// ============================================================================

type ReflectionGeneratePromptsRequest struct {
	Activity      Activity               `json:"activity"`
	UserContext   map[string]interface{} `json:"userContext"`
	CourseContext map[string]interface{} `json:"courseContext"`
}

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
	Reflection    string      `json:"reflection"`
	Activity      Activity    `json:"activity"`
	UserProfile   UserProfile `json:"userProfile"`
	LearningGoals []string    `json:"learningGoals"`
}

type ReflectionAnalyzeResponse struct {
	Insights       string   `json:"insights"`
	Strengths      []string `json:"strengths"`
	AreasForGrowth []string `json:"areasForGrowth"`
	Suggestions    []string `json:"suggestions"`
	DepthScore     int      `json:"depthScore"`
}

type ReflectionLearningInsightsRequest struct {
	UserID        string       `json:"userId"`
	Timeframe     string       `json:"timeframe"` // "week", "month", "quarter", "year"
	Reflections   []Reflection `json:"reflections"`
	Activities    []Activity   `json:"activities"`
	LearningGoals []string     `json:"learningGoals"`
}

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
	UserID              string     `json:"userId"`
	StartDate           string     `json:"startDate"`
	EndDate             string     `json:"endDate"`
	CompletedActivities []Activity `json:"completedActivities"`
	Reflections         []string   `json:"reflections"`
	Achievements        []string   `json:"achievements"`
	Challenges          []string   `json:"challenges"`
}

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
	UserID            string   `json:"userId"`
	LearningStyle     string   `json:"learningStyle"`
	CurrentChallenges []string `json:"currentChallenges"`
	LearningGoals     []string `json:"learningGoals"`
	RecentActivities  []string `json:"recentActivities"`
}

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
	UserID         string                 `json:"userId"`
	TaskID         string                 `json:"taskId"`
	TaskInfo       TaskInfo               `json:"taskInfo"`
	TimeData       TimeData               `json:"timeData"`
	CompletionData CompletionData         `json:"completionData"`
	UserResponses  map[string]string      `json:"userResponses"`
	Context        map[string]interface{} `json:"context"`
}

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
	TaskType string `json:"taskType"`
}

type TaskPromptsResponse struct {
	Prompts []string `json:"prompts"`
}

type AnalyzeTaskRequest struct {
	UserResponses map[string]string `json:"userResponses"`
}

type AnalyzeTaskResponse struct {
	Insights       []string `json:"insights"`
	Strengths      []string `json:"strengths"`
	AreasForGrowth []string `json:"areasForGrowth"`
	Suggestions    []string `json:"suggestions"`
	DepthScore     int      `json:"depthScore"`
}
