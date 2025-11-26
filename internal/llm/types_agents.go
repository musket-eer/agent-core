package llm

// ============================================================================
// Professor Agent Types
// ============================================================================

type ProfessorChatRequest struct {
	PersonaRequest
	ActivityID     string        `json:"activityId"`
	LessonTitle    string        `json:"lessonTitle"`
	Leader         string        `json:"leader"`
	StudentMessage string        `json:"studentMessage"`
	StudentProfile UserProfile   `json:"studentProfile"`
	CourseDetails  CourseDetails `json:"courseDetails"`
	Notes          interface{}   `json:"notes"`
	ChatHistory    []ChatMessage `json:"chatHistory"`
}

func (r ProfessorChatRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
}

type ProfessorChatResponse struct {
	Response string `json:"response"`
}

type GenerateNotesRequest struct {
	PersonaRequest
	CourseTitle           string `json:"courseTitle"`
	CourseDescription     string `json:"courseDescription"`
	CourseDifficultyLevel string `json:"courseDifficultyLevel"`
	LessonTitle           string `json:"lessonTitle"`
	LessonObjective       string `json:"lessonObjective"`
	LessonAssessment      string `json:"lessonAssessment"`
	ProfileName           string `json:"profileName"`
	EducationLevel        string `json:"educationLevel"`
	CareerGoals           string `json:"careerGoals"`
	Interests             string `json:"interests"`
}

func (r GenerateNotesRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
}

type GenerateNotesResponse struct {
	Notes string `json:"notes"`
}

type GenerateLessonRequest struct {
	PersonaRequest
	CourseTitle       string      `json:"courseTitle"`
	CourseDescription string      `json:"courseDescription"`
	LessonTitle       string      `json:"lessonTitle"`
	LessonObjective   string      `json:"lessonObjective"`
	StudentProfile    UserProfile `json:"studentProfile"`
	AdditionalContext string      `json:"additionalContext"`
}

func (r GenerateLessonRequest) GetPersona() (string, *Persona) {
	return r.AgentID, r.Persona
}

type GenerateLessonResponse struct {
	LessonPlan LessonPlanDetail `json:"lessonPlan"`
}

type LessonPlanDetail struct {
	LessonMetadata LessonMetadata   `json:"lessonMetadata"`
	Activities     []ActivityDetail `json:"activities"`
}

type LessonMetadata struct {
	Title     string `json:"title"`
	Objective string `json:"objective"`
}

type ActivityDetail struct {
	ID      string                 `json:"id"`
	Title   string                 `json:"title"`
	Leader  string                 `json:"leader"`
	Type    string                 `json:"type"`
	Content map[string]interface{} `json:"content"`
}

// ============================================================================
// Examiner Agent Types
// ============================================================================

type ExaminerChatRequest struct {
	Question        string                 `json:"question"`
	CurrentQuestion map[string]interface{} `json:"currentQuestion"`
	ChatHistory     []ChatMessage          `json:"chatHistory"`
	StudentProfile  UserProfile            `json:"studentProfile"`
	ExamContext     ExamContext            `json:"examContext"`
	QuestionPhase   string                 `json:"questionPhase"` // "pre_exam", "during_exam", "post_exam", "reviewing"
}

type ExamContext struct {
	IsExamActive         bool `json:"isExamActive"`
	CurrentQuestionIndex int  `json:"currentQuestionIndex"`
	TotalQuestions       int  `json:"totalQuestions"`
	TimeRemaining        int  `json:"timeRemaining"`
	Score                int  `json:"score"`
}

type ExaminerChatResponse struct {
	Response string `json:"response"`
}

type GenerateExamQuestionsRequest struct {
	LessonTitle    string        `json:"lessonTitle"`
	StudentProfile UserProfile   `json:"studentProfile"`
	CourseDetails  CourseDetails `json:"courseDetails"`
	QuestionCount  int           `json:"questionCount"`
}

type GenerateExamQuestionsResponse struct {
	Questions []ExamQuestion `json:"questions"`
}

type ExamQuestion struct {
	ID            string   `json:"id"`
	Question      string   `json:"question"`
	Type          string   `json:"type"` // "multiple_choice", "short_answer", "essay", "true_false"
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correctAnswer"`
	Explanation   string   `json:"explanation"`
	Difficulty    string   `json:"difficulty"` // "easy", "medium", "hard"
}

type SubmitExamRequest struct {
	ExamID         string          `json:"examId"`
	StudentAnswers []StudentAnswer `json:"studentAnswers"`
}

type StudentAnswer struct {
	QuestionID string `json:"questionId"`
	Answer     string `json:"answer"`
}

type SubmitExamResponse struct {
	Score           int              `json:"score"`
	TotalQuestions  int              `json:"totalQuestions"`
	CorrectAnswers  int              `json:"correctAnswers"`
	Feedback        []string         `json:"feedback"`
	DetailedResults []QuestionResult `json:"detailedResults"`
}

type QuestionResult struct {
	QuestionID    string `json:"questionId"`
	Correct       bool   `json:"correct"`
	StudentAnswer string `json:"studentAnswer"`
	CorrectAnswer string `json:"correctAnswer"`
	Feedback      string `json:"feedback"`
}

type ExamDetailsResponse struct {
	ExamID    string                 `json:"examId"`
	Questions []ExamQuestion         `json:"questions"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ============================================================================
// Librarian Agent Types
// ============================================================================

type GenerateResourcesRequest struct {
	Topic          string                 `json:"topic"`
	StudentProfile UserProfile            `json:"studentProfile"`
	TopicInfo      map[string]interface{} `json:"topicInfo"`
}

type GenerateResourcesResponse struct {
	Resources []LearningResource `json:"resources"`
}

type LearningResource struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"` // "article", "video", "book", "tutorial", "documentation"
	URL         string `json:"url"`
	Description string `json:"description"`
	Relevance   int    `json:"relevance"`
}

