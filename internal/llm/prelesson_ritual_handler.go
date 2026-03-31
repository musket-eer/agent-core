package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"

	"github.com/google/uuid"
)

const generatePrelessonRitualEndpoint = "/generate-prelesson-ritual"

const defaultPrelessonRitualPrompt = `You generate a short pre-lesson ritual as strict JSON only.

This ritual appears before the real lesson and must feel like part of the lesson journey.
It is not the lesson itself. It should establish readiness, reduce stress, and direct attention.

Return exactly this shape:
{
  "title": "string",
  "subtitle": "string",
  "estimatedDurationSec": 24,
  "steps": [
    {
      "id": "step_focus",
      "kind": "focus|orientation|readiness",
      "headline": "string",
      "body": "string"
    }
  ],
  "closingCue": "string"
}

Rules:
- Return only student-facing ritual content.
- Do not mention AI, generation, loading, latency, backend systems, or hidden evaluation logic.
- steps length must be between 3 and 5.
- headline must be short.
- body must be concise, concrete, and actionable.
- estimatedDurationSec should stay between 18 and 40.
- no quizzes, no required learner input, no free-response tasks.
- calm, encouraging, anticipatory tone.
- frame attention; do not teach the whole lesson.

Priority order:
1. lessonType
2. unitObjective
3. unitTitle
4. lessonTitle
5. courseTitle
6. learnerProfile

Type guidance:
- exam: prime recall, reduce panic, focus on structure and careful reading.
- text: orient the learner to relationships, definitions, and progression.
- visual: direct attention to patterns, contrasts, and structures.
- scene: direct attention to transitions, changes, and sequences unfolding over time.`

var allowedPrelessonTypes = map[string]struct{}{
	"exam":   {},
	"text":   {},
	"visual": {},
	"scene":  {},
}

type GeneratePrelessonRitualRequest struct {
	LessonType     string                 `json:"lessonType"`
	CourseID       string                 `json:"courseId,omitempty"`
	CourseTitle    string                 `json:"courseTitle,omitempty"`
	UnitID         string                 `json:"unitId,omitempty"`
	UnitTitle      string                 `json:"unitTitle,omitempty"`
	UnitObjective  string                 `json:"unitObjective,omitempty"`
	LessonTitle    string                 `json:"lessonTitle,omitempty"`
	LearnerProfile map[string]interface{} `json:"learnerProfile,omitempty"`
}

type PrelessonRitualResponse struct {
	RitualID             string              `json:"ritualId"`
	LessonType           string              `json:"lessonType"`
	Title                string              `json:"title"`
	Subtitle             string              `json:"subtitle"`
	EstimatedDurationSec int                 `json:"estimatedDurationSec"`
	Steps                []PrelessonStep     `json:"steps"`
	ClosingCue           string              `json:"closingCue"`
}

type PrelessonStep struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Headline string `json:"headline"`
	Body     string `json:"body"`
}

func GeneratePrelessonRitualHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeExamContractError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}

	var req GeneratePrelessonRitualRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeExamContractError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	req.normalize()
	if err := req.validate(); err != nil {
		writeExamContractError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	resp, err := generatePrelessonRitual(context.Background(), req)
	if err != nil {
		writeExamContractError(w, http.StatusBadGateway, "INTERNAL_ERROR", "failed to generate prelesson ritual")
		return
	}
	writeExamContractJSON(w, http.StatusOK, resp)
}

func (r *GeneratePrelessonRitualRequest) normalize() {
	r.LessonType = strings.ToLower(strings.TrimSpace(r.LessonType))
	r.CourseID = strings.TrimSpace(r.CourseID)
	r.CourseTitle = strings.TrimSpace(r.CourseTitle)
	r.UnitID = strings.TrimSpace(r.UnitID)
	r.UnitTitle = strings.TrimSpace(r.UnitTitle)
	r.UnitObjective = strings.TrimSpace(r.UnitObjective)
	r.LessonTitle = strings.TrimSpace(r.LessonTitle)
}

func (r GeneratePrelessonRitualRequest) validate() error {
	if _, ok := allowedPrelessonTypes[r.LessonType]; !ok {
		return fmt.Errorf("lessonType must be one of exam, text, visual, or scene")
	}
	return nil
}

