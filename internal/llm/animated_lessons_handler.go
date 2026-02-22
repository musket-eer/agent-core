package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

const (
	animatedKnowledgeGraphEndpoint = "/lessons/animated-knowledge-graph"
	animatedSlideContentEndpoint   = "/lessons/animated-slide-content"
	animatedMaxRequestBytes        = 1 << 20 // 1MB
	animatedDefaultRateLimit       = 30
	animatedDefaultRateWindow      = time.Minute
	animatedMaxConcurrentSlides    = 16
)

var knowledgeNodeIDPattern = regexp.MustCompile(`^[a-z0-9_\-]{2,64}$`)

const defaultAnimatedSystemPrompt = `You generate educational animated lessons.
Return only JSON matching the required schema.
Rules:
1) Output nodes in level-order traversal.
2) Every prerequisite must appear earlier than dependent nodes.
3) Keep each lesson step concise and pedagogically clear.
4) Include 1-4 keyPoints and 0-3 examples per node.
5) Use ids matching ^[a-z0-9_\-]{2,64}$.
6) Ensure knowledgeGraph has at least 2 nodes.
7) No markdown, no commentary, no extra keys.
8) The lesson content must sound flowing and natural for a learner.
9) Do not mention internal structure words in learner-facing text: knowledge graph, graph, node, edge, prerequisite.`

const defaultAnimatedSlideSystemPrompt = `You generate detailed content for one animated lesson topic.
Return only JSON matching the required schema.
Rules:
1) Focus only on the current topic.
2) Keep language natural, concise, and classroom-ready.
3) Use previous and next topic only for smooth transitions.
4) Keep output frontend-interpretable only.
5) Text and demonstration are separate child regions; never describe overlay/stacking between them.
6) Choose exactly one primary demo modality for this topic: imageUrl OR sketchPath OR equations.
7) demoPlan.kind must be one of: NONE, SKETCH_PATH, EQUATION_REVEAL, IMAGE_DRAW, PHYSICS_GRAPH.
8) demoPlan.kind must agree with modality fields:
   SKETCH_PATH or PHYSICS_GRAPH -> sketchPath
   EQUATION_REVEAL -> equations
   IMAGE_DRAW -> imageUrl
   NONE -> no modality field
9) If using sketchPath, keep geometry simple and non-overlapping.
10) No markdown, no commentary, no extra keys.
11) Do not mention internal structure words in learner-facing text: knowledge graph, graph, node, edge, prerequisite.`

var allowedDemoPlanKinds = map[string]struct{}{
	"NONE":            {},
	"SKETCH_PATH":     {},
	"EQUATION_REVEAL": {},
	"IMAGE_DRAW":      {},
	"PHYSICS_GRAPH":   {},
}

var allowedDemoPlanEmphasis = map[string]struct{}{
	"":                 {},
	"progressive_draw": {},
	"step_trace":       {},
	"settle_motion":    {},
	"reveal":           {},
}

type animatedAPIError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type animatedRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

func newAnimatedRateLimiter(limit int, window time.Duration) *animatedRateLimiter {
	rps := rate.Every(window / time.Duration(limit))
	return &animatedRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    limit,
	}
}

func (rl *animatedRateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, ok := rl.limiters[key]
	if !ok {
		limiter = rate.NewLimiter(rl.rps, rl.burst)
		rl.limiters[key] = limiter
	}
	return limiter.Allow()
}

var animatedEndpointLimiter = newAnimatedRateLimiter(animatedDefaultRateLimit, animatedDefaultRateWindow)
var animatedSlideSemaphore = make(chan struct{}, animatedMaxConcurrentSlides)

