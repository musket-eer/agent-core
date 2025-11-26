package llm

// ============================================================================
// Dashboard Agent Types
// ============================================================================

type DashboardInsightsRequest struct {
	UserID           string     `json:"userId"`
	UserName         string     `json:"userName"`
	EducationLevel   string     `json:"educationLevel"`
	CareerGoals      string     `json:"careerGoals"`
	Interests        []string   `json:"interests"`
	RecentActivities []Activity `json:"recentActivities"`
	CompletedCourses []string   `json:"completedCourses"`
	CurrentCourses   []string   `json:"currentCourses"`
	LearningStreak   int        `json:"learningStreak"`
	TotalStudyTime   int        `json:"totalStudyTime"`
}

type DashboardInsightsResponse struct {
	Insights            InsightsSummary     `json:"insights"`
	MotivationalMessage string              `json:"motivationalMessage"`
	ProgressHighlights  []ProgressHighlight `json:"progressHighlights"`
}

type InsightsSummary struct {
	Summary         string   `json:"summary"`
	Strengths       []string `json:"strengths"`
	Recommendations []string `json:"recommendations"`
	NextSteps       []string `json:"nextSteps"`
}

type ProgressHighlight struct {
	Metric string `json:"metric"`
	Value  string `json:"value"`
	Trend  string `json:"trend"` // "up", "down", "stable"
}

type DashboardRecommendationsRequest struct {
	UserID           string   `json:"userId"`
	Interests        []string `json:"interests"`
	CareerGoals      string   `json:"careerGoals"`
	CompletedCourses []string `json:"completedCourses"`
	SkillLevel       string   `json:"skillLevel"`
	LearningStyle    string   `json:"learningStyle"`
}

type DashboardRecommendationsResponse struct {
	Recommendations []CourseRecommendation `json:"recommendations"`
}

type CourseRecommendation struct {
	CourseID          string `json:"courseId"`
	Title             string `json:"title"`
	RelevanceScore    int    `json:"relevanceScore"`
	Reasoning         string `json:"reasoning"`
	EstimatedDuration int    `json:"estimatedDuration"`
	Difficulty        string `json:"difficulty"`
}

type DashboardAnalyzeProgressRequest struct {
	UserID              string            `json:"userId"`
	Timeframe           string            `json:"timeframe"` // "week", "month", "all-time"
	CompletedActivities []Activity        `json:"completedActivities"`
	CoursesInProgress   []string          `json:"coursesInProgress"`
	AssessmentScores    []AssessmentScore `json:"assessmentScores"`
	StudyTimeData       StudyTimeData     `json:"studyTimeData"`
}

type AssessmentScore struct {
	CourseID string `json:"courseId"`
	Score    int    `json:"score"`
	Date     string `json:"date"`
}

type StudyTimeData struct {
	Daily  []int `json:"daily"`
	Weekly []int `json:"weekly"`
}

type DashboardAnalyzeProgressResponse struct {
	Analysis ProgressAnalysis `json:"analysis"`
}

type ProgressAnalysis struct {
	Summary             string   `json:"summary"`
	Trends              []Trend  `json:"trends"`
	Achievements        []string `json:"achievements"`
	AreasForImprovement []string `json:"areasForImprovement"`
	Recommendations     []string `json:"recommendations"`
}

type Trend struct {
	Metric    string `json:"metric"`
	Direction string `json:"direction"` // "improving", "declining", "stable"
	Insight   string `json:"insight"`
}

type DashboardChatRequest struct {
	PersonaRequest
	Message          string                 `json:"message"`
	UserID           string                 `json:"userId"`
	UserName         string                 `json:"userName"`
	UserProfile      UserProfile            `json:"userProfile"`
	ChatHistory      []ChatMessage          `json:"chatHistory"`
	DashboardContext map[string]interface{} `json:"dashboardContext"`
}

func (r DashboardChatRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
}

type DashboardChatResponse struct {
	Response    string   `json:"response"`
	Suggestions []string `json:"suggestions"`
}
