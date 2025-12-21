package types

// ============================================================================
// Professor Agent V2 Types
// ============================================================================

// ProfessorChatV2Request represents a chat request to the Professor agent
type ProfessorChatV2Request struct {
	V2BaseRequest
	LessonTitle    string                 `json:"lessonTitle"`
	StudentMessage string                 `json:"studentMessage"`
	CourseDetails  map[string]interface{} `json:"courseDetails"`
	StudentProfile map[string]interface{} `json:"studentProfile"`
}

// ProfessorChatV2Response represents the chat response
type ProfessorChatV2Response struct {
	Response string             `json:"response"`
	Metadata V2ResponseMetadata `json:"metadata"`
}

// ProfessorGenerateNotesV2Request represents a request to generate lesson notes
type ProfessorGenerateNotesV2Request struct {
	V2BaseRequest
	CourseTitle           string `json:"courseTitle"`
	CourseDescription     string `json:"courseDescription"`
	CourseDifficultyLevel string `json:"courseDifficultyLevel"`
	LessonTitle           string `json:"lessonTitle"`
	LessonObjective       string `json:"lessonObjective"`
	EducationLevel        string `json:"educationLevel"`
	LearningStyle         string `json:"learningStyle"`
}

// ProfessorGenerateNotesV2Response represents the notes generation response
type ProfessorGenerateNotesV2Response struct {
	Notes    string             `json:"notes"`
	Metadata V2ResponseMetadata `json:"metadata"`
}

// ProfessorInsightsV2Request represents a request for teaching insights
type ProfessorInsightsV2Request struct {
	V2BaseRequest
	Timeframe string `json:"timeframe"` // e.g., "week", "month", "all"
}

// ProfessorInsightsV2Response represents the insights response
type ProfessorInsightsV2Response struct {
	Insights InsightsData       `json:"insights"`
	Metadata V2ResponseMetadata `json:"metadata"`
}

// InsightsData contains teaching insights
type InsightsData struct {
	LearningPace        string   `json:"learningPace"`
	ComprehensionLevel  string   `json:"comprehensionLevel"`
	Strengths           []string `json:"strengths"`
	AreasForImprovement []string `json:"areasForImprovement"`
	RecommendedTopics   []string `json:"recommendedTopics"`
	TeachingStrategy    string   `json:"teachingStrategy"`
}

// ProfessorExplainV2Request represents a request to explain a concept
type ProfessorExplainV2Request struct {
	V2BaseRequest
	Concept        string                 `json:"concept"`
	ConceptContext map[string]interface{} `json:"conceptContext"`
	Difficulty     string                 `json:"difficulty"`
}

// ProfessorExplainV2Response represents the explanation response
type ProfessorExplainV2Response struct {
	Explanation string             `json:"explanation"`
	Metadata    V2ResponseMetadata `json:"metadata"`
}

// ProfessorPracticeQuestionsV2Request represents a request for practice questions
type ProfessorPracticeQuestionsV2Request struct {
	V2BaseRequest
	Topic         string `json:"topic"`
	Difficulty    string `json:"difficulty"`
	QuestionCount int    `json:"questionCount"`
	QuestionType  string `json:"questionType"` // "multiple-choice", "short-answer", "essay"
}

// ProfessorPracticeQuestionsV2Response represents the practice questions response
type ProfessorPracticeQuestionsV2Response struct {
	Questions []PracticeQuestion `json:"questions"`
	Metadata  V2ResponseMetadata `json:"metadata"`
}

// PracticeQuestion represents a single practice question
type PracticeQuestion struct {
	Question      string   `json:"question"`
	Type          string   `json:"type"`
	Options       []string `json:"options,omitempty"`
	CorrectAnswer string   `json:"correctAnswer,omitempty"`
	Explanation   string   `json:"explanation"`
}
