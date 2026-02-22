package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"

	"github.com/google/uuid"
)

const (
	generateCourseGraphEndpoint     = "/generate-course-graph"
	generateLessonIndexPlanEndpoint = "/generate-lesson-index-plan"
)

var tokenRe = regexp.MustCompile(`[a-z0-9]+`)

type graphValidationStats struct {
	ObjectiveCoverageScore float64
	OffDomainDriftScore    float64
	RepairInvoked          bool
}

type courseGraphAPIError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func GenerateCourseGraphHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeCourseGraphError(w, 405, "INVALID_REQUEST", "method not allowed")
		return
	}

	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = uuid.NewString()
	}

	var req GenerateCourseGraphRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeCourseGraphError(w, 400, "INVALID_JSON", "invalid request body")
		return
	}
	if err := validateGenerateCourseGraphRequest(&req); err != nil {
		writeCourseGraphError(w, 400, "INVALID_REQUEST", err.Error())
		return
	}

	start := time.Now()
	resp, model, status, stats, err := generateCourseGraph(context.Background(), req)
	latency := time.Since(start).Milliseconds()
	log.Printf("[GenerateCourseGraph] request_id=%s topic=%q model=%s latency_ms=%d objective_coverage=%.2f drift=%.2f validation=%s repair=%t",
		requestID, req.CourseRequest.Topic, model, latency, stats.ObjectiveCoverageScore, stats.OffDomainDriftScore, status, stats.RepairInvoked)

	if err != nil {
		if strings.Contains(err.Error(), "LLM_INVALID_OUTPUT") {
			writeCourseGraphError(w, 502, "LLM_INVALID_OUTPUT", "failed to produce valid course graph")
			return
		}
		writeCourseGraphError(w, 502, "LLM_FAILURE", "failed to generate course graph")
		return
	}
	writeCourseGraphJSON(w, 200, resp)
}

func GenerateLessonIndexPlanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeCourseGraphError(w, 405, "INVALID_REQUEST", "method not allowed")
		return
	}

	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = uuid.NewString()
	}

	var req GenerateLessonIndexPlanRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeCourseGraphError(w, 400, "INVALID_JSON", "invalid request body")
		return
	}
	if err := validateGenerateLessonIndexPlanRequest(&req); err != nil {
		writeCourseGraphError(w, 400, "INVALID_REQUEST", err.Error())
		return
	}

	start := time.Now()
	resp, model, status, repair, err := generateLessonIndexPlan(context.Background(), req)
	latency := time.Since(start).Milliseconds()
	log.Printf("[GenerateLessonIndexPlan] request_id=%s topic=%q model=%s latency_ms=%d validation=%s repair=%t",
		requestID, req.Lesson.Title, model, latency, status, repair)

	if err != nil {
		if strings.Contains(err.Error(), "LLM_INVALID_OUTPUT") {
			writeCourseGraphError(w, 502, "LLM_INVALID_OUTPUT", "failed to produce valid lesson index plan")
			return
		}
		writeCourseGraphError(w, 502, "LLM_FAILURE", "failed to generate lesson index plan")
		return
	}
	writeCourseGraphJSON(w, 200, resp)
}

