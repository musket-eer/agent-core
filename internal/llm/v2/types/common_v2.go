package types

// ============================================================================
// V2 Common Types - Memory Architecture Integration
// ============================================================================

// V2BaseRequest is the standard request structure for all V2 endpoints
type V2BaseRequest struct {
	AgentID       string                 `json:"agentId"`
	UserID        string                 `json:"userId"`
	SessionID     string                 `json:"sessionId,omitempty"`
	ComponentType string                 `json:"componentType,omitempty"`
	ComponentID   string                 `json:"componentId,omitempty"`
	Context       V2Context              `json:"context"`
	Messages      []V2Message            `json:"messages"`
	RequestData   map[string]interface{} `json:"requestData,omitempty"`
}

// V2Context contains all memory-derived context
type V2Context struct {
	UserInstructions      string `json:"userInstructions,omitempty"`
	SessionMetadata       string `json:"sessionMetadata,omitempty"`
	UserMemory            string `json:"userMemory,omitempty"`
	ConversationSummaries string `json:"conversationSummaries,omitempty"`
	CurrentSession        string `json:"currentSession,omitempty"`
	ComponentContext      string `json:"componentContext,omitempty"`
}

// V2Message represents a chat message in the conversation
type V2Message struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// V2BaseResponse is the standard response structure for all V2 endpoints
type V2BaseResponse struct {
	Success  bool               `json:"success"`
	Data     interface{}        `json:"data"`
	Metadata V2ResponseMetadata `json:"metadata"`
	Error    *V2Error           `json:"error,omitempty"`
}

// V2ResponseMetadata contains response metadata
type V2ResponseMetadata struct {
	TokensUsed             int      `json:"tokensUsed"`
	ProcessingTime         float64  `json:"processingTime"`
	CacheHit               bool     `json:"cacheHit"`
	ContextSources         []string `json:"contextSources,omitempty"`
	PersonalizationApplied []string `json:"personalizationApplied,omitempty"`
}

// V2Error represents an error response
type V2Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// V2FeatureFlags controls V2 feature rollout
type V2FeatureFlags struct {
	UseV2Endpoints      bool     `json:"useV2Endpoints"`
	V2RolloutPercentage float64  `json:"v2RolloutPercentage"`
	EnabledAgents       []string `json:"enabledAgents"`
}

// V2SessionInfo tracks session state
type V2SessionInfo struct {
	SessionID     string `json:"sessionId"`
	UserID        string `json:"userId"`
	AgentID       string `json:"agentId"`
	ComponentType string `json:"componentType"`
	ComponentID   string `json:"componentId"`
	StartTime     int64  `json:"startTime"`
	MessageCount  int    `json:"messageCount"`
}