// GenerateAnimatedKnowledgeGraphHandler handles POST /lessons/animated-knowledge-graph.
func GenerateAnimatedKnowledgeGraphHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAnimatedError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}

	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = uuid.NewString()
	}

	clientIP := clientIPFromRequest(r)
	if !animatedEndpointLimiter.allow(clientIP) {
		writeAnimatedError(w, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, animatedMaxRequestBytes)

	var req AnimatedLessonRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeAnimatedError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if err := validateAnimatedRequest(&req); err != nil {
		writeAnimatedError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	start := time.Now()
	resp, modelName, validationStatus, err := generateAnimatedKnowledgeGraph(context.Background(), req)
	latency := time.Since(start).Milliseconds()
	log.Printf("[GenerateAnimatedKnowledgeGraph] request_id=%s topic=%q model=%s latency_ms=%d token_usage=%s validation=%s",
		requestID, req.Topic, modelName, latency, "unavailable", validationStatus)

	if err != nil {
		if strings.Contains(err.Error(), "LLM_INVALID_OUTPUT") {
			writeAnimatedError(w, http.StatusBadGateway, "LLM_INVALID_OUTPUT", "failed to produce valid knowledge graph")
			return
		}
		writeAnimatedError(w, http.StatusBadGateway, "LLM_FAILURE", "failed to generate knowledge graph")
		return
	}

	writeAnimatedJSON(w, http.StatusOK, resp)
}

// GenerateAnimatedSlideContentHandler handles POST /lessons/animated-slide-content.
func GenerateAnimatedSlideContentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAnimatedError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}

	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = uuid.NewString()
	}

	clientIP := clientIPFromRequest(r)
	if !animatedEndpointLimiter.allow(clientIP) {
		writeAnimatedError(w, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded")
		return
	}

	select {
	case animatedSlideSemaphore <- struct{}{}:
		defer func() { <-animatedSlideSemaphore }()
	default:
		writeAnimatedError(w, http.StatusTooManyRequests, "CONCURRENCY_LIMITED", "too many slide content requests in progress")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, animatedMaxRequestBytes)

	var req AnimatedSlideContentRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeAnimatedError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if err := validateAnimatedSlideContentRequest(&req); err != nil {
		writeAnimatedError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	start := time.Now()
	resp, modelName, validationStatus, err := generateAnimatedSlideContent(context.Background(), req)
	latency := time.Since(start).Milliseconds()
	log.Printf("[GenerateAnimatedSlideContent] request_id=%s topic_id=%q model=%s latency_ms=%d token_usage=%s validation=%s",
		requestID, req.CurrentTopic.ID, modelName, latency, "unavailable", validationStatus)

	if err != nil {
		if strings.Contains(err.Error(), "LLM_INVALID_OUTPUT") {
			writeAnimatedError(w, http.StatusBadGateway, "LLM_INVALID_OUTPUT", "failed to produce valid slide content")
			return
		}
		writeAnimatedError(w, http.StatusBadGateway, "LLM_FAILURE", "failed to generate slide content")
		return
	}

	writeAnimatedJSON(w, http.StatusOK, resp)
}

func generateAnimatedKnowledgeGraph(ctx context.Context, req AnimatedLessonRequest) (AnimatedLessonResponse, string, string, error) {
	cfg, ok := configs.GetDeepseekConfig(animatedKnowledgeGraphEndpoint)
	if !ok {
		cfg, _ = configs.GetDeepseekConfig("/chat")
	}

	clientName := cfg.ClientName
	if clientName == "" {
		clientName = "general"
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		return AnimatedLessonResponse{}, cfg.ModelId, "failed", err
	}

	seed := stableSeedFromRequest(req)
	userPrompt := buildAnimatedUserPrompt(req)
	systemPrompt := loadAnimatedSystemPrompt()

	rawOutput, err := callLLMWithTimeout(ctx, client, animatedKnowledgeGraphEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  2200,
		"seed":        seed,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": systemPrompt + "\n\n" + userPrompt,
		},
	})
	if err != nil {
		return AnimatedLessonResponse{}, cfg.ModelId, "failed", err
	}

	resp, errs, parseErr := parseAndValidateAnimatedLLMOutput(rawOutput, req.Animation.IncludePrerequisiteEdges)
	if parseErr == nil && len(errs) == 0 {
		return resp, cfg.ModelId, "pass", nil
	}

	repairPrompt := buildAnimatedRepairPrompt(req, rawOutput, errs, parseErr)
	repairedOutput, repairErr := callLLMWithTimeout(ctx, client, animatedKnowledgeGraphEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  2200,
		"seed":        seed + 1,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": systemPrompt + "\n\n" + repairPrompt,
		},
	})
	if repairErr != nil {
		return AnimatedLessonResponse{}, cfg.ModelId, "failed", repairErr
	}

	finalResp, finalErrs, finalParseErr := parseAndValidateAnimatedLLMOutput(repairedOutput, req.Animation.IncludePrerequisiteEdges)
	if finalParseErr != nil || len(finalErrs) > 0 {
		log.Printf("[GenerateAnimatedKnowledgeGraph] repair validation failed parse_error=%v errors=%v", finalParseErr, finalErrs)
		return AnimatedLessonResponse{}, cfg.ModelId, "failed", fmt.Errorf("LLM_INVALID_OUTPUT")
	}

	return finalResp, cfg.ModelId, "pass_after_repair", nil
}

func generateAnimatedSlideContent(ctx context.Context, req AnimatedSlideContentRequest) (AnimatedSlideContentResponse, string, string, error) {
	cfg, ok := configs.GetDeepseekConfig(animatedSlideContentEndpoint)
	if !ok {
		cfg, _ = configs.GetDeepseekConfig("/chat")
	}

	clientName := cfg.ClientName
	if clientName == "" {
		clientName = "general"
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		return AnimatedSlideContentResponse{}, cfg.ModelId, "failed", err
	}

	seed := stableSeedFromSlideRequest(req)
	systemPrompt := loadAnimatedSlideSystemPrompt()
	userPrompt := buildAnimatedSlideUserPrompt(req)

	rawOutput, err := callLLMWithTimeout(ctx, client, animatedSlideContentEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  1400,
		"seed":        seed,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": systemPrompt + "\n\n" + userPrompt,
		},
	})
	if err != nil {
		return AnimatedSlideContentResponse{}, cfg.ModelId, "failed", err
	}

	resp, errs, parseErr, sanitized := parseAndValidateAnimatedSlideOutput(rawOutput, req.CurrentTopic.ID)
	if parseErr == nil && len(errs) == 0 {
		if sanitized {
			return resp, cfg.ModelId, "pass_sanitized", nil
		}
		return resp, cfg.ModelId, "pass", nil
	}

	repairPrompt := buildAnimatedSlideRepairPrompt(req, rawOutput, errs, parseErr)
	repairedOutput, repairErr := callLLMWithTimeout(ctx, client, animatedSlideContentEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  1400,
		"seed":        seed + 1,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": systemPrompt + "\n\n" + repairPrompt,
		},
	})
	if repairErr != nil {
		return AnimatedSlideContentResponse{}, cfg.ModelId, "failed", repairErr
	}

	finalResp, finalErrs, finalParseErr, finalSanitized := parseAndValidateAnimatedSlideOutput(repairedOutput, req.CurrentTopic.ID)
	if finalParseErr != nil || len(finalErrs) > 0 {
		log.Printf("[GenerateAnimatedSlideContent] repair validation failed parse_error=%v errors=%v", finalParseErr, finalErrs)
		return AnimatedSlideContentResponse{}, cfg.ModelId, "failed", fmt.Errorf("LLM_INVALID_OUTPUT")
	}

	if finalSanitized {
		return finalResp, cfg.ModelId, "pass_after_repair_sanitized", nil
	}
	return finalResp, cfg.ModelId, "pass_after_repair", nil
}

func loadAnimatedSystemPrompt() string {
	systemPrompt, err := prompt.GetPrompt("animatedKnowledgeGraphPrompt")
	if err != nil || strings.TrimSpace(systemPrompt) == "" {
		return defaultAnimatedSystemPrompt
	}
	return systemPrompt
}