func generateCourseGraph(ctx context.Context, req GenerateCourseGraphRequest) (GenerateCourseGraphResponse, string, string, graphValidationStats, error) {
	cfg, ok := configs.GetDeepseekConfig(generateCourseGraphEndpoint)
	if !ok {
		cfg, _ = configs.GetDeepseekConfig("/generate-course")
	}
	client, err := getLLMClientByConfig(cfg)
	if err != nil {
		return GenerateCourseGraphResponse{}, cfg.ModelId, "failed", graphValidationStats{}, err
	}

	systemPrompt := loadPromptOrDefault("courseGraphPrompt", defaultCourseGraphPrompt)
	userPrompt := buildCourseGraphUserPrompt(req)
	seed := stableSeed(req)

	raw, err := callLLMWithTimeout(ctx, client, generateCourseGraphEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  2800,
		"seed":        seed,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{"prompt": systemPrompt + "\n\n" + userPrompt},
	})
	if err != nil {
		return GenerateCourseGraphResponse{}, cfg.ModelId, "failed", graphValidationStats{}, err
	}

	resp, errs, parseErr, stats := parseAndValidateCourseGraph(raw, req)
	if parseErr == nil && len(errs) == 0 {
		return resp, cfg.ModelId, "pass", stats, nil
	}

	stats.RepairInvoked = true
	repairPrompt := buildRepairPrompt(userPrompt, raw, errs, parseErr)
	repairedRaw, repairErr := callLLMWithTimeout(ctx, client, generateCourseGraphEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  2800,
		"seed":        seed + 1,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{"prompt": systemPrompt + "\n\n" + repairPrompt},
	})
	if repairErr != nil {
		return GenerateCourseGraphResponse{}, cfg.ModelId, "failed", stats, repairErr
	}

	finalResp, finalErrs, finalParseErr, finalStats := parseAndValidateCourseGraph(repairedRaw, req)
	finalStats.RepairInvoked = true
	if finalParseErr != nil || len(finalErrs) > 0 {
		log.Printf("[GenerateCourseGraph] repair failed parse_error=%v violations=%v", finalParseErr, finalErrs)
		return GenerateCourseGraphResponse{}, cfg.ModelId, "failed", finalStats, fmt.Errorf("LLM_INVALID_OUTPUT")
	}
	return finalResp, cfg.ModelId, "pass_after_repair", finalStats, nil
}

func generateLessonIndexPlan(ctx context.Context, req GenerateLessonIndexPlanRequest) (GenerateLessonIndexPlanResponse, string, string, bool, error) {
	cfg, ok := configs.GetDeepseekConfig(generateLessonIndexPlanEndpoint)
	if !ok {
		cfg, _ = configs.GetDeepseekConfig("/generate-lesson")
	}
	client, err := getLLMClientByConfig(cfg)
	if err != nil {
		return GenerateLessonIndexPlanResponse{}, cfg.ModelId, "failed", false, err
	}

	systemPrompt := loadPromptOrDefault("lessonIndexPlanPrompt", defaultLessonIndexPrompt)
	userPrompt := buildLessonIndexUserPrompt(req)
	seed := stableSeed(req)

	raw, err := callLLMWithTimeout(ctx, client, generateLessonIndexPlanEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  2400,
		"seed":        seed,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{"prompt": systemPrompt + "\n\n" + userPrompt},
	})
	if err != nil {
		return GenerateLessonIndexPlanResponse{}, cfg.ModelId, "failed", false, err
	}

	resp, errs, parseErr := parseAndValidateLessonIndexPlan(raw, req)
	if parseErr == nil && len(errs) == 0 {
		return resp, cfg.ModelId, "pass", false, nil
	}

	repairPrompt := buildRepairPrompt(userPrompt, raw, errs, parseErr)
	repairedRaw, repairErr := callLLMWithTimeout(ctx, client, generateLessonIndexPlanEndpoint, map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  2400,
		"seed":        seed + 1,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{"prompt": systemPrompt + "\n\n" + repairPrompt},
	})
	if repairErr != nil {
		return GenerateLessonIndexPlanResponse{}, cfg.ModelId, "failed", true, repairErr
	}

	finalResp, finalErrs, finalParseErr := parseAndValidateLessonIndexPlan(repairedRaw, req)
	if finalParseErr != nil || len(finalErrs) > 0 {
		log.Printf("[GenerateLessonIndexPlan] repair failed parse_error=%v violations=%v", finalParseErr, finalErrs)
		return GenerateLessonIndexPlanResponse{}, cfg.ModelId, "failed", true, fmt.Errorf("LLM_INVALID_OUTPUT")
	}
	return finalResp, cfg.ModelId, "pass_after_repair", true, nil
}

