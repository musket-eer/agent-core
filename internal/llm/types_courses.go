package llm

// ============================================================================
// Course Browse Agent Types
// ============================================================================

type CourseIntelligentSearchRequest struct {
	Query       string                 `json:"query"`
	UserID      string                 `json:"userId"`
	UserProfile UserProfile            `json:"userProfile"`
	Filters     SearchFilters          `json:"filters"`
	UserContext map[string]interface{} `json:"userContext"`
}

type SearchFilters struct {
	Difficulty string   `json:"difficulty,omitempty"`
	Category   string   `json:"category,omitempty"`
	Duration   int      `json:"duration,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

type CourseIntelligentSearchResponse struct {
	Results        []SearchResult `json:"results"`
	SearchInsights SearchInsights `json:"searchInsights"`
}

type SearchResult struct {
	CourseID       string   `json:"courseId"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	RelevanceScore int      `json:"relevanceScore"`
	MatchReason    string   `json:"matchReason"`
	Highlights     []string `json:"highlights"`
}

type SearchInsights struct {
	InterpretedQuery     string   `json:"interpretedQuery"`
	SuggestedRefinements []string `json:"suggestedRefinements"`
}

type CourseExplainRelevanceRequest struct {
	Course      CourseDetails `json:"course"`
	UserProfile UserProfile   `json:"userProfile"`
}

type CourseExplainRelevanceResponse struct {
	Explanation string `json:"explanation"`
}

type CourseCompareRequest struct {
	Courses     []CourseDetails        `json:"courses"`
	UserContext map[string]interface{} `json:"userContext"`
}

type CourseCompareResponse struct {
	Comparison CourseComparison `json:"comparison"`
}

type CourseComparison struct {
	Summary         string             `json:"summary"`
	Recommendations string             `json:"recommendations"`
	ComparisonTable []ComparisonAspect `json:"comparisonTable"`
	BestFor         BestForCategories  `json:"bestFor"`
}

type ComparisonAspect struct {
	Aspect  string              `json:"aspect"`
	Courses []CourseAspectValue `json:"courses"`
}

type CourseAspectValue struct {
	CourseID string `json:"courseId"`
	Value    string `json:"value"`
	Score    int    `json:"score"`
}

type BestForCategories struct {
	Beginners       string `json:"beginners"`
	TimeConstrained string `json:"timeConstrained"`
	Comprehensive   string `json:"comprehensive"`
}

type CourseSuggestPathRequest struct {
	UserID              string   `json:"userId"`
	TargetGoal          string   `json:"targetGoal"`
	CurrentSkills       []string `json:"currentSkills"`
	TimeCommitment      string   `json:"timeCommitment"`      // "flexible", "part-time", "full-time"
	PreferredDifficulty string   `json:"preferredDifficulty"` // "progressive", "challenging", "comfortable"
	Interests           []string `json:"interests"`
}

type CourseSuggestPathResponse struct {
	LearningPath LearningPath `json:"learningPath"`
}

type LearningPath struct {
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	EstimatedDuration int             `json:"estimatedDuration"`
	Phases            []LearningPhase `json:"phases"`
	Milestones        []string        `json:"milestones"`
	Reasoning         string          `json:"reasoning"`
}

type LearningPhase struct {
	Phase      int      `json:"phase"`
	Title      string   `json:"title"`
	Courses    []string `json:"courses"`
	Duration   int      `json:"duration"`
	Objectives []string `json:"objectives"`
}

type CourseChatRequest struct {
	Message       string                 `json:"message"`
	UserID        string                 `json:"userId"`
	UserProfile   UserProfile            `json:"userProfile"`
	ChatHistory   []ChatMessage          `json:"chatHistory"`
	BrowseContext map[string]interface{} `json:"browseContext"`
}

type CourseChatResponse struct {
	Response         string             `json:"response"`
	SuggestedCourses []string           `json:"suggestedCourses"`
	SuggestedFilters []FilterSuggestion `json:"suggestedFilters"`
}

type FilterSuggestion struct {
	FilterType string `json:"filterType"`
	Value      string `json:"value"`
	Reason     string `json:"reason"`
}
