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
	Canvas             *CanvasNodeSpec         `json:"canvas,omitempty"`
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
	Points []DiagramGraphNode       `json:"points"`
	Curves []DiagramCoordinateCurve `json:"curves"`
}

type DiagramCoordinateCurve struct {
	ID     string                `json:"id"`
	Points []DiagramCoordinatePt `json:"points"`
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

// CanvasArtifactRendererKind mirrors the low-level registered artifact renderers
// in the frontend canvas engine artifact registry.
type CanvasArtifactRendererKind string

const (
	CanvasArtifactRendererChalkTextBlock CanvasArtifactRendererKind = "chalkTextBlock"
	CanvasArtifactRendererPrimitive2D    CanvasArtifactRendererKind = "primitive2d"
)

// Primitive2DKind represents the supported primitive2d renderer shapes.
type Primitive2DKind string

const (
	Primitive2DCircle   Primitive2DKind = "circle"
	Primitive2DLine     Primitive2DKind = "line"
	Primitive2DArrow    Primitive2DKind = "arrow"
	Primitive2DRect     Primitive2DKind = "rect"
	Primitive2DPolyline Primitive2DKind = "polyline"
	Primitive2DSpring   Primitive2DKind = "spring"
	Primitive2DPendulum Primitive2DKind = "pendulum"
)

// CanvasArtifactAnimationKind mirrors the low-level registered artifact animations
// in the frontend canvas engine artifact registry.
type CanvasArtifactAnimationKind string

const (
	CanvasArtifactAnimationTextTimeline CanvasArtifactAnimationKind = "textTimeline"
	CanvasArtifactAnimationKinematics2D CanvasArtifactAnimationKind = "kinematics2d"
)

// Kinematics2DMotionKind represents supported motions in the kinematics2d animator.
type Kinematics2DMotionKind string

const (
	Kinematics2DFreeFall          Kinematics2DMotionKind = "free_fall"
	Kinematics2DBounceBall        Kinematics2DMotionKind = "bounce_ball"
	Kinematics2DLightPulse        Kinematics2DMotionKind = "light_pulse"
	Kinematics2DWavefrontExpand   Kinematics2DMotionKind = "wavefront_expand"
	Kinematics2DPendulumSwing     Kinematics2DMotionKind = "pendulum_swing"
	Kinematics2DSpringOscillation Kinematics2DMotionKind = "spring_oscillation"
)

// CanvasLessonNodeKind mirrors the higher-level lesson node types registered in
// the frontend canvas engine runtime.
type CanvasLessonNodeKind string

const (
	CanvasLessonNodeText               CanvasLessonNodeKind = "TEXT"
	CanvasLessonNodeDivider            CanvasLessonNodeKind = "DIVIDER"
	CanvasLessonNodeTableFill          CanvasLessonNodeKind = "TABLE_FILL"
	CanvasLessonNodeDiagramStroke      CanvasLessonNodeKind = "DIAGRAM_STROKE"
	CanvasLessonNodeMultiRegion        CanvasLessonNodeKind = "MULTI_REGION"
	CanvasLessonNodeMultiColumn        CanvasLessonNodeKind = "MULTI_COLUMN"
	CanvasLessonNodePhysicsDiagram     CanvasLessonNodeKind = "PHYSICS_DIAGRAM"
	CanvasLessonNodePedagogicalDiagram CanvasLessonNodeKind = "PEDAGOGICAL_DIAGRAM"
	CanvasLessonNodeRegionCompose      CanvasLessonNodeKind = "REGION_COMPOSE"
	CanvasLessonNodeImageDraw          CanvasLessonNodeKind = "IMAGE_DRAW"
	CanvasLessonNodeSketchPath         CanvasLessonNodeKind = "SKETCH_PATH"
	CanvasLessonNodeEquation           CanvasLessonNodeKind = "EQUATION"
	CanvasLessonNodeBoxedDemo          CanvasLessonNodeKind = "BOXED_DEMO"
)

// CanvasArtifactSpec provides a backend-side schema for low-level canvas artifacts.
type CanvasArtifactSpec struct {
	Renderer  *CanvasRendererSpec  `json:"renderer,omitempty"`
	Animation *CanvasAnimationSpec `json:"animation,omitempty"`
}

type CanvasNodeSpec struct {
	Kind      CanvasLessonNodeKind `json:"kind"`
	Artifacts []CanvasArtifactSpec `json:"artifacts,omitempty"`
}

type CanvasRendererSpec struct {
	Kind        CanvasArtifactRendererKind `json:"kind"`
	ChalkText   *ChalkTextBlockSpec        `json:"chalkText,omitempty"`
	Primitive2D *Primitive2DSpec           `json:"primitive2d,omitempty"`
}

type ChalkTextBlockSpec struct {
	Text      string   `json:"text,omitempty"`
	Lines     []string `json:"lines,omitempty"`
	Highlight []int    `json:"highlight,omitempty"`
}

type Primitive2DSpec struct {
	Kind       Primitive2DKind        `json:"kind"`
	Label      string                 `json:"label,omitempty"`
	Style      map[string]interface{} `json:"style,omitempty"`
	Points     []DiagramCoordinatePt  `json:"points,omitempty"`
	Anchor     *DiagramCoordinatePt   `json:"anchor,omitempty"`
	From       *DiagramCoordinatePt   `json:"from,omitempty"`
	To         *DiagramCoordinatePt   `json:"to,omitempty"`
	Center     *DiagramCoordinatePt   `json:"center,omitempty"`
	RadiusK    float64                `json:"radiusK,omitempty"`
	WidthK     float64                `json:"widthK,omitempty"`
	HeightK    float64                `json:"heightK,omitempty"`
	AmplitudeK float64                `json:"amplitudeK,omitempty"`
	LengthK    float64                `json:"lengthK,omitempty"`
}

type CanvasAnimationSpec struct {
	Kind         CanvasArtifactAnimationKind `json:"kind"`
	TextTimeline *TextTimelineSpec           `json:"textTimeline,omitempty"`
	Kinematics2D *Kinematics2DSpec           `json:"kinematics2d,omitempty"`
}

type TextTimelineSpec struct {
	Mode        string   `json:"mode,omitempty"`
	Steps       []string `json:"steps,omitempty"`
	RevealOrder []int    `json:"revealOrder,omitempty"`
}

type Kinematics2DSpec struct {
	Motion      Kinematics2DMotionKind `json:"motion"`
	DurationSec float64                `json:"durationSec,omitempty"`
	Loop        bool                   `json:"loop,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
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