func parseAndValidateCourseGraph(raw string, req GenerateCourseGraphRequest) (GenerateCourseGraphResponse, []string, error, graphValidationStats) {
	var resp GenerateCourseGraphResponse
	cleaned := utils.CleanJSONResponse(raw)
	dec := json.NewDecoder(bytes.NewReader([]byte(cleaned)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&resp); err != nil {
		return GenerateCourseGraphResponse{}, nil, err, graphValidationStats{}
	}
	errs, stats := validateCourseGraphResponse(resp, req)
	return resp, errs, nil, stats
}

func parseAndValidateLessonIndexPlan(raw string, req GenerateLessonIndexPlanRequest) (GenerateLessonIndexPlanResponse, []string, error) {
	var resp GenerateLessonIndexPlanResponse
	cleaned := utils.CleanJSONResponse(raw)
	dec := json.NewDecoder(bytes.NewReader([]byte(cleaned)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&resp); err != nil {
		return GenerateLessonIndexPlanResponse{}, nil, err
	}
	errs := validateLessonIndexPlanResponse(resp, req)
	return resp, errs, nil
}

func validateGenerateCourseGraphRequest(req *GenerateCourseGraphRequest) error {
	req.CourseRequest.Title = strings.TrimSpace(req.CourseRequest.Title)
	req.CourseRequest.Topic = strings.TrimSpace(req.CourseRequest.Topic)
	req.CourseRequest.LearnerQuestion = strings.TrimSpace(req.CourseRequest.LearnerQuestion)
	if req.CourseRequest.Topic == "" && req.CourseRequest.Title == "" {
		return fmt.Errorf("courseRequest.topic or courseRequest.title is required")
	}
	if len(req.CourseRequest.Objectives) == 0 {
		return fmt.Errorf("courseRequest.objectives is required")
	}
	if req.CourseRequest.TargetLessonCount <= 0 {
		return fmt.Errorf("courseRequest.targetLessonCount must be > 0")
	}
	if len(req.CourseRequest.Topic) > 200 || len(req.CourseRequest.Title) > 200 {
		return fmt.Errorf("courseRequest.title/topic too long")
	}
	if len(req.CourseRequest.LearnerQuestion) > 400 {
		return fmt.Errorf("courseRequest.learnerQuestion too long")
	}
	if len(req.LearnerProfile.Interests) > 240 {
		return fmt.Errorf("learnerProfile.interests too long")
	}
	return nil
}

func validateGenerateLessonIndexPlanRequest(req *GenerateLessonIndexPlanRequest) error {
	req.Mode = strings.TrimSpace(req.Mode)
	req.Lesson.ID = strings.TrimSpace(req.Lesson.ID)
	req.Lesson.Title = strings.TrimSpace(req.Lesson.Title)
	req.Lesson.KGTopicID = strings.TrimSpace(req.Lesson.KGTopicID)
	if req.Lesson.Title == "" || req.Lesson.KGTopicID == "" {
		return fmt.Errorf("lesson.title and lesson.kgTopicId are required")
	}
	if len(req.Course.KnowledgeGraph) == 0 {
		return fmt.Errorf("course.knowledgeGraph is required")
	}
	return nil
}

func validateCourseGraphResponse(resp GenerateCourseGraphResponse, req GenerateCourseGraphRequest) ([]string, graphValidationStats) {
	var errs []string
	stats := graphValidationStats{}
	course := resp.Course
	if strings.TrimSpace(course.ID) == "" || strings.TrimSpace(course.Title) == "" {
		errs = append(errs, "course.id and course.title are required")
	}
	if len(course.KnowledgeGraph) < 2 {
		errs = append(errs, "knowledgeGraph must contain at least 2 nodes")
		return errs, stats
	}
	if len(course.Units) != len(course.KnowledgeGraph) {
		errs = append(errs, "units.length must equal knowledgeGraph.length")
	}

	nodeByID := map[string]CourseGraphNode{}
	indexByID := map[string]int{}
	for i, node := range course.KnowledgeGraph {
		id := strings.TrimSpace(node.ID)
		if id == "" {
			errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].id is required", i))
			continue
		}
		if _, exists := nodeByID[id]; exists {
			errs = append(errs, fmt.Sprintf("duplicate node id: %s", id))
		}
		nodeByID[id] = node
		indexByID[id] = i
		if strings.TrimSpace(node.Title) == "" {
			errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].title is required", i))
		}
		if node.Order != i {
			errs = append(errs, fmt.Sprintf("knowledgeGraph[%d].order must equal index", i))
		}
	}

	for i, node := range course.KnowledgeGraph {
		for _, p := range node.Prerequisites {
			p = strings.TrimSpace(p)
			if p == node.ID {
				errs = append(errs, fmt.Sprintf("knowledgeGraph[%d] has self prerequisite", i))
				continue
			}
			if _, ok := nodeByID[p]; !ok {
				errs = append(errs, fmt.Sprintf("knowledgeGraph[%d] prerequisite %s missing", i, p))
			}
			if idx, ok := indexByID[p]; ok && idx >= i {
				errs = append(errs, fmt.Sprintf("knowledgeGraph[%d] prerequisite %s must appear earlier", i, p))
			}
		}
	}

	levels := map[string]int{}
	state := map[string]int{}
	var dfs func(string) (int, error)
	dfs = func(id string) (int, error) {
		if state[id] == 1 {
			return 0, fmt.Errorf("cycle detected at %s", id)
		}
		if state[id] == 2 {
			return levels[id], nil
		}
		state[id] = 1
		level := 0
		for _, p := range nodeByID[id].Prerequisites {
			l, err := dfs(p)
			if err != nil {
				return 0, err
			}
			if l+1 > level {
				level = l + 1
			}
		}
		state[id] = 2
		levels[id] = level
		return level, nil
	}
	for _, n := range course.KnowledgeGraph {
		if _, err := dfs(n.ID); err != nil {
			errs = append(errs, err.Error())
			break
		}
	}
	prevLevel := -1
	for i, n := range course.KnowledgeGraph {
		level := levels[n.ID]
		if level < prevLevel {
			errs = append(errs, fmt.Sprintf("knowledgeGraph not level-ordered at index %d", i))
			break
		}
		prevLevel = level
	}

	unitsByKG := map[string]int{}
	for i, u := range course.Units {
		if strings.TrimSpace(u.KGTopicID) == "" {
			errs = append(errs, fmt.Sprintf("units[%d].kgTopicId is required", i))
			continue
		}
		unitsByKG[u.KGTopicID]++
		if _, ok := nodeByID[u.KGTopicID]; !ok {
			errs = append(errs, fmt.Sprintf("units[%d].kgTopicId references unknown node", i))
		}
	}
	for id := range nodeByID {
		if unitsByKG[id] != 1 {
			errs = append(errs, fmt.Sprintf("exactly one unit must map to kgTopicId=%s", id))
		}
	}

	stats.OffDomainDriftScore = calculateOffDomainDriftScore(req, course.KnowledgeGraph)
	if stats.OffDomainDriftScore > 0.20 {
		errs = append(errs, fmt.Sprintf("off-domain drift exceeds threshold: %.2f", stats.OffDomainDriftScore))
	}

	missingVerbs, coverage := objectiveCoverage(req.CourseRequest.Objectives, course.KnowledgeGraph)
	stats.ObjectiveCoverageScore = coverage
	if len(missingVerbs) > 0 {
		errs = append(errs, fmt.Sprintf("objective verbs missing from graph coverage: %v", missingVerbs))
	}

	maxInterestInfluence := req.GenerationPolicy.Personalization.MaxInterestInfluence
	if maxInterestInfluence <= 0 {
		maxInterestInfluence = 0.25
	}
	influence := interestTitleInfluence(req.LearnerProfile.Interests, course.KnowledgeGraph)
	if influence > maxInterestInfluence || influence > 0.20 {
		errs = append(errs, fmt.Sprintf("interest influence on core topic titles too high: %.2f", influence))
	}

	return errs, stats
}

