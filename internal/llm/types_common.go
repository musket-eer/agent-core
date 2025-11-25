package llm

// ============================================================================
// Common Types
// ============================================================================

type ChatMessage struct {
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

type UserProfile struct {
	Name           string   `json:"name"`
	EducationLevel string   `json:"education_level"`
	CareerGoals    string   `json:"careerGoals"`
	Interests      []string `json:"interests"`
	LearningStyle  string   `json:"learningStyle,omitempty"`
	CurrentSkills  []string `json:"currentSkills,omitempty"`
	LearningGoals  []string `json:"learningGoals,omitempty"`
}

type CourseDetails struct {
	ID              string   `json:"id,omitempty"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	DifficultyLevel string   `json:"difficulty_level"`
	Topics          []string `json:"topics,omitempty"`
	Skills          []string `json:"skills,omitempty"`
	Duration        int      `json:"duration,omitempty"`
	Prerequisites   []string `json:"prerequisites,omitempty"`
	Progress        int      `json:"progress,omitempty"`
}

type Activity struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type,omitempty"`
	Completed   bool   `json:"completed,omitempty"`
	CompletedAt string `json:"completedAt,omitempty"`
	Difficulty  string `json:"difficulty,omitempty"`
	Duration    int    `json:"duration,omitempty"`
	Date        string `json:"date,omitempty"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type Resource struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Link        string `json:"link,omitempty"`
	URL         string `json:"url,omitempty"`
	Type        string `json:"type,omitempty"` // "article", "video", "exercise"
}