func loadAnimatedSlideSystemPrompt() string {
	systemPrompt, err := prompt.GetPrompt("animatedSlideContentPrompt")
	if err != nil || strings.TrimSpace(systemPrompt) == "" {
		return defaultAnimatedSlideSystemPrompt
	}
	return systemPrompt
}

func parseAndValidateAnimatedLLMOutput(raw string, includePrerequisiteEdges bool) (AnimatedLessonResponse, []string, error) {
	var resp AnimatedLessonResponse
	cleaned := utils.CleanJSONResponse(raw)
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return AnimatedLessonResponse{}, nil, err
	}

	errs := validateAnimatedLessonResponse(resp, includePrerequisiteEdges)
	return resp, errs, nil
}

func parseAndValidateAnimatedSlideOutput(raw string, expectedTopicID string) (AnimatedSlideContentResponse, []string, error, bool) {
	cleaned := utils.CleanJSONResponse(raw)
	var decoded map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader([]byte(cleaned)))
	if err := decoder.Decode(&decoded); err != nil {
		return AnimatedSlideContentResponse{}, nil, err, false
	}
	resp, sanitized := sanitizeAnimatedSlideWhitelist(decoded)
	normalizeAndBoundDetail(&resp, &sanitized)
	normalizeDemoPlan(&resp, &sanitized)
	resolveSingleModality(&resp, &sanitized)
	errs := validateAnimatedSlideContentResponse(resp, expectedTopicID)
	return resp, errs, nil, sanitized
}

func validateAnimatedRequest(req *AnimatedLessonRequest) error {
	normalizeLessonSharedFields(&req.CourseTitle, &req.LessonObjective, &req.Learner, &req.Animation)
	req.Topic = strings.TrimSpace(req.Topic)

	if req.Topic == "" {
		return fmt.Errorf("topic is required")
	}
	if len(req.Topic) > 200 {
		return fmt.Errorf("topic is too long (max 200)")
	}
	if err := validateLessonSharedFieldLengths(req.CourseTitle, req.LessonObjective, req.Learner, req.Animation); err != nil {
		return err
	}

	return nil
}

func validateAnimatedSlideContentRequest(req *AnimatedSlideContentRequest) error {
	normalizeLessonSharedFields(&req.CourseTitle, &req.LessonObjective, &req.Learner, &req.Animation)
	req.LessonID = strings.TrimSpace(req.LessonID)
	req.CurrentTopic.ID = strings.TrimSpace(req.CurrentTopic.ID)
	req.CurrentTopic.Title = strings.TrimSpace(req.CurrentTopic.Title)
	req.CurrentTopic.Summary = strings.TrimSpace(req.CurrentTopic.Summary)
	if req.PreviousTopic != nil {
		req.PreviousTopic.ID = strings.TrimSpace(req.PreviousTopic.ID)
		req.PreviousTopic.Title = strings.TrimSpace(req.PreviousTopic.Title)
	}
	if req.NextTopic != nil {
		req.NextTopic.ID = strings.TrimSpace(req.NextTopic.ID)
		req.NextTopic.Title = strings.TrimSpace(req.NextTopic.Title)
	}

	if req.CurrentTopic.ID == "" || req.CurrentTopic.Title == "" {
		return fmt.Errorf("currentTopic.id and currentTopic.title are required")
	}
	if !knowledgeNodeIDPattern.MatchString(req.CurrentTopic.ID) {
		return fmt.Errorf("currentTopic.id must match ^[a-z0-9_\\-]{2,64}$")
	}
	if len(req.LessonID) > 100 {
		return fmt.Errorf("lessonId is too long (max 100)")
	}
	if len(req.CurrentTopic.Title) > 120 {
		return fmt.Errorf("currentTopic.title is too long (max 120)")
	}
	if len(req.CurrentTopic.Summary) > 400 {
		return fmt.Errorf("currentTopic.summary is too long (max 400)")
	}
	if len(req.CurrentTopic.KeyPoints) > 8 {
		return fmt.Errorf("currentTopic.keyPoints is too long (max 8)")
	}
	if len(req.CurrentTopic.Examples) > 8 {
		return fmt.Errorf("currentTopic.examples is too long (max 8)")
	}
	if req.PreviousTopic != nil {
		if len(req.PreviousTopic.ID) > 64 {
			return fmt.Errorf("previousTopic.id is too long (max 64)")
		}
		if len(req.PreviousTopic.Title) > 120 {
			return fmt.Errorf("previousTopic.title is too long (max 120)")
		}
	}
	if req.NextTopic != nil {
		if len(req.NextTopic.ID) > 64 {
			return fmt.Errorf("nextTopic.id is too long (max 64)")
		}
		if len(req.NextTopic.Title) > 120 {
			return fmt.Errorf("nextTopic.title is too long (max 120)")
		}
	}

	if err := validateLessonSharedFieldLengths(req.CourseTitle, req.LessonObjective, req.Learner, req.Animation); err != nil {
		return err
	}

	return nil
}

func normalizeLessonSharedFields(courseTitle *string, lessonObjective *string, learner *LearnerProfile, animation *AnimationPrefs) {
	*courseTitle = strings.TrimSpace(*courseTitle)
	*lessonObjective = strings.TrimSpace(*lessonObjective)
	learner.Name = strings.TrimSpace(learner.Name)
	learner.EducationLevel = strings.TrimSpace(learner.EducationLevel)
	learner.Interests = strings.TrimSpace(learner.Interests)
	learner.CareerGoals = strings.TrimSpace(learner.CareerGoals)
	animation.Mode = strings.TrimSpace(animation.Mode)
}

