package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"agent-core/internal/cache"
	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"

	"github.com/google/uuid"
)

const (
	generateExamContractEndpoint = "/generate-exam"
	examBundleTTLSeconds         = 30 * 24 * 60 * 60
	defaultExamAttemptCap        = 3
	maxLLMGradingInputChars      = 2000
)

var examQuestionIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-]{2,64}$`)

var allowedQuestionTypes = map[string]struct{}{
	"single_choice": {},
	"multi_select":  {},
	"short_text":    {},
	"long_text":     {},
	"numeric":       {},
	"boolean":       {},
	"code":          {},
}

var allowedQuestionDifficulty = map[string]struct{}{
	"easy":   {},
	"medium": {},
	"hard":   {},
}

const defaultMostLearnedExamPrompt = `You generate exam payloads in strict JSON.
Return this schema exactly:
{
  "title": string,
  "instructions": string,
  "questions": [
    {
      "id": string,
      "type": "single_choice|multi_select|short_text|long_text|numeric|boolean|code",
      "prompt": string,
      "points": number,
      "required": boolean,
      "difficulty": "easy|medium|hard",
      "options": string[],
      "ui": {
        "placeholder": string,
        "rows": number,
        "min": number,
        "max": number
      },
      "metadata": {},
      "answerKey": any,
      "rubric": string[],
      "modelAnswer": string
    }
  ]
}
Rules:
- keep topic fidelity
- produce exactly requested question count
- include answerKey for every question
- no markdown, no commentary, no extra keys`

type GenerateExamContractRequest struct {
	SessionID     string `json:"sessionId"`
	Topic         string `json:"topic"`
	QuestionCount int    `json:"questionCount"`
}

type SubmitSessionContractRequest struct {
	SessionID string                 `json:"sessionId"`
	ExamID    string                 `json:"examId"`
	UserID    string                 `json:"userId"`
	Answers   map[string]interface{} `json:"answers"`
}

type GenerateExamContractResponse struct {
	SessionID   string          `json:"sessionId"`
	ExamID      string          `json:"examId"`
	StudentView ExamStudentView `json:"studentView"`
}

type ExamStudentView struct {
	ExamID       string                `json:"examId"`
	Title        string                `json:"title"`
	Instructions string                `json:"instructions"`
	TotalPoints  int                   `json:"totalPoints"`
	Questions    []ExamStudentQuestion `json:"questions"`
}

type ExamStudentQuestion struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Prompt     string                 `json:"prompt"`
	Points     int                    `json:"points"`
	Required   bool                   `json:"required"`
	Difficulty string                 `json:"difficulty"`
	Options    []string               `json:"options,omitempty"`
	UI         map[string]interface{} `json:"ui,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type SubmitSessionContractResponse struct {
	SessionID   string                  `json:"sessionId"`
	ExamID      string                  `json:"examId"`
	Status      string                  `json:"status"`
	ScoreRaw    float64                 `json:"scoreRaw"`
	ScoreMax    float64                 `json:"scoreMax"`
	ScorePct    float64                 `json:"scorePct"`
	Grade       string                  `json:"grade"`
	PerQuestion []ExamQuestionResultDTO `json:"perQuestion"`
}

type ExamQuestionResultDTO struct {
	QuestionID     string      `json:"questionId"`
	MaxPoints      float64     `json:"maxPoints"`
	EarnedPoints   float64     `json:"earnedPoints"`
	Method         string      `json:"method"`
	Feedback       string      `json:"feedback"`
	ExpectedAnswer interface{} `json:"expectedAnswer"`
	StudentAnswer  interface{} `json:"studentAnswer"`
}

type examContractError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type llmExamQuestion struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Prompt      string                 `json:"prompt"`
	Points      int                    `json:"points"`
	Required    bool                   `json:"required"`
	Difficulty  string                 `json:"difficulty"`
	Options     []string               `json:"options"`
	UI          map[string]interface{} `json:"ui"`
	Metadata    map[string]interface{} `json:"metadata"`
	AnswerKey   interface{}            `json:"answerKey"`
	Rubric      []string               `json:"rubric"`
	ModelAnswer string                 `json:"modelAnswer"`
}

type llmExamEnvelope struct {
	Title        string            `json:"title"`
	Instructions string            `json:"instructions"`
	Questions    []llmExamQuestion `json:"questions"`
}

type gradingKey struct {
	Type        string      `json:"type"`
	Points      float64     `json:"points"`
	Required    bool        `json:"required"`
	Options     []string    `json:"options,omitempty"`
	AnswerKey   interface{} `json:"answerKey"`
	Rubric      []string    `json:"rubric,omitempty"`
	ModelAnswer string      `json:"modelAnswer,omitempty"`
}

type gradingBundle struct {
	ExamID      string                `json:"examId"`
	QuestionIDs []string              `json:"questionIds"`
	QuestionBy  map[string]gradingKey `json:"questionBy"`
	ScoreMax    float64               `json:"scoreMax"`
}

type generatedExamArtifact struct {
	ExamID        string          `json:"examId"`
	StudentView   ExamStudentView `json:"studentView"`
	GradingBundle gradingBundle   `json:"gradingBundle"`
}

type sessionExamMap struct {
	SessionID  string `json:"sessionId"`
	ExamID     string `json:"examId"`
	UserID     string `json:"userId,omitempty"`
	Attempts   int    `json:"attempts"`
	AttemptCap int    `json:"attemptCap"`
	CreatedAt  int64  `json:"createdAt"`
	ExpiresAt  int64  `json:"expiresAt"`
}

