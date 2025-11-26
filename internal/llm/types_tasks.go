package llm

// ============================================================================
// Tasks Agent Types
// ============================================================================

type TasksGenerateSuggestionsRequest struct {
	UserID              string          `json:"userId"`
	LearningGoals       []string        `json:"learningGoals"`
	CurrentCourses      []CourseDetails `json:"currentCourses"`
	AvailableTime       string          `json:"availableTime"`       // "flexible", "limited", "abundant"
	PreferredDifficulty string          `json:"preferredDifficulty"` // "easy", "medium", "hard"
	Interests           []string        `json:"interests"`
	UpcomingDeadlines   []Deadline      `json:"upcomingDeadlines"`
}

type Deadline struct {
	CourseID string `json:"courseId"`
	Deadline string `json:"deadline"`
	Type     string `json:"type"`
}

type TasksGenerateSuggestionsResponse struct {
	SuggestedTasks []SuggestedTask `json:"suggestedTasks"`
}

type SuggestedTask struct {
	Who               string `json:"who"`
	What              string `json:"what"`
	When              string `json:"when"`
	Where             string `json:"where"`
	Why               string `json:"why"`
	How               string `json:"how"`
	UrgencyLevel      string `json:"urgencyLevel"` // "low", "medium", "high", "critical"
	LearningGoal      string `json:"learningGoal"`
	EstimatedDuration int    `json:"estimatedDuration"`
	Reasoning         string `json:"reasoning"`
}

type TasksPrioritizeRequest struct {
	UserID      string                 `json:"userId"`
	Tasks       []Task                 `json:"tasks"`
	UserContext map[string]interface{} `json:"userContext"`
}

type Task struct {
	ID                string   `json:"id"`
	What              string   `json:"what"`
	UrgencyLevel      string   `json:"urgencyLevel"`
	ScheduledTime     string   `json:"scheduledTime"`
	EstimatedDuration int      `json:"estimatedDuration"`
	LearningGoal      string   `json:"learningGoal"`
	Dependencies      []string `json:"dependencies"`
}

type TasksPrioritizeResponse struct {
	PrioritizedTasks []PrioritizedTask `json:"prioritizedTasks"`
	Reasoning        string            `json:"reasoning"`
	TimeAllocation   TimeAllocation    `json:"timeAllocation"`
	Suggestions      []string          `json:"suggestions"`
}

type PrioritizedTask struct {
	TaskID        string `json:"taskId"`
	Priority      int    `json:"priority"`
	Reasoning     string `json:"reasoning"`
	SuggestedTime string `json:"suggestedTime"`
}

type TimeAllocation struct {
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
	Evening   []string `json:"evening"`
}

type TasksBreakDownRequest struct {
	Task        TaskDetail             `json:"task"`
	UserContext map[string]interface{} `json:"userContext"`
}

type TaskDetail struct {
	What         string `json:"what"`
	Why          string `json:"why"`
	LearningGoal string `json:"learningGoal"`
	Difficulty   string `json:"difficulty"`
}

type TasksBreakDownResponse struct {
	Subtasks []Subtask `json:"subtasks"`
}

type Subtask struct {
	Sequence          int      `json:"sequence"`
	What              string   `json:"what"`
	EstimatedDuration int      `json:"estimatedDuration"`
	Difficulty        string   `json:"difficulty"`
	Resources         []string `json:"resources"`
	Checkpoints       []string `json:"checkpoints"`
}

type TasksProductivityInsightsRequest struct {
	UserID              string               `json:"userId"`
	Timeframe           string               `json:"timeframe"` // "day", "week", "month"
	CompletedTasks      []CompletedTask      `json:"completedTasks"`
	PendingTasks        []PendingTask        `json:"pendingTasks"`
	TaskCompletionTimes []TaskCompletionTime `json:"taskCompletionTimes"`
	UrgencyDistribution UrgencyDistribution  `json:"urgencyDistribution"`
}

type CompletedTask struct {
	ID           string `json:"id"`
	CompletedAt  string `json:"completedAt"`
	UrgencyLevel string `json:"urgencyLevel"`
	Duration     int    `json:"duration"`
}