func validateLessonSharedFieldLengths(courseTitle string, lessonObjective string, learner LearnerProfile, animation AnimationPrefs) error {
	if len(courseTitle) > 200 {
		return fmt.Errorf("courseTitle is too long (max 200)")
	}
	if len(lessonObjective) > 400 {
		return fmt.Errorf("lessonObjective is too long (max 400)")
	}
	if len(learner.Name) > 80 {
		return fmt.Errorf("learner.name is too long (max 80)")
	}
	if len(learner.EducationLevel) > 80 {
		return fmt.Errorf("learner.educationLevel is too long (max 80)")
	}
	if len(learner.Interests) > 300 {
		return fmt.Errorf("learner.interests is too long (max 300)")
	}
	if len(learner.CareerGoals) > 300 {
		return fmt.Errorf("learner.careerGoals is too long (max 300)")
	}
	if len(animation.Mode) > 40 {
		return fmt.Errorf("animation.mode is too long (max 40)")
	}

	return nil
}

func validateAnimatedLessonResponse(resp AnimatedLessonResponse, includePrerequisiteEdges bool) []string {
	var errs []string

	if strings.TrimSpace(resp.LessonID) == "" {
		errs = append(errs, "lessonId is required")
	}
	if strings.TrimSpace(resp.Title) == "" {
		errs = append(errs, "title is required")
	}
	if len(resp.KnowledgeGraph) < 2 {
		errs = append(errs, "knowledgeGraph must contain at least 2 nodes")
		return errs
	}

	indexByID := make(map[string]int, len(resp.KnowledgeGraph))
	nodeByID := make(map[string]KnowledgeNode, len(resp.KnowledgeGraph))
	for i, node := range resp.KnowledgeGraph {
		id := strings.TrimSpace(node.ID)
		if !knowledgeNodeIDPattern.MatchString(id) {
			errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].id must match ^[a-z0-9_\\-]{2,64}$", i))
		}
		if _, exists := indexByID[id]; exists {
			errs = append(errs, fmt.Sprintf("duplicate node id: %s", id))
		}
		if strings.TrimSpace(node.Title) == "" {
			errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].title is required", i))
		}
		if strings.TrimSpace(node.Summary) == "" {
			errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].summary is required", i))
		}
		if len(node.KeyPoints) < 1 || len(node.KeyPoints) > 4 {
			errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].keyPoints must contain 1-4 items", i))
		}
		if len(node.Examples) > 3 {
			errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].examples must contain 0-3 items", i))
		}
		if !includePrerequisiteEdges && len(node.Prerequisites) > 0 {
			errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].prerequisites must be empty when includePrerequisiteEdges=false", i))
		}
		indexByID[id] = i
		nodeByID[id] = node
	}

	for i, node := range resp.KnowledgeGraph {
		for _, prereqID := range node.Prerequisites {
			prereqID = strings.TrimSpace(prereqID)
			if prereqID == node.ID {
				errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].prerequisites cannot reference self", i))
				continue
			}
			if _, ok := nodeByID[prereqID]; !ok {
				errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].prerequisite %s does not exist", i, prereqID))
			}
			if prereqIdx, ok := indexByID[prereqID]; ok && prereqIdx >= i {
				errs = append(errs, fmt.Sprintf("knowledgeGraph[%d] prerequisite %s must appear earlier in response order", i, prereqID))
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	levels := make(map[string]int, len(resp.KnowledgeGraph))
	state := make(map[string]int, len(resp.KnowledgeGraph))
	var computeLevel func(id string) (int, error)
	computeLevel = func(id string) (int, error) {
		switch state[id] {
		case 1:
			return 0, fmt.Errorf("cycle detected involving node id: %s", id)
		case 2:
			return levels[id], nil
		}

		state[id] = 1
		node := nodeByID[id]
		level := 0
		for _, prereqID := range node.Prerequisites {
			prereqLevel, err := computeLevel(prereqID)
			if err != nil {
				return 0, err
			}
			if prereqLevel+1 > level {
				level = prereqLevel + 1
			}
		}
		state[id] = 2
		levels[id] = level
		return level, nil
	}

	for _, node := range resp.KnowledgeGraph {
		if _, err := computeLevel(node.ID); err != nil {
			errs = append(errs, err.Error())
			return errs
		}
	}

	prevLevel := -1
	for i, node := range resp.KnowledgeGraph {
		level := levels[node.ID]
		if level < prevLevel {
			errs = append(errs, fmt.Sprintf("knowledgeGraph is not in level-order at index %d", i))
			break
		}
		prevLevel = level
	}

	return errs
}

