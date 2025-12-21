package types

// ============================================================================
// Reflection Agent V2 Types
// ============================================================================

// V2ReflectionGenerateQuestionsRequest - Generate reflection questions
type V2ReflectionGenerateQuestionsRequest struct {
	V2BaseRequest
	ActivityType  string   `json:"activityType"`
	ActivityTitle string   `json:"activityTitle"`
	LearningGoals []string `json:"learningGoals,omitempty"`
	Depth         string   `json:"depth,omitempty"` // "quick", "standard", "deep"
}

// V2ReflectionGenerateQuestionsResponse - Reflection questions
type V2ReflectionGenerateQuestionsResponse struct {
	Questions            []V2ReflectionQuestion `json:"questions"`
	Reasoning            string                 `json:"reasoning,omitempty"`
	PersonalizationNotes string                 `json:"personalizationNotes,omitempty"`
}

// V2ReflectionQuestion - Individual reflection question
type V2ReflectionQuestion struct {
	ID       string `json:"id"`
	Question string `json:"question"`
	Category string `json:"category"` // "understanding", "application", "metacognition", "emotion"
	FollowUp string `json:"followUp,omitempty"`
	Depth    int    `json:"depth"` // 1-3 (surface to deep)
}

// V2ReflectionGenerateJournalRequest - Generate journal entry
type V2ReflectionGenerateJournalRequest struct {
	V2BaseRequest
	ActivityType  string            `json:"activityType"`
	ActivityTitle string            `json:"activityTitle"`
	UserResponses map[string]string `json:"userResponses"`
	Prompts       []string          `json:"prompts,omitempty"`
}

// V2ReflectionGenerateJournalResponse - Journal entry
type V2ReflectionGenerateJournalResponse struct {
	JournalEntry         string   `json:"journalEntry"`
	KeyInsights          []string `json:"keyInsights"`
	EmotionalTone        string   `json:"emotionalTone"`
	GrowthAreas          []string `json:"growthAreas,omitempty"`
	PersonalizationNotes string   `json:"personalizationNotes,omitempty"`
}

// V2ReflectionInsightsRequest - Get reflection insights
type V2ReflectionInsightsRequest struct {
	V2BaseRequest
	TimeFrame   string              `json:"timeFrame"`
	Reflections []V2ReflectionEntry `json:"reflections"`
}

// V2ReflectionEntry - Individual reflection
type V2ReflectionEntry struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	ActivityID string `json:"activityId"`
	Date       string `json:"date"`
	Type       string `json:"type,omitempty"`
}

// V2ReflectionInsightsResponse - Reflection insights
type V2ReflectionInsightsResponse struct {
	Patterns             []V2LearningPattern `json:"patterns"`
	GrowthAreas          []string            `json:"growthAreas"`
	Recommendations      []string            `json:"recommendations"`
	Milestones           []V2Milestone       `json:"milestones"`
	MetacognitiveScore   float64             `json:"metacognitiveScore"`
	TrendAnalysis        V2ReflectionTrend   `json:"trendAnalysis,omitempty"`
	PersonalizationNotes string              `json:"personalizationNotes,omitempty"`
}

// V2LearningPattern - Observed learning pattern
type V2LearningPattern struct {
	Pattern    string  `json:"pattern"`
	Frequency  string  `json:"frequency"`
	Insight    string  `json:"insight"`
	Confidence float64 `json:"confidence,omitempty"`
}

// V2Milestone - Learning milestone
type V2Milestone struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Impact      string `json:"impact,omitempty"`
}

// V2ReflectionTrend - Reflection trends over time
type V2ReflectionTrend struct {
	DepthImprovement    string `json:"depthImprovement"`
	SelfAwarenessGrowth string `json:"selfAwarenessGrowth"`
	ReflectionFrequency string `json:"reflectionFrequency"`
}

// V2ReflectionQuickReflectionRequest - Quick task reflection
type V2ReflectionQuickReflectionRequest struct {
	V2BaseRequest
	TaskID         string            `json:"taskId"`
	TaskInfo       V2TaskInfo        `json:"taskInfo"`
	TimeData       V2TimeData        `json:"timeData"`
	CompletionData V2CompletionData  `json:"completionData"`
	UserResponses  map[string]string `json:"userResponses"`
}

// V2TaskInfo - Task information
type V2TaskInfo struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Priority     string `json:"priority,omitempty"`
	UrgencyLevel string `json:"urgencyLevel,omitempty"`
	LearningGoal string `json:"learningGoal,omitempty"`
	Why          string `json:"why,omitempty"`
	How          string `json:"how,omitempty"`
}

// V2TimeData - Time tracking data
type V2TimeData struct {
	EstimatedMinutes int    `json:"estimatedMinutes"`
	TimeSpent        int    `json:"timeSpent"` // in seconds
	ScheduledTime    string `json:"scheduledTime,omitempty"`
	StartedAt        string `json:"startedAt,omitempty"`
	CompletedAt      string `json:"completedAt,omitempty"`
}

// V2CompletionData - Task completion data
type V2CompletionData struct {
	Completed            bool   `json:"completed"`
	TaskOutcome          string `json:"taskOutcome,omitempty"`
	CompletionPercentage int    `json:"completionPercentage"`
	ProductivityRating   int    `json:"productivityRating,omitempty"`
	FocusQuality         int    `json:"focusQuality,omitempty"`
}

// V2ReflectionQuickReflectionResponse - Quick reflection response
type V2ReflectionQuickReflectionResponse struct {
	Reflection           string                `json:"reflection"`
	Insights             []string              `json:"insights"`
	Patterns             []string              `json:"patterns"`
	Recommendations      []string              `json:"recommendations"`
	DepthScore           float64               `json:"depthScore"`
	EmotionalTone        string                `json:"emotionalTone"`
	KeyLearnings         []string              `json:"keyLearnings"`
	ChallengesFaced      []string              `json:"challengesFaced"`
	SolutionsApplied     []string              `json:"solutionsApplied"`
	SkillsPracticed      []string              `json:"skillsPracticed"`
	TrendAnalysis        V2TaskReflectionTrend `json:"trendAnalysis,omitempty"`
	PersonalizationNotes string                `json:"personalizationNotes,omitempty"`
}

// V2TaskReflectionTrend - Task reflection trends
type V2TaskReflectionTrend struct {
	EstimationAccuracy string `json:"estimationAccuracy"`
	ProductivityTrend  string `json:"productivityTrend"`
	SkillGrowth        string `json:"skillGrowth"`
}

// V2ReflectionChatRequest - Chat with reflection coach
type V2ReflectionChatRequest struct {
	V2BaseRequest
	Message           string `json:"message"`
	ReflectionContext string `json:"reflectionContext,omitempty"`
}

// V2ReflectionChatResponse - Chat response
type V2ReflectionChatResponse struct {
	Response             string   `json:"response"`
	FollowUpQuestions    []string `json:"followUpQuestions,omitempty"`
	Insights             []string `json:"insights,omitempty"`
	PersonalizationNotes string   `json:"personalizationNotes,omitempty"`
}
