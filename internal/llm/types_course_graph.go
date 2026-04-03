package llm

type GenerateCourseGraphRequest struct {
	CourseRequest    CourseGraphInputCourse    `json:"courseRequest"`
	LearnerProfile   CourseGraphLearnerProfile `json:"learnerProfile"`
	GenerationPolicy CourseGraphPolicy         `json:"generationPolicy"`
}

type CourseGraphInputCourse struct {
	Title             string   `json:"title"`
	Topic             string   `json:"topic"`
	LearnerQuestion   string   `json:"learnerQuestion"`
	Objectives        []string `json:"objectives"`
	PriorKnowledge    []string `json:"priorKnowledge"`
	DifficultyLevel   string   `json:"difficultyLevel"`
	WeeklyHours       int      `json:"weeklyHours"`
	TargetLessonCount int      `json:"targetLessonCount"`
	ReferenceBooks    []string `json:"referenceBooks"`
}

type CourseGraphLearnerProfile struct {
	Name           string `json:"name"`
	EducationLevel string `json:"educationLevel"`
	CareerGoals    string `json:"careerGoals"`
	Interests      string `json:"interests"`
}

type CourseGraphPolicy struct {
	Mode                        string                     `json:"mode"`
	RequireTopicAnchoring       bool                       `json:"requireTopicAnchoring"`
	RequireObjectiveCoverage    bool                       `json:"requireObjectiveCoverage"`
	RequirePrerequisiteOrdering bool                       `json:"requirePrerequisiteOrdering"`
	Personalization             CourseGraphPersonalization `json:"personalization"`
}

type CourseGraphPersonalization struct {
	InterestsRole         string  `json:"interestsRole"`
	MaxInterestInfluence  float64 `json:"maxInterestInfluence"`
	DoNotReplaceCoreTopic bool    `json:"doNotReplaceCoreTopic"`
}

type GenerateCourseGraphResponse struct {
	Course CourseGraphCourse `json:"course"`
}

type CourseGraphCourse struct {
	ID              string              `json:"id"`
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	Objectives      []string            `json:"objectives"`
	PriorKnowledge  []string            `json:"priorKnowledge"`
	DifficultyLevel string              `json:"difficultyLevel"`
	WeeklyHours     int                 `json:"weeklyHours"`
	Lessons         int                 `json:"lessons"`
	ReferenceBooks  []string            `json:"referenceBooks"`
	Units           []CourseGraphUnit   `json:"units"`
	KnowledgeGraph  []CourseGraphNode   `json:"knowledgeGraph"`
	Metadata        CourseGraphMetadata `json:"metadata"`
}

type CourseGraphUnit struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Objective      string `json:"objective"`
	Assessment     string `json:"assessment"`
	Position       int    `json:"position"`
	KGTopicID      string `json:"kgTopicId"`
	DeepDivePrompt string `json:"deepDivePrompt"`
}

type CourseGraphNode struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	KeyPoints      []string `json:"keyPoints"`
	Examples       []string `json:"examples"`
	Prerequisites  []string `json:"prerequisites"`
	Order          int      `json:"order"`
	DeepDivePrompt string   `json:"deepDivePrompt,omitempty"`
}

type CourseGraphMetadata struct {
	GenerationMode string `json:"generationMode"`
	InterestUsage  string `json:"interestUsage"`
}

type GenerateLessonIndexPlanRequest struct {
	Mode    string                    `json:"mode"`
	Course  LessonIndexCourseContext  `json:"course"`
	Lesson  LessonIndexLessonContext  `json:"lesson"`
	Learner CourseGraphLearnerProfile `json:"learner"`
}

type LessonIndexCourseContext struct {
	ID              string              `json:"id"`
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	DifficultyLevel string              `json:"difficultyLevel"`
	Objectives      []string            `json:"objectives"`
	KnowledgeGraph  []LessonIndexKGNode `json:"knowledgeGraph"`
}

type LessonIndexKGNode struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Prerequisites []string `json:"prerequisites,omitempty"`
}

type LessonIndexLessonContext struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Objective      string `json:"objective"`
	KGTopicID      string `json:"kgTopicId"`
	DeepDivePrompt string `json:"deepDivePrompt"`
}

type GenerateLessonIndexPlanResponse struct {
	LessonIndexPlan LessonIndexPlan `json:"lessonIndexPlan"`
}

type LessonIndexPlan struct {
	LessonTitle     string            `json:"lessonTitle"`
	LessonObjective string            `json:"lessonObjective"`
	LessonIndex     []LessonIndexItem `json:"lessonIndex"`
}

type LessonIndexItem struct {
	ID       string              `json:"id"`
	Title    string              `json:"title"`
	Leader   string              `json:"leader"`
	DeepDive LessonIndexDeepDive `json:"deepDive"`
}

type LessonIndexDeepDive struct {
	Summary       string   `json:"summary"`
	LearningGoals []string `json:"learningGoals"`
	KeyQuestions  []string `json:"keyQuestions"`
	Deliverables  []string `json:"deliverables"`
	References    []string `json:"references,omitempty"`
}