func validateAnimatedSlideContentResponse(resp AnimatedSlideContentResponse, expectedTopicID string) []string {
	var errs []string

	resp.TopicID = strings.TrimSpace(resp.TopicID)
	resp.Explanation = strings.TrimSpace(resp.Explanation)
	resp.BridgeFromPrevious = strings.TrimSpace(resp.BridgeFromPrevious)
	resp.BridgeToNext = strings.TrimSpace(resp.BridgeToNext)
	resp.ImageURL = strings.TrimSpace(resp.ImageURL)

	if resp.TopicID == "" {
		errs = append(errs, "topicId is required")
	}
	if expectedTopicID != "" && resp.TopicID != expectedTopicID {
		errs = append(errs, fmt.Sprintf("topicId must equal currentTopic.id (%s)", expectedTopicID))
	}
	if resp.Explanation == "" {
		errs = append(errs, "explanation is required")
	}
	if len(resp.Explanation) > 1200 {
		errs = append(errs, "explanation is too long (max 1200)")
	}
	if len(resp.KeyPoints) < 1 || len(resp.KeyPoints) > 6 {
		errs = append(errs, "keyPoints must contain 1-6 items")
	}
	if len(resp.Examples) > 4 {
		errs = append(errs, "examples must contain 0-4 items")
	}
	if len(resp.Equations) > 6 {
		errs = append(errs, "equations must contain 0-6 items")
	}
	for i, kp := range resp.KeyPoints {
		if strings.TrimSpace(kp) == "" {
			errs = append(errs, fmt.Sprintf("keyPoints[%d] must be non-empty", i))
		}
		if len(strings.TrimSpace(kp)) > 220 {
			errs = append(errs, fmt.Sprintf("keyPoints[%d] is too long (max 220)", i))
		}
	}
	for i, ex := range resp.Examples {
		if strings.TrimSpace(ex) == "" {
			errs = append(errs, fmt.Sprintf("examples[%d] must be non-empty", i))
		}
		if len(strings.TrimSpace(ex)) > 260 {
			errs = append(errs, fmt.Sprintf("examples[%d] is too long (max 260)", i))
		}
	}
	for i, eq := range resp.Equations {
		eq = strings.TrimSpace(eq)
		if eq == "" {
			errs = append(errs, fmt.Sprintf("equations[%d] must be non-empty", i))
		}
		if len(eq) > 120 {
			errs = append(errs, fmt.Sprintf("equations[%d] is too long (max 120)", i))
		}
	}
	if len(resp.BridgeFromPrevious) > 240 {
		errs = append(errs, "bridgeFromPrevious is too long (max 240)")
	}
	if len(resp.BridgeToNext) > 240 {
		errs = append(errs, "bridgeToNext is too long (max 240)")
	}
	if resp.ImageURL != "" {
		parsed, err := url.Parse(resp.ImageURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			errs = append(errs, "imageUrl must be a valid absolute URL")
		}
	}
	if resp.DemoPlan != nil {
		kind := strings.TrimSpace(resp.DemoPlan.Kind)
		if _, ok := allowedDemoPlanKinds[kind]; !ok {
			errs = append(errs, "demoPlan.kind must be one of NONE|SKETCH_PATH|EQUATION_REVEAL|IMAGE_DRAW|PHYSICS_GRAPH")
		}
		if len(strings.TrimSpace(resp.DemoPlan.Reason)) > 240 {
			errs = append(errs, "demoPlan.reason is too long (max 240)")
		}
		if len(strings.TrimSpace(resp.DemoPlan.AnimationEmphasis)) > 40 {
			errs = append(errs, "demoPlan.animationEmphasis is too long (max 40)")
		}
		if _, ok := allowedDemoPlanEmphasis[strings.TrimSpace(resp.DemoPlan.AnimationEmphasis)]; !ok {
			errs = append(errs, "demoPlan.animationEmphasis must be one of progressive_draw|step_trace|settle_motion|reveal")
		}
	}

	if resp.SketchPath != nil {
		errs = append(errs, validateSketchPath(resp.SketchPath)...)
	}
	errs = append(errs, validateDemoPlanConsistency(resp)...)

	if impliesOverlayBehavior(resp.Explanation) || impliesOverlayBehavior(resp.BridgeFromPrevious) || impliesOverlayBehavior(resp.BridgeToNext) {
		errs = append(errs, "text must not imply overlay/stacking between text and demo regions")
	}

	return errs
}

func validateDemoPlanConsistency(resp AnimatedSlideContentResponse) []string {
	var errs []string
	if resp.DemoPlan == nil {
		return errs
	}

	kind := strings.TrimSpace(resp.DemoPlan.Kind)
	hasImage := resp.ImageURL != ""
	hasSketch := resp.SketchPath != nil && len(resp.SketchPath.Nodes) > 0
	hasEquations := len(resp.Equations) > 0

	switch kind {
	case "NONE":
		if hasImage || hasSketch || hasEquations {
			errs = append(errs, "demoPlan.kind NONE requires no modality fields")
		}
	case "SKETCH_PATH", "PHYSICS_GRAPH":
		if !hasSketch {
			errs = append(errs, "demoPlan.kind SKETCH_PATH/PHYSICS_GRAPH requires sketchPath")
		}
	case "EQUATION_REVEAL":
		if !hasEquations {
			errs = append(errs, "demoPlan.kind EQUATION_REVEAL requires equations")
		}
	case "IMAGE_DRAW":
		if !hasImage {
			errs = append(errs, "demoPlan.kind IMAGE_DRAW requires imageUrl")
		}
	}
	return errs
}

func validateSketchPath(sketch *SketchPath) []string {
	var errs []string
	if sketch == nil {
		return errs
	}
	if strings.TrimSpace(sketch.Title) == "" {
		errs = append(errs, "sketchPath.title is required")
	}
	if len(sketch.Nodes) < 2 || len(sketch.Nodes) > 12 {
		errs = append(errs, "sketchPath.nodes must contain 2-12 items")
	}
	if len(sketch.Edges) < 1 || len(sketch.Edges) > 20 {
		errs = append(errs, "sketchPath.edges must contain 1-20 items")
	}
	if sketch.Draw.SecPerEdge <= 0 || sketch.Draw.SecPerEdge > 5 {
		errs = append(errs, "sketchPath.draw.secPerEdge must be in (0,5]")
	}
	if sketch.Draw.SecPerNode <= 0 || sketch.Draw.SecPerNode > 5 {
		errs = append(errs, "sketchPath.draw.secPerNode must be in (0,5]")
	}

	nodeIDs := make(map[string]struct{}, len(sketch.Nodes))
	for i, node := range sketch.Nodes {
		id := strings.TrimSpace(node.ID)
		if id == "" {
			errs = append(errs, fmt.Sprintf("sketchPath.nodes[%d].id is required", i))
			continue
		}
		if _, exists := nodeIDs[id]; exists {
			errs = append(errs, fmt.Sprintf("duplicate sketchPath node id: %s", id))
		}
		nodeIDs[id] = struct{}{}
		if strings.TrimSpace(node.Label) == "" {
			errs = append(errs, fmt.Sprintf("sketchPath.nodes[%d].label is required", i))
		}
		if node.XK < 0 || node.XK > 1 {
			errs = append(errs, fmt.Sprintf("sketchPath.nodes[%d].xk must be in [0,1]", i))
		}
		if node.YK < 0 || node.YK > 1 {
			errs = append(errs, fmt.Sprintf("sketchPath.nodes[%d].yk must be in [0,1]", i))
		}
	}

	for i, edge := range sketch.Edges {
		from := strings.TrimSpace(edge.From)
		to := strings.TrimSpace(edge.To)
		if from == "" || to == "" {
			errs = append(errs, fmt.Sprintf("sketchPath.edges[%d].from and to are required", i))
			continue
		}
		if _, ok := nodeIDs[from]; !ok {
			errs = append(errs, fmt.Sprintf("sketchPath.edges[%d].from references unknown node id: %s", i, from))
		}
		if _, ok := nodeIDs[to]; !ok {
			errs = append(errs, fmt.Sprintf("sketchPath.edges[%d].to references unknown node id: %s", i, to))
		}
	}

	return errs
}