type enrichedSubmission struct {
	SessionID string
	ExamID    string
	UserID    string
	Answers   map[string]interface{}
	Bundle    gradingBundle
}

type gradingOutput struct {
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
}

type examSecretStore interface {
	StoreWithTTL(key string, value []byte, ttl int) error
	Retrieve(key string) ([]byte, error)
	Check(key string) (bool, error)
}

type upstashExamSecretStore struct{}

func (upstashExamSecretStore) StoreWithTTL(key string, value []byte, ttl int) error {
	if cache.Upstash == nil {
		return fmt.Errorf("upstash not configured")
	}
	return cache.Upstash.StoreWithTTL(key, value, ttl)
}

func (upstashExamSecretStore) Retrieve(key string) ([]byte, error) {
	if cache.Upstash == nil {
		return nil, fmt.Errorf("upstash not configured")
	}
	return cache.Upstash.Retrieve(key)
}

func (upstashExamSecretStore) Check(key string) (bool, error) {
	if cache.Upstash == nil {
		return false, fmt.Errorf("upstash not configured")
	}
	return cache.Upstash.Check(key)
}

var secretStore examSecretStore = upstashExamSecretStore{}

func GenerateExamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeExamContractError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}
	var req GenerateExamContractRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeExamContractError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Topic = strings.TrimSpace(req.Topic)
	if req.SessionID == "" || req.Topic == "" || req.QuestionCount < 1 || req.QuestionCount > 30 {
		writeExamContractError(w, http.StatusBadRequest, "INVALID_REQUEST", "sessionId, topic, and valid questionCount are required")
		return
	}
	requestID := requestIDFromHeader(r)

	cachePayload, _ := json.Marshal(struct {
		Topic         string `json:"topic"`
		QuestionCount int    `json:"questionCount"`
	}{Topic: strings.ToLower(req.Topic), QuestionCount: req.QuestionCount})
	cacheKey := generateExamContractEndpoint + ":" + string(cachePayload)

	var artifact generatedExamArtifact
	if cached, found, _ := cache.Get().CheckByPayload(cacheKey); found {
		if err := json.Unmarshal(cached, &artifact); err == nil {
			if err := persistExamSecretState(req.SessionID, "", artifact); err != nil {
				log.Printf("[GenerateExam] request_id=%s session_id=%s exam_id=%s persist_bundle=failed", requestID, req.SessionID, artifact.ExamID)
				writeExamContractError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to persist grading state")
				return
			}
			log.Printf("[GenerateExam] request_id=%s session_id=%s exam_id=%s source=cache generation=success", requestID, req.SessionID, artifact.ExamID)
			writeExamContractJSON(w, http.StatusOK, GenerateExamContractResponse{
				SessionID:   req.SessionID,
				ExamID:      artifact.ExamID,
				StudentView: artifact.StudentView,
			})
			return
		}
	}

	start := time.Now()
	artifact, model, validation, err := generateExamArtifact(context.Background(), req)
	log.Printf("[GenerateExam] request_id=%s session_id=%s exam_id=%s model=%s latency_ms=%d validation=%s",
		requestID, req.SessionID, artifact.ExamID, model, time.Since(start).Milliseconds(), validation)
	if err != nil {
		if strings.Contains(err.Error(), "LLM_INVALID_OUTPUT") {
			writeExamContractError(w, http.StatusBadGateway, "INTERNAL_ERROR", "failed to produce valid exam")
			return
		}
		writeExamContractError(w, http.StatusBadGateway, "INTERNAL_ERROR", "failed to generate exam")
		return
	}

	if payload, err := json.Marshal(artifact); err == nil {
		_ = cache.Get().CacheByPayload(cacheKey, payload)
	}

	if err := persistExamSecretState(req.SessionID, "", artifact); err != nil {
		log.Printf("[GenerateExam] request_id=%s session_id=%s exam_id=%s persist_bundle=failed", requestID, req.SessionID, artifact.ExamID)
		writeExamContractError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to persist grading state")
		return
	}
	log.Printf("[GenerateExam] request_id=%s session_id=%s exam_id=%s persist_bundle=ok generation=success", requestID, req.SessionID, artifact.ExamID)

	writeExamContractJSON(w, http.StatusOK, GenerateExamContractResponse{
		SessionID:   req.SessionID,
		ExamID:      artifact.ExamID,
		StudentView: artifact.StudentView,
	})
}

func SubmitSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeExamContractError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}

	var req SubmitSessionContractRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeExamContractError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.ExamID = strings.TrimSpace(req.ExamID)
	req.UserID = strings.TrimSpace(req.UserID)
	if req.SessionID == "" || req.ExamID == "" || req.UserID == "" || len(req.Answers) == 0 {
		writeExamContractError(w, http.StatusBadRequest, "INVALID_REQUEST", "sessionId, examId, userId, and answers are required")
		return
	}
	requestID := requestIDFromHeader(r)

	enriched, code, msg, err := enrichSubmission(req)
	if err != nil {
		log.Printf("[SubmitSession] request_id=%s session_id=%s exam_id=%s bundle_fetch=failed failure_code=%s",
			requestID, req.SessionID, req.ExamID, code)
		writeExamContractErrorByCode(w, code, msg)
		return
	}

	start := time.Now()
	resp, detCount, llmCount, code, msg, err := gradeSubmission(r.Context(), enriched)
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		log.Printf("[SubmitSession] request_id=%s session_id=%s exam_id=%s bundle_fetch=ok grading_latency_ms=%d deterministic=%d llm=%d failure_code=%s",
			requestID, enriched.SessionID, enriched.ExamID, latencyMs, detCount, llmCount, code)
		writeExamContractErrorByCode(w, code, msg)
		return
	}
	if err := validateGradedResponse(resp); err != nil {
		log.Printf("[SubmitSession] request_id=%s session_id=%s exam_id=%s failure_code=LLM_INVALID_OUTPUT err=%v",
			requestID, enriched.SessionID, enriched.ExamID, err)
		writeExamContractErrorByCode(w, "LLM_INVALID_OUTPUT", "graded response missing required fields")
		return
	}
	if err := persistExamResult(enriched, resp); err != nil {
		writeExamContractError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to persist result")
		return
	}
	if err := persistUserRollup(enriched.UserID, resp); err != nil {
		writeExamContractError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to persist progress")
		return
	}
	if err := updateSessionAttempt(enriched.SessionID, enriched.UserID, enriched.ExamID); err != nil {
		writeExamContractError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to persist attempts")
		return
	}

	log.Printf("[SubmitSession] request_id=%s session_id=%s exam_id=%s bundle_fetch=ok grading_latency_ms=%d deterministic=%d llm=%d failure_code=none",
		requestID, enriched.SessionID, enriched.ExamID, latencyMs, detCount, llmCount)
	writeExamContractJSON(w, http.StatusOK, resp)
}

