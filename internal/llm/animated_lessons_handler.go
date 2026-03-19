package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"agent-core/internal/cache"
	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

const (
	animatedKnowledgeGraphEndpoint = "/lessons/animated-knowledge-graph"
	animatedSlideContentEndpoint   = "/lessons/animated-slide-content"
	animatedSlideSceneEndpoint     = "/lessons/animated-slide-scene-content"
	animatedTreeSlideEndpoint      = "/animated-video/tree-slide"
	animatedNodeContentEndpoint    = "/animated-video/node-content"
	animatedMaxRequestBytes        = 1 << 20 // 1MB
	animatedDefaultRateLimit       = 30
	animatedSceneRateLimit         = 300
	animatedDefaultRateWindow      = time.Minute
	animatedMaxConcurrentSlides    = 16
	slideExplanationMaxChars       = 260
	slideBridgeMaxChars            = 90
	slideKeyPointsMaxItems         = 3
	slideKeyPointMaxChars          = 90
	slideExamplesMaxItems          = 1
	slideExampleMaxChars           = 110
	slideEquationsMaxItems         = 4
	slideEquationMaxChars          = 100
	slideDemoPlanReasonMaxChars    = 140
)

type slideValidationPolicy struct {
	RequireScene         bool
	RequireTimeline      bool
	RequireAnimationPlan bool
}

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
2.1) Keep textual content compact for slide UI: explanation <= 260 chars, each keyPoint <= 90 chars, each example <= 110 chars, bridge lines <= 90 chars.
2.2) Prefer 2-3 keyPoints and 0-1 examples.
3) Use previous and next topic only for smooth transitions.
4) Keep output frontend-interpretable only.
5) Text and demonstration are separate child regions; never describe overlay/stacking between them.
6) Choose exactly one primary demo modality for this topic: imageUrl OR sketchPath OR equations OR pedagogicalDiagram.
7) demoPlan.kind must be one of: NONE, SKETCH_PATH, EQUATION_REVEAL, IMAGE_DRAW, PHYSICS_GRAPH.
8) demoPlan.kind must agree with modality fields:
   SKETCH_PATH or PHYSICS_GRAPH -> sketchPath
   EQUATION_REVEAL -> equations
   IMAGE_DRAW -> imageUrl
   NONE -> no modality field
