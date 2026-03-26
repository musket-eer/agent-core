package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

type AnimatedLessonRequest struct {
	Topic            string         `json:"topic"`
	CourseTitle      string         `json:"courseTitle"`
	LessonObjective  string         `json:"lessonObjective"`
	Learner          LearnerProfile `json:"learner"`
	Animation        AnimationPrefs `json:"animation"`
	PresentationMode string         `json:"presentationMode,omitempty"`
	ScenePolicy      ScenePolicy    `json:"scenePolicy,omitempty"`
}

type LearnerProfile struct {
	Name           string `json:"name"`
	EducationLevel string `json:"educationLevel"`
	Interests      string `json:"interests"`
	CareerGoals    string `json:"careerGoals"`
}

func (lp *LearnerProfile) UnmarshalJSON(data []byte) error {
	type alias struct {
		Name           string          `json:"name"`
		EducationLevel string          `json:"educationLevel"`
		Interests      json.RawMessage `json:"interests"`
		CareerGoals    json.RawMessage `json:"careerGoals"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	interests, err := decodeFlexibleString(a.Interests)
	if err != nil {
		return fmt.Errorf("learner.interests: %w", err)
	}
	careerGoals, err := decodeFlexibleString(a.CareerGoals)
	if err != nil {
		return fmt.Errorf("learner.careerGoals: %w", err)
	}

	lp.Name = strings.TrimSpace(a.Name)
	lp.EducationLevel = strings.TrimSpace(a.EducationLevel)
	lp.Interests = interests
	lp.CareerGoals = careerGoals
	return nil
}

func decodeFlexibleString(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}

	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return strings.TrimSpace(single), nil
	}

	var many []string
	if err := json.Unmarshal(raw, &many); err == nil {
		out := make([]string, 0, len(many))
		for _, v := range many {
			tv := strings.TrimSpace(v)
			if tv != "" {
				out = append(out, tv)
			}
		}
		return strings.Join(out, ", "), nil
	}

	return "", fmt.Errorf("must be string or string[]")
}

type AnimationPrefs struct {
	Mode                     string `json:"mode"`
	RequireLevelOrder        bool   `json:"requireLevelOrder"`
	IncludePrerequisiteEdges bool   `json:"includePrerequisiteEdges"`
}

type KnowledgeNode struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Type          string   `json:"type,omitempty"`
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
	LessonID         string         `json:"lessonId"`
	CourseTitle      string         `json:"courseTitle"`
	LessonObjective  string         `json:"lessonObjective"`
	Learner          LearnerProfile `json:"learner"`
	Animation        AnimationPrefs `json:"animation"`
	CurrentTopic     KnowledgeNode  `json:"currentTopic"`
	PreviousTopic    *TopicRef      `json:"previousTopic,omitempty"`
	NextTopic        *TopicRef      `json:"nextTopic,omitempty"`
	PresentationMode string         `json:"presentationMode,omitempty"`
	ScenePolicy      ScenePolicy    `json:"scenePolicy,omitempty"`
}

type AnimatedSlideContentResponse struct {
	TopicID            string                  `json:"topicId"`
	Explanation        string                  `json:"explanation"`
	KeyPoints          []string                `json:"keyPoints"`
	Examples           []string                `json:"examples"`
	BridgeFromPrevious string                  `json:"bridgeFromPrevious"`
	BridgeToNext       string                  `json:"bridgeToNext"`
	DemoPlan           *DemoPlan               `json:"demoPlan,omitempty"`
	ImageURL           string                  `json:"imageUrl,omitempty"`
	Equations          []string                `json:"equations,omitempty"`
	SketchPath         *SketchPath             `json:"sketchPath,omitempty"`
	PedagogicalDiagram *PedagogicalDiagramNode `json:"pedagogicalDiagram,omitempty"`
	Scene              *SceneSpec              `json:"scene,omitempty"`
}

type DemoPlan struct {
	Kind              string `json:"kind"` // NONE|SKETCH_PATH|EQUATION_REVEAL|IMAGE_DRAW|PHYSICS_GRAPH|PHYSICS_DIAGRAM|PEDAGOGICAL_DIAGRAM
	Reason            string `json:"reason,omitempty"`
	AnimationEmphasis string `json:"animationEmphasis,omitempty"` // progressive_draw|step_trace|settle_motion|reveal|pendulum_swing|spring_oscillation
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

type SceneSpec struct {
	Layout        *SceneLayout         `json:"layout,omitempty"`
	AnimationPlan []SceneAnimationStep `json:"animationPlan,omitempty"`
	Timeline      *SceneTimeline       `json:"timeline,omitempty"`
	Physics       *ScenePhysics        `json:"physics,omitempty"`
	Layers        *SceneLayers         `json:"layers,omitempty"`
}

type SceneLayout struct {
	Mode   string                 `json:"mode"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type SceneAnimationStep struct {
	Step          int     `json:"step"`
	Target        string  `json:"target"`
	Effect        string  `json:"effect"`
	DurationSec   float64 `json:"durationSec,omitempty"`
	WaitForSignal bool    `json:"waitForSignal,omitempty"`
}

type SceneTimeline struct {
	WaitForAnimationPlan bool    `json:"waitForAnimationPlan,omitempty"`
	RequireSignalDone    bool    `json:"requireSignalDone,omitempty"`
	MaxWaitSec           float64 `json:"maxWaitSec,omitempty"`
}

type ScenePhysics struct {
	Enabled bool                   `json:"enabled,omitempty"`
	Mode    string                 `json:"mode,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type SceneLayers struct {
	Background bool `json:"background,omitempty"`
	Content    bool `json:"content,omitempty"`
	Highlight  bool `json:"highlight,omitempty"`
	Transition bool `json:"transition,omitempty"`
	Debug      bool `json:"debug,omitempty"`
}

type ScenePolicy struct {
	RequireScene         bool `json:"requireScene,omitempty"`
	RequireTimeline      bool `json:"requireTimeline,omitempty"`
	RequireAnimationPlan bool `json:"requireAnimationPlan,omitempty"`
}

type PedagogicalDiagramNode struct {
	ID      string                    `json:"id"`
	Kind    string                    `json:"kind"`
	Deps    []string                  `json:"deps,omitempty"`
	Payload PedagogicalDiagramPayload `json:"payload"`
	Scene   *SceneSpec                `json:"scene,omitempty"`
}

type PedagogicalDiagramPayload struct {
	Variant     string  `json:"variant"`
	Pattern     string  `json:"pattern,omitempty"`
	Title       string  `json:"title,omitempty"`
	Subtitle    string  `json:"subtitle,omitempty"`
	DurationSec float64 `json:"durationSec,omitempty"`

	Shapes          []DiagramShape          `json:"shapes,omitempty"`
	Characters      []DiagramCharacter      `json:"characters,omitempty"`
	Table           *DiagramTable           `json:"table,omitempty"`
	Chart           *DiagramChart           `json:"chart,omitempty"`
	Graph           *DiagramGraph           `json:"graph,omitempty"`
	Timeline        *DiagramTimeline        `json:"timeline,omitempty"`
	Tree            *DiagramTree            `json:"tree,omitempty"`
	StackQueue      *DiagramStackQueue      `json:"stackQueue,omitempty"`
	CoordinatePlane *DiagramCoordinatePlane `json:"coordinatePlane,omitempty"`
	FlowDiagram     *DiagramFlowDiagram     `json:"flowDiagram,omitempty"`
	SetDiagram      *DiagramSetDiagram      `json:"setDiagram,omitempty"`
	Heatmap         *DiagramHeatmap         `json:"heatmap,omitempty"`
	Field           *DiagramField           `json:"field,omitempty"`
	Chat            *DiagramChat            `json:"chat,omitempty"`
}

type DiagramShape struct {
	Kind  string  `json:"kind"`
	XK    float64 `json:"xk"`
	YK    float64 `json:"yk"`
	WK    float64 `json:"wk"`
	HK    float64 `json:"hk"`
	Label string  `json:"label,omitempty"`
}

type DiagramCharacter struct {
	Name string  `json:"name"`
	XK   float64 `json:"xk"`
	YK   float64 `json:"yk"`
	Pose string  `json:"pose,omitempty"`
}

type DiagramTable struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
	Caption string     `json:"caption,omitempty"`
}

type DiagramChart struct {
	Type   string               `json:"type"`
	Labels []string             `json:"labels"`
	Series []DiagramChartSeries `json:"series"`
}

type DiagramChartSeries struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
}

type DiagramGraph struct {
	Nodes []DiagramGraphNode `json:"nodes"`
	Edges []DiagramGraphEdge `json:"edges"`
}

type DiagramGraphNode struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	XK    float64 `json:"xk"`
	YK    float64 `json:"yk"`
}

type DiagramGraphEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Label    string `json:"label,omitempty"`
	Directed bool   `json:"directed,omitempty"`
}

type DiagramTimeline struct {
	Events []DiagramTimelineEvent `json:"events"`
}

type DiagramTimelineEvent struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	AtK   float64 `json:"atK"`
}

type DiagramTree struct {
	Nodes []DiagramGraphNode `json:"nodes"`
	Edges []DiagramGraphEdge `json:"edges"`
}

type DiagramStackQueue struct {
	Mode  string   `json:"mode"`
	Items []string `json:"items"`
}

type DiagramCoordinatePlane struct {
	Points []DiagramGraphNode         `json:"points"`
	Curves []DiagramCoordinateCurve   `json:"curves"`
}

type DiagramCoordinateCurve struct {
	ID     string                 `json:"id"`
	Points []DiagramCoordinatePt  `json:"points"`
}

type DiagramCoordinatePt struct {
	XK float64 `json:"xk"`
	YK float64 `json:"yk"`
}

type DiagramFlowDiagram struct {
	Stages []DiagramFlowStage `json:"stages"`
	Links  []DiagramFlowLink  `json:"links"`
}

type DiagramFlowStage struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	XK    float64 `json:"xk"`
	YK    float64 `json:"yk"`
}

type DiagramFlowLink struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}

type DiagramSetDiagram struct {
	Sets []DiagramSet `json:"sets"`
}

type DiagramSet struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	XK    float64 `json:"xk"`
	YK    float64 `json:"yk"`
	RK    float64 `json:"rk"`
}

type DiagramHeatmap struct {
	LabelsX []string    `json:"labelsX"`
	LabelsY []string    `json:"labelsY"`
	Values  [][]float64 `json:"values"`
}

type DiagramField struct {
	Vectors []DiagramVector `json:"vectors"`
}

type DiagramVector struct {
	ID  string  `json:"id"`
	XK  float64 `json:"xk"`
	YK  float64 `json:"yk"`
	DXK float64 `json:"dxk"`
	DYK float64 `json:"dyk"`
}

type DiagramChat struct {
	Participants []DiagramParticipant `json:"participants"`
	Messages     []DiagramMessage     `json:"messages"`
}

type DiagramParticipant struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	XK   float64 `json:"xk"`
}

type DiagramMessage struct {
	SpeakerID string  `json:"speakerId"`
	Text      string  `json:"text"`
	AtSec     float64 `json:"atSec"`
}

type AnimatedTreeSlideResponse struct {
	LessonID   string          `json:"lessonId"`
	Title      string          `json:"title"`
	TreeSlides []TreeSlideNode `json:"treeSlides"`
}

type TreeSlideNode struct {
	ID       string          `json:"id"`
	Title    string          `json:"title"`
	Type     string          `json:"type"`
	Physics  *ScenePhysics   `json:"physics"`
	Timeline *SceneTimeline  `json:"timeline"`
	Children []TreeSlideNode `json:"children"`
}