// GetExamResultHandler returns latest safe graded result for session.
func GetExamResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeExamContractError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}
	sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId"))
	if sessionID == "" {
		writeExamContractError(w, http.StatusBadRequest, "INVALID_REQUEST", "sessionId is required")
		return
	}
	raw, err := retrieveOptionalKey("exam:latest_result:" + sessionID)
	if err != nil {
		writeExamContractError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to read exam result")
		return
	}
	if len(raw) == 0 {
		writeExamContractError(w, http.StatusNotFound, "INVALID_REQUEST", "result not found")
		return
	}
	var resp SubmitSessionContractResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		writeExamContractError(w, http.StatusInternalServerError, "GRADING_BUNDLE_CORRUPT", "result payload is corrupt")
		return
	}
	writeExamContractJSON(w, http.StatusOK, resp)
}

// GetExamProgressHandler returns safe progress rollup for a user.
func GetExamProgressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeExamContractError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}
	userID := strings.TrimSpace(r.URL.Query().Get("userId"))
	if userID == "" {
		writeExamContractError(w, http.StatusBadRequest, "INVALID_REQUEST", "userId is required")
		return
	}
	raw, err := retrieveOptionalKey("progress:rollup:user:" + userID)
	if err != nil {
		writeExamContractError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to read user progress")
		return
	}
	if len(raw) == 0 {
		writeExamContractJSON(w, http.StatusOK, map[string]interface{}{
			"userId":       userID,
			"attempts":     0,
			"averagePct":   0,
			"lastScorePct": 0,
		})
		return
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		writeExamContractError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "progress payload is corrupt")
		return
	}
	writeExamContractJSON(w, http.StatusOK, payload)
}

func generateExamArtifact(ctx context.Context, req GenerateExamContractRequest) (generatedExamArtifact, string, string, error) {
	cfg, ok := configs.GetDeepseekConfig(generateExamContractEndpoint)
	if !ok {
		cfg, _ = configs.GetDeepseekConfig("/generate-exams")
	}
	client, err := getLLMClientByConfig(cfg)
	if err != nil {
		return generatedExamArtifact{}, cfg.ModelId, "failed", err
	}

	systemPrompt := loadExamPrompt()
	userPrompt := buildExamPrompt(req)
	examID := "exam_" + uuid.NewString()
	seed := stableExamSeed(req.Topic, req.QuestionCount)

	raw, err := callLLMWithTimeout(ctx, client, generateExamContractEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  2600,
		"seed":        seed,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": systemPrompt + "\n\n" + userPrompt,
		},
	})
	if err != nil {
		return generatedExamArtifact{}, cfg.ModelId, "failed", err
	}

	artifact, violations, parseErr := parseExamOutput(raw, req, examID)
	if parseErr == nil && len(violations) == 0 {
		return artifact, cfg.ModelId, "pass", nil
	}

	repairPrompt := buildExamRepairPrompt(req, raw, violations, parseErr)
	repairedRaw, repairErr := callLLMWithTimeout(ctx, client, generateExamContractEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  2600,
		"seed":        seed + 1,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": systemPrompt + "\n\n" + repairPrompt,
		},
	})
	if repairErr != nil {
		return generatedExamArtifact{}, cfg.ModelId, "failed", repairErr
	}

	finalArtifact, finalViolations, finalParseErr := parseExamOutput(repairedRaw, req, examID)
	if finalParseErr != nil || len(finalViolations) > 0 {
		log.Printf("[GenerateExam] repair failed parse_error=%v violations=%v", finalParseErr, finalViolations)
		return generatedExamArtifact{}, cfg.ModelId, "failed", fmt.Errorf("LLM_INVALID_OUTPUT")
	}
	return finalArtifact, cfg.ModelId, "pass_after_repair", nil
}