func validateLessonIndexPlanResponse(resp GenerateLessonIndexPlanResponse, req GenerateLessonIndexPlanRequest) []string {
	var errs []string
	plan := resp.LessonIndexPlan
	if strings.TrimSpace(plan.LessonTitle) == "" || strings.TrimSpace(plan.LessonObjective) == "" {
		errs = append(errs, "lessonIndexPlan.lessonTitle and lessonObjective are required")
	}
	if len(plan.LessonIndex) < 3 || len(plan.LessonIndex) > 10 {
		errs = append(errs, "lessonIndex length must be between 3 and 10")
	}

	allowedLeaders := map[string]bool{"Professor": true, "Librarian": true, "Examiner": true}
	neighborhoodTokens := lessonNeighborhoodTokens(req)
	for i, item := range plan.LessonIndex {
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Title) == "" {
			errs = append(errs, fmt.Sprintf("lessonIndex[%d].id/title required", i))
		}
		if !allowedLeaders[item.Leader] {
			errs = append(errs, fmt.Sprintf("lessonIndex[%d].leader must be Professor|Librarian|Examiner", i))
		}
		if strings.TrimSpace(item.DeepDive.Summary) == "" || len(item.DeepDive.LearningGoals) == 0 || len(item.DeepDive.KeyQuestions) == 0 || len(item.DeepDive.Deliverables) == 0 {
			errs = append(errs, fmt.Sprintf("lessonIndex[%d].deepDive is incomplete", i))
		}
		itemTokens := tokenize(strings.ToLower(item.Title + " " + item.DeepDive.Summary))
		if !hasTokenOverlap(itemTokens, neighborhoodTokens) {
			errs = append(errs, fmt.Sprintf("lessonIndex[%d] appears outside lesson topic neighborhood", i))
		}
	}
	return errs
}