type PendingTask struct {
	ID           string `json:"id"`
	UrgencyLevel string `json:"urgencyLevel"`
	Overdue      bool   `json:"overdue"`
}

type TaskCompletionTime struct {
	TaskID            string `json:"taskId"`
	EstimatedDuration int    `json:"estimatedDuration"`
	ActualDuration    int    `json:"actualDuration"`
}

type UrgencyDistribution struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

type TasksProductivityInsightsResponse struct {
	CompletionRate   float64               `json:"completionRate"`
	Patterns         []ProductivityPattern `json:"patterns"`
	Strengths        []string              `json:"strengths"`
	Improvements     []string              `json:"improvements"`
	Recommendations  []string              `json:"recommendations"`
	OptimalWorkTimes []OptimalWorkTime     `json:"optimalWorkTimes"`
}

type ProductivityPattern struct {
	Pattern   string `json:"pattern"`
	Insight   string `json:"insight"`
	Frequency string `json:"frequency"`
}

type OptimalWorkTime struct {
	TimeOfDay         string `json:"timeOfDay"`
	ProductivityScore int    `json:"productivityScore"`
	Reasoning         string `json:"reasoning"`
}

type TasksGenerateScheduleRequest struct {
	UserID      string              `json:"userId"`
	Tasks       []Task              `json:"tasks"`
	Constraints ScheduleConstraints `json:"constraints"`
	Preferences SchedulePreferences `json:"preferences"`
}

type ScheduleConstraints struct {
	AvailableTimeSlots []TimeSlot       `json:"availableTimeSlots"`
	PreferredWorkHours []string         `json:"preferredWorkHours"`
	BreakPreferences   BreakPreferences `json:"breakPreferences"`
	Deadlines          []Deadline       `json:"deadlines"`
}

type TimeSlot struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type BreakPreferences struct {
	Frequency int `json:"frequency"`
	Duration  int `json:"duration"`
}

type SchedulePreferences struct {
	TaskBatching           bool   `json:"taskBatching"`
	DifficultyDistribution string `json:"difficultyDistribution"` // "front-loaded", "balanced", "back-loaded"
	BufferTime             int    `json:"bufferTime"`
}

type TasksGenerateScheduleResponse struct {
	Schedule     []ScheduleItem        `json:"schedule"`
	Reasoning    string                `json:"reasoning"`
	Alternatives []ScheduleAlternative `json:"alternatives"`
}

type ScheduleItem struct {
	TaskID    string `json:"taskId"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Type      string `json:"type"` // "task", "break"
	Reasoning string `json:"reasoning"`
}

type ScheduleAlternative struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Schedule    []ScheduleItem `json:"schedule"`
}

type TasksChatRequest struct {
	PersonaRequest
	Message      string                 `json:"message"`
	UserID       string                 `json:"userId"`
	UserName     string                 `json:"userName"`
	UserProfile  UserProfile            `json:"userProfile"`
	ChatHistory  []ChatMessage          `json:"chatHistory"`
	TasksContext map[string]interface{} `json:"tasksContext"`
}

func (r TasksChatRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
}

type TasksChatResponse struct {
	Response            string               `json:"response"`
	SuggestedActions    []SuggestedAction    `json:"suggestedActions"`
	TaskRecommendations []TaskRecommendation `json:"taskRecommendations"`
}

type SuggestedAction struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	Priority    string `json:"priority"` // "low", "medium", "high"
}

type TaskRecommendation struct {
	What         string `json:"what"`
	Why          string `json:"why"`
	UrgencyLevel string `json:"urgencyLevel"`
}

type TasksMotivationRequest struct {
	UserID             string   `json:"userId"`
	CompletedToday     int      `json:"completedToday"`
	Streak             int      `json:"streak"`
	RecentAchievements []string `json:"recentAchievements"`
	CurrentChallenges  []string `json:"currentChallenges"`
}

type TasksMotivationResponse struct {
	Message string `json:"message"`
}