func sanitizeAnimatedSlideWhitelist(raw map[string]interface{}) (AnimatedSlideContentResponse, bool) {
	out := AnimatedSlideContentResponse{}
	sanitized := false
	allowedKeys := map[string]struct{}{
		"topicId":            {},
		"explanation":        {},
		"keyPoints":          {},
		"examples":           {},
		"bridgeFromPrevious": {},
		"bridgeToNext":       {},
		"demoPlan":           {},
		"imageUrl":           {},
		"equations":          {},
		"sketchPath":         {},
	}
	for k := range raw {
		if _, ok := allowedKeys[k]; !ok {
			sanitized = true
		}
	}

	out.TopicID = asString(raw["topicId"])
	out.Explanation = asString(raw["explanation"])
	out.KeyPoints = asStringSlice(raw["keyPoints"])
	out.Examples = asStringSlice(raw["examples"])
	out.BridgeFromPrevious = asString(raw["bridgeFromPrevious"])
	out.BridgeToNext = asString(raw["bridgeToNext"])
	out.DemoPlan = sanitizeDemoPlan(raw["demoPlan"], &sanitized)
	out.ImageURL = asString(raw["imageUrl"])
	out.Equations = asStringSlice(raw["equations"])
	out.SketchPath = sanitizeSketchPath(raw["sketchPath"], &sanitized)
	return out, sanitized
}

func sanitizeDemoPlan(v interface{}, sanitized *bool) *DemoPlan {
	m, ok := v.(map[string]interface{})
	if !ok || m == nil {
		return nil
	}
	return &DemoPlan{
		Kind:              asString(m["kind"]),
		Reason:            asString(m["reason"]),
		AnimationEmphasis: asString(m["animationEmphasis"]),
	}
}

func sanitizeSketchPath(v interface{}, sanitized *bool) *SketchPath {
	m, ok := v.(map[string]interface{})
	if !ok || m == nil {
		return nil
	}
	sketch := &SketchPath{
		Title: asString(m["title"]),
		Draw: SketchDraw{
			SecPerEdge:     0.45,
			SecPerNode:     0.25,
			ShowArrowheads: true,
		},
	}

	if rawNodes, ok := m["nodes"].([]interface{}); ok {
		sketch.Nodes = make([]SketchNode, 0, len(rawNodes))
		for _, rv := range rawNodes {
			nm, ok := rv.(map[string]interface{})
			if !ok {
				*sanitized = true
				continue
			}
			sketch.Nodes = append(sketch.Nodes, SketchNode{
				ID:    asString(nm["id"]),
				Label: asString(nm["label"]),
				XK:    clamp01(asFloat(nm["xk"])),
				YK:    clamp01(asFloat(nm["yk"])),
			})
		}
	}

	if rawEdges, ok := m["edges"].([]interface{}); ok {
		sketch.Edges = make([]SketchEdge, 0, len(rawEdges))
		for _, rv := range rawEdges {
			em, ok := rv.(map[string]interface{})
			if !ok {
				*sanitized = true
				continue
			}
			sketch.Edges = append(sketch.Edges, SketchEdge{
				From:  asString(em["from"]),
				To:    asString(em["to"]),
				Label: asString(em["label"]),
			})
		}
	}

	if drawRaw, ok := m["draw"].(map[string]interface{}); ok {
		if _, exists := drawRaw["secPerEdge"]; exists {
			sketch.Draw.SecPerEdge = asFloat(drawRaw["secPerEdge"])
		} else {
			*sanitized = true
		}
		if _, exists := drawRaw["secPerNode"]; exists {
			sketch.Draw.SecPerNode = asFloat(drawRaw["secPerNode"])
		} else {
			*sanitized = true
		}
		if v, exists := drawRaw["showArrowheads"]; exists {
			if b, ok := v.(bool); ok {
				sketch.Draw.ShowArrowheads = b
			} else {
				*sanitized = true
			}
		} else {
			*sanitized = true
		}
	} else {
		*sanitized = true
	}
	return sketch
}