func lessonNeighborhoodTokens(req GenerateLessonIndexPlanRequest) map[string]struct{} {
	tokens := tokenize(strings.ToLower(req.Lesson.Title + " " + req.Lesson.Objective))
	byID := map[string]LessonIndexKGNode{}
	dependents := map[string][]string{}
	for _, n := range req.Course.KnowledgeGraph {
		byID[n.ID] = n
		for _, p := range n.Prerequisites {
			dependents[p] = append(dependents[p], n.ID)
		}
	}
	if center, ok := byID[req.Lesson.KGTopicID]; ok {
		for t := range tokenize(strings.ToLower(center.Title)) {
			tokens[t] = struct{}{}
		}
		for _, p := range center.Prerequisites {
			if n, ok := byID[p]; ok {
				for t := range tokenize(strings.ToLower(n.Title)) {
					tokens[t] = struct{}{}
				}
			}
		}
		for _, d := range dependents[center.ID] {
			if n, ok := byID[d]; ok {
				for t := range tokenize(strings.ToLower(n.Title)) {
					tokens[t] = struct{}{}
				}
			}
		}
	}
	return tokens
}

func calculateOffDomainDriftScore(req GenerateCourseGraphRequest, nodes []CourseGraphNode) float64 {
	domain := tokenize(strings.ToLower(req.CourseRequest.Topic + " " + req.CourseRequest.Title + " " + req.CourseRequest.LearnerQuestion))
	if len(nodes) == 0 {
		return 0
	}
	off := 0
	for _, n := range nodes {
		titleTokens := tokenize(strings.ToLower(n.Title))
		if !hasTokenOverlap(titleTokens, domain) {
			off++
		}
	}
	return float64(off) / float64(len(nodes))
}

func objectiveCoverage(objectives []string, nodes []CourseGraphNode) ([]string, float64) {
	corpus := tokenize(strings.ToLower(serializeGraph(nodes)))
	missing := make([]string, 0)
	if len(objectives) == 0 {
		return missing, 1
	}
	covered := 0
	for _, obj := range objectives {
		verb := firstToken(strings.ToLower(obj))
		if verb == "" {
			continue
		}
		if _, ok := corpus[verb]; ok {
			covered++
		} else {
			missing = append(missing, verb)
		}
	}
	return missing, float64(covered) / float64(len(objectives))
}