func generatePrelessonRitual(ctx context.Context, req GeneratePrelessonRitualRequest) (PrelessonRitualResponse, error) {
	cfg, ok := configs.GetDeepseekConfig(generatePrelessonRitualEndpoint)
	if !ok {
		return PrelessonRitualResponse{}, fmt.Errorf("missing config")
	}
	client, err := getLLMClientByConfig(cfg)
	if err != nil {
		return PrelessonRitualResponse{}, err
	}

	systemPrompt := loadPrelessonRitualPrompt()
	userPrompt := buildPrelessonRitualPrompt(req)
	raw, err := callLLMWithTimeout(ctx, client, generatePrelessonRitualEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": cfg.Temperature,
		"max_tokens":  cfg.MaxTokens,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": systemPrompt + "\n\n" + userPrompt,
		},
	})
	if err != nil {
		return PrelessonRitualResponse{}, err
	}

	var out PrelessonRitualResponse
	if err := json.Unmarshal([]byte(utils.CleanJSONResponse(raw)), &out); err != nil {
		return PrelessonRitualResponse{}, err
	}
	out.RitualID = "ritual_" + uuid.NewString()
	out.LessonType = req.LessonType
	if err := validatePrelessonRitual(out); err != nil {
		return PrelessonRitualResponse{}, err
	}
	return out, nil
}

func loadPrelessonRitualPrompt() string {
	p, err := prompt.GetPrompt("prelessonRitualPrompt")
	if err != nil || strings.TrimSpace(p) == "" {
		return defaultPrelessonRitualPrompt
	}
	return p
}

func buildPrelessonRitualPrompt(req GeneratePrelessonRitualRequest) string {
	var sb strings.Builder
	sb.WriteString("Generate a compact pre-lesson ritual.\n")
	sb.WriteString("lessonType: ")
	sb.WriteString(req.LessonType)
	sb.WriteString("\n")
	if req.CourseTitle != "" || req.CourseID != "" {
		sb.WriteString("courseTitle: ")
		sb.WriteString(fallbackText(req.CourseTitle, "(not provided)"))
		sb.WriteString("\ncourseId: ")
		sb.WriteString(fallbackText(req.CourseID, "(not provided)"))
		sb.WriteString("\n")
	}
	if req.UnitTitle != "" || req.UnitID != "" || req.UnitObjective != "" {
		sb.WriteString("unitTitle: ")
		sb.WriteString(fallbackText(req.UnitTitle, "(not provided)"))
		sb.WriteString("\nunitId: ")
		sb.WriteString(fallbackText(req.UnitID, "(not provided)"))
		sb.WriteString("\nunitObjective: ")
		sb.WriteString(fallbackText(req.UnitObjective, "(not provided)"))
		sb.WriteString("\n")
	}
	if req.LessonTitle != "" {
		sb.WriteString("lessonTitle: ")
		sb.WriteString(req.LessonTitle)
		sb.WriteString("\n")
	}
	if len(req.LearnerProfile) > 0 {
		if b, err := json.Marshal(req.LearnerProfile); err == nil {
			sb.WriteString("learnerProfile: ")
			sb.Write(b)
			sb.WriteString("\n")
		}
	}
	sb.WriteString(`
Keep the ritual compact. Focus attention, orientation, and readiness. Do not ask the learner to answer anything.`)
	return sb.String()
}

func validatePrelessonRitual(resp PrelessonRitualResponse) error {
	resp.Title = strings.TrimSpace(resp.Title)
	resp.Subtitle = strings.TrimSpace(resp.Subtitle)
	resp.ClosingCue = strings.TrimSpace(resp.ClosingCue)
	if resp.Title == "" || resp.Subtitle == "" || resp.ClosingCue == "" {
		return fmt.Errorf("missing required ritual fields")
	}
	if resp.EstimatedDurationSec < 18 || resp.EstimatedDurationSec > 40 {
		return fmt.Errorf("estimatedDurationSec out of range")
	}
	if len(resp.Steps) < 3 || len(resp.Steps) > 5 {
		return fmt.Errorf("steps length out of range")
	}
	for i, step := range resp.Steps {
		if strings.TrimSpace(step.ID) == "" || strings.TrimSpace(step.Kind) == "" || strings.TrimSpace(step.Headline) == "" || strings.TrimSpace(step.Body) == "" {
			return fmt.Errorf("steps[%d] missing required fields", i)
		}
		if len([]rune(strings.TrimSpace(step.Headline))) > 48 {
			return fmt.Errorf("steps[%d].headline too long", i)
		}
		if containsProhibitedRitualText(step.Headline) || containsProhibitedRitualText(step.Body) {
			return fmt.Errorf("steps[%d] contains prohibited language", i)
		}
	}
	if containsProhibitedRitualText(resp.Title) || containsProhibitedRitualText(resp.Subtitle) || containsProhibitedRitualText(resp.ClosingCue) {
		return fmt.Errorf("ritual contains prohibited language")
	}
	return nil
}

func containsProhibitedRitualText(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return false
	}
	for _, banned := range []string{
		" ai ",
		"generation",
		"loading",
		"latency",
		"backend",
		"hidden evaluation",
		"grading logic",
		"quiz",
		"free-response",
	} {
		if strings.Contains(" "+normalized+" ", banned) {
			return true
		}
	}
	return false
}