func normalizeAndBoundDetail(out *AnimatedSlideContentResponse, sanitized *bool) {
	out.TopicID = strings.TrimSpace(out.TopicID)
	out.Explanation = truncate(strings.TrimSpace(out.Explanation), 1200, sanitized)
	out.BridgeFromPrevious = truncate(strings.TrimSpace(out.BridgeFromPrevious), 240, sanitized)
	out.BridgeToNext = truncate(strings.TrimSpace(out.BridgeToNext), 240, sanitized)
	out.ImageURL = strings.TrimSpace(out.ImageURL)

	out.KeyPoints = boundStringArray(out.KeyPoints, 6, 220, sanitized)
	out.Examples = boundStringArray(out.Examples, 4, 260, sanitized)
	out.Equations = boundStringArray(out.Equations, 6, 120, sanitized)

	if out.SketchPath != nil {
		out.SketchPath.Title = truncate(strings.TrimSpace(out.SketchPath.Title), 120, sanitized)
		if out.SketchPath.Draw.SecPerEdge <= 0 {
			out.SketchPath.Draw.SecPerEdge = 0.45
			*sanitized = true
		}
		if out.SketchPath.Draw.SecPerNode <= 0 {
			out.SketchPath.Draw.SecPerNode = 0.25
			*sanitized = true
		}
		for i := range out.SketchPath.Nodes {
			node := &out.SketchPath.Nodes[i]
			node.ID = truncate(strings.TrimSpace(node.ID), 64, sanitized)
			node.Label = truncate(strings.TrimSpace(node.Label), 80, sanitized)
			oldX, oldY := node.XK, node.YK
			node.XK, node.YK = clamp01(node.XK), clamp01(node.YK)
			if oldX != node.XK || oldY != node.YK {
				*sanitized = true
			}
		}
		for i := range out.SketchPath.Edges {
			edge := &out.SketchPath.Edges[i]
			edge.From = truncate(strings.TrimSpace(edge.From), 64, sanitized)
			edge.To = truncate(strings.TrimSpace(edge.To), 64, sanitized)
			edge.Label = truncate(strings.TrimSpace(edge.Label), 80, sanitized)
		}
	}
}

func normalizeDemoPlan(out *AnimatedSlideContentResponse, sanitized *bool) {
	if out.DemoPlan == nil {
		return
	}
	out.DemoPlan.Kind = strings.ToUpper(strings.TrimSpace(out.DemoPlan.Kind))
	out.DemoPlan.Reason = truncate(strings.TrimSpace(out.DemoPlan.Reason), 240, sanitized)
	out.DemoPlan.AnimationEmphasis = strings.TrimSpace(strings.ToLower(out.DemoPlan.AnimationEmphasis))

	if out.DemoPlan.Kind == "" {
		out.DemoPlan.Kind = "NONE"
		*sanitized = true
	}
	if _, ok := allowedDemoPlanKinds[out.DemoPlan.Kind]; !ok {
		*sanitized = true
		out.DemoPlan.Kind = "NONE"
	}
	if _, ok := allowedDemoPlanEmphasis[out.DemoPlan.AnimationEmphasis]; !ok {
		out.DemoPlan.AnimationEmphasis = ""
		*sanitized = true
	}
}

func resolveSingleModality(out *AnimatedSlideContentResponse, sanitized *bool) {
	if out.DemoPlan != nil {
		kind := out.DemoPlan.Kind
		switch kind {
		case "NONE":
			if out.ImageURL != "" || len(out.Equations) > 0 || out.SketchPath != nil {
				out.ImageURL = ""
				out.Equations = nil
				out.SketchPath = nil
				*sanitized = true
			}
			return
		case "SKETCH_PATH", "PHYSICS_GRAPH":
			if out.ImageURL != "" {
				out.ImageURL = ""
				*sanitized = true
			}
			if len(out.Equations) > 0 {
				out.Equations = nil
				*sanitized = true
			}
			return
		case "EQUATION_REVEAL":
			if out.ImageURL != "" {
				out.ImageURL = ""
				*sanitized = true
			}
			if out.SketchPath != nil {
				out.SketchPath = nil
				*sanitized = true
			}
			return
		case "IMAGE_DRAW":
			if len(out.Equations) > 0 {
				out.Equations = nil
				*sanitized = true
			}
			if out.SketchPath != nil {
				out.SketchPath = nil
				*sanitized = true
			}
			return
		}
	}

	hasImage := out.ImageURL != ""
	hasSketch := out.SketchPath != nil && len(out.SketchPath.Nodes) > 0
	hasEquations := len(out.Equations) > 0

	if hasSketch {
		if hasImage {
			out.ImageURL = ""
			*sanitized = true
		}
		if hasEquations {
			out.Equations = nil
			*sanitized = true
		}
		return
	}
	if hasImage && hasEquations {
		out.Equations = nil
		*sanitized = true
	}
}

func boundStringArray(values []string, maxItems int, maxLen int, sanitized *bool) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		tv := truncate(strings.TrimSpace(v), maxLen, sanitized)
		if tv == "" {
			*sanitized = true
			continue
		}
		out = append(out, tv)
	}
	if len(out) > maxItems {
		out = out[:maxItems]
		*sanitized = true
	}
	return out
}

func truncate(s string, max int, sanitized *bool) string {
	if len(s) <= max {
		return s
	}
	*sanitized = true
	return s[:max]
}