func interestTitleInfluence(interests string, nodes []CourseGraphNode) float64 {
	interestTokens := tokenize(strings.ToLower(interests))
	if len(nodes) == 0 || len(interestTokens) == 0 {
		return 0
	}
	hits := 0
	for _, n := range nodes {
		if hasTokenOverlap(tokenize(strings.ToLower(n.Title)), interestTokens) {
			hits++
		}
	}
	return float64(hits) / float64(len(nodes))
}

func tokenize(s string) map[string]struct{} {
	stop := map[string]struct{}{"the": {}, "and": {}, "for": {}, "with": {}, "into": {}, "from": {}, "that": {}, "this": {}, "your": {}, "what": {}, "want": {}, "understand": {}, "foundations": {}, "introduction": {}, "to": {}, "of": {}, "a": {}, "an": {}, "in": {}, "on": {}}
	out := map[string]struct{}{}
	for _, m := range tokenRe.FindAllString(strings.ToLower(s), -1) {
		if len(m) < 3 {
			continue
		}
		if _, ok := stop[m]; ok {
			continue
		}
		out[m] = struct{}{}
	}
	return out
}

func hasTokenOverlap(a, b map[string]struct{}) bool {
	for t := range a {
		if _, ok := b[t]; ok {
			return true
		}
	}
	return false
}

func serializeGraph(nodes []CourseGraphNode) string {
	var b strings.Builder
	for _, n := range nodes {
		b.WriteString(n.Title)
		b.WriteByte(' ')
		b.WriteString(n.Summary)
		b.WriteByte(' ')
		for _, k := range n.KeyPoints {
			b.WriteString(k)
			b.WriteByte(' ')
		}
	}
	return b.String()
}

func firstToken(s string) string {
	tokens := tokenRe.FindAllString(s, -1)
	if len(tokens) == 0 {
		return ""
	}
	return tokens[0]
}

func getLLMClientByConfig(cfg configs.DeepseekConfig) (LLMClient, error) {
	clientName := cfg.ClientName
	if clientName == "" {
		clientName = "general"
	}
	return GetLLMClient(clientName)
}

func loadPromptOrDefault(key string, fallback string) string {
	p, err := prompt.GetPrompt(key)
	if err != nil || strings.TrimSpace(p) == "" {
		return fallback
	}
	return p
}

func buildCourseGraphUserPrompt(req GenerateCourseGraphRequest) string {
	return fmt.Sprintf(`Build a domain-anchored course knowledge graph.
Title: %s
Topic: %s
Learner question: %s
Objectives: %v
Prior knowledge: %v
Difficulty: %s
Weekly hours: %d
Target lessons: %d
Reference books: %v
Learner profile:
- Education: %s
- Career goals: %s
- Interests: %s
Policy:
- mode: %s
- requireTopicAnchoring: %t
- requireObjectiveCoverage: %t
- requirePrerequisiteOrdering: %t
- interestsRole: %s
- maxInterestInfluence: %.2f
- doNotReplaceCoreTopic: %t

Return JSON:
{
  "course": {
    "id": string,
    "title": string,
    "description": string,
    "objectives": string[],
    "priorKnowledge": string[],
    "difficultyLevel": string,
    "weeklyHours": number,
    "lessons": number,
    "referenceBooks": string[],
    "units": [{
      "id": string,
      "title": string,
      "objective": string,
      "assessment": string,
      "position": number,
      "kgTopicId": string,
      "deepDivePrompt": string
    }],
    "knowledgeGraph": [{
      "id": string,
      "title": string,
      "summary": string,
      "keyPoints": string[],
      "examples": string[],
      "prerequisites": string[],
      "order": number
    }],
    "metadata": {"generationMode": string, "interestUsage": string}
  }
}`, req.CourseRequest.Title, req.CourseRequest.Topic, req.CourseRequest.LearnerQuestion, req.CourseRequest.Objectives, req.CourseRequest.PriorKnowledge, req.CourseRequest.DifficultyLevel, req.CourseRequest.WeeklyHours, req.CourseRequest.TargetLessonCount, req.CourseRequest.ReferenceBooks, req.LearnerProfile.EducationLevel, req.LearnerProfile.CareerGoals, req.LearnerProfile.Interests, req.GenerationPolicy.Mode, req.GenerationPolicy.RequireTopicAnchoring, req.GenerationPolicy.RequireObjectiveCoverage, req.GenerationPolicy.RequirePrerequisiteOrdering, req.GenerationPolicy.Personalization.InterestsRole, req.GenerationPolicy.Personalization.MaxInterestInfluence, req.GenerationPolicy.Personalization.DoNotReplaceCoreTopic)
}