9) If using sketchPath, keep geometry simple and non-overlapping.
10) If presentationMode is scene, include scene with layout, animationPlan, and timeline.
11) Optional scene metadata is allowed only as declarative JSON under scene.
12) scene.animationPlan must contain 1-5 simple steps from allowed effects only.
13) Do not emit scripts, event handlers, draw calls, or imperative renderer commands.
14) For pedagogicalDiagram, set kind=PEDAGOGICAL_DIAGRAM and payload.variant in {graph, table, timeline, tree, stack_queue, coordinate_plane, flow_diagram, set_diagram, heatmap, field, shapes, character_scene, chart, chat}.
15) Optional payload.pattern must be one of: PROGRESSIVE_REVEAL, BUILD_UP_CONSTRUCTION, STEP_EXECUTION, STATE_TRANSITION, DEPENDENCY_PROPAGATION, VALUE_ACCUMULATION, FOCUS_SHIFT, STRUCTURAL_TRANSFORMATION, MORPHING, SIMULATION_LOOP, PARALLEL_COMPARISON, COUNTEREXAMPLE_REVEAL, PATH_TRACING, COLLAPSE_EXPAND, ENERGY_FLOW_SIMULATION, GRAPH_TRAVERSAL, DP_FILL.
16) If pedagogicalDiagram.payload.variant=shapes, each shapes[].kind must be one of: rect, oval, triangle, diamond.
17) No markdown, no commentary, no extra keys.
18) Do not mention internal structure words in learner-facing text: knowledge graph, graph, node, edge, prerequisite.`

const defaultAnimatedSlideTextFallbackSystemPrompt = `Return plain text only using the exact template fields requested.
Do not return JSON.
Do not include markdown fences.
Do not include extra headings or commentary.`

const defaultAnimatedTreeSlideSystemPrompt = `You generate scene-first animated lesson skeleton trees.
Return only JSON matching the required schema.
Rules:
1) Return skeleton tree only under treeSlides.
2) Allowed node keys only: id, title, type, physics, timeline, children.
3) Do not include content, layout, or animationPlan in tree nodes.
4) Each node must include physics and timeline.
5) children must be an array of valid node objects.
6) Use ids matching ^[a-z0-9_\-]{2,64}$.
7) Keep output deterministic, concise, and strict JSON only.
8) No markdown, no commentary, no extra keys.`

var allowedDemoPlanKinds = map[string]struct{}{
	"NONE":                {},
	"SKETCH_PATH":         {},
	"EQUATION_REVEAL":     {},
	"IMAGE_DRAW":          {},
	"PHYSICS_GRAPH":       {},
	"PEDAGOGICAL_DIAGRAM": {},
}

var allowedDemoPlanEmphasis = map[string]struct{}{
	"":                 {},
	"progressive_draw": {},
	"step_trace":       {},
	"settle_motion":    {},
	"reveal":           {},
}

var allowedSceneLayoutModes = map[string]struct{}{
	"full_text":            {},
	"equation_centered":    {},
	"split":                {},
	"grid":                 {},
	"stacked_progressive":  {},
	"overlay_layered":      {},
	"canvas_based_dynamic": {},
}

var allowedSceneAnimationEffects = map[string]struct{}{
	"fade-in":          {},
	"typewriter":       {},
	"draw":             {},
	"progressive_draw": {},
	"stagger_in":       {},
	"highlight":        {},
	"morph":            {},
	"zoom":             {},
}

var allowedScenePhysicsModes = map[string]struct{}{
	"none":           {},
	"force_graph":    {},
	"spring_expand":  {},
	"gravity_reveal": {},
	"elastic_drop":   {},
}

var allowedPedagogicalVariants = map[string]struct{}{
	"graph":            {},
	"table":            {},
	"timeline":         {},
	"tree":             {},
	"stack_queue":      {},
	"coordinate_plane": {},
	"flow_diagram":     {},
	"set_diagram":      {},
	"heatmap":          {},
	"field":            {},
	"shapes":           {},
	"character_scene":  {},
	"chart":            {},
	"chat":             {},
}

var allowedPedagogicalPatterns = map[string]struct{}{
	"":                         {},
	"PROGRESSIVE_REVEAL":       {},
	"BUILD_UP_CONSTRUCTION":    {},
	"STEP_EXECUTION":           {},
	"STATE_TRANSITION":         {},
	"DEPENDENCY_PROPAGATION":   {},
	"VALUE_ACCUMULATION":       {},
	"FOCUS_SHIFT":              {},
	"STRUCTURAL_TRANSFORMATION": {},
	"MORPHING":                 {},
	"SIMULATION_LOOP":          {},
	"PARALLEL_COMPARISON":      {},
	"COUNTEREXAMPLE_REVEAL":    {},
	"PATH_TRACING":             {},
	"COLLAPSE_EXPAND":          {},
	"ENERGY_FLOW_SIMULATION":   {},
	"GRAPH_TRAVERSAL":          {},
	"DP_FILL":                  {},
}

var allowedShapeKinds = map[string]struct{}{
	"rect":     {},
	"oval":     {},
	"triangle": {},
	"diamond":  {},
}

var shapeKindAliases = map[string]string{
	"rectangle": "rect",
	"box":       "rect",
	"square":    "rect",
	"circle":    "oval",
	"ellipse":   "oval",
	"rhombus":   "diamond",
}

var allowedCharacterPoses = map[string]struct{}{
	"idle":  {},
	"wave":  {},
	"point": {},
}

var allowedChartTypes = map[string]struct{}{
	"bar":  {},
	"line": {},
}

type animatedAPIError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type animatedRetryError struct {
	Error        string `json:"error"`
	Code         string `json:"code"`
	RetryAfterMs int    `json:"retryAfterMs,omitempty"`
	NodeID       string `json:"nodeId,omitempty"`
}

type animatedRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

type animatedRequestQueue struct {
	mu       sync.Mutex
	inFlight map[string]chan struct{}
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

func newAnimatedRequestQueue() *animatedRequestQueue {
	return &animatedRequestQueue{
		inFlight: make(map[string]chan struct{}),
	}
}

func (q *animatedRequestQueue) begin(key string) (waitCh chan struct{}, leader bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if ch, ok := q.inFlight[key]; ok {
		return ch, false
	}
	ch := make(chan struct{})
	q.inFlight[key] = ch
	return ch, true
}

func (q *animatedRequestQueue) end(key string) {
	q.mu.Lock()
	ch, ok := q.inFlight[key]
	if ok {
		delete(q.inFlight, key)
	}
	q.mu.Unlock()
	if ok {
		close(ch)
	}
}

var animatedEndpointLimiter = newAnimatedRateLimiter(animatedDefaultRateLimit, animatedDefaultRateWindow)
var animatedSceneEndpointLimiter = newAnimatedRateLimiter(animatedSceneRateLimit, animatedDefaultRateWindow)
var animatedSlideSemaphore = make(chan struct{}, animatedMaxConcurrentSlides)
var animatedQueue = newAnimatedRequestQueue()

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

	cachePayload, _ := json.Marshal(req)
	cacheKey := animatedKnowledgeGraphEndpoint + ":" + string(cachePayload)
	if cached, found, _ := cache.Get().CheckByPayload(cacheKey); found {
		var cachedResp AnimatedLessonResponse
		if err := json.Unmarshal(cached, &cachedResp); err == nil {
			writeAnimatedJSON(w, http.StatusOK, cachedResp)
			return
		}
	}

	var releaseQueue func()
	for {
		leader, release := beginQueuedRequest(cacheKey)
		if leader {
			releaseQueue = release
			break
		}
		if cached, found, _ := cache.Get().CheckByPayload(cacheKey); found {
			var cachedResp AnimatedLessonResponse
			if err := json.Unmarshal(cached, &cachedResp); err == nil {
				writeAnimatedJSON(w, http.StatusOK, cachedResp)
				return
			}
		}
	}
	defer releaseQueue()
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

	if b, err := json.Marshal(resp); err == nil {
		_ = cache.Get().CacheByPayload(cacheKey, b)
	}
	writeAnimatedJSON(w, http.StatusOK, resp)
}

// GenerateAnimatedTreeSlideHandler handles POST /animated-video/tree-slide.
func GenerateAnimatedTreeSlideHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAnimatedError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}
	if !hasJSONContentType(r) {
		writeAnimatedError(w, http.StatusBadRequest, "INVALID_REQUEST", "content-type must be application/json")
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
	if err := decoder.Decode(&req); err != nil {
		log.Printf("[GenerateAnimatedTreeSlide] invalid_json err=%v", err)
		writeAnimatedError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if err := validateAnimatedTreeSlideRequest(&req); err != nil {
		log.Printf("[GenerateAnimatedTreeSlide] invalid_request err=%v", err)
		writeAnimatedError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	cachePayload, _ := json.Marshal(req)
	cacheKey := animatedTreeSlideEndpoint + ":" + string(cachePayload)
	if cached, found, _ := cache.Get().CheckByPayload(cacheKey); found {
		var cachedResp AnimatedTreeSlideResponse
		if err := json.Unmarshal(cached, &cachedResp); err == nil {
			writeAnimatedJSON(w, http.StatusOK, cachedResp)
			return
		}
	}

	var releaseQueue func()
	for {
		leader, release := beginQueuedRequest(cacheKey)
		if leader {
			releaseQueue = release
			break
		}
		if cached, found, _ := cache.Get().CheckByPayload(cacheKey); found {
			var cachedResp AnimatedTreeSlideResponse
			if err := json.Unmarshal(cached, &cachedResp); err == nil {
				writeAnimatedJSON(w, http.StatusOK, cachedResp)
				return
			}
		}
	}
	defer releaseQueue()
	start := time.Now()
	resp, modelName, validationStatus, err := generateAnimatedTreeSlide(context.Background(), req)
	latency := time.Since(start).Milliseconds()
	log.Printf("[GenerateAnimatedTreeSlide] request_id=%s topic=%q model=%s latency_ms=%d token_usage=%s validation=%s",
		requestID, req.Topic, modelName, latency, "unavailable", validationStatus)

	if err != nil {
		if strings.Contains(err.Error(), "LLM_INVALID_OUTPUT") {
			writeAnimatedError(w, http.StatusBadGateway, "LLM_INVALID_OUTPUT", "failed to produce valid tree slide content")
			return
		}
		writeAnimatedError(w, http.StatusBadGateway, "LLM_FAILURE", "failed to generate tree slide content")
		return
	}

	if b, err := json.Marshal(resp); err == nil {
		_ = cache.Get().CacheByPayload(cacheKey, b)
	}
	writeAnimatedJSON(w, http.StatusOK, resp)
}

// GenerateAnimatedSlideContentHandler handles POST /lessons/animated-slide-content.
func GenerateAnimatedSlideContentHandler(w http.ResponseWriter, r *http.Request) {
	handleAnimatedSlideContent(w, r, animatedSlideContentEndpoint, false)
}

// GenerateAnimatedSlideSceneContentHandler handles POST /lessons/animated-slide-scene-content.
func GenerateAnimatedSlideSceneContentHandler(w http.ResponseWriter, r *http.Request) {
	handleAnimatedSlideContent(w, r, animatedSlideSceneEndpoint, true)
}

// GenerateAnimatedNodeContentHandler handles POST /animated-video/node-content.
func GenerateAnimatedNodeContentHandler(w http.ResponseWriter, r *http.Request) {
	handleAnimatedSlideContent(w, r, animatedNodeContentEndpoint, true)
}

func handleAnimatedSlideContent(w http.ResponseWriter, r *http.Request, endpoint string, requireSceneEndpoint bool) {
	if r.Method != http.MethodPost {
		writeAnimatedError(w, http.StatusMethodNotAllowed, "INVALID_REQUEST", "method not allowed")
		return
	}
	if !hasJSONContentType(r) {
		writeAnimatedError(w, http.StatusBadRequest, "INVALID_REQUEST", "content-type must be application/json")
		return
	}

	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = uuid.NewString()
	}

	clientIP := clientIPFromRequest(r)
	limiter := animatedEndpointLimiter
	if endpoint == animatedNodeContentEndpoint || endpoint == animatedTreeSlideEndpoint {
		limiter = animatedSceneEndpointLimiter
	}
	if !limiter.allow(clientIP) {
		if endpoint == animatedNodeContentEndpoint {
			writeAnimatedRetryError(w, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded", 1500, "")
			return
		}
		writeAnimatedError(w, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded")
		return
	}

	select {
	case animatedSlideSemaphore <- struct{}{}:
		defer func() { <-animatedSlideSemaphore }()
	default:
		if endpoint == animatedNodeContentEndpoint {
			writeAnimatedRetryError(w, http.StatusServiceUnavailable, "SCENE_NODE_NOT_READY", "node not ready", 1500, "")
			return
		}
		writeAnimatedError(w, http.StatusTooManyRequests, "CONCURRENCY_LIMITED", "too many slide content requests in progress")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, animatedMaxRequestBytes)

	var req AnimatedSlideContentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Printf("[GenerateAnimatedSlideContent] invalid_json endpoint=%s err=%v", endpoint, err)
		writeAnimatedError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if err := validateAnimatedSlideContentRequest(&req); err != nil {
		log.Printf("[GenerateAnimatedSlideContent] invalid_request endpoint=%s err=%v", endpoint, err)
		writeAnimatedError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	policy := resolveSlideValidationPolicy(req, requireSceneEndpoint)

	cachePayload, _ := json.Marshal(req)
	cacheKey := endpoint + ":" + string(cachePayload)
	if endpoint == animatedNodeContentEndpoint {
		cacheKey = endpoint + ":" + strings.TrimSpace(req.LessonID) + ":" + strings.TrimSpace(req.CurrentTopic.ID)
	}
	if cached, found, _ := cache.Get().CheckByPayload(cacheKey); found {
		var cachedResp AnimatedSlideContentResponse
		if err := json.Unmarshal(cached, &cachedResp); err == nil {
			writeAnimatedJSON(w, http.StatusOK, cachedResp)
			return
		}
	}

	var releaseQueue func()
	for {
		leader, release := beginQueuedRequest(cacheKey)
		if leader {
			releaseQueue = release
			break
		}
		if cached, found, _ := cache.Get().CheckByPayload(cacheKey); found {
			var cachedResp AnimatedSlideContentResponse
			if err := json.Unmarshal(cached, &cachedResp); err == nil {
				writeAnimatedJSON(w, http.StatusOK, cachedResp)
				return
			}
		}
	}
	defer releaseQueue()
	start := time.Now()
	resp, modelName, validationStatus, err := generateAnimatedSlideContent(context.Background(), req, endpoint, policy)
	latency := time.Since(start).Milliseconds()
	log.Printf("[GenerateAnimatedSlideContent] request_id=%s endpoint=%s topic_id=%q model=%s latency_ms=%d token_usage=%s validation=%s",
		requestID, endpoint, req.CurrentTopic.ID, modelName, latency, "unavailable", validationStatus)

	if err != nil {
		if strings.Contains(err.Error(), "LLM_INVALID_OUTPUT") {
			writeAnimatedError(w, http.StatusBadGateway, "LLM_INVALID_OUTPUT", "failed to produce valid slide content")
			return
		}
		if endpoint == animatedNodeContentEndpoint {
			writeAnimatedRetryError(w, http.StatusServiceUnavailable, "SCENE_NODE_NOT_READY", "node not ready", 2000, req.CurrentTopic.ID)
			return
		}
		writeAnimatedError(w, http.StatusBadGateway, "LLM_FAILURE", "failed to generate slide content")
		return
	}

	if b, err := json.Marshal(resp); err == nil {
		_ = cache.Get().CacheByPayload(cacheKey, b)
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
	forceModel := shouldForceModelOverride(clientName)
	modelForLogs := cfg.ModelId
	if !forceModel {
		modelForLogs = "provider_default(" + clientName + ")"
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		return AnimatedLessonResponse{}, modelForLogs, "failed", err
	}

	seed := stableSeedFromRequest(req)
	userPrompt := buildAnimatedUserPrompt(req)
	systemPrompt := loadAnimatedSystemPrompt()

	payload := map[string]interface{}{
		"temperature": 0.2,
		"max_tokens":  2200,
		"seed":        seed,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": systemPrompt + "\n\n" + userPrompt,
		},
	}
	if forceModel {
		payload["model"] = cfg.ModelId
	}
	rawOutput, err := callLLMWithTimeout(ctx, client, animatedKnowledgeGraphEndpoint, payload)
	if err != nil {
		return AnimatedLessonResponse{}, modelForLogs, "failed", err
	}

	resp, errs, parseErr := parseAndValidateAnimatedLLMOutput(rawOutput, req.Animation.IncludePrerequisiteEdges)
	if parseErr == nil && len(errs) == 0 {
		return resp, modelForLogs, "pass", nil
	}
	log.Printf("[GenerateAnimatedKnowledgeGraph] validation failed parse_error=%v errors=%v", parseErr, errs)
	return AnimatedLessonResponse{}, modelForLogs, "failed", fmt.Errorf("LLM_INVALID_OUTPUT")
}

func generateAnimatedTreeSlide(ctx context.Context, req AnimatedLessonRequest) (AnimatedTreeSlideResponse, string, string, error) {
	cfg, ok := configs.GetDeepseekConfig(animatedTreeSlideEndpoint)
	if !ok {
		cfg, _ = configs.GetDeepseekConfig("/chat")
	}

	clientName := cfg.ClientName
	if clientName == "" {
		clientName = "general"
	}
	forceModel := shouldForceModelOverride(clientName)
	modelForLogs := cfg.ModelId
	if !forceModel {
		modelForLogs = "provider_default(" + clientName + ")"
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		return AnimatedTreeSlideResponse{}, modelForLogs, "failed", err
	}

	seed := stableSeedFromRequest(req)
	userPrompt := buildAnimatedTreeSlideUserPrompt(req)
	systemPrompt := loadAnimatedTreeSlideSystemPrompt()

	payload := map[string]interface{}{
		"temperature": 0.2,
		"max_tokens":  2200,
		"seed":        seed,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": systemPrompt + "\n\n" + userPrompt,
		},
	}
	if forceModel {
		payload["model"] = cfg.ModelId
	}
	rawOutput, err := callLLMWithTimeout(ctx, client, animatedTreeSlideEndpoint, payload)
	if err != nil {
		return AnimatedTreeSlideResponse{}, modelForLogs, "failed", err
	}

	resp, errs, parseErr, sanitized := parseAndValidateAnimatedTreeSlideOutput(rawOutput)
	if parseErr == nil && len(errs) == 0 {
		if sanitized {
			return resp, modelForLogs, "pass_sanitized", nil
		}
		return resp, modelForLogs, "pass", nil
	}
	log.Printf("[GenerateAnimatedTreeSlide] validation failed parse_error=%v errors=%v", parseErr, errs)
	return AnimatedTreeSlideResponse{}, modelForLogs, "failed", fmt.Errorf("LLM_INVALID_OUTPUT")
}

func generateAnimatedSlideContent(ctx context.Context, req AnimatedSlideContentRequest, endpoint string, policy slideValidationPolicy) (AnimatedSlideContentResponse, string, string, error) {
	cfg, ok := configs.GetDeepseekConfig(endpoint)
	if !ok {
		cfg, _ = configs.GetDeepseekConfig("/chat")
	}

	clientName := cfg.ClientName
	if clientName == "" {
		clientName = "general"
	}
	forceModel := shouldForceModelOverride(clientName)
	modelForLogs := cfg.ModelId
	if !forceModel {
		modelForLogs = "provider_default(" + clientName + ")"
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		return AnimatedSlideContentResponse{}, modelForLogs, "failed", err
	}

	seed := stableSeedFromSlideRequest(req)
	systemPrompt := loadAnimatedSlideSystemPrompt()
	userPrompt := buildAnimatedSlideUserPrompt(req)

	payload := map[string]interface{}{
		"temperature": 0.2,
		"max_tokens":  1400,
		"seed":        seed,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{
			"prompt": systemPrompt + "\n\n" + userPrompt,
		},
	}
	if forceModel {
		payload["model"] = cfg.ModelId
	}
	rawOutput, err := callLLMWithTimeout(ctx, client, endpoint, payload)
	if err != nil {
		return AnimatedSlideContentResponse{}, modelForLogs, "failed", err
	}

	resp, errs, parseErr, sanitized := parseAndValidateAnimatedSlideOutput(rawOutput, req.CurrentTopic.ID, policy)
	if parseErr == nil && len(errs) == 0 {
		if sanitized {
			return resp, modelForLogs, "pass_sanitized", nil
		}
		return resp, modelForLogs, "pass", nil
	}
	log.Printf("[GenerateAnimatedSlideContent] validation failed parse_error=%v errors=%v", parseErr, errs)
	return AnimatedSlideContentResponse{}, modelForLogs, "failed", fmt.Errorf("LLM_INVALID_OUTPUT")
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

func loadAnimatedTreeSlideSystemPrompt() string {
	systemPrompt, err := prompt.GetPrompt("animatedTreeSlidePrompt")
	if err != nil || strings.TrimSpace(systemPrompt) == "" {
		return defaultAnimatedTreeSlideSystemPrompt
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

func parseAndValidateAnimatedTreeSlideOutput(raw string) (AnimatedTreeSlideResponse, []string, error, bool) {
	cleaned := utils.CleanJSONResponse(raw)
	var decoded map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader([]byte(cleaned)))
	if err := decoder.Decode(&decoded); err != nil {
		return AnimatedTreeSlideResponse{}, nil, err, false
	}
	resp, sanitized := sanitizeAnimatedTreeSlideWhitelist(decoded)
	normalizeTreeSlides(&resp, &sanitized)
	errs := validateAnimatedTreeSlideResponse(resp)
	return resp, errs, nil, sanitized
}

func parseAndValidateAnimatedSlideOutput(raw string, expectedTopicID string, policy slideValidationPolicy) (AnimatedSlideContentResponse, []string, error, bool) {
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
	errs := validateAnimatedSlideContentResponse(resp, expectedTopicID, policy)
	return resp, errs, nil, sanitized
}

func parseAnimatedSlideTextFallbackOutput(raw string, req AnimatedSlideContentRequest, policy slideValidationPolicy) (AnimatedSlideContentResponse, []string) {
	resp := AnimatedSlideContentResponse{
		TopicID:            req.CurrentTopic.ID,
		Explanation:        req.CurrentTopic.Summary,
		KeyPoints:          append([]string{}, req.CurrentTopic.KeyPoints...),
		Examples:           append([]string{}, req.CurrentTopic.Examples...),
		BridgeFromPrevious: "",
		BridgeToNext:       "",
		DemoPlan:           &DemoPlan{Kind: "NONE"},
	}

	if len(resp.KeyPoints) == 0 {
		resp.KeyPoints = []string{"Understand the core idea of this topic."}
	}

	section := ""
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "TOPIC_ID:"):
			resp.TopicID = strings.TrimSpace(strings.TrimPrefix(line, "TOPIC_ID:"))
			section = ""
		case strings.HasPrefix(line, "EXPLANATION:"):
			resp.Explanation = strings.TrimSpace(strings.TrimPrefix(line, "EXPLANATION:"))
			section = ""
		case strings.HasPrefix(line, "BRIDGE_FROM_PREVIOUS:"):
			resp.BridgeFromPrevious = strings.TrimSpace(strings.TrimPrefix(line, "BRIDGE_FROM_PREVIOUS:"))
			section = ""
		case strings.HasPrefix(line, "BRIDGE_TO_NEXT:"):
			resp.BridgeToNext = strings.TrimSpace(strings.TrimPrefix(line, "BRIDGE_TO_NEXT:"))
			section = ""
		case strings.HasPrefix(line, "DEMO_KIND:"):
			kind := strings.TrimSpace(strings.TrimPrefix(line, "DEMO_KIND:"))
			resp.DemoPlan = &DemoPlan{Kind: strings.ToUpper(kind)}
			section = ""
		case strings.HasPrefix(line, "IMAGE_URL:"):
			resp.ImageURL = strings.TrimSpace(strings.TrimPrefix(line, "IMAGE_URL:"))
			section = ""
		case strings.HasPrefix(line, "KEY_POINTS:"):
			resp.KeyPoints = nil
			section = "key_points"
		case strings.HasPrefix(line, "EXAMPLES:"):
			resp.Examples = nil
			section = "examples"
		case strings.HasPrefix(line, "EQUATIONS:"):
			resp.Equations = nil
			section = "equations"
		case strings.HasPrefix(line, "SCENE_LAYOUT_MODE:"):
			mode := strings.TrimSpace(strings.TrimPrefix(line, "SCENE_LAYOUT_MODE:"))
			if resp.Scene == nil {
				resp.Scene = &SceneSpec{}
			}
			if resp.Scene.Layout == nil {
				resp.Scene.Layout = &SceneLayout{}
			}
			resp.Scene.Layout.Mode = mode
			section = ""
		case strings.HasPrefix(line, "SCENE_TIMELINE_MAX_WAIT_SEC:"):
			val := strings.TrimSpace(strings.TrimPrefix(line, "SCENE_TIMELINE_MAX_WAIT_SEC:"))
			f, _ := strconv.ParseFloat(val, 64)
			if resp.Scene == nil {
				resp.Scene = &SceneSpec{}
			}
			if resp.Scene.Timeline == nil {
				resp.Scene.Timeline = &SceneTimeline{}
			}
			resp.Scene.Timeline.MaxWaitSec = f
			resp.Scene.Timeline.WaitForAnimationPlan = true
		case strings.HasPrefix(line, "SCENE_ANIMATION_PLAN:"):
			if resp.Scene == nil {
				resp.Scene = &SceneSpec{}
			}
			resp.Scene.AnimationPlan = nil
			section = "scene_plan"
		case strings.HasPrefix(line, "- "):
			item := strings.TrimSpace(strings.TrimPrefix(line, "- "))
			switch section {
			case "key_points":
				resp.KeyPoints = append(resp.KeyPoints, item)
			case "examples":
				resp.Examples = append(resp.Examples, item)
			case "equations":
				resp.Equations = append(resp.Equations, item)
			case "scene_plan":
				step := parseScenePlanStep(item)
				if step.Step > 0 {
					resp.Scene.AnimationPlan = append(resp.Scene.AnimationPlan, step)
				}
			}
		}
	}

	if policy.RequireScene || req.PresentationMode == "scene" || req.ScenePolicy.RequireScene {
		if resp.Scene == nil {
			resp.Scene = &SceneSpec{}
		}
		if resp.Scene.Layout == nil {
			resp.Scene.Layout = &SceneLayout{Mode: "split"}
		}
		if resp.Scene.Timeline == nil {
			resp.Scene.Timeline = &SceneTimeline{WaitForAnimationPlan: true, RequireSignalDone: false, MaxWaitSec: 10}
		}
		if (policy.RequireAnimationPlan || req.ScenePolicy.RequireAnimationPlan) && len(resp.Scene.AnimationPlan) == 0 {
			resp.Scene.AnimationPlan = []SceneAnimationStep{
				{Step: 1, Target: "title", Effect: "fade-in", DurationSec: 0.3},
			}
		}
	}

	sanitized := false
	normalizeAndBoundDetail(&resp, &sanitized)
	normalizeDemoPlan(&resp, &sanitized)
	resolveSingleModality(&resp, &sanitized)
	errs := validateAnimatedSlideContentResponse(resp, req.CurrentTopic.ID, policy)
	return resp, errs
}

func parseScenePlanStep(item string) SceneAnimationStep {
	step := SceneAnimationStep{}
	parts := strings.Split(item, "|")
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(strings.ToLower(kv[0]))
		v := strings.TrimSpace(kv[1])
		switch k {
		case "step":
			i, _ := strconv.Atoi(v)
			step.Step = i
		case "target":
			step.Target = v
		case "effect":
			step.Effect = v
		case "durationsec":
			f, _ := strconv.ParseFloat(v, 64)
			step.DurationSec = f
		}
	}
	return step
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

func validateAnimatedTreeSlideRequest(req *AnimatedLessonRequest) error {
	normalizeLessonSharedFields(&req.CourseTitle, &req.LessonObjective, &req.Learner, &req.Animation)
	req.Topic = strings.TrimSpace(req.Topic)
	req.PresentationMode = strings.TrimSpace(strings.ToLower(req.PresentationMode))

	if req.Topic == "" {
		return fmt.Errorf("topic is required")
	}
	if len(req.Topic) > 200 {
		return fmt.Errorf("topic is too long (max 200)")
	}
	if req.PresentationMode != "" && req.PresentationMode != "scene" {
		return fmt.Errorf("presentationMode must be scene when provided")
	}
	if err := validateLessonSharedFieldLengths(req.CourseTitle, req.LessonObjective, req.Learner, req.Animation); err != nil {
		return err
	}
	return nil
}

func validateAnimatedSlideContentRequest(req *AnimatedSlideContentRequest) error {
	normalizeLessonSharedFields(&req.CourseTitle, &req.LessonObjective, &req.Learner, &req.Animation)
	req.LessonID = strings.TrimSpace(req.LessonID)
	req.PresentationMode = strings.TrimSpace(strings.ToLower(req.PresentationMode))
	req.CurrentTopic.ID = strings.TrimSpace(req.CurrentTopic.ID)
	req.CurrentTopic.Title = strings.TrimSpace(req.CurrentTopic.Title)
	req.CurrentTopic.Type = strings.TrimSpace(req.CurrentTopic.Type)
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
	if len(req.CurrentTopic.Type) > 40 {
		return fmt.Errorf("currentTopic.type is too long (max 40)")
	}
	if req.PresentationMode != "" && req.PresentationMode != "scene" {
		return fmt.Errorf("presentationMode must be scene when provided")
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

func resolveSlideValidationPolicy(req AnimatedSlideContentRequest, requireSceneEndpoint bool) slideValidationPolicy {
	p := slideValidationPolicy{
		RequireScene:         requireSceneEndpoint,
		RequireTimeline:      requireSceneEndpoint,
		RequireAnimationPlan: requireSceneEndpoint,
	}

	if req.PresentationMode == "scene" || req.ScenePolicy.RequireScene {
		p.RequireScene = true
	}
	if req.ScenePolicy.RequireTimeline {
		p.RequireTimeline = true
	}
	if req.ScenePolicy.RequireAnimationPlan {
		p.RequireAnimationPlan = true
	}
	if p.RequireTimeline || p.RequireAnimationPlan {
		p.RequireScene = true
	}

	return p
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

func validateAnimatedTreeSlideResponse(resp AnimatedTreeSlideResponse) []string {
	var errs []string
	if strings.TrimSpace(resp.LessonID) == "" {
		errs = append(errs, "lessonId is required")
	}
	if strings.TrimSpace(resp.Title) == "" {
		errs = append(errs, "title is required")
	}
	if len(resp.TreeSlides) < 1 {
		errs = append(errs, "treeSlides must contain at least 1 node")
		return errs
	}

	seen := map[string]struct{}{}
	var walk func(node TreeSlideNode, path string)
	walk = func(node TreeSlideNode, path string) {
		id := strings.TrimSpace(node.ID)
		if id == "" {
			errs = append(errs, path+".id is required")
		} else {
			if !knowledgeNodeIDPattern.MatchString(id) {
				errs = append(errs, path+".id must match ^[a-z0-9_\\-]{2,64}$")
			}
			if _, ok := seen[id]; ok {
				errs = append(errs, "duplicate tree node id: "+id)
			}
			seen[id] = struct{}{}
		}

		if strings.TrimSpace(node.Title) == "" {
			errs = append(errs, path+".title is required")
		}
		if strings.TrimSpace(node.Type) == "" {
			errs = append(errs, path+".type is required")
		}

		if node.Physics == nil {
			errs = append(errs, path+".physics is required")
		} else {
			mode := strings.TrimSpace(strings.ToLower(node.Physics.Mode))
			if _, ok := allowedScenePhysicsModes[mode]; !ok {
				errs = append(errs, path+".physics.mode invalid")
			}
			if !node.Physics.Enabled && mode != "none" {
				errs = append(errs, path+".physics.mode invalid for enabled=false")
			}
		}

		if node.Timeline == nil {
			errs = append(errs, path+".timeline is required")
		} else if node.Timeline.MaxWaitSec < 1 || node.Timeline.MaxWaitSec > 30 {
			errs = append(errs, fmt.Sprintf("%s.timeline.maxWaitSec out of bounds: %.2f", path, node.Timeline.MaxWaitSec))
		}

		for i, child := range node.Children {
			walk(child, fmt.Sprintf("%s.children[%d]", path, i))
		}
	}

	for i, node := range resp.TreeSlides {
		walk(node, fmt.Sprintf("treeSlides[%d]", i))
	}
	return errs
}

func validateAnimatedSlideContentResponse(resp AnimatedSlideContentResponse, expectedTopicID string, policy slideValidationPolicy) []string {
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
	if len(resp.Explanation) > slideExplanationMaxChars {
		errs = append(errs, fmt.Sprintf("explanation is too long (max %d)", slideExplanationMaxChars))
	}
	if len(resp.KeyPoints) < 1 || len(resp.KeyPoints) > slideKeyPointsMaxItems {
		errs = append(errs, fmt.Sprintf("keyPoints must contain 1-%d items", slideKeyPointsMaxItems))
	}
	if len(resp.Examples) > slideExamplesMaxItems {
		errs = append(errs, fmt.Sprintf("examples must contain 0-%d items", slideExamplesMaxItems))
	}
	if len(resp.Equations) > slideEquationsMaxItems {
		errs = append(errs, fmt.Sprintf("equations must contain 0-%d items", slideEquationsMaxItems))
	}
	for i, kp := range resp.KeyPoints {
		if strings.TrimSpace(kp) == "" {
			errs = append(errs, fmt.Sprintf("keyPoints[%d] must be non-empty", i))
		}
		if len(strings.TrimSpace(kp)) > slideKeyPointMaxChars {
			errs = append(errs, fmt.Sprintf("keyPoints[%d] is too long (max %d)", i, slideKeyPointMaxChars))
		}
	}
	for i, ex := range resp.Examples {
		if strings.TrimSpace(ex) == "" {
			errs = append(errs, fmt.Sprintf("examples[%d] must be non-empty", i))
		}
		if len(strings.TrimSpace(ex)) > slideExampleMaxChars {
			errs = append(errs, fmt.Sprintf("examples[%d] is too long (max %d)", i, slideExampleMaxChars))
		}
	}
	for i, eq := range resp.Equations {
		eq = strings.TrimSpace(eq)
		if eq == "" {
			errs = append(errs, fmt.Sprintf("equations[%d] must be non-empty", i))
		}
		if len(eq) > slideEquationMaxChars {
			errs = append(errs, fmt.Sprintf("equations[%d] is too long (max %d)", i, slideEquationMaxChars))
		}
	}
	if len(resp.BridgeFromPrevious) > slideBridgeMaxChars {
		errs = append(errs, fmt.Sprintf("bridgeFromPrevious is too long (max %d)", slideBridgeMaxChars))
	}
	if len(resp.BridgeToNext) > slideBridgeMaxChars {
		errs = append(errs, fmt.Sprintf("bridgeToNext is too long (max %d)", slideBridgeMaxChars))
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
			errs = append(errs, "demoPlan.kind must be one of NONE|SKETCH_PATH|EQUATION_REVEAL|IMAGE_DRAW|PHYSICS_GRAPH|PEDAGOGICAL_DIAGRAM")
		}
		if len(strings.TrimSpace(resp.DemoPlan.Reason)) > slideDemoPlanReasonMaxChars {
			errs = append(errs, fmt.Sprintf("demoPlan.reason is too long (max %d)", slideDemoPlanReasonMaxChars))
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
	if resp.PedagogicalDiagram != nil {
		errs = append(errs, validatePedagogicalDiagram(resp.PedagogicalDiagram)...)
	}
	if policy.RequireScene && resp.Scene == nil {
		errs = append(errs, "scene is required")
	}
	if resp.Scene != nil {
		errs = append(errs, validateScene(resp.Scene, resp.DemoPlan, policy)...)
	}
	errs = append(errs, validateDemoPlanConsistency(resp)...)

	if impliesOverlayBehavior(resp.Explanation) || impliesOverlayBehavior(resp.BridgeFromPrevious) || impliesOverlayBehavior(resp.BridgeToNext) {
		errs = append(errs, "text must not imply overlay/stacking between text and demo regions")
	}

	return errs
}

func validatePedagogicalDiagram(d *PedagogicalDiagramNode) []string {
	var errs []string
	if d == nil {
		return errs
	}
	if strings.TrimSpace(d.Kind) != "PEDAGOGICAL_DIAGRAM" {
		errs = append(errs, "pedagogicalDiagram.kind must equal PEDAGOGICAL_DIAGRAM")
	}
	if strings.TrimSpace(d.ID) == "" {
		errs = append(errs, "pedagogicalDiagram.id is required")
	}
	variant := strings.TrimSpace(strings.ToLower(d.Payload.Variant))
	if _, ok := allowedPedagogicalVariants[variant]; !ok {
		errs = append(errs, "pedagogicalDiagram.payload.variant invalid")
		return errs
	}
	pattern := strings.TrimSpace(strings.ToUpper(d.Payload.Pattern))
	if _, ok := allowedPedagogicalPatterns[pattern]; !ok {
		errs = append(errs, "pedagogicalDiagram.payload.pattern invalid")
	}

	switch variant {
	case "shapes":
		if len(d.Payload.Shapes) < 1 || len(d.Payload.Shapes) > 32 {
			errs = append(errs, "pedagogicalDiagram.payload.shapes must contain 1-32 items")
		}
		for i, s := range d.Payload.Shapes {
			if _, ok := allowedShapeKinds[s.Kind]; !ok {
				errs = append(errs, fmt.Sprintf("pedagogicalDiagram.payload.shapes[%d].kind invalid", i))
			}
		}
	case "character_scene":
		if len(d.Payload.Characters) < 1 || len(d.Payload.Characters) > 8 {
			errs = append(errs, "pedagogicalDiagram.payload.characters must contain 1-8 items")
		}
		for i, c := range d.Payload.Characters {
			if strings.TrimSpace(c.Name) == "" {
				errs = append(errs, fmt.Sprintf("pedagogicalDiagram.payload.characters[%d].name is required", i))
			}
			if _, ok := allowedCharacterPoses[c.Pose]; !ok && c.Pose != "" {
				errs = append(errs, fmt.Sprintf("pedagogicalDiagram.payload.characters[%d].pose invalid", i))
			}
		}
	case "table":
		if d.Payload.Table == nil {
			errs = append(errs, "pedagogicalDiagram.payload.table is required for variant table")
		} else {
			if len(d.Payload.Table.Headers) < 1 || len(d.Payload.Table.Headers) > 8 {
				errs = append(errs, "pedagogicalDiagram.payload.table.headers must contain 1-8 items")
			}
			if len(d.Payload.Table.Rows) < 1 || len(d.Payload.Table.Rows) > 10 {
				errs = append(errs, "pedagogicalDiagram.payload.table.rows must contain 1-10 items")
			}
		}
	case "chart":
		if d.Payload.Chart == nil {
			errs = append(errs, "pedagogicalDiagram.payload.chart is required for variant chart")
		} else {
			if _, ok := allowedChartTypes[d.Payload.Chart.Type]; !ok {
				errs = append(errs, "pedagogicalDiagram.payload.chart.type invalid")
			}
			if len(d.Payload.Chart.Labels) < 1 || len(d.Payload.Chart.Labels) > 10 {
				errs = append(errs, "pedagogicalDiagram.payload.chart.labels must contain 1-10 items")
			}
			if len(d.Payload.Chart.Series) < 1 || len(d.Payload.Chart.Series) > 4 {
				errs = append(errs, "pedagogicalDiagram.payload.chart.series must contain 1-4 items")
			}
		}
	case "graph":
		if d.Payload.Graph == nil {
			errs = append(errs, "pedagogicalDiagram.payload.graph is required for variant graph")
		} else {
			if len(d.Payload.Graph.Nodes) < 1 || len(d.Payload.Graph.Nodes) > 20 {
				errs = append(errs, "pedagogicalDiagram.payload.graph.nodes must contain 1-20 items")
			}
			if len(d.Payload.Graph.Edges) > 30 {
				errs = append(errs, "pedagogicalDiagram.payload.graph.edges must contain 0-30 items")
			}
			nodeIDs := make(map[string]struct{}, len(d.Payload.Graph.Nodes))
			for i, n := range d.Payload.Graph.Nodes {
				if strings.TrimSpace(n.ID) == "" {
					errs = append(errs, fmt.Sprintf("pedagogicalDiagram.payload.graph.nodes[%d].id is required", i))
				}
				nodeIDs[n.ID] = struct{}{}
			}
			for i, e := range d.Payload.Graph.Edges {
				if _, ok := nodeIDs[e.From]; !ok {
					errs = append(errs, fmt.Sprintf("pedagogicalDiagram.payload.graph.edges[%d].from unknown node", i))
				}
				if _, ok := nodeIDs[e.To]; !ok {
					errs = append(errs, fmt.Sprintf("pedagogicalDiagram.payload.graph.edges[%d].to unknown node", i))
				}
			}
		}
	case "timeline":
		if d.Payload.Timeline == nil {
			errs = append(errs, "pedagogicalDiagram.payload.timeline is required for variant timeline")
		} else if len(d.Payload.Timeline.Events) < 1 || len(d.Payload.Timeline.Events) > 14 {
			errs = append(errs, "pedagogicalDiagram.payload.timeline.events must contain 1-14 items")
		}
	case "tree":
		if d.Payload.Tree == nil {
			errs = append(errs, "pedagogicalDiagram.payload.tree is required for variant tree")
		} else {
			if len(d.Payload.Tree.Nodes) < 1 || len(d.Payload.Tree.Nodes) > 20 {
				errs = append(errs, "pedagogicalDiagram.payload.tree.nodes must contain 1-20 items")
			}
			if len(d.Payload.Tree.Edges) > 30 {
				errs = append(errs, "pedagogicalDiagram.payload.tree.edges must contain 0-30 items")
			}
		}
	case "stack_queue":
		if d.Payload.StackQueue == nil {
			errs = append(errs, "pedagogicalDiagram.payload.stackQueue is required for variant stack_queue")
		} else {
			mode := strings.TrimSpace(strings.ToLower(d.Payload.StackQueue.Mode))
			if mode != "stack" && mode != "queue" {
				errs = append(errs, "pedagogicalDiagram.payload.stackQueue.mode must be stack or queue")
			}
			if len(d.Payload.StackQueue.Items) < 1 || len(d.Payload.StackQueue.Items) > 10 {
				errs = append(errs, "pedagogicalDiagram.payload.stackQueue.items must contain 1-10 items")
			}
		}
	case "coordinate_plane":
		if d.Payload.CoordinatePlane == nil {
			errs = append(errs, "pedagogicalDiagram.payload.coordinatePlane is required for variant coordinate_plane")
		} else {
			if len(d.Payload.CoordinatePlane.Points) > 16 {
				errs = append(errs, "pedagogicalDiagram.payload.coordinatePlane.points must contain 0-16 items")
			}
			if len(d.Payload.CoordinatePlane.Curves) > 4 {
				errs = append(errs, "pedagogicalDiagram.payload.coordinatePlane.curves must contain 0-4 items")
			}
			for i, c := range d.Payload.CoordinatePlane.Curves {
				if len(c.Points) < 2 || len(c.Points) > 24 {
					errs = append(errs, fmt.Sprintf("pedagogicalDiagram.payload.coordinatePlane.curves[%d].points must contain 2-24 items", i))
				}
			}
		}
	case "flow_diagram":
		if d.Payload.FlowDiagram == nil {
			errs = append(errs, "pedagogicalDiagram.payload.flowDiagram is required for variant flow_diagram")
		} else {
			if len(d.Payload.FlowDiagram.Stages) < 1 || len(d.Payload.FlowDiagram.Stages) > 12 {
				errs = append(errs, "pedagogicalDiagram.payload.flowDiagram.stages must contain 1-12 items")
			}
			if len(d.Payload.FlowDiagram.Links) > 20 {
				errs = append(errs, "pedagogicalDiagram.payload.flowDiagram.links must contain 0-20 items")
			}
		}
	case "set_diagram":
		if d.Payload.SetDiagram == nil {
			errs = append(errs, "pedagogicalDiagram.payload.setDiagram is required for variant set_diagram")
		} else if len(d.Payload.SetDiagram.Sets) < 1 || len(d.Payload.SetDiagram.Sets) > 4 {
			errs = append(errs, "pedagogicalDiagram.payload.setDiagram.sets must contain 1-4 items")
		}
	case "heatmap":
		if d.Payload.Heatmap == nil {
			errs = append(errs, "pedagogicalDiagram.payload.heatmap is required for variant heatmap")
		} else {
			if len(d.Payload.Heatmap.Values) < 1 || len(d.Payload.Heatmap.Values) > 10 {
				errs = append(errs, "pedagogicalDiagram.payload.heatmap.values must contain 1-10 rows")
			}
			for i, row := range d.Payload.Heatmap.Values {
				if len(row) < 1 || len(row) > 10 {
					errs = append(errs, fmt.Sprintf("pedagogicalDiagram.payload.heatmap.values[%d] must contain 1-10 columns", i))
				}
			}
		}
	case "field":
		if d.Payload.Field == nil {
			errs = append(errs, "pedagogicalDiagram.payload.field is required for variant field")
		} else if len(d.Payload.Field.Vectors) < 1 || len(d.Payload.Field.Vectors) > 24 {
			errs = append(errs, "pedagogicalDiagram.payload.field.vectors must contain 1-24 items")
		}
	case "chat":
		if d.Payload.Chat == nil {
			errs = append(errs, "pedagogicalDiagram.payload.chat is required for variant chat")
		} else {
			if len(d.Payload.Chat.Participants) < 1 || len(d.Payload.Chat.Participants) > 6 {
				errs = append(errs, "pedagogicalDiagram.payload.chat.participants must contain 1-6 items")
			}
			if len(d.Payload.Chat.Messages) < 1 || len(d.Payload.Chat.Messages) > 12 {
				errs = append(errs, "pedagogicalDiagram.payload.chat.messages must contain 1-12 items")
			}
		}
	}

	if d.Scene != nil {
		errs = append(errs, validateScene(d.Scene, nil, slideValidationPolicy{})...)
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
	hasPedagogical := resp.PedagogicalDiagram != nil

	switch kind {
	case "NONE":
		if hasImage || hasSketch || hasEquations || hasPedagogical {
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
	case "PEDAGOGICAL_DIAGRAM":
		if !hasPedagogical {
			errs = append(errs, "demoPlan.kind PEDAGOGICAL_DIAGRAM requires pedagogicalDiagram")
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

func validateScene(scene *SceneSpec, demoPlan *DemoPlan, policy slideValidationPolicy) []string {
	var errs []string
	if scene == nil {
		return errs
	}

	if scene.Layout == nil || strings.TrimSpace(scene.Layout.Mode) == "" {
		errs = append(errs, "scene.layout.mode is required")
	} else if _, ok := allowedSceneLayoutModes[strings.TrimSpace(strings.ToLower(scene.Layout.Mode))]; !ok {
		errs = append(errs, fmt.Sprintf("scene.layout.mode invalid: %q", scene.Layout.Mode))
	}

	if policy.RequireAnimationPlan && len(scene.AnimationPlan) < 1 {
		errs = append(errs, "scene.animationPlan must contain 1-5 steps")
	}
	if len(scene.AnimationPlan) > 5 {
		errs = append(errs, "scene.animationPlan must contain 1-5 steps")
	}
	seenSteps := map[int]struct{}{}
	prev := -1
	for i, step := range scene.AnimationPlan {
		if step.Step <= 0 {
			errs = append(errs, fmt.Sprintf("scene.animationPlan[%d].step must be positive", i))
		}
		if _, exists := seenSteps[step.Step]; exists {
			errs = append(errs, fmt.Sprintf("scene.animationPlan.step duplicate: %d", step.Step))
		}
		seenSteps[step.Step] = struct{}{}
		if step.Step < prev {
			errs = append(errs, "scene.animationPlan must be sorted ascending by step")
		}
		prev = step.Step
		if strings.TrimSpace(step.Target) == "" {
			errs = append(errs, fmt.Sprintf("scene.animationPlan[%d].target is required", i))
		}
		if _, ok := allowedSceneAnimationEffects[strings.TrimSpace(strings.ToLower(step.Effect))]; !ok {
			errs = append(errs, fmt.Sprintf("scene.animationPlan[%d].effect invalid: %q", i, step.Effect))
		}
		if step.DurationSec <= 0 || step.DurationSec > 20 {
			errs = append(errs, fmt.Sprintf("scene.animationPlan[%d].durationSec must be in (0,20]", i))
		}
	}

	if policy.RequireTimeline && scene.Timeline == nil {
		errs = append(errs, "scene.timeline is required")
	}
	if scene.Timeline != nil {
		if scene.Timeline.MaxWaitSec < 1 || scene.Timeline.MaxWaitSec > 30 {
			errs = append(errs, fmt.Sprintf("scene.timeline.maxWaitSec out of bounds: %.2f", scene.Timeline.MaxWaitSec))
		}
	}

	if scene.Physics != nil {
		mode := strings.TrimSpace(strings.ToLower(scene.Physics.Mode))
		if _, ok := allowedScenePhysicsModes[mode]; !ok {
			errs = append(errs, fmt.Sprintf("scene.physics.mode invalid: %q", scene.Physics.Mode))
		}
		if !scene.Physics.Enabled && mode != "none" {
			errs = append(errs, "scene.physics.mode invalid for enabled=false")
		}
	}

	if demoPlan != nil {
		layoutMode := ""
		if scene.Layout != nil {
			layoutMode = strings.TrimSpace(strings.ToLower(scene.Layout.Mode))
		}
		kind := strings.TrimSpace(demoPlan.Kind)
		switch kind {
		case "SKETCH_PATH", "PHYSICS_GRAPH":
			if layoutMode != "" && layoutMode != "split" && layoutMode != "canvas_based_dynamic" {
				errs = append(errs, "scene.layout.mode should be split or canvas_based_dynamic for SKETCH_PATH/PHYSICS_GRAPH")
			}
		case "EQUATION_REVEAL":
			if layoutMode != "" && layoutMode != "equation_centered" && layoutMode != "split" {
				errs = append(errs, "scene.layout.mode should be equation_centered or split for EQUATION_REVEAL")
			}
		case "NONE":
			for i, step := range scene.AnimationPlan {
				target := strings.ToLower(strings.TrimSpace(step.Target))
				if strings.Contains(target, "demo") || strings.Contains(target, "right_demo") {
					errs = append(errs, fmt.Sprintf("scene.animationPlan[%d].target demo-only element invalid for demoPlan.kind=NONE", i))
				}
			}
		}
	}

	return errs
}

func sanitizeAnimatedTreeSlideWhitelist(raw map[string]interface{}) (AnimatedTreeSlideResponse, bool) {
	out := AnimatedTreeSlideResponse{}
	sanitized := false
	allowedKeys := map[string]struct{}{
		"lessonId":   {},
		"title":      {},
		"treeSlides": {},
	}
	for k := range raw {
		if _, ok := allowedKeys[k]; !ok {
			sanitized = true
		}
	}

	out.LessonID = asString(raw["lessonId"])
	out.Title = asString(raw["title"])
	if rawSlides, ok := raw["treeSlides"].([]interface{}); ok {
		out.TreeSlides = make([]TreeSlideNode, 0, len(rawSlides))
		for _, item := range rawSlides {
			node, ok := sanitizeTreeSlideNode(item, &sanitized)
			if !ok {
				continue
			}
			out.TreeSlides = append(out.TreeSlides, node)
		}
	}
	return out, sanitized
}

func sanitizeTreeSlideNode(v interface{}, sanitized *bool) (TreeSlideNode, bool) {
	m, ok := v.(map[string]interface{})
	if !ok || m == nil {
		*sanitized = true
		return TreeSlideNode{}, false
	}
	allowedKeys := map[string]struct{}{
		"id":       {},
		"title":    {},
		"type":     {},
		"physics":  {},
		"timeline": {},
		"children": {},
	}
	for k := range m {
		if _, ok := allowedKeys[k]; !ok {
			*sanitized = true
		}
	}

	out := TreeSlideNode{
		ID:    asString(m["id"]),
		Title: asString(m["title"]),
		Type:  asString(m["type"]),
	}

	if phy, ok := m["physics"].(map[string]interface{}); ok {
		out.Physics = &ScenePhysics{
			Enabled: asBool(phy["enabled"]),
			Mode:    asString(phy["mode"]),
			Params:  asAnyMap(phy["params"]),
		}
	}

	if timeline, ok := m["timeline"].(map[string]interface{}); ok {
		out.Timeline = &SceneTimeline{
			WaitForAnimationPlan: asBool(timeline["waitForAnimationPlan"]),
			RequireSignalDone:    asBool(timeline["requireSignalDone"]),
			MaxWaitSec:           asFloat(timeline["maxWaitSec"]),
		}
	}

	if rawChildren, ok := m["children"].([]interface{}); ok {
		out.Children = make([]TreeSlideNode, 0, len(rawChildren))
		for _, child := range rawChildren {
			n, ok := sanitizeTreeSlideNode(child, sanitized)
			if !ok {
				continue
			}
			out.Children = append(out.Children, n)
		}
	}

	return out, true
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
		"pedagogicalDiagram": {},
		"scene":              {},
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
	out.PedagogicalDiagram = sanitizePedagogicalDiagram(raw["pedagogicalDiagram"], &sanitized)
	out.Scene = sanitizeScene(raw["scene"], &sanitized)
	return out, sanitized
}

func sanitizePedagogicalDiagram(v interface{}, sanitized *bool) *PedagogicalDiagramNode {
	m, ok := v.(map[string]interface{})
	if !ok || m == nil {
		return nil
	}
	allowedKeys := map[string]struct{}{
		"id":      {},
		"kind":    {},
		"deps":    {},
		"payload": {},
		"scene":   {},
	}
	for k := range m {
		if _, ok := allowedKeys[k]; !ok {
			*sanitized = true
		}
	}

	out := &PedagogicalDiagramNode{
		ID:   asString(m["id"]),
		Kind: asString(m["kind"]),
		Deps: asStringSlice(m["deps"]),
	}

	if payloadRaw, ok := m["payload"].(map[string]interface{}); ok {
		out.Payload = sanitizePedagogicalPayload(payloadRaw, sanitized)
	}
	out.Scene = sanitizeScene(m["scene"], sanitized)
	return out
}

func sanitizePedagogicalPayload(m map[string]interface{}, sanitized *bool) PedagogicalDiagramPayload {
	allowedKeys := map[string]struct{}{
		"variant":     {},
		"pattern":     {},
		"title":       {},
		"subtitle":    {},
		"durationSec": {},
		"shapes":      {},
		"characters":  {},
		"table":       {},
		"chart":       {},
		"graph":       {},
		"timeline":    {},
		"tree":        {},
		"stackQueue":  {},
		"coordinatePlane": {},
		"flowDiagram": {},
		"setDiagram":  {},
		"heatmap":     {},
		"field":       {},
		"chat":        {},
	}
	for k := range m {
		if _, ok := allowedKeys[k]; !ok {
			*sanitized = true
		}
	}

	out := PedagogicalDiagramPayload{
		Variant:     asString(m["variant"]),
		Pattern:     asString(m["pattern"]),
		Title:       asString(m["title"]),
		Subtitle:    asString(m["subtitle"]),
		DurationSec: asFloat(m["durationSec"]),
	}

	if rawShapes, ok := m["shapes"].([]interface{}); ok {
		out.Shapes = make([]DiagramShape, 0, len(rawShapes))
		for _, item := range rawShapes {
			sm, ok := item.(map[string]interface{})
			if !ok {
				*sanitized = true
				continue
			}
			out.Shapes = append(out.Shapes, DiagramShape{
				Kind:  asString(sm["kind"]),
				XK:    clamp01(asFloat(sm["xk"])),
				YK:    clamp01(asFloat(sm["yk"])),
				WK:    clamp01(asFloat(sm["wk"])),
				HK:    clamp01(asFloat(sm["hk"])),
				Label: asString(sm["label"]),
			})
		}
	}
	if rawChars, ok := m["characters"].([]interface{}); ok {
		out.Characters = make([]DiagramCharacter, 0, len(rawChars))
		for _, item := range rawChars {
			cm, ok := item.(map[string]interface{})
			if !ok {
				*sanitized = true
				continue
			}
			out.Characters = append(out.Characters, DiagramCharacter{
				Name: asString(cm["name"]),
				XK:   clamp01(asFloat(cm["xk"])),
				YK:   clamp01(asFloat(cm["yk"])),
				Pose: asString(cm["pose"]),
			})
		}
	}
	if tm, ok := m["table"].(map[string]interface{}); ok {
		table := &DiagramTable{
			Headers: asStringSlice(tm["headers"]),
			Caption: asString(tm["caption"]),
		}
		if rowsRaw, ok := tm["rows"].([]interface{}); ok {
			table.Rows = make([][]string, 0, len(rowsRaw))
			for _, rowItem := range rowsRaw {
				rowRaw, ok := rowItem.([]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				row := make([]string, 0, len(rowRaw))
				for _, cell := range rowRaw {
					row = append(row, asString(cell))
				}
				table.Rows = append(table.Rows, row)
			}
		}
		out.Table = table
	}
	if cm, ok := m["chart"].(map[string]interface{}); ok {
		chart := &DiagramChart{
			Type:   asString(cm["type"]),
			Labels: asStringSlice(cm["labels"]),
		}
		if seriesRaw, ok := cm["series"].([]interface{}); ok {
			chart.Series = make([]DiagramChartSeries, 0, len(seriesRaw))
			for _, s := range seriesRaw {
				sm, ok := s.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				ds := DiagramChartSeries{Name: asString(sm["name"])}
				if valsRaw, ok := sm["values"].([]interface{}); ok {
					ds.Values = make([]float64, 0, len(valsRaw))
					for _, v := range valsRaw {
						ds.Values = append(ds.Values, asFloat(v))
					}
				}
				chart.Series = append(chart.Series, ds)
			}
		}
		out.Chart = chart
	}
	if gm, ok := m["graph"].(map[string]interface{}); ok {
		graph := &DiagramGraph{}
		if nodesRaw, ok := gm["nodes"].([]interface{}); ok {
			graph.Nodes = make([]DiagramGraphNode, 0, len(nodesRaw))
			for _, n := range nodesRaw {
				nm, ok := n.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				graph.Nodes = append(graph.Nodes, DiagramGraphNode{
					ID:    asString(nm["id"]),
					Label: asString(nm["label"]),
					XK:    clamp01(asFloat(nm["xk"])),
					YK:    clamp01(asFloat(nm["yk"])),
				})
			}
		}
		if edgesRaw, ok := gm["edges"].([]interface{}); ok {
			graph.Edges = make([]DiagramGraphEdge, 0, len(edgesRaw))
			for _, e := range edgesRaw {
				em, ok := e.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				graph.Edges = append(graph.Edges, DiagramGraphEdge{
					From:     asString(em["from"]),
					To:       asString(em["to"]),
					Label:    asString(em["label"]),
					Directed: asBool(em["directed"]),
				})
			}
		}
		out.Graph = graph
	}
	if tm, ok := m["timeline"].(map[string]interface{}); ok {
		timeline := &DiagramTimeline{}
		if eventsRaw, ok := tm["events"].([]interface{}); ok {
			timeline.Events = make([]DiagramTimelineEvent, 0, len(eventsRaw))
			for _, e := range eventsRaw {
				em, ok := e.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				timeline.Events = append(timeline.Events, DiagramTimelineEvent{
					ID:    asString(em["id"]),
					Label: asString(em["label"]),
					AtK:   clamp01(asFloat(em["atK"])),
				})
			}
		}
		out.Timeline = timeline
	}
	if trm, ok := m["tree"].(map[string]interface{}); ok {
		tree := &DiagramTree{}
		if nodesRaw, ok := trm["nodes"].([]interface{}); ok {
			tree.Nodes = make([]DiagramGraphNode, 0, len(nodesRaw))
			for _, n := range nodesRaw {
				nm, ok := n.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				tree.Nodes = append(tree.Nodes, DiagramGraphNode{
					ID:    asString(nm["id"]),
					Label: asString(nm["label"]),
					XK:    clamp01(asFloat(nm["xk"])),
					YK:    clamp01(asFloat(nm["yk"])),
				})
			}
		}
		if edgesRaw, ok := trm["edges"].([]interface{}); ok {
			tree.Edges = make([]DiagramGraphEdge, 0, len(edgesRaw))
			for _, e := range edgesRaw {
				em, ok := e.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				tree.Edges = append(tree.Edges, DiagramGraphEdge{
					From:     asString(em["from"]),
					To:       asString(em["to"]),
					Label:    asString(em["label"]),
					Directed: asBool(em["directed"]),
				})
			}
		}
		out.Tree = tree
	}
	if sqm, ok := m["stackQueue"].(map[string]interface{}); ok {
		out.StackQueue = &DiagramStackQueue{
			Mode:  asString(sqm["mode"]),
			Items: asStringSlice(sqm["items"]),
		}
	}
	if cpm, ok := m["coordinatePlane"].(map[string]interface{}); ok {
		cp := &DiagramCoordinatePlane{}
		if pointsRaw, ok := cpm["points"].([]interface{}); ok {
			cp.Points = make([]DiagramGraphNode, 0, len(pointsRaw))
			for _, p := range pointsRaw {
				pm, ok := p.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				cp.Points = append(cp.Points, DiagramGraphNode{
					ID:    asString(pm["id"]),
					Label: asString(pm["label"]),
					XK:    clamp01(asFloat(pm["xk"])),
					YK:    clamp01(asFloat(pm["yk"])),
				})
			}
		}
		if curvesRaw, ok := cpm["curves"].([]interface{}); ok {
			cp.Curves = make([]DiagramCoordinateCurve, 0, len(curvesRaw))
			for _, c := range curvesRaw {
				cm, ok := c.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				curve := DiagramCoordinateCurve{ID: asString(cm["id"])}
				if cPointsRaw, ok := cm["points"].([]interface{}); ok {
					curve.Points = make([]DiagramCoordinatePt, 0, len(cPointsRaw))
					for _, cpItem := range cPointsRaw {
						cpm, ok := cpItem.(map[string]interface{})
						if !ok {
							*sanitized = true
							continue
						}
						curve.Points = append(curve.Points, DiagramCoordinatePt{
							XK: clamp01(asFloat(cpm["xk"])),
							YK: clamp01(asFloat(cpm["yk"])),
						})
					}
				}
				cp.Curves = append(cp.Curves, curve)
			}
		}
		out.CoordinatePlane = cp
	}
	if fdm, ok := m["flowDiagram"].(map[string]interface{}); ok {
		fd := &DiagramFlowDiagram{}
		if stagesRaw, ok := fdm["stages"].([]interface{}); ok {
			fd.Stages = make([]DiagramFlowStage, 0, len(stagesRaw))
			for _, s := range stagesRaw {
				sm, ok := s.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				fd.Stages = append(fd.Stages, DiagramFlowStage{
					ID:    asString(sm["id"]),
					Label: asString(sm["label"]),
					XK:    clamp01(asFloat(sm["xk"])),
					YK:    clamp01(asFloat(sm["yk"])),
				})
			}
		}
		if linksRaw, ok := fdm["links"].([]interface{}); ok {
			fd.Links = make([]DiagramFlowLink, 0, len(linksRaw))
			for _, l := range linksRaw {
				lm, ok := l.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				fd.Links = append(fd.Links, DiagramFlowLink{
					From:  asString(lm["from"]),
					To:    asString(lm["to"]),
					Label: asString(lm["label"]),
				})
			}
		}
		out.FlowDiagram = fd
	}
	if sdm, ok := m["setDiagram"].(map[string]interface{}); ok {
		sd := &DiagramSetDiagram{}
		if setsRaw, ok := sdm["sets"].([]interface{}); ok {
			sd.Sets = make([]DiagramSet, 0, len(setsRaw))
			for _, s := range setsRaw {
				sm, ok := s.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				sd.Sets = append(sd.Sets, DiagramSet{
					ID:    asString(sm["id"]),
					Label: asString(sm["label"]),
					XK:    clamp01(asFloat(sm["xk"])),
					YK:    clamp01(asFloat(sm["yk"])),
					RK:    clamp01(asFloat(sm["rk"])),
				})
			}
		}
		out.SetDiagram = sd
	}
	if hm, ok := m["heatmap"].(map[string]interface{}); ok {
		h := &DiagramHeatmap{
			LabelsX: asStringSlice(hm["labelsX"]),
			LabelsY: asStringSlice(hm["labelsY"]),
		}
		if valsRaw, ok := hm["values"].([]interface{}); ok {
			h.Values = make([][]float64, 0, len(valsRaw))
			for _, rowRaw := range valsRaw {
				rowAny, ok := rowRaw.([]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				row := make([]float64, 0, len(rowAny))
				for _, c := range rowAny {
					row = append(row, asFloat(c))
				}
				h.Values = append(h.Values, row)
			}
		}
		out.Heatmap = h
	}
	if fm, ok := m["field"].(map[string]interface{}); ok {
		field := &DiagramField{}
		if vectorsRaw, ok := fm["vectors"].([]interface{}); ok {
			field.Vectors = make([]DiagramVector, 0, len(vectorsRaw))
			for _, v := range vectorsRaw {
				vm, ok := v.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				field.Vectors = append(field.Vectors, DiagramVector{
					ID:  asString(vm["id"]),
					XK:  clamp01(asFloat(vm["xk"])),
					YK:  clamp01(asFloat(vm["yk"])),
					DXK: asFloat(vm["dxk"]),
					DYK: asFloat(vm["dyk"]),
				})
			}
		}
		out.Field = field
	}
	if chm, ok := m["chat"].(map[string]interface{}); ok {
		chat := &DiagramChat{}
		if partsRaw, ok := chm["participants"].([]interface{}); ok {
			chat.Participants = make([]DiagramParticipant, 0, len(partsRaw))
			for _, p := range partsRaw {
				pm, ok := p.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				chat.Participants = append(chat.Participants, DiagramParticipant{
					ID:   asString(pm["id"]),
					Name: asString(pm["name"]),
					XK:   clamp01(asFloat(pm["xk"])),
				})
			}
		}
		if msgsRaw, ok := chm["messages"].([]interface{}); ok {
			chat.Messages = make([]DiagramMessage, 0, len(msgsRaw))
			for _, m := range msgsRaw {
				mm, ok := m.(map[string]interface{})
				if !ok {
					*sanitized = true
					continue
				}
				chat.Messages = append(chat.Messages, DiagramMessage{
					SpeakerID: asString(mm["speakerId"]),
					Text:      asString(mm["text"]),
					AtSec:     asFloat(mm["atSec"]),
				})
			}
		}
		out.Chat = chat
	}

	return out
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

func sanitizeScene(v interface{}, sanitized *bool) *SceneSpec {
	m, ok := v.(map[string]interface{})
	if !ok || m == nil {
		return nil
	}
	scene := &SceneSpec{}

	if layoutRaw, ok := m["layout"].(map[string]interface{}); ok {
		scene.Layout = &SceneLayout{
			Mode:   asString(layoutRaw["mode"]),
			Params: asAnyMap(layoutRaw["params"]),
		}
	}

	if planRaw, ok := m["animationPlan"].([]interface{}); ok {
		scene.AnimationPlan = make([]SceneAnimationStep, 0, len(planRaw))
		for _, item := range planRaw {
			stepMap, ok := item.(map[string]interface{})
			if !ok {
				*sanitized = true
				continue
			}
			scene.AnimationPlan = append(scene.AnimationPlan, SceneAnimationStep{
				Step:          asInt(stepMap["step"]),
				Target:        asString(stepMap["target"]),
				Effect:        asString(stepMap["effect"]),
				DurationSec:   asFloat(stepMap["durationSec"]),
				WaitForSignal: asBool(stepMap["waitForSignal"]),
			})
		}
	}

	if timelineRaw, ok := m["timeline"].(map[string]interface{}); ok {
		scene.Timeline = &SceneTimeline{
			WaitForAnimationPlan: asBool(timelineRaw["waitForAnimationPlan"]),
			RequireSignalDone:    asBool(timelineRaw["requireSignalDone"]),
			MaxWaitSec:           asFloat(timelineRaw["maxWaitSec"]),
		}
	}

	if physicsRaw, ok := m["physics"].(map[string]interface{}); ok {
		scene.Physics = &ScenePhysics{
			Enabled: asBool(physicsRaw["enabled"]),
			Mode:    asString(physicsRaw["mode"]),
			Params:  asAnyMap(physicsRaw["params"]),
		}
	}

	if layersRaw, ok := m["layers"].(map[string]interface{}); ok {
		scene.Layers = &SceneLayers{
			Background: getBoolOrDefault(layersRaw, "background", true),
			Content:    getBoolOrDefault(layersRaw, "content", true),
			Highlight:  getBoolOrDefault(layersRaw, "highlight", true),
			Transition: getBoolOrDefault(layersRaw, "transition", true),
			Debug:      getBoolOrDefault(layersRaw, "debug", false),
		}
	}

	return scene
}

func normalizeTreeSlides(out *AnimatedTreeSlideResponse, sanitized *bool) {
	out.LessonID = truncate(strings.TrimSpace(out.LessonID), 100, sanitized)
	out.Title = truncate(strings.TrimSpace(out.Title), 200, sanitized)
	for i := range out.TreeSlides {
		normalizeTreeSlideNode(&out.TreeSlides[i], sanitized)
	}
}

func normalizeTreeSlideNode(node *TreeSlideNode, sanitized *bool) {
	node.ID = truncate(strings.TrimSpace(node.ID), 64, sanitized)
	node.Title = truncate(strings.TrimSpace(node.Title), 140, sanitized)
	node.Type = truncate(strings.TrimSpace(strings.ToLower(node.Type)), 40, sanitized)

	if node.Physics == nil {
		node.Physics = &ScenePhysics{
			Enabled: false,
			Mode:    "none",
			Params:  map[string]interface{}{},
		}
		*sanitized = true
	} else {
		node.Physics.Mode = strings.TrimSpace(strings.ToLower(node.Physics.Mode))
		if node.Physics.Params == nil {
			node.Physics.Params = map[string]interface{}{}
		}
		if !node.Physics.Enabled && node.Physics.Mode != "none" {
			node.Physics.Mode = "none"
			*sanitized = true
		}
		if node.Physics.Enabled && node.Physics.Mode == "" {
			node.Physics.Mode = "force_graph"
			*sanitized = true
		}
	}

	if node.Timeline == nil {
		node.Timeline = &SceneTimeline{
			WaitForAnimationPlan: true,
			RequireSignalDone:    false,
			MaxWaitSec:           10,
		}
		*sanitized = true
	} else {
		if node.Timeline.MaxWaitSec <= 0 {
			node.Timeline.MaxWaitSec = 10
			*sanitized = true
		}
		if node.Timeline.MaxWaitSec < 1 {
			node.Timeline.MaxWaitSec = 1
			*sanitized = true
		}
		if node.Timeline.MaxWaitSec > 30 {
			node.Timeline.MaxWaitSec = 30
			*sanitized = true
		}
	}

	if node.Children == nil {
		node.Children = []TreeSlideNode{}
		*sanitized = true
	}
	for i := range node.Children {
		normalizeTreeSlideNode(&node.Children[i], sanitized)
	}
}

func normalizeAndBoundDetail(out *AnimatedSlideContentResponse, sanitized *bool) {
	out.TopicID = strings.TrimSpace(out.TopicID)
	out.Explanation = truncate(strings.TrimSpace(out.Explanation), slideExplanationMaxChars, sanitized)
	out.BridgeFromPrevious = truncate(strings.TrimSpace(out.BridgeFromPrevious), slideBridgeMaxChars, sanitized)
	out.BridgeToNext = truncate(strings.TrimSpace(out.BridgeToNext), slideBridgeMaxChars, sanitized)
	out.ImageURL = strings.TrimSpace(out.ImageURL)

	out.KeyPoints = boundStringArray(out.KeyPoints, slideKeyPointsMaxItems, slideKeyPointMaxChars, sanitized)
	out.Examples = boundStringArray(out.Examples, slideExamplesMaxItems, slideExampleMaxChars, sanitized)
	out.Equations = boundStringArray(out.Equations, slideEquationsMaxItems, slideEquationMaxChars, sanitized)

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
	if out.PedagogicalDiagram != nil {
		normalizePedagogicalDiagram(out.PedagogicalDiagram, sanitized)
	}

	if out.Scene != nil {
		normalizeScene(out.Scene, sanitized)
	}
}

func normalizePedagogicalDiagram(d *PedagogicalDiagramNode, sanitized *bool) {
	if d == nil {
		return
	}
	d.ID = truncate(strings.TrimSpace(d.ID), 64, sanitized)
	d.Kind = strings.TrimSpace(strings.ToUpper(d.Kind))
	for i := range d.Deps {
		d.Deps[i] = truncate(strings.TrimSpace(d.Deps[i]), 64, sanitized)
	}
	d.Payload.Variant = strings.TrimSpace(strings.ToLower(d.Payload.Variant))
	d.Payload.Pattern = strings.TrimSpace(strings.ToUpper(d.Payload.Pattern))
	d.Payload.Title = truncate(strings.TrimSpace(d.Payload.Title), 140, sanitized)
	d.Payload.Subtitle = truncate(strings.TrimSpace(d.Payload.Subtitle), 180, sanitized)
	if d.Payload.DurationSec < 0 {
		d.Payload.DurationSec = 0
		*sanitized = true
	}
	for i := range d.Payload.Shapes {
		s := &d.Payload.Shapes[i]
		s.Kind = strings.TrimSpace(strings.ToLower(s.Kind))
		if mapped, ok := shapeKindAliases[s.Kind]; ok {
			s.Kind = mapped
			*sanitized = true
		}
		s.Label = truncate(strings.TrimSpace(s.Label), 80, sanitized)
		s.XK, s.YK, s.WK, s.HK = clamp01(s.XK), clamp01(s.YK), clamp01(s.WK), clamp01(s.HK)
	}
	for i := range d.Payload.Characters {
		c := &d.Payload.Characters[i]
		c.Name = truncate(strings.TrimSpace(c.Name), 80, sanitized)
		c.Pose = strings.TrimSpace(strings.ToLower(c.Pose))
		c.XK, c.YK = clamp01(c.XK), clamp01(c.YK)
	}
	if d.Payload.Table != nil {
		for i := range d.Payload.Table.Headers {
			d.Payload.Table.Headers[i] = truncate(strings.TrimSpace(d.Payload.Table.Headers[i]), 60, sanitized)
		}
		for i := range d.Payload.Table.Rows {
			for j := range d.Payload.Table.Rows[i] {
				d.Payload.Table.Rows[i][j] = truncate(strings.TrimSpace(d.Payload.Table.Rows[i][j]), 80, sanitized)
			}
		}
		d.Payload.Table.Caption = truncate(strings.TrimSpace(d.Payload.Table.Caption), 200, sanitized)
	}
	if d.Payload.Chart != nil {
		d.Payload.Chart.Type = strings.TrimSpace(strings.ToLower(d.Payload.Chart.Type))
		for i := range d.Payload.Chart.Labels {
			d.Payload.Chart.Labels[i] = truncate(strings.TrimSpace(d.Payload.Chart.Labels[i]), 50, sanitized)
		}
		for i := range d.Payload.Chart.Series {
			d.Payload.Chart.Series[i].Name = truncate(strings.TrimSpace(d.Payload.Chart.Series[i].Name), 60, sanitized)
		}
	}
	if d.Payload.Graph != nil {
		for i := range d.Payload.Graph.Nodes {
			n := &d.Payload.Graph.Nodes[i]
			n.ID = truncate(strings.TrimSpace(n.ID), 64, sanitized)
			n.Label = truncate(strings.TrimSpace(n.Label), 80, sanitized)
			n.XK, n.YK = clamp01(n.XK), clamp01(n.YK)
		}
		for i := range d.Payload.Graph.Edges {
			e := &d.Payload.Graph.Edges[i]
			e.From = truncate(strings.TrimSpace(e.From), 64, sanitized)
			e.To = truncate(strings.TrimSpace(e.To), 64, sanitized)
			e.Label = truncate(strings.TrimSpace(e.Label), 80, sanitized)
		}
	}
	if d.Payload.Timeline != nil {
		for i := range d.Payload.Timeline.Events {
			e := &d.Payload.Timeline.Events[i]
			e.ID = truncate(strings.TrimSpace(e.ID), 64, sanitized)
			e.Label = truncate(strings.TrimSpace(e.Label), 90, sanitized)
			e.AtK = clamp01(e.AtK)
		}
	}
	if d.Payload.Tree != nil {
		for i := range d.Payload.Tree.Nodes {
			n := &d.Payload.Tree.Nodes[i]
			n.ID = truncate(strings.TrimSpace(n.ID), 64, sanitized)
			n.Label = truncate(strings.TrimSpace(n.Label), 80, sanitized)
			n.XK, n.YK = clamp01(n.XK), clamp01(n.YK)
		}
		for i := range d.Payload.Tree.Edges {
			e := &d.Payload.Tree.Edges[i]
			e.From = truncate(strings.TrimSpace(e.From), 64, sanitized)
			e.To = truncate(strings.TrimSpace(e.To), 64, sanitized)
			e.Label = truncate(strings.TrimSpace(e.Label), 80, sanitized)
		}
	}
	if d.Payload.StackQueue != nil {
		d.Payload.StackQueue.Mode = strings.TrimSpace(strings.ToLower(d.Payload.StackQueue.Mode))
		for i := range d.Payload.StackQueue.Items {
			d.Payload.StackQueue.Items[i] = truncate(strings.TrimSpace(d.Payload.StackQueue.Items[i]), 80, sanitized)
		}
	}
	if d.Payload.CoordinatePlane != nil {
		for i := range d.Payload.CoordinatePlane.Points {
			p := &d.Payload.CoordinatePlane.Points[i]
			p.ID = truncate(strings.TrimSpace(p.ID), 64, sanitized)
			p.Label = truncate(strings.TrimSpace(p.Label), 80, sanitized)
			p.XK, p.YK = clamp01(p.XK), clamp01(p.YK)
		}
		for i := range d.Payload.CoordinatePlane.Curves {
			c := &d.Payload.CoordinatePlane.Curves[i]
			c.ID = truncate(strings.TrimSpace(c.ID), 64, sanitized)
			for j := range c.Points {
				c.Points[j].XK = clamp01(c.Points[j].XK)
				c.Points[j].YK = clamp01(c.Points[j].YK)
			}
		}
	}
	if d.Payload.FlowDiagram != nil {
		for i := range d.Payload.FlowDiagram.Stages {
			s := &d.Payload.FlowDiagram.Stages[i]
			s.ID = truncate(strings.TrimSpace(s.ID), 64, sanitized)
			s.Label = truncate(strings.TrimSpace(s.Label), 80, sanitized)
			s.XK, s.YK = clamp01(s.XK), clamp01(s.YK)
		}
		for i := range d.Payload.FlowDiagram.Links {
			l := &d.Payload.FlowDiagram.Links[i]
			l.From = truncate(strings.TrimSpace(l.From), 64, sanitized)
			l.To = truncate(strings.TrimSpace(l.To), 64, sanitized)
			l.Label = truncate(strings.TrimSpace(l.Label), 80, sanitized)
		}
	}
	if d.Payload.SetDiagram != nil {
		for i := range d.Payload.SetDiagram.Sets {
			s := &d.Payload.SetDiagram.Sets[i]
			s.ID = truncate(strings.TrimSpace(s.ID), 64, sanitized)
			s.Label = truncate(strings.TrimSpace(s.Label), 80, sanitized)
			s.XK, s.YK, s.RK = clamp01(s.XK), clamp01(s.YK), clamp01(s.RK)
		}
	}
	if d.Payload.Heatmap != nil {
		for i := range d.Payload.Heatmap.LabelsX {
			d.Payload.Heatmap.LabelsX[i] = truncate(strings.TrimSpace(d.Payload.Heatmap.LabelsX[i]), 50, sanitized)
		}
		for i := range d.Payload.Heatmap.LabelsY {
			d.Payload.Heatmap.LabelsY[i] = truncate(strings.TrimSpace(d.Payload.Heatmap.LabelsY[i]), 50, sanitized)
		}
	}
	if d.Payload.Field != nil {
		for i := range d.Payload.Field.Vectors {
			v := &d.Payload.Field.Vectors[i]
			v.ID = truncate(strings.TrimSpace(v.ID), 64, sanitized)
			v.XK, v.YK = clamp01(v.XK), clamp01(v.YK)
		}
	}
	if d.Payload.Chat != nil {
		for i := range d.Payload.Chat.Participants {
			p := &d.Payload.Chat.Participants[i]
			p.ID = truncate(strings.TrimSpace(p.ID), 64, sanitized)
			p.Name = truncate(strings.TrimSpace(p.Name), 80, sanitized)
			p.XK = clamp01(p.XK)
		}
		for i := range d.Payload.Chat.Messages {
			m := &d.Payload.Chat.Messages[i]
			m.SpeakerID = truncate(strings.TrimSpace(m.SpeakerID), 64, sanitized)
			m.Text = truncate(strings.TrimSpace(m.Text), 220, sanitized)
			if m.AtSec < 0 {
				m.AtSec = 0
				*sanitized = true
			}
		}
	}
	if d.Scene != nil {
		normalizeScene(d.Scene, sanitized)
	}
}

func normalizeScene(scene *SceneSpec, sanitized *bool) {
	if scene == nil {
		return
	}
	if scene.Layout != nil {
		scene.Layout.Mode = strings.TrimSpace(strings.ToLower(scene.Layout.Mode))
		switch scene.Layout.Mode {
		case "split_vertical", "split-horizontal", "split_horizontal":
			scene.Layout.Mode = "split"
			*sanitized = true
		}
		if scene.Layout.Params == nil {
			scene.Layout.Params = map[string]interface{}{}
		}
	}

	if len(scene.AnimationPlan) > 5 {
		scene.AnimationPlan = scene.AnimationPlan[:5]
		*sanitized = true
	}
	for i := range scene.AnimationPlan {
		scene.AnimationPlan[i].Target = truncate(strings.TrimSpace(scene.AnimationPlan[i].Target), 80, sanitized)
		scene.AnimationPlan[i].Effect = strings.TrimSpace(strings.ToLower(scene.AnimationPlan[i].Effect))
		if scene.AnimationPlan[i].DurationSec <= 0 {
			scene.AnimationPlan[i].DurationSec = 0.5
			*sanitized = true
		}
		if scene.AnimationPlan[i].DurationSec > 20 {
			scene.AnimationPlan[i].DurationSec = 20
			*sanitized = true
		}
	}
	sort.Slice(scene.AnimationPlan, func(i, j int) bool {
		return scene.AnimationPlan[i].Step < scene.AnimationPlan[j].Step
	})

	if scene.Timeline != nil {
		if scene.Timeline.MaxWaitSec <= 0 {
			scene.Timeline.MaxWaitSec = 10
			*sanitized = true
		}
		if scene.Timeline.MaxWaitSec < 1 {
			scene.Timeline.MaxWaitSec = 1
			*sanitized = true
		}
		if scene.Timeline.MaxWaitSec > 30 {
			scene.Timeline.MaxWaitSec = 30
			*sanitized = true
		}
	}

	if scene.Physics != nil {
		scene.Physics.Mode = strings.TrimSpace(strings.ToLower(scene.Physics.Mode))
		if scene.Physics.Params == nil {
			scene.Physics.Params = map[string]interface{}{}
		}
		if !scene.Physics.Enabled {
			if scene.Physics.Mode != "none" {
				scene.Physics.Mode = "none"
				*sanitized = true
			}
		} else if scene.Physics.Mode == "" {
			scene.Physics.Mode = "force_graph"
			*sanitized = true
		}
	}

	if scene.Layers == nil {
		scene.Layers = &SceneLayers{
			Background: true,
			Content:    true,
			Highlight:  true,
			Transition: true,
			Debug:      false,
		}
		*sanitized = true
	}
}

func normalizeDemoPlan(out *AnimatedSlideContentResponse, sanitized *bool) {
	if out.DemoPlan == nil {
		return
	}
	out.DemoPlan.Kind = strings.ToUpper(strings.TrimSpace(out.DemoPlan.Kind))
	out.DemoPlan.Reason = truncate(strings.TrimSpace(out.DemoPlan.Reason), slideDemoPlanReasonMaxChars, sanitized)
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
			if out.ImageURL != "" || len(out.Equations) > 0 || out.SketchPath != nil || out.PedagogicalDiagram != nil {
				out.ImageURL = ""
				out.Equations = nil
				out.SketchPath = nil
				out.PedagogicalDiagram = nil
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
			if out.PedagogicalDiagram != nil {
				out.PedagogicalDiagram = nil
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
			if out.PedagogicalDiagram != nil {
				out.PedagogicalDiagram = nil
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
			if out.PedagogicalDiagram != nil {
				out.PedagogicalDiagram = nil
				*sanitized = true
			}
			return
		case "PEDAGOGICAL_DIAGRAM":
			if out.ImageURL != "" {
				out.ImageURL = ""
				*sanitized = true
			}
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
	hasPedagogical := out.PedagogicalDiagram != nil

	if hasPedagogical {
		if hasImage {
			out.ImageURL = ""
			*sanitized = true
		}
		if hasSketch {
			out.SketchPath = nil
			*sanitized = true
		}
		if hasEquations {
			out.Equations = nil
			*sanitized = true
		}
		return
	}
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

func asInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	default:
		return 0
	}
}

func asBool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func asAnyMap(v interface{}) map[string]interface{} {
	m, ok := v.(map[string]interface{})
	if !ok || m == nil {
		return nil
	}
	return m
}

func getBoolOrDefault(m map[string]interface{}, key string, fallback bool) bool {
	v, exists := m[key]
	if !exists {
		return fallback
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return fallback
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

func buildAnimatedTreeSlideUserPrompt(req AnimatedLessonRequest) string {
	return fmt.Sprintf(`Create a scene-first lesson skeleton tree.
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
- Presentation mode: %s
- Scene policy: requireScene=%t requireTimeline=%t requireAnimationPlan=%t

