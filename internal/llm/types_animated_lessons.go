package llm

type AnimatedLessonRequest struct {
	Topic           string         `json:"topic"`
	CourseTitle     string         `json:"courseTitle"`
	LessonObjective string         `json:"lessonObjective"`
	Learner         LearnerProfile `json:"learner"`
	Animation       AnimationPrefs `json:"animation"`
}

type LearnerProfile struct {
	Name           string `json:"name"`
	EducationLevel string `json:"educationLevel"`
	Interests      string `json:"interests"`
	CareerGoals    string `json:"careerGoals"`
}

type AnimationPrefs struct {
	Mode                     string `json:"mode"`
	RequireLevelOrder        bool   `json:"requireLevelOrder"`
	IncludePrerequisiteEdges bool   `json:"includePrerequisiteEdges"`
}

type KnowledgeNode struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	KeyPoints     []string `json:"keyPoints"`
	Examples      []string `json:"examples"`
	Prerequisites []string `json:"prerequisites"`
}

type AnimatedLessonResponse struct {
	LessonID       string          `json:"lessonId"`
	Title          string          `json:"title"`
	KnowledgeGraph []KnowledgeNode `json:"knowledgeGraph"`
}

type TopicRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type AnimatedSlideContentRequest struct {
	LessonID        string         `json:"lessonId"`
	CourseTitle     string         `json:"courseTitle"`
	LessonObjective string         `json:"lessonObjective"`
	Learner         LearnerProfile `json:"learner"`
	Animation       AnimationPrefs `json:"animation"`
	CurrentTopic    KnowledgeNode  `json:"currentTopic"`
	PreviousTopic   *TopicRef      `json:"previousTopic,omitempty"`
	NextTopic       *TopicRef      `json:"nextTopic,omitempty"`
}

type AnimatedSlideContentResponse struct {
	TopicID            string      `json:"topicId"`
	Explanation        string      `json:"explanation"`
	KeyPoints          []string    `json:"keyPoints"`
	Examples           []string    `json:"examples"`
	BridgeFromPrevious string      `json:"bridgeFromPrevious"`
	BridgeToNext       string      `json:"bridgeToNext"`
	DemoPlan           *DemoPlan   `json:"demoPlan,omitempty"`
	ImageURL           string      `json:"imageUrl,omitempty"`
	Equations          []string    `json:"equations,omitempty"`
	SketchPath         *SketchPath `json:"sketchPath,omitempty"`
}

type DemoPlan struct {
	Kind              string `json:"kind"` // NONE|SKETCH_PATH|EQUATION_REVEAL|IMAGE_DRAW|PHYSICS_GRAPH
	Reason            string `json:"reason,omitempty"`
	AnimationEmphasis string `json:"animationEmphasis,omitempty"` // progressive_draw|step_trace|settle_motion|reveal
}

type SketchPath struct {
	Title string       `json:"title"`
	Nodes []SketchNode `json:"nodes"`
	Edges []SketchEdge `json:"edges"`
	Draw  SketchDraw   `json:"draw"`
}

type SketchNode struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	XK    float64 `json:"xk"`
	YK    float64 `json:"yk"`
}

type SketchEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

type SketchDraw struct {
	SecPerEdge     float64 `json:"secPerEdge"`
	SecPerNode     float64 `json:"secPerNode"`
	ShowArrowheads bool    `json:"showArrowheads"`
}
