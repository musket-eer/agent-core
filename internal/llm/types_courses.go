package llm

import (
	"agent-core/internal/llm/v2/types"
)

// ============================================================================
// Course Browse Agent Types
// ============================================================================

type CourseIntelligentSearchRequest struct {
	// Legacy V1 fields
	Query       string                 `json:"query"`
	UserID      string                 `json:"userId"`
	UserProfile UserProfile            `json:"userProfile"`
	Filters     SearchFilters          `json:"filters"`
	UserContext map[string]interface{} `json:"userContext"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r CourseIntelligentSearchRequest) GetV2Context() *types.V2Context { return r.Context }
func (r CourseIntelligentSearchRequest) GetUserID() string              { return r.UserID }
func (r CourseIntelligentSearchRequest) GetAgentID() string             { return r.AgentID }

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
	// Legacy V1 fields
	Course      CourseDetails `json:"course"`
	UserProfile UserProfile   `json:"userProfile"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r CourseExplainRelevanceRequest) GetV2Context() *types.V2Context { return r.Context }
func (r CourseExplainRelevanceRequest) GetUserID() string              { return "" }
func (r CourseExplainRelevanceRequest) GetAgentID() string             { return r.AgentID }

type CourseExplainRelevanceResponse struct {
	Explanation string `json:"explanation"`
}

type CourseCompareRequest struct {
	// Legacy V1 fields
	Courses     []CourseDetails        `json:"courses"`
	UserContext map[string]interface{} `json:"userContext"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r CourseCompareRequest) GetV2Context() *types.V2Context { return r.Context }
func (r CourseCompareRequest) GetUserID() string              { return "" }
func (r CourseCompareRequest) GetAgentID() string             { return r.AgentID }

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
	// Legacy V1 fields
	UserID              string   `json:"userId"`
	TargetGoal          string   `json:"targetGoal"`
	CurrentSkills       []string `json:"currentSkills"`
	TimeCommitment      string   `json:"timeCommitment"`      // "flexible", "part-time", "full-time"
	PreferredDifficulty string   `json:"preferredDifficulty"` // "progressive", "challenging", "comfortable"
	Interests           []string `json:"interests"`

	// V2 fields (optional)
	AgentID       string            `json:"agentId,omitempty"`
	SessionID     string            `json:"sessionId,omitempty"`
	ComponentType string            `json:"componentType,omitempty"`
	ComponentID   string            `json:"componentId,omitempty"`
	Context       *types.V2Context  `json:"context,omitempty"`
	Messages      []types.V2Message `json:"messages,omitempty"`
}

func (r CourseSuggestPathRequest) GetV2Context() *types.V2Context { return r.Context }
func (r CourseSuggestPathRequest) GetUserID() string              { return r.UserID }
func (r CourseSuggestPathRequest) GetAgentID() string             { return r.AgentID }

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
	PersonaRequest
	Message       string                 `json:"message"`
	UserID        string                 `json:"userId"`
	UserProfile   UserProfile            `json:"userProfile"`
	ChatHistory   []ChatMessage          `json:"chatHistory"`
	BrowseContext map[string]interface{} `json:"browseContext"`
}

func (r CourseChatRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
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