func buildLessonIndexUserPrompt(req GenerateLessonIndexPlanRequest) string {
	return fmt.Sprintf(`Generate a deep-dive lesson index plan for one lesson.
Mode: %s
Course:
- id: %s
- title: %s
- description: %s
- difficulty: %s
- objectives: %v
- knowledgeGraph: %v
Lesson:
- id: %s
- title: %s
- objective: %s
- kgTopicId: %s
- deepDivePrompt: %s
Learner:
- education: %s
- career goals: %s
- interests: %s

Return JSON:
{
  "lessonIndexPlan": {
    "lessonTitle": string,
    "lessonObjective": string,
    "lessonIndex": [{
      "id": string,
      "title": string,
      "leader": string,
      "deepDive": {
        "summary": string,
        "learningGoals": string[],
        "keyQuestions": string[],
        "deliverables": string[],
        "references": string[]
      }
    }]
  }
}`, req.Mode, req.Course.ID, req.Course.Title, req.Course.Description, req.Course.DifficultyLevel, req.Course.Objectives, req.Course.KnowledgeGraph, req.Lesson.ID, req.Lesson.Title, req.Lesson.Objective, req.Lesson.KGTopicID, req.Lesson.DeepDivePrompt, req.Learner.EducationLevel, req.Learner.CareerGoals, req.Learner.Interests)
}

func buildRepairPrompt(originalPrompt string, raw string, violations []string, parseErr error) string {
	var b strings.Builder
	b.WriteString("Regenerate schema-correct JSON only.\n")
	if parseErr != nil {
		b.WriteString("Violation: json_parse_error=" + parseErr.Error() + "\n")
	}
	for _, v := range violations {
		b.WriteString("Violation: " + v + "\n")
	}
	b.WriteString("\nOriginal Task:\n")
	b.WriteString(originalPrompt)
	b.WriteString("\nPrevious Output:\n")
	b.WriteString(utils.CleanJSONResponse(raw))
	return b.String()
}

func stableSeed(v interface{}) int {
	b, _ := json.Marshal(v)
	h := fnv.New32a()
	_, _ = h.Write(b)
	return int(h.Sum32() & 0x7fffffff)
}

func writeCourseGraphError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(courseGraphAPIError{
		Error: message,
		Code:  code,
	})
}

func writeCourseGraphJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

const defaultCourseGraphPrompt = `You generate domain-anchored educational course knowledge graphs.
Return only strict JSON matching schema.
Rules:
1) Keep graph in-domain with requested topic/title.
2) Use learner interests only in examples/context hooks.
3) Ensure prerequisite ordering and level-order traversal.
4) Ensure every objective is covered by the graph.
5) No markdown or extra keys.`

const defaultLessonIndexPrompt = `You generate deep-dive lesson index plans.
Return only strict JSON matching schema.
Rules:
1) Provide 3-10 index items.
2) Each item must include deepDive with summary, learningGoals, keyQuestions, deliverables.
3) Keep content within lesson topic and prerequisite neighborhood.
4) Favor conceptual depth over shallow slide phrasing.
5) No markdown or extra keys.`