type ChatVideoRequest struct {
	Message string       `json:"message"`
	Context VideoContext `json:"context"`
}

type VideoContext struct {
	Transcript       string `json:"transcript"`
	CurrentTime      int    `json:"currentTime"`
	VideoTitle       string `json:"videoTitle"`
	VideoDescription string `json:"videoDescription"`
	CourseName       string `json:"courseName"`
}

type ChatVideoResponse struct {
	Response  string `json:"response"`
	Timestamp int    `json:"timestamp"`
}

type RecommendVideosRequest struct {
	CourseID     string                 `json:"courseId"`
	CourseTopic  string                 `json:"courseTopic"`
	UserProgress map[string]interface{} `json:"userProgress"`
}

type RecommendVideosResponse struct {
	Recommendations []VideoRecommendation `json:"recommendations"`
}

type VideoRecommendation struct {
	VideoID   string `json:"videoId"`
	Title     string `json:"title"`
	Relevance int    `json:"relevance"`
	Reasoning string `json:"reasoning"`
}

type SummarizeVideoRequest struct {
	Transcript string `json:"transcript"`
	Duration   int    `json:"duration"`
}

type SummarizeVideoResponse struct {
	Summary VideoSummary `json:"summary"`
}

type VideoSummary struct {
	Overview   string           `json:"overview"`
	KeyPoints  []string         `json:"keyPoints"`
	Timestamps []TimestampTopic `json:"timestamps"`
}

type TimestampTopic struct {
	Time  int    `json:"time"`
	Topic string `json:"topic"`
}

type ExtractConceptsRequest struct {
	Transcript string `json:"transcript"`
}

type ExtractConceptsResponse struct {
	Concepts []Concept `json:"concepts"`
}

type Concept struct {
	Concept     string `json:"concept"`
	Explanation string `json:"explanation"`
	Timestamp   int    `json:"timestamp"`
}

type GenerateQuizRequest struct {
	Transcript string `json:"transcript"`
	Difficulty string `json:"difficulty"` // "easy", "medium", "hard"
}

type GenerateQuizResponse struct {
	Questions []QuizQuestion `json:"questions"`
}

type QuizQuestion struct {
	ID            string   `json:"id"`
	Question      string   `json:"question"`
	Type          string   `json:"type"` // "multiple_choice", "true_false"
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correctAnswer"`
}

type FindTopicRequest struct {
	Transcript string `json:"transcript"`
	Topic      string `json:"topic"`
}

type FindTopicResponse struct {
	Results TopicResults `json:"results"`
}

type TopicResults struct {
	Timestamps []int    `json:"timestamps"`
	Context    []string `json:"context"`
}

// ============================================================================
// Project Assistant Agent Types
// ============================================================================

type ProjectsChatRequest struct {
	Message        string                 `json:"message"`
	UserID         string                 `json:"userId"`
	UserName       string                 `json:"userName"`
	ChatHistory    []ChatMessage          `json:"chatHistory"`
	ProjectContext map[string]interface{} `json:"projectContext"`
}

type ProjectsChatResponse struct {
	Response    string   `json:"response"`
	Suggestions []string `json:"suggestions"`
}

type ProjectsSuggestionsRequest struct {
	UserID           string   `json:"userId"`
	Interests        []string `json:"interests"`
	CareerGoals      string   `json:"careerGoals"`
	CompletedCourses []string `json:"completedCourses"`
	SkillLevel       string   `json:"skillLevel"` // "beginner", "intermediate", "advanced"
}

type ProjectsSuggestionsResponse struct {
	Suggestions []ProjectSuggestion `json:"suggestions"`
}

type ProjectSuggestion struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Difficulty     string   `json:"difficulty"` // "beginner", "intermediate", "advanced"
	Technologies   []string `json:"technologies"`
	EstimatedHours int      `json:"estimatedHours"`
	Reasoning      string   `json:"reasoning"`
}

type ProjectsHelpRequest struct {
	UserID      string                 `json:"userId"`
	Project     map[string]interface{} `json:"project"`
	Question    string                 `json:"question"`
	ChatHistory []ChatMessage          `json:"chatHistory"`
}

type ProjectsHelpResponse struct {
	Response  string     `json:"response"`
	Resources []Resource `json:"resources"`
	NextSteps []string   `json:"nextSteps"`
}

type ProjectsBreakDownRequest struct {
	UserID  string                 `json:"userId"`
	Project map[string]interface{} `json:"project"`
}

type ProjectsBreakDownResponse struct {
	Milestones []ProjectMilestone `json:"milestones"`
}

type ProjectMilestone struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	EstimatedHours int      `json:"estimatedHours"`
	Tasks          []string `json:"tasks"`
	Deliverables   []string `json:"deliverables"`
}

type ProjectsResourcesRequest struct {
	UserID    string                 `json:"userId"`
	Project   map[string]interface{} `json:"project"`
	Milestone map[string]interface{} `json:"milestone"`
}

type ProjectsResourcesResponse struct {
	Resources []ProjectResource `json:"resources"`
}

type ProjectResource struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Type        string `json:"type"` // "tutorial", "documentation", "article", "video"
	Relevance   int    `json:"relevance"`
	Description string `json:"description"`
}