func parseExamOutput(raw string, req GenerateExamContractRequest, examID string) (generatedExamArtifact, []string, error) {
	var out llmExamEnvelope
	cleaned := utils.CleanJSONResponse(raw)
	if err := json.Unmarshal([]byte(cleaned), &out); err != nil {
		return generatedExamArtifact{}, nil, err
	}

	student := ExamStudentView{
		ExamID:       examID,
		Title:        strings.TrimSpace(out.Title),
		Instructions: strings.TrimSpace(out.Instructions),
	}
	if student.Title == "" {
		student.Title = req.Topic + " Assessment"
	}
	if student.Instructions == "" {
		student.Instructions = "Answer all questions."
	}

	bundle := gradingBundle{
		ExamID:      examID,
		QuestionBy:  make(map[string]gradingKey, len(out.Questions)),
		QuestionIDs: make([]string, 0, len(out.Questions)),
	}

	var violations []string
	if len(out.Questions) != req.QuestionCount {
		violations = append(violations, fmt.Sprintf("question count mismatch: expected %d got %d", req.QuestionCount, len(out.Questions)))
	}

	seen := make(map[string]struct{}, len(out.Questions))
	for i, q := range out.Questions {
		id := strings.TrimSpace(q.ID)
		if id == "" {
			id = fmt.Sprintf("q_%d", i+1)
		}
		if !examQuestionIDPattern.MatchString(id) {
			violations = append(violations, fmt.Sprintf("questions[%d].id invalid", i))
		}
		if _, exists := seen[id]; exists {
			violations = append(violations, fmt.Sprintf("duplicate question id: %s", id))
		}
		seen[id] = struct{}{}

		typ := strings.TrimSpace(q.Type)
		if _, ok := allowedQuestionTypes[typ]; !ok {
			violations = append(violations, fmt.Sprintf("questions[%d].type invalid", i))
		}
		if strings.TrimSpace(q.Prompt) == "" {
			violations = append(violations, fmt.Sprintf("questions[%d].prompt missing", i))
		}
		if q.Points <= 0 || q.Points > 20 {
			violations = append(violations, fmt.Sprintf("questions[%d].points invalid", i))
		}

		diff := strings.ToLower(strings.TrimSpace(q.Difficulty))
		if _, ok := allowedQuestionDifficulty[diff]; !ok {
			diff = "medium"
		}

		options := normalizeExamOptions(q.Options)
		if (typ == "single_choice" || typ == "multi_select") && len(options) < 2 {
			violations = append(violations, fmt.Sprintf("questions[%d].options must have at least 2 items", i))
		}
		if typ != "single_choice" && typ != "multi_select" {
			options = nil
		}

		key := gradingKey{
			Type:        typ,
			Points:      float64(q.Points),
			Required:    q.Required,
			Options:     options,
			AnswerKey:   q.AnswerKey,
			Rubric:      q.Rubric,
			ModelAnswer: strings.TrimSpace(q.ModelAnswer),
		}
		key = projectGradingBundleKey(key)
		if missingRequiredGradingKey(key) {
			violations = append(violations, fmt.Sprintf("questions[%d] missing grading keys", i))
		}

		studentQuestion := projectStudentViewQuestion(ExamStudentQuestion{
			ID:         id,
			Type:       typ,
			Prompt:     strings.TrimSpace(q.Prompt),
			Points:     q.Points,
			Required:   q.Required,
			Difficulty: diff,
			Options:    options,
			UI:         sanitizeSecretMap(q.UI),
			Metadata:   sanitizeSecretMap(q.Metadata),
		})

		student.Questions = append(student.Questions, studentQuestion)
		bundle.QuestionIDs = append(bundle.QuestionIDs, id)
		bundle.QuestionBy[id] = key
		student.TotalPoints += q.Points
		bundle.ScoreMax += float64(q.Points)
	}

	if len(student.Questions) == 0 {
		violations = append(violations, "questions must not be empty")
	}

	return generatedExamArtifact{
		ExamID:        examID,
		StudentView:   student,
		GradingBundle: bundle,
	}, violations, nil
}

func gradeSubmission(ctx context.Context, enriched enrichedSubmission) (SubmitSessionContractResponse, int, int, string, string, error) {
	req := SubmitSessionContractRequest{
		SessionID: enriched.SessionID,
		ExamID:    enriched.ExamID,
		UserID:    enriched.UserID,
		Answers:   enriched.Answers,
	}
	bundle := enriched.Bundle
	deterministicCount := 0
	llmCount := 0

	if strings.TrimSpace(bundle.ExamID) == "" || len(bundle.QuestionBy) == 0 {
		return SubmitSessionContractResponse{}, deterministicCount, llmCount, "GRADING_BUNDLE_CORRUPT", "grading bundle is corrupt", fmt.Errorf("corrupt bundle")
	}
	for _, qid := range bundle.QuestionIDs {
		k, ok := bundle.QuestionBy[qid]
		if !ok || missingRequiredGradingKey(k) {
			return SubmitSessionContractResponse{}, deterministicCount, llmCount, "GRADING_KEYS_MISSING", "grading keys missing", fmt.Errorf("missing keys")
		}
	}

	for answerID := range req.Answers {
		if _, ok := bundle.QuestionBy[answerID]; !ok {
			return SubmitSessionContractResponse{}, deterministicCount, llmCount, "INVALID_QUESTION_ID", "unknown question id in answers", fmt.Errorf("unknown id")
		}
	}

	scoreRaw := 0.0
	scoreMax := bundle.ScoreMax
	results := make([]ExamQuestionResultDTO, 0, len(bundle.QuestionIDs))
	for _, qid := range bundle.QuestionIDs {
		key := bundle.QuestionBy[qid]
		answer, provided := req.Answers[qid]
		earned, feedback := gradeDeterministic(answer, provided, key)
		method := "deterministic"
		deterministicCount++
		if shouldUseLLMGrading(key, provided) {
			if llmScore, llmFeedback, ok := gradeWithLLM(ctx, answer, key); ok {
				earned = llmScore
				feedback = llmFeedback
				method = "llm"
				deterministicCount--
				llmCount++
			}
		}
		earned = clampToRange(earned, 0, key.Points)
		scoreRaw += earned
		results = append(results, ExamQuestionResultDTO{
			QuestionID:     qid,
			MaxPoints:      key.Points,
			EarnedPoints:   earned,
			Method:         method,
			Feedback:       feedback,
			ExpectedAnswer: expectedAnswerForReveal(key),
			StudentAnswer:  normalizeStudentAnswer(answer),
		})
	}
	scorePct := 0.0
	if scoreMax > 0 {
		scorePct = (scoreRaw / scoreMax) * 100
	}
	return SubmitSessionContractResponse{
		SessionID:   req.SessionID,
		ExamID:      req.ExamID,
		Status:      "graded",
		ScoreRaw:    round2(scoreRaw),
		ScoreMax:    round2(scoreMax),
		ScorePct:    round2(scorePct),
		Grade:       gradeFromPercentage(scorePct),
		PerQuestion: results,
	}, deterministicCount, llmCount, "", "", nil
}