Return JSON:
{
  "lessonId": string,
  "title": string,
  "treeSlides": [
    {
      "id": string,
      "title": string,
      "type": string,
      "physics": { "enabled": boolean, "mode": string, "params": object },
      "timeline": { "waitForAnimationPlan": boolean, "requireSignalDone": boolean, "maxWaitSec": number },
      "children": []
    }
  ]
}

Hard constraints:
- Return skeleton only; no content/layout/animationPlan in tree nodes.
- Children must be valid node objects with the same keys.
- Use strict JSON only.`,
		req.Topic,
		req.CourseTitle,
		req.LessonObjective,
		req.Learner.EducationLevel,
		req.Learner.Interests,
		req.Learner.CareerGoals,
		req.Animation.Mode,
		req.Animation.RequireLevelOrder,
		req.Animation.IncludePrerequisiteEdges,
		req.PresentationMode,
		req.ScenePolicy.RequireScene,
		req.ScenePolicy.RequireTimeline,
		req.ScenePolicy.RequireAnimationPlan,
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

func buildAnimatedTreeSlideRepairPrompt(req AnimatedLessonRequest, rawOutput string, validationErrors []string, parseErr error) string {
	var sb strings.Builder
	sb.WriteString("The previous output failed validation. Regenerate and return ONLY corrected JSON.\n")
	sb.WriteString("Original task:\n")
	sb.WriteString(buildAnimatedTreeSlideUserPrompt(req))
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
- type: %s
- summary: %s
- key points seed: %v
- examples seed: %v

Previous topic (for continuity): %s
Next topic (for continuity): %s

Hard constraints:
- Return strict JSON only.
- If presentationMode is scene, include a populated scene object.
- Keep output frontend-interpretable (no custom node DSL, no canvas scripting).
- Text and demo are separate child regions; do not describe overlay behavior.
- Keep text concise for fixed slide regions:
  - explanation max 260 chars
  - keyPoints: 2-3 items, each <= 90 chars
  - examples: 0-1 item, each <= 110 chars
  - bridgeFromPrevious and bridgeToNext: each <= 90 chars
- Choose one primary demo modality for this topic:
  1) imageUrl
  2) sketchPath
  3) equations
  4) pedagogicalDiagram
- If using sketchPath, keep geometry simple and non-overlapping.
- Do not include keys outside the schema below.
- Optional scene metadata is allowed only as declarative JSON under scene.
- scene.animationPlan must contain 1-5 steps.
- Use effects only from: fade-in, typewriter, draw, progressive_draw, stagger_in, highlight, morph, zoom.
- Do not emit scripts, event handlers, or draw commands.
- demoPlan.kind must be one of:
  - NONE
  - SKETCH_PATH
  - EQUATION_REVEAL
  - IMAGE_DRAW
  - PHYSICS_GRAPH
  - PEDAGOGICAL_DIAGRAM
- demoPlan.kind must agree with emitted modality fields:
  - SKETCH_PATH -> provide sketchPath
  - EQUATION_REVEAL -> provide equations
  - IMAGE_DRAW -> provide imageUrl
  - PHYSICS_GRAPH -> provide sketchPath
  - PEDAGOGICAL_DIAGRAM -> provide pedagogicalDiagram
  - NONE -> provide no modality field
- When using pedagogicalDiagram, infer the best single variant for the concept domain:
  - Algorithms: graph, table, tree, stack_queue
  - History: timeline, graph
  - Math: coordinate_plane, chart, table
  - Systems: flow_diagram, graph
  - Probability/Logic: set_diagram, tree, heatmap
- pedagogicalDiagram.payload.pattern is optional, but if present must be one of:
  PROGRESSIVE_REVEAL, BUILD_UP_CONSTRUCTION, STEP_EXECUTION, STATE_TRANSITION, DEPENDENCY_PROPAGATION, VALUE_ACCUMULATION, FOCUS_SHIFT, STRUCTURAL_TRANSFORMATION, MORPHING, SIMULATION_LOOP, PARALLEL_COMPARISON, COUNTEREXAMPLE_REVEAL, PATH_TRACING, COLLAPSE_EXPAND, ENERGY_FLOW_SIMULATION, GRAPH_TRAVERSAL, DP_FILL

Scene request context:
- presentationMode: %s
- scenePolicy.requireScene: %t
- scenePolicy.requireTimeline: %t
- scenePolicy.requireAnimationPlan: %t

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
  },
  "pedagogicalDiagram": {
    "id": string,
    "kind": "PEDAGOGICAL_DIAGRAM",
    "deps": string[],
    "payload": {
      "variant": string,
      "pattern": string,
      "title": string,
      "subtitle": string,
      "durationSec": number,
      "shapes": [{ "kind": string, "xk": number, "yk": number, "wk": number, "hk": number, "label": string }],
      "characters": [{ "name": string, "xk": number, "yk": number, "pose": string }],
      "table": { "headers": string[], "rows": string[][], "caption": string },
      "chart": { "type": string, "labels": string[], "series": [{ "name": string, "values": number[] }] },
      "timeline": { "events": [{ "id": string, "label": string, "atK": number }] },
      "tree": {
        "nodes": [{ "id": string, "label": string, "xk": number, "yk": number }],
        "edges": [{ "from": string, "to": string, "label": string, "directed": boolean }]
      },
      "stackQueue": { "mode": string, "items": string[] },
      "coordinatePlane": {
        "points": [{ "id": string, "label": string, "xk": number, "yk": number }],
        "curves": [{ "id": string, "points": [{ "xk": number, "yk": number }] }]
      },
      "flowDiagram": {
        "stages": [{ "id": string, "label": string, "xk": number, "yk": number }],
        "links": [{ "from": string, "to": string, "label": string }]
      },
      "setDiagram": { "sets": [{ "id": string, "label": string, "xk": number, "yk": number, "rk": number }] },
      "heatmap": { "labelsX": string[], "labelsY": string[], "values": number[][] },
      "field": { "vectors": [{ "id": string, "xk": number, "yk": number, "dxk": number, "dyk": number }] },
      "graph": {
        "nodes": [{ "id": string, "label": string, "xk": number, "yk": number }],
        "edges": [{ "from": string, "to": string, "label": string, "directed": boolean }]
      },
      "chat": {
        "participants": [{ "id": string, "name": string, "xk": number }],
        "messages": [{ "speakerId": string, "text": string, "atSec": number }]
      }
    },
    "scene": object
  },
  "scene": {
    "layout": { "mode": string, "params": object },
    "animationPlan": [{ "step": number, "target": string, "effect": string, "durationSec": number, "waitForSignal": boolean }],
    "timeline": { "waitForAnimationPlan": boolean, "requireSignalDone": boolean, "maxWaitSec": number },
    "physics": { "enabled": boolean, "mode": string, "params": object },
    "layers": { "background": boolean, "content": boolean, "highlight": boolean, "transition": boolean, "debug": boolean }
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
		req.CurrentTopic.Type,
		req.CurrentTopic.Summary,
		req.CurrentTopic.KeyPoints,
		req.CurrentTopic.Examples,
		previousTitle,
		nextTitle,
		req.PresentationMode,
		req.ScenePolicy.RequireScene,
		req.ScenePolicy.RequireTimeline,
		req.ScenePolicy.RequireAnimationPlan,
	)
}

func buildAnimatedSlideTextFallbackPrompt(req AnimatedSlideContentRequest) string {
	previousTitle := ""
	if req.PreviousTopic != nil {
		previousTitle = req.PreviousTopic.Title
	}
	nextTitle := ""
	if req.NextTopic != nil {
		nextTitle = req.NextTopic.Title
	}

	return fmt.Sprintf(`Fill this template exactly with plain text values.