func asString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func asStringSlice(v interface{}) []string {
	raw, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func asFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func impliesOverlayBehavior(text string) bool {
	l := strings.ToLower(text)
	banned := []string{
		"overlay",
		"overlaid",
		"on top of",
		"over the text",
		"above the text",
		"stacked on",
		"superimpose",
	}
	for _, token := range banned {
		if strings.Contains(l, token) {
			return true
		}
	}
	return false
}

func buildAnimatedUserPrompt(req AnimatedLessonRequest) string {
	return fmt.Sprintf(`Create a flowing animated lesson.
Topic: %s
Course context: %s
Objective: %s
Learner profile:
- Education: %s
- Interests: %s
- Goals: %s
Animation:
- Mode: %s
- Require level-order: %t
- Include prerequisite edges: %t

Output JSON:
{
  "lessonId": string,
  "title": string,
  "knowledgeGraph": [
    {
      "id": string,
      "title": string,
      "summary": string,
      "keyPoints": string[],
      "examples": string[],
      "prerequisites": string[]
    }
  ]
}

Important:
- Use natural lesson language that reads like a smooth teaching sequence.
- Do not mention: knowledge graph, graph, node, edge, prerequisite.`,
		req.Topic,
		req.CourseTitle,
		req.LessonObjective,
		req.Learner.EducationLevel,
		req.Learner.Interests,
		req.Learner.CareerGoals,
		req.Animation.Mode,
		req.Animation.RequireLevelOrder,
		req.Animation.IncludePrerequisiteEdges,
	)
}

func buildAnimatedRepairPrompt(req AnimatedLessonRequest, rawOutput string, validationErrors []string, parseErr error) string {
	var sb strings.Builder
	sb.WriteString("The previous output failed validation. Regenerate and return ONLY corrected JSON.\n")
	sb.WriteString("Original task:\n")
	sb.WriteString(buildAnimatedUserPrompt(req))
	sb.WriteString("\n\nValidation issues:\n")

	if parseErr != nil {
		sb.WriteString("- JSON parse error: ")
		sb.WriteString(parseErr.Error())
		sb.WriteString("\n")
	}
	for _, e := range validationErrors {
		sb.WriteString("- ")
		sb.WriteString(e)
		sb.WriteString("\n")
	}

	sb.WriteString("\nPrevious output:\n")
	sb.WriteString(utils.CleanJSONResponse(rawOutput))
	return sb.String()
}

func buildAnimatedSlideUserPrompt(req AnimatedSlideContentRequest) string {
	previousTitle := ""
	if req.PreviousTopic != nil {
		previousTitle = req.PreviousTopic.Title
	}
	nextTitle := ""
	if req.NextTopic != nil {
		nextTitle = req.NextTopic.Title
	}

	return fmt.Sprintf(`Generate detailed slide content for ONE topic in an animated lesson.
Course context: %s
Objective: %s
Learner profile:
- Education: %s
- Interests: %s
- Goals: %s
Animation mode: %s

Current topic:
- id: %s
- title: %s
- summary: %s
- key points seed: %v
- examples seed: %v

Previous topic (for continuity): %s
Next topic (for continuity): %s

Hard constraints:
- Return strict JSON only.
- Keep output frontend-interpretable (no custom node DSL, no canvas scripting).
- Text and demo are separate child regions; do not describe overlay behavior.
- Choose one primary demo modality for this topic:
  1) imageUrl
  2) sketchPath
  3) equations
- If using sketchPath, keep geometry simple and non-overlapping.
- Do not include keys outside the schema below.
- demoPlan.kind must be one of:
  - NONE
  - SKETCH_PATH
  - EQUATION_REVEAL
  - IMAGE_DRAW
  - PHYSICS_GRAPH
- demoPlan.kind must agree with emitted modality fields:
  - SKETCH_PATH -> provide sketchPath
  - EQUATION_REVEAL -> provide equations
  - IMAGE_DRAW -> provide imageUrl
  - PHYSICS_GRAPH -> provide sketchPath
  - NONE -> provide no modality field

Return JSON:
{
  "topicId": string,
  "explanation": string,
  "keyPoints": string[],
  "examples": string[],
  "bridgeFromPrevious": string,
  "bridgeToNext": string,
  "demoPlan": {
    "kind": string,
    "reason": string,
    "animationEmphasis": string
  },
  "imageUrl": string,
  "equations": string[],
  "sketchPath": {
    "title": string,
    "nodes": [{ "id": string, "label": string, "xk": number, "yk": number }],
    "edges": [{ "from": string, "to": string, "label": string }],
    "draw": { "secPerEdge": number, "secPerNode": number, "showArrowheads": boolean }
  }
}

Important:
- Write content as a flowing natural lesson narrative.
- Do not mention: knowledge graph, graph, node, edge, prerequisite.`,
		req.CourseTitle,
		req.LessonObjective,
		req.Learner.EducationLevel,
		req.Learner.Interests,
		req.Learner.CareerGoals,
		req.Animation.Mode,
		req.CurrentTopic.ID,
		req.CurrentTopic.Title,
		req.CurrentTopic.Summary,
		req.CurrentTopic.KeyPoints,
		req.CurrentTopic.Examples,
		previousTitle,
		nextTitle,
	)
}

func buildAnimatedSlideRepairPrompt(req AnimatedSlideContentRequest, rawOutput string, validationErrors []string, parseErr error) string {
	var sb strings.Builder
	sb.WriteString("The previous output failed validation. Regenerate and return ONLY corrected JSON.\n")
	sb.WriteString("Original task:\n")
	sb.WriteString(buildAnimatedSlideUserPrompt(req))
	sb.WriteString("\n\nValidation issues:\n")

	if parseErr != nil {
		sb.WriteString("- JSON parse error: ")
		sb.WriteString(parseErr.Error())
		sb.WriteString("\n")
	}
	for _, e := range validationErrors {
		sb.WriteString("- ")
		sb.WriteString(e)
		sb.WriteString("\n")
	}

	sb.WriteString("\nPrevious output:\n")
	sb.WriteString(utils.CleanJSONResponse(rawOutput))
	return sb.String()
}

func callLLMWithTimeout(ctx context.Context, client LLMClient, endpoint string, payload map[string]interface{}) (string, error) {
	type result struct {
		output string
		err    error
	}

	ch := make(chan result, 1)
	go func() {
		out, err := client.Call(endpoint, payload)
		ch <- result{output: out, err: err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-ch:
		return res.output, res.err
	}
}

func stableSeedFromRequest(req AnimatedLessonRequest) int {
	serialized, _ := json.Marshal(req)
	hasher := fnv.New32a()
	_, _ = hasher.Write(serialized)
	return int(hasher.Sum32() & 0x7fffffff)
}

func stableSeedFromSlideRequest(req AnimatedSlideContentRequest) int {
	serialized, _ := json.Marshal(req)
	hasher := fnv.New32a()
	_, _ = hasher.Write(serialized)
	return int(hasher.Sum32() & 0x7fffffff)
}

func writeAnimatedError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(animatedAPIError{
		Error: message,
		Code:  code,
	})
}

func writeAnimatedJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func clientIPFromRequest(r *http.Request) string {
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