func gradeDeterministic(answer interface{}, provided bool, key gradingKey) (float64, string) {
	if !provided {
		if key.Required {
			return 0, "No answer provided."
		}
		return 0, "Optional question skipped."
	}
	switch key.Type {
	case "single_choice":
		expected := strings.TrimSpace(valueToString(answerFromKey(key.AnswerKey)))
		got := strings.TrimSpace(valueToString(answer))
		if got != "" && strings.EqualFold(got, expected) {
			return key.Points, "Correct."
		}
		return 0, "Incorrect choice."
	case "multi_select":
		expected := normalizeExamSet(toStringSlice(answerFromKey(key.AnswerKey)))
		got := normalizeExamSet(toStringSlice(answer))
		if len(expected) == 0 {
			return 0, "Invalid grading key."
		}
		if slices.Equal(expected, got) {
			return key.Points, "Correct selection."
		}
		overlap := countExamOverlap(expected, got)
		return key.Points * float64(overlap) / float64(len(expected)), "Partially correct."
	case "boolean":
		expected := strings.ToLower(strings.TrimSpace(valueToString(answerFromKey(key.AnswerKey))))
		got := strings.ToLower(strings.TrimSpace(valueToString(answer)))
		if (expected == "true" || expected == "false") && expected == got {
			return key.Points, "Correct."
		}
		return 0, "Incorrect."
	case "numeric":
		got, ok := toFloatExam(answer)
		if !ok {
			return 0, "Answer must be numeric."
		}
		expected, tol := parseNumericKey(answerFromKey(key.AnswerKey))
		if tol <= 0 {
			tol = 0.0001
		}
		if abs(got-expected) <= tol {
			return key.Points, "Correct numeric answer."
		}
		return 0, "Numeric answer out of tolerance."
	case "short_text", "long_text", "code":
		got := strings.ToLower(strings.TrimSpace(valueToString(answer)))
		if got == "" {
			return 0, "No answer provided."
		}
		keywords := toStringSlice(answerFromKey(key.AnswerKey))
		if len(keywords) == 0 {
			return 0, "No deterministic key available."
		}
		hits := 0
		for _, kw := range keywords {
			kw = strings.ToLower(strings.TrimSpace(kw))
			if kw != "" && strings.Contains(got, kw) {
				hits++
			}
		}
		if hits == 0 {
			return 0, "Missing key concepts."
		}
		return key.Points * float64(hits) / float64(len(keywords)), "Partial/complete concept match."
	default:
		return 0, "Unsupported question type."
	}
}

func persistExamSecretState(sessionID string, userID string, artifact generatedExamArtifact) error {
	bundlePayload, _ := json.Marshal(artifact.GradingBundle)
	if err := secretStore.StoreWithTTL("exam:grading:"+artifact.ExamID, bundlePayload, examBundleTTLSeconds); err != nil {
		return err
	}
	now := time.Now().Unix()
	cap := examAttemptCap()
	sessionPayload, _ := json.Marshal(sessionExamMap{
		SessionID:  sessionID,
		ExamID:     artifact.ExamID,
		UserID:     userID,
		Attempts:   0,
		AttemptCap: cap,
		CreatedAt:  now,
		ExpiresAt:  now + int64(examBundleTTLSeconds),
	})
	if err := secretStore.StoreWithTTL("session:exam:"+sessionID, sessionPayload, examBundleTTLSeconds); err != nil {
		return err
	}
	metaPayload, _ := json.Marshal(map[string]interface{}{
		"examId":    artifact.ExamID,
		"title":     artifact.StudentView.Title,
		"questions": len(artifact.StudentView.Questions),
	})
	return secretStore.StoreWithTTL("exam:meta:"+artifact.ExamID, metaPayload, examBundleTTLSeconds)
}

func retrieveRequiredKey(key string) ([]byte, error) {
	exists, err := secretStore.Check(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("missing key")
	}
	return secretStore.Retrieve(key)
}

func loadExamPrompt() string {
	p, err := prompt.GetPrompt("mostlearnedExamPrompt")
	if err != nil || strings.TrimSpace(p) == "" {
		return defaultMostLearnedExamPrompt
	}
	return p
}

func buildExamPrompt(req GenerateExamContractRequest) string {
	return fmt.Sprintf("Topic: %s\nQuestion count: %d\nGenerate strict JSON output.", req.Topic, req.QuestionCount)
}

func buildExamRepairPrompt(req GenerateExamContractRequest, raw string, violations []string, parseErr error) string {
	var sb strings.Builder
	sb.WriteString("Previous output invalid. Return corrected JSON only.\n")
	sb.WriteString(buildExamPrompt(req))
	sb.WriteString("\nViolations:\n")
	if parseErr != nil {
		sb.WriteString("- parse_error: ")
		sb.WriteString(parseErr.Error())
		sb.WriteString("\n")
	}
	for _, v := range violations {
		sb.WriteString("- ")
		sb.WriteString(v)
		sb.WriteString("\n")
	}
	sb.WriteString("\nPrevious output:\n")
	sb.WriteString(utils.CleanJSONResponse(raw))
	return sb.String()
}

