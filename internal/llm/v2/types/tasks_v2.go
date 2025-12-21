package types

// ============================================================================
// Tasks Agent V2 Types
// ============================================================================

// V2TasksGenerateSuggestionsRequest - Generate task suggestions with memory context
type V2TasksGenerateSuggestionsRequest struct {
	V2BaseRequest
	LearningGoals  []string               `json:"learningGoals"`
	AvailableTime  string                 `json:"availableTime"`
	CurrentCourses []string               `json:"currentCourses,omitempty"`
	Preferences    map[string]interface{} `json:"preferences,omitempty"`
}

// V2TasksGenerateSuggestionsResponse - Task suggestions with personalization
type V2TasksGenerateSuggestionsResponse struct {
	SuggestedTasks       []V2SuggestedTask `json:"suggestedTasks"`
	PersonalizationNotes string            `json:"personalizationNotes,omitempty"`
	Reasoning            string            `json:"reasoning,omitempty"`
}

// V2SuggestedTask - Enhanced task with personalization
type V2SuggestedTask struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	EstimatedTime    int      `json:"estimatedTime"`
	Priority         string   `json:"priority"`
	Reasoning        string   `json:"reasoning,omitempty"`
	VisualResources  []string `json:"visualResources,omitempty"`
	Prerequisites    []string `json:"prerequisites,omitempty"`
	LearningOutcomes []string `json:"learningOutcomes,omitempty"`
}

// V2TasksPrioritizeRequest - Prioritize tasks with memory context
type V2TasksPrioritizeRequest struct {
	V2BaseRequest
	Tasks          []V2TaskInput `json:"tasks"`
	AvailableTime  string        `json:"availableTime"`
	EnergyLevel    string        `json:"energyLevel,omitempty"`
	CurrentContext string        `json:"currentContext,omitempty"`
}

// V2TaskInput - Task to be prioritized
type V2TaskInput struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"dueDate,omitempty"`
	Urgency     string `json:"urgency,omitempty"`
	Importance  string `json:"importance,omitempty"`
}

// V2TasksPrioritizeResponse - Prioritized tasks with reasoning
type V2TasksPrioritizeResponse struct {
	PrioritizedTasks     []V2PrioritizedTask `json:"prioritizedTasks"`
	Reasoning            string              `json:"reasoning"`
	TimeAllocation       V2TimeAllocation    `json:"timeAllocation"`
	Suggestions          []string            `json:"suggestions"`
	PersonalizationNotes string              `json:"personalizationNotes,omitempty"`
}

// V2PrioritizedTask - Task with priority and reasoning
type V2PrioritizedTask struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Priority      int    `json:"priority"`
	Reasoning     string `json:"reasoning"`
	SuggestedTime string `json:"suggestedTime,omitempty"`
	EstimatedTime int    `json:"estimatedTime"`
}

// V2TimeAllocation - Time-based task allocation
type V2TimeAllocation struct {
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
	Evening   []string `json:"evening"`
}

// V2TasksBreakDownRequest - Break down complex task
type V2TasksBreakDownRequest struct {
	V2BaseRequest
	TaskTitle       string `json:"taskTitle"`
	TaskDescription string `json:"taskDescription"`
	Complexity      string `json:"complexity,omitempty"`
	TimeAvailable   string `json:"timeAvailable,omitempty"`
}

// V2TasksBreakDownResponse - Subtasks with dependencies
type V2TasksBreakDownResponse struct {
	Subtasks             []V2Subtask `json:"subtasks"`
	Reasoning            string      `json:"reasoning"`
	EstimatedTotalTime   int         `json:"estimatedTotalTime"`
	PersonalizationNotes string      `json:"personalizationNotes,omitempty"`
}

// V2Subtask - Individual subtask
type V2Subtask struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	EstimatedTime int      `json:"estimatedTime"`
	Order         int      `json:"order"`
	Dependencies  []string `json:"dependencies,omitempty"`
	Tips          []string `json:"tips,omitempty"`
}

// V2TasksInsightsRequest - Get productivity insights
type V2TasksInsightsRequest struct {
	V2BaseRequest
	TimeFrame      string            `json:"timeFrame"`
	CompletedTasks []V2CompletedTask `json:"completedTasks"`
}

// V2CompletedTask - Task completion data
type V2CompletedTask struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	EstimatedTime      int    `json:"estimatedTime"`
	ActualTime         int    `json:"actualTime"`
	CompletedAt        string `json:"completedAt"`
	ProductivityRating int    `json:"productivityRating,omitempty"`
}

// V2TasksInsightsResponse - Productivity insights
type V2TasksInsightsResponse struct {
	CompletionRate       float64                 `json:"completionRate"`
	Patterns             []V2ProductivityPattern `json:"patterns"`
	Strengths            []string                `json:"strengths"`
	Improvements         []string                `json:"improvements"`
	Recommendations      []string                `json:"recommendations"`
	OptimalWorkTimes     []V2OptimalWorkTime     `json:"optimalWorkTimes"`
	TrendAnalysis        V2TrendAnalysis         `json:"trendAnalysis,omitempty"`
	PersonalizationNotes string                  `json:"personalizationNotes,omitempty"`
}

// V2ProductivityPattern - Observed pattern
type V2ProductivityPattern struct {
	Pattern    string  `json:"pattern"`
	Frequency  string  `json:"frequency"`
	Impact     string  `json:"impact"`
	Confidence float64 `json:"confidence,omitempty"`
}

// V2OptimalWorkTime - Best time for work
type V2OptimalWorkTime struct {
	TimeOfDay         string   `json:"timeOfDay"`
	ProductivityScore float64  `json:"productivityScore"`
	TaskTypes         []string `json:"taskTypes"`
}

// V2TrendAnalysis - Trend over time
type V2TrendAnalysis struct {
	EstimationAccuracy string `json:"estimationAccuracy"`
	ProductivityTrend  string `json:"productivityTrend"`
	SkillGrowth        string `json:"skillGrowth"`
}

// V2TasksChatRequest - Chat with task assistant
type V2TasksChatRequest struct {
	V2BaseRequest
	Message string `json:"message"`
}

// V2TasksChatResponse - Chat response
type V2TasksChatResponse struct {
	Response             string   `json:"response"`
	Suggestions          []string `json:"suggestions,omitempty"`
	ActionItems          []string `json:"actionItems,omitempty"`
	PersonalizationNotes string   `json:"personalizationNotes,omitempty"`
}