Topic context:
- currentTopic.id: %s
- currentTopic.title: %s
- currentTopic.summary: %s
- previousTopic.title: %s
- nextTopic.title: %s
- presentationMode: %s

Output template:
TOPIC_ID: <string>
EXPLANATION: <single paragraph>
KEY_POINTS:
- <point 1>
- <point 2>
EXAMPLES:
- <example 1>
BRIDGE_FROM_PREVIOUS: <string>
BRIDGE_TO_NEXT: <string>
DEMO_KIND: <NONE|IMAGE_DRAW|EQUATION_REVEAL>
IMAGE_URL: <absolute-url-or-empty>
EQUATIONS:
- <equation-1>
SCENE_LAYOUT_MODE: <split|equation_centered|full_text|grid|stacked_progressive|overlay_layered|canvas_based_dynamic>
SCENE_TIMELINE_MAX_WAIT_SEC: <number>
SCENE_ANIMATION_PLAN:
- step=1|target=title|effect=fade-in|durationSec=0.30

Rules:
- Plain text only, no JSON.
- Do not add keys beyond this template.
- If DEMO_KIND is NONE, keep IMAGE_URL empty and EQUATIONS empty.
`, req.CurrentTopic.ID, req.CurrentTopic.Title, req.CurrentTopic.Summary, previousTitle, nextTitle, req.PresentationMode)
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
	q := getLLMCallQueue()
	modelName := inferModelName(payload)
	clientAlias := inferClientAlias(endpoint)
	contentKind := inferContentKind(endpoint)
	start := time.Now()
	log.Printf("[LLMDispatch] endpoint=%s content=%s client=%s model=%s", endpoint, contentKind, clientAlias, modelName)

	out, err := q.submit(ctx, func() (string, error) {
		return client.Call(endpoint, payload)
	})
	latencyMs := time.Since(start).Milliseconds()
	if err != nil {
		if strings.Contains(err.Error(), "LLM_QUEUE_OVERLOADED") {
			log.Printf("[LLMResult] endpoint=%s content=%s client=%s model=%s latency_ms=%d status=queue_overloaded",
				endpoint, contentKind, clientAlias, modelName, latencyMs)
		}
		log.Printf("[LLMResult] endpoint=%s content=%s client=%s model=%s latency_ms=%d status=error err=%v",
			endpoint, contentKind, clientAlias, modelName, latencyMs, err)
		return "", err
	}
	log.Printf("[LLMResult] endpoint=%s content=%s client=%s model=%s latency_ms=%d status=ok",
		endpoint, contentKind, clientAlias, modelName, latencyMs)
	return out, nil
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

func writeAnimatedRetryError(w http.ResponseWriter, status int, code string, message string, retryAfterMs int, nodeID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(animatedRetryError{
		Error:        message,
		Code:         code,
		RetryAfterMs: retryAfterMs,
		NodeID:       strings.TrimSpace(nodeID),
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

func hasJSONContentType(r *http.Request) bool {
	ct := strings.TrimSpace(r.Header.Get("Content-Type"))
	if ct == "" {
		// Accept missing content-type for compatibility as long as body parses as JSON.
		return true
	}
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return false
	}
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func beginQueuedRequest(cacheKey string) (bool, func()) {
	hash := cache.HashPayload(cacheKey)
	waitCh, leader := animatedQueue.begin(hash)
	if !leader {
		<-waitCh
		return false, nil
	}
	markQueueStatus(hash, "inflight")
	return true, func() {
		markQueueStatus(hash, "done")
		animatedQueue.end(hash)
	}
}

func markQueueStatus(hash string, status string) {
	c := cache.Get()
	if c == nil || c.Upstash == nil {
		return
	}
	key := "queue:inflight:" + hash
	_ = c.Upstash.StoreWithTTL(key, []byte(fmt.Sprintf(`{"status":"%s"}`, status)), 30)
}

func inferModelName(payload map[string]interface{}) string {
	if payload == nil {
		return "unknown"
	}
	if model, ok := payload["model"].(string); ok && strings.TrimSpace(model) != "" {
		return model
	}
	return "unknown"
}

func inferClientAlias(endpoint string) string {
	if cfg, ok := configs.GetDeepseekConfig(endpoint); ok && strings.TrimSpace(cfg.ClientName) != "" {
		return cfg.ClientName
	}
	return "unknown"
}

func inferContentKind(endpoint string) string {
	switch endpoint {
	case animatedKnowledgeGraphEndpoint, animatedTreeSlideEndpoint:
		return "skeleton_generation"
	case animatedSlideContentEndpoint, animatedSlideSceneEndpoint, animatedNodeContentEndpoint:
		return "node_content_generation"
	case "/generate-course-graph":
		return "course_graph_generation"
	case "/generate-lesson-index-plan":
		return "lesson_index_generation"
	case "/generate-exam":
		return "exam_generation"
	case "/submit-session":
		return "grading"
	case "/reflection/generate-task-reflection":
		return "task_reflection_generation"
	default:
		return "generic_generation"
	}
}

func shouldForceModelOverride(clientName string) bool {
	// "general"/"chat" are virtual routed aliases and may map to different providers.
	// Let provider clients pick their own configured model for these routes.
	switch strings.TrimSpace(strings.ToLower(clientName)) {
	case "general", "chat":
		return false
	default:
		return true
	}
}