func requestIDFromHeader(r *http.Request) string {
	rid := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if rid == "" {
		rid = uuid.NewString()
	}
	return rid
}

func stableExamSeed(topic string, count int) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(topic))))
	_, _ = h.Write([]byte(fmt.Sprintf(":%d", count)))
	return int(h.Sum32() & 0x7fffffff)
}

func missingRequiredGradingKey(k gradingKey) bool {
	if strings.TrimSpace(k.Type) == "" || k.Points <= 0 {
		return true
	}
	if k.AnswerKey == nil {
		return true
	}
	return false
}

func sanitizeSecretMap(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	blocked := map[string]struct{}{
		"answerkey":   {},
		"rubric":      {},
		"modelanswer": {},
	}
	out := make(map[string]interface{})
	for k, v := range in {
		if _, ok := blocked[strings.ToLower(strings.TrimSpace(k))]; ok {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizeExamOptions(options []string) []string {
	if len(options) == 0 {
		return nil
	}
	out := make([]string, 0, len(options))
	for _, opt := range options {
		opt = strings.TrimSpace(opt)
		if opt != "" {
			out = append(out, opt)
		}
	}
	return out
}

func projectStudentViewQuestion(in ExamStudentQuestion) ExamStudentQuestion {
	in.Prompt = strings.TrimSpace(in.Prompt)
	in.Options = normalizeExamOptions(in.Options)
	in.UI = sanitizeSecretMap(in.UI)
	in.Metadata = sanitizeSecretMap(in.Metadata)
	return in
}

func projectGradingBundleKey(in gradingKey) gradingKey {
	in.ModelAnswer = strings.TrimSpace(in.ModelAnswer)
	in.Rubric = normalizeExamOptions(in.Rubric)
	in.Options = normalizeExamOptions(in.Options)
	return in
}

func answerFromKey(v interface{}) interface{} { return v }

func toStringSlice(v interface{}) []string {
	switch t := v.(type) {
	case []string:
		out := make([]string, 0, len(t))
		for _, s := range t {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, it := range t {
			if s, ok := it.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					out = append(out, s)
				}
			}
		}
		return out
	default:
		return nil
	}
}

func normalizeExamSet(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		v = strings.ToLower(strings.TrimSpace(v))
		if v != "" {
			seen[v] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	slices.Sort(out)
	return out
}

func countExamOverlap(expected []string, got []string) int {
	set := make(map[string]struct{}, len(got))
	for _, g := range got {
		set[g] = struct{}{}
	}
	hits := 0
	for _, e := range expected {
		if _, ok := set[e]; ok {
			hits++
		}
	}
	return hits
}

func valueToString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%v", t)
	case int:
		return fmt.Sprintf("%d", t)
	default:
		return ""
	}
}

func parseNumericKey(v interface{}) (float64, float64) {
	if m, ok := v.(map[string]interface{}); ok {
		value, _ := toFloatExam(m["value"])
		tol, _ := toFloatExam(m["tolerance"])
		return value, tol
	}
	value, _ := toFloatExam(v)
	return value, 0
}

func toFloatExam(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func clampToRange(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func examAttemptCap() int {
	raw := strings.TrimSpace(os.Getenv("EXAM_ATTEMPT_CAP"))
	if raw == "" {
		return defaultExamAttemptCap
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultExamAttemptCap
	}
	return n
}

func enrichSubmission(req SubmitSessionContractRequest) (enrichedSubmission, string, string, error) {
	if containsForbiddenFields(req.Answers) {
		return enrichedSubmission{}, "INVALID_REQUEST", "answers contain forbidden grading metadata", fmt.Errorf("forbidden answer fields")
	}

	sessionRaw, err := retrieveRequiredKey("session:exam:" + req.SessionID)
	if err != nil {
		return enrichedSubmission{}, "INVALID_REQUEST", "session not found or expired", err
	}
	var session sessionExamMap
	if err := json.Unmarshal(sessionRaw, &session); err != nil {
		return enrichedSubmission{}, "GRADING_BUNDLE_CORRUPT", "session map is corrupt", err
	}
	if strings.TrimSpace(session.ExamID) == "" || session.ExamID != req.ExamID {
		return enrichedSubmission{}, "INVALID_REQUEST", "sessionId does not match examId", fmt.Errorf("session exam mismatch")
	}
	if session.UserID != "" && session.UserID != req.UserID {
		return enrichedSubmission{}, "INVALID_REQUEST", "session bound to different user", fmt.Errorf("session user mismatch")
	}
	if session.ExpiresAt > 0 && time.Now().Unix() > session.ExpiresAt {
		return enrichedSubmission{}, "INVALID_REQUEST", "session expired", fmt.Errorf("expired session")
	}
	attemptCap := session.AttemptCap
	if attemptCap <= 0 {
		attemptCap = examAttemptCap()
	}
	if session.Attempts >= attemptCap {
		return enrichedSubmission{}, "INVALID_REQUEST", "attempt cap reached", fmt.Errorf("attempt cap reached")
	}

	bundleRaw, err := retrieveRequiredKey("exam:grading:" + req.ExamID)
	if err != nil {
		if strings.Contains(err.Error(), "missing") {
			return enrichedSubmission{}, "GRADING_KEYS_MISSING", "grading bundle not found", err
		}
		return enrichedSubmission{}, "INTERNAL_ERROR", "failed to read grading bundle", err
	}
	var bundle gradingBundle
	if err := json.Unmarshal(bundleRaw, &bundle); err != nil {
		return enrichedSubmission{}, "GRADING_BUNDLE_CORRUPT", "grading bundle is corrupt", err
	}
	if strings.TrimSpace(bundle.ExamID) == "" || len(bundle.QuestionBy) == 0 {
		return enrichedSubmission{}, "GRADING_BUNDLE_CORRUPT", "grading bundle is corrupt", fmt.Errorf("invalid bundle")
	}
	for _, qid := range bundle.QuestionIDs {
		if _, ok := bundle.QuestionBy[qid]; !ok {
			return enrichedSubmission{}, "GRADING_BUNDLE_CORRUPT", "grading bundle is corrupt", fmt.Errorf("missing question key")
		}
	}
	for answerID := range req.Answers {
		if _, ok := bundle.QuestionBy[answerID]; !ok {
			return enrichedSubmission{}, "INVALID_QUESTION_ID", "unknown question id in answers", fmt.Errorf("unknown question id")
		}
	}

	return enrichedSubmission{
		SessionID: req.SessionID,
		ExamID:    req.ExamID,
		UserID:    req.UserID,
		Answers:   req.Answers,
		Bundle:    bundle,
	}, "", "", nil
}

func shouldUseLLMGrading(key gradingKey, provided bool) bool {
	if !provided {
		return false
	}
	switch key.Type {
	case "short_text", "long_text", "code":
		return true
	default:
		return false
	}
}

func gradeWithLLM(ctx context.Context, answer interface{}, key gradingKey) (float64, string, bool) {
	answerText := strings.TrimSpace(valueToString(answer))
	if answerText == "" {
		return 0, "", false
	}
	if len(answerText) > maxLLMGradingInputChars {
		answerText = answerText[:maxLLMGradingInputChars]
	}

	cfg, ok := configs.GetDeepseekConfig("/submit-session")
	if !ok {
		cfg, _ = configs.GetDeepseekConfig(generateExamContractEndpoint)
	}
	client, err := getLLMClientByConfig(cfg)
	if err != nil {
		return 0, "", false
	}

	promptText := fmt.Sprintf(`Grade this answer and return strict JSON only.
Schema:
{"score": number, "feedback": string}
Constraints:
- score in [0, %.2f]
- concise feedback
QuestionType: %s
Rubric: %v
ModelAnswer: %s
AnswerKey: %v
StudentAnswer: %s`,
		key.Points, key.Type, key.Rubric, key.ModelAnswer, key.AnswerKey, answerText)

	raw, err := callLLMWithTimeout(ctx, client, "/submit-session", map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.1,
		"max_tokens":  300,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": promptText,
		},
	})
	if err != nil {
		return 0, "", false
	}

	var out gradingOutput
	if err := json.Unmarshal([]byte(utils.CleanJSONResponse(raw)), &out); err != nil {
		return 0, "", false
	}
	out.Feedback = strings.TrimSpace(out.Feedback)
	if out.Feedback == "" {
		return 0, "", false
	}
	out.Score = clampToRange(out.Score, 0, key.Points)
	return out.Score, out.Feedback, true
}

func persistExamResult(enriched enrichedSubmission, resp SubmitSessionContractResponse) error {
	sessionRaw, err := retrieveRequiredKey("session:exam:" + enriched.SessionID)
	if err != nil {
		return err
	}
	var session sessionExamMap
	if err := json.Unmarshal(sessionRaw, &session); err != nil {
		return err
	}
	attempt := session.Attempts + 1

	resultPayload, _ := json.Marshal(resp)
	if err := secretStore.StoreWithTTL(fmt.Sprintf("exam:result:%s:%d", enriched.SessionID, attempt), resultPayload, examBundleTTLSeconds); err != nil {
		return err
	}
	if err := secretStore.StoreWithTTL("exam:latest_result:"+enriched.SessionID, resultPayload, examBundleTTLSeconds); err != nil {
		return err
	}

	eventPayload, _ := json.Marshal(map[string]interface{}{
		"sessionId": enriched.SessionID,
		"examId":    enriched.ExamID,
		"userId":    enriched.UserID,
		"attempt":   attempt,
		"scoreRaw":  resp.ScoreRaw,
		"scoreMax":  resp.ScoreMax,
		"scorePct":  resp.ScorePct,
		"ts":        time.Now().Unix(),
	})
	return secretStore.StoreWithTTL(fmt.Sprintf("exam:score_event:%s:%d", enriched.SessionID, attempt), eventPayload, examBundleTTLSeconds)
}

func persistUserRollup(userID string, resp SubmitSessionContractResponse) error {
	if strings.TrimSpace(userID) == "" {
		return nil
	}
	rollupKey := "progress:rollup:user:" + userID
	var rollup struct {
		UserID       string  `json:"userId"`
		Attempts     int     `json:"attempts"`
		AveragePct   float64 `json:"averagePct"`
		LastScorePct float64 `json:"lastScorePct"`
		LastExamID   string  `json:"lastExamId"`
		LastUpdated  int64   `json:"lastUpdated"`
	}
	if raw, err := retrieveOptionalKey(rollupKey); err == nil && len(raw) > 0 {
		_ = json.Unmarshal(raw, &rollup)
	}
	if rollup.UserID == "" {
		rollup.UserID = userID
	}
	rollup.Attempts++
	if rollup.Attempts == 1 {
		rollup.AveragePct = resp.ScorePct
	} else {
		rollup.AveragePct = ((rollup.AveragePct * float64(rollup.Attempts-1)) + resp.ScorePct) / float64(rollup.Attempts)
	}
	rollup.AveragePct = round2(rollup.AveragePct)
	rollup.LastScorePct = round2(resp.ScorePct)
	rollup.LastExamID = resp.ExamID
	rollup.LastUpdated = time.Now().Unix()
	payload, _ := json.Marshal(rollup)
	return secretStore.StoreWithTTL(rollupKey, payload, examBundleTTLSeconds)
}

func updateSessionAttempt(sessionID, userID, examID string) error {
	sessionRaw, err := retrieveRequiredKey("session:exam:" + sessionID)
	if err != nil {
		return err
	}
	var session sessionExamMap
	if err := json.Unmarshal(sessionRaw, &session); err != nil {
		return err
	}
	session.Attempts++
	if session.UserID == "" {
		session.UserID = userID
	}
	if session.ExamID == "" {
		session.ExamID = examID
	}
	if session.AttemptCap <= 0 {
		session.AttemptCap = examAttemptCap()
	}
	payload, _ := json.Marshal(session)
	return secretStore.StoreWithTTL("session:exam:"+sessionID, payload, examBundleTTLSeconds)
}

func retrieveOptionalKey(key string) ([]byte, error) {
	exists, err := secretStore.Check(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	return secretStore.Retrieve(key)
}

func containsForbiddenFields(v interface{}) bool {
	forbidden := map[string]struct{}{
		"answerkey":   {},
		"rubric":      {},
		"modelanswer": {},
	}
	var walk func(interface{}) bool
	walk = func(item interface{}) bool {
		switch t := item.(type) {
		case map[string]interface{}:
			for k, val := range t {
				if _, blocked := forbidden[strings.ToLower(strings.TrimSpace(k))]; blocked {
					return true
				}
				if walk(val) {
					return true
				}
			}
		case []interface{}:
			for _, val := range t {
				if walk(val) {
					return true
				}
			}
		}
		return false
	}
	return walk(v)
}

func expectedAnswerForReveal(key gradingKey) interface{} {
	switch key.Type {
	case "short_text", "long_text", "code":
		if strings.TrimSpace(key.ModelAnswer) != "" {
			return strings.TrimSpace(key.ModelAnswer)
		}
		if len(key.Rubric) > 0 {
			joined := strings.Join(key.Rubric, "; ")
			if len(joined) > 300 {
				return joined[:300]
			}
			return joined
		}
		return "Use the key rubric criteria for full credit."
	case "multi_select":
		return toStringSlice(answerFromKey(key.AnswerKey))
	default:
		return strings.TrimSpace(valueToString(answerFromKey(key.AnswerKey)))
	}
}

func normalizeStudentAnswer(v interface{}) interface{} {
	switch t := v.(type) {
	case []interface{}:
		return t
	case []string:
		return t
	default:
		return v
	}
}

func gradeFromPercentage(pct float64) string {
	switch {
	case pct >= 90:
		return "A"
	case pct >= 80:
		return "B"
	case pct >= 70:
		return "C"
	case pct >= 60:
		return "D"
	default:
		return "F"
	}
}

func writeExamContractErrorByCode(w http.ResponseWriter, code string, message string) {
	switch code {
	case "INVALID_REQUEST":
		writeExamContractError(w, http.StatusBadRequest, code, message)
	case "INVALID_QUESTION_ID":
		writeExamContractError(w, http.StatusBadRequest, code, message)
	case "GRADING_KEYS_MISSING":
		writeExamContractError(w, http.StatusConflict, code, message)
	case "GRADING_BUNDLE_CORRUPT":
		writeExamContractError(w, http.StatusInternalServerError, code, message)
	case "LLM_INVALID_OUTPUT":
		writeExamContractError(w, http.StatusBadGateway, code, message)
	default:
		writeExamContractError(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
	}
}

func validateGradedResponse(resp SubmitSessionContractResponse) error {
	if strings.TrimSpace(resp.SessionID) == "" || strings.TrimSpace(resp.ExamID) == "" {
		return fmt.Errorf("missing sessionId or examId")
	}
	if resp.Status != "graded" {
		return fmt.Errorf("status must be graded")
	}
	if resp.ScoreMax <= 0 || resp.ScoreRaw < 0 || resp.ScoreRaw > resp.ScoreMax {
		return fmt.Errorf("invalid score totals")
	}
	if len(resp.PerQuestion) == 0 {
		return fmt.Errorf("perQuestion must not be empty")
	}
	for i, q := range resp.PerQuestion {
		if strings.TrimSpace(q.QuestionID) == "" {
			return fmt.Errorf("perQuestion[%d].questionId missing", i)
		}
		if q.MaxPoints <= 0 {
			return fmt.Errorf("perQuestion[%d].maxPoints invalid", i)
		}
		if q.EarnedPoints < 0 || q.EarnedPoints > q.MaxPoints {
			return fmt.Errorf("perQuestion[%d].earnedPoints invalid", i)
		}
		if q.Method != "deterministic" && q.Method != "llm" {
			return fmt.Errorf("perQuestion[%d].method invalid", i)
		}
		if strings.TrimSpace(q.Feedback) == "" {
			return fmt.Errorf("perQuestion[%d].feedback missing", i)
		}
		if q.ExpectedAnswer == nil {
			return fmt.Errorf("perQuestion[%d].expectedAnswer missing", i)
		}
		if q.StudentAnswer == nil {
			return fmt.Errorf("perQuestion[%d].studentAnswer missing", i)
		}
	}
	return nil
}

func writeExamContractError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(examContractError{
		Error: message,
		Code:  code,
	})
}

func writeExamContractJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
