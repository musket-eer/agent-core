package llm

import (
	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// ReflectionGeneratePromptsHandler - POST /reflection/generate-prompts
func ReflectionGeneratePromptsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ReflectionGeneratePromptsHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ReflectionGeneratePromptsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ReflectionGeneratePromptsHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "reflectionPromptsPrompt", "/reflection/generate-prompts")
	if response != nil {
		var resp ReflectionGeneratePromptsResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ReflectionGeneratePromptsResponse{Prompts: []ReflectionPrompt{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// GenerateTaskReflectionHandler - POST /reflection/generate-task-reflection
func GenerateTaskReflectionHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GenerateTaskReflectionHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req TaskReflectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GenerateTaskReflectionHandler] Failed to decode request: %v", err)
		writeTaskReflectionError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
		return
	}

	if strings.TrimSpace(req.UserID) == "" || strings.TrimSpace(req.TaskID) == "" {
		writeTaskReflectionError(w, http.StatusBadRequest, "INVALID_REQUEST", "userId and taskId are required")
		return
	}

	resp, err := generateTaskReflectionCanonical(context.Background(), req)
	if err != nil {
		if strings.Contains(err.Error(), "LLM_INVALID_OUTPUT") {
			writeTaskReflectionError(w, http.StatusBadGateway, "LLM_INVALID_OUTPUT", "invalid reflection output")
			return
		}
		writeTaskReflectionError(w, http.StatusBadGateway, "LLM_FAILURE", "failed to generate task reflection")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func generateTaskReflectionCanonical(ctx context.Context, req TaskReflectionRequest) (TaskReflectionResponse, error) {
	cfg, ok := configs.GetDeepseekConfig("/reflection/generate-task-reflection")
	if !ok {
		cfg, _ = configs.GetDeepseekConfig("/chat")
	}

	clientName := cfg.ClientName
	if clientName == "" {
		clientName = "general"
	}
	client, err := GetLLMClient(clientName)
	if err != nil {
		return TaskReflectionResponse{}, err
	}

	systemPrompt, err := prompt.GetPrompt("taskReflectionPrompt")
	if err != nil {
		return TaskReflectionResponse{}, err
	}

	reqJSON, _ := json.Marshal(req)
	basePrompt := systemPrompt + "\n" + string(reqJSON)

	raw, err := callLLMWithTimeout(ctx, client, "/reflection/generate-task-reflection", map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  1400,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{"prompt": basePrompt},
	})
	if err != nil {
		return TaskReflectionResponse{}, err
	}

	resp, violations, parseErr := normalizeTaskReflectionOutput(raw)
	if parseErr == nil && len(violations) == 0 {
		return resp, nil
	}

	repairPrompt := buildTaskReflectionRepairPrompt(basePrompt, raw, violations, parseErr)
	repairedRaw, repairErr := callLLMWithTimeout(ctx, client, "/reflection/generate-task-reflection", map[string]interface{}{
		"model":       cfg.ModelId,
		"temperature": 0.2,
		"max_tokens":  1400,
		"response_format": map[string]interface{}{
			"type": "json_object",
		},
		"input": map[string]string{"prompt": repairPrompt},
	})
	if repairErr != nil {
		return TaskReflectionResponse{}, repairErr
	}

	finalResp, finalViolations, finalParseErr := normalizeTaskReflectionOutput(repairedRaw)
	if finalParseErr != nil || len(finalViolations) > 0 {
		log.Printf("[GenerateTaskReflectionHandler] repair failed parse_error=%v violations=%v", finalParseErr, finalViolations)
		return TaskReflectionResponse{}, fmt.Errorf("LLM_INVALID_OUTPUT")
	}
	return finalResp, nil
}

func normalizeTaskReflectionOutput(raw string) (TaskReflectionResponse, []string, error) {
	cleaned := utils.CleanJSONResponse(raw)
	var payload map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader([]byte(cleaned)))
	if err := decoder.Decode(&payload); err != nil {
		return TaskReflectionResponse{}, nil, err
	}

	resp := TaskReflectionResponse{
		Insights:         toStringArray(payload["insights"]),
		Patterns:         toStringArray(payload["patterns"]),
		Recommendations:  toStringArray(payload["recommendations"]),
		KeyLearnings:     toStringArray(payload["keyLearnings"]),
		ChallengesFaced:  toStringArray(payload["challengesFaced"]),
		SolutionsApplied: toStringArray(payload["solutionsApplied"]),
		SkillsPracticed:  toStringArray(payload["skillsPracticed"]),
		EmotionalTone:    normalizeTone(toString(payload["emotionalTone"])),
		DepthScore:       toFloat(payload["depthScore"]),
	}

	markdown := toString(payload["reflectionMarkdown"])
	legacyText := toString(payload["reflection"])
	if legacyText == "" {
		legacyText = toString(payload["aiReflection"])
	}
	if legacyText == "" {
		legacyText = toString(payload["response"])
	}

	if markdown == "" {
		markdown = legacyText
	}

	if looksLikeJSONObject(markdown) {
		var nested map[string]interface{}
		if err := json.Unmarshal([]byte(markdown), &nested); err == nil {
			if len(resp.Insights) == 0 {
				resp.Insights = toStringArray(nested["insights"])
			}
			if len(resp.Patterns) == 0 {
				resp.Patterns = toStringArray(nested["patterns"])
			}
			if len(resp.Recommendations) == 0 {
				resp.Recommendations = toStringArray(nested["recommendations"])
			}
			if len(resp.KeyLearnings) == 0 {
				resp.KeyLearnings = toStringArray(nested["keyLearnings"])
			}
			if len(resp.ChallengesFaced) == 0 {
				resp.ChallengesFaced = toStringArray(nested["challengesFaced"])
			}
			if len(resp.SolutionsApplied) == 0 {
				resp.SolutionsApplied = toStringArray(nested["solutionsApplied"])
			}
			if len(resp.SkillsPracticed) == 0 {
				resp.SkillsPracticed = toStringArray(nested["skillsPracticed"])
			}
			markdown = toString(nested["reflection"])
		}
	}

	resp.ReflectionMarkdown = buildTaskReflectionMarkdown(markdown, resp)
	resp.Reflection = resp.ReflectionMarkdown
	resp.AIReflection = resp.ReflectionMarkdown
	resp.Response = resp.ReflectionMarkdown

	violations := validateTaskReflectionCanonical(resp)
	return resp, violations, nil
}

func validateTaskReflectionCanonical(resp TaskReflectionResponse) []string {
	var violations []string
	if strings.TrimSpace(resp.ReflectionMarkdown) == "" {
		violations = append(violations, "reflectionMarkdown missing")
	}
	if looksLikeJSONObject(resp.ReflectionMarkdown) {
		violations = append(violations, "reflectionMarkdown contains raw JSON object")
	}
	if len(resp.ReflectionMarkdown) > 1800 {
		violations = append(violations, "reflectionMarkdown exceeds 1800 chars")
	}
	if !isStringSliceValid(resp.Insights) {
		violations = append(violations, "insights must be string[]")
	}
	if !isStringSliceValid(resp.Patterns) {
		violations = append(violations, "patterns must be string[]")
	}
	if !isStringSliceValid(resp.Recommendations) {
		violations = append(violations, "recommendations must be string[]")
	}
	if !isStringSliceValid(resp.KeyLearnings) {
		violations = append(violations, "keyLearnings must be string[]")
	}
	if !isStringSliceValid(resp.ChallengesFaced) {
		violations = append(violations, "challengesFaced must be string[]")
	}
	if !isStringSliceValid(resp.SolutionsApplied) {
		violations = append(violations, "solutionsApplied must be string[]")
	}
	if !isStringSliceValid(resp.SkillsPracticed) {
		violations = append(violations, "skillsPracticed must be string[]")
	}
	return violations
}

func buildTaskReflectionMarkdown(reflection string, resp TaskReflectionResponse) string {
	reflection = strings.TrimSpace(reflection)
	if reflection == "" {
		reflection = "Task reflection captured and normalized."
	}

	var b strings.Builder
	b.WriteString("**Reflection:**\n")
	b.WriteString(reflection)
	b.WriteString("\n\n")

	writeMarkdownListSection(&b, "Key Insights", resp.Insights)
	writeMarkdownListSection(&b, "Patterns Observed", resp.Patterns)
	writeMarkdownListSection(&b, "Recommendations", resp.Recommendations)
	writeMarkdownListSection(&b, "Key Learnings", resp.KeyLearnings)

	md := strings.TrimSpace(b.String())
	if len(md) > 1800 {
		md = md[:1800]
	}
	return md
}

func writeMarkdownListSection(b *strings.Builder, title string, values []string) {
	if len(values) == 0 {
		return
	}
	b.WriteString("**")
	b.WriteString(title)
	b.WriteString(":**\n")
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		b.WriteString("- ")
		b.WriteString(v)
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func buildTaskReflectionRepairPrompt(basePrompt string, raw string, violations []string, parseErr error) string {
	var b strings.Builder
	b.WriteString("Return strict JSON with canonical task reflection fields only.\n")
	if parseErr != nil {
		b.WriteString("Violation: json_parse_error=" + parseErr.Error() + "\n")
	}
	for _, v := range violations {
		b.WriteString("Violation: " + v + "\n")
	}
	b.WriteString("Required fields: reflectionMarkdown, insights, patterns, recommendations, keyLearnings, challengesFaced, solutionsApplied, skillsPracticed, emotionalTone, depthScore.\n")
	b.WriteString("reflectionMarkdown must be Markdown (no JSON object text).\n")
	b.WriteString("\nOriginal prompt:\n")
	b.WriteString(basePrompt)
	b.WriteString("\nPrevious output:\n")
	b.WriteString(utils.CleanJSONResponse(raw))
	return b.String()
}

func writeTaskReflectionError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
		"code":  code,
	})
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func toStringArray(v interface{}) []string {
	raw, ok := v.([]interface{})
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(raw))
	for _, it := range raw {
		if s, ok := it.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

func normalizeTone(tone string) string {
	t := strings.ToLower(strings.TrimSpace(tone))
	switch t {
	case "positive", "neutral", "challenging", "mixed", "constructive":
		return t
	default:
		return "neutral"
	}
}

func looksLikeJSONObject(s string) bool {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return false
	}
	var obj map[string]interface{}
	return json.Unmarshal([]byte(s), &obj) == nil
}

func isStringSliceValid(items []string) bool {
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			return false
		}
	}
	return true
}

// TaskPromptsHandler - POST /reflection/task-prompts
func TaskPromptsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TaskPromptsHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req TaskPromptsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TaskPromptsHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var prompts []string
	switch req.TaskType {
	case "learning":
		prompts = []string{
			"What concepts did you grasp?",
			"What remains unclear?",
			"How will you apply this knowledge?",
		}
	case "project":
		prompts = []string{
			"What milestones were achieved?",
			"What blockers did you encounter?",
			"What's the next priority?",
		}
	default:
		prompts = []string{
			"How did it go?",
			"What did you learn?",
			"What would you do differently?",
		}
	}

	resp := TaskPromptsResponse{Prompts: prompts}
	finalBytes, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}

// AnalyzeTaskHandler - POST /reflection/analyze-task
func AnalyzeTaskHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[AnalyzeTaskHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req AnalyzeTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[AnalyzeTaskHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For now, return a simple stub analysis; can be upgraded to LLM-backed later.
	resp := AnalyzeTaskResponse{
		Insights:       []string{},
		Strengths:      []string{},
		AreasForGrowth: []string{},
		Suggestions:    []string{},
		DepthScore:     7,
	}
	finalBytes, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)
}

// ReflectionAnalyzeHandler - POST /reflection/analyze
func ReflectionAnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ReflectionAnalyzeHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ReflectionAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ReflectionAnalyzeHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "reflectionAnalyzePrompt", "/reflection/analyze")
	if response != nil {
		var resp ReflectionAnalyzeResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ReflectionAnalyzeResponse{
				Insights: string(response), Strengths: []string{}, AreasForGrowth: []string{},
				Suggestions: []string{}, DepthScore: 0,
			}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// ReflectionLearningInsightsHandler - POST /reflection/learning-insights
func ReflectionLearningInsightsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ReflectionLearningInsightsHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ReflectionLearningInsightsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ReflectionLearningInsightsHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "reflectionInsightsPrompt", "/reflection/learning-insights")
	if response != nil {
		var resp ReflectionLearningInsightsResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ReflectionLearningInsightsResponse{
				Patterns: []LearningPattern{}, GrowthAreas: []string{},
				Recommendations: []string{}, Milestones: []Milestone{}, MetacognitiveScore: 0,
			}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// ReflectionPeriodSummaryHandler - POST /reflection/period-summary
func ReflectionPeriodSummaryHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ReflectionPeriodSummaryHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ReflectionPeriodSummaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ReflectionPeriodSummaryHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "reflectionSummaryPrompt", "/reflection/period-summary")
	if response != nil {
		var resp ReflectionPeriodSummaryResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ReflectionPeriodSummaryResponse{Summary: string(response)}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// ReflectionChatHandler - POST /reflection/chat
func ReflectionChatHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ReflectionChatHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ReflectionChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ReflectionChatHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "reflectionChatPrompt", "/reflection/chat")
	if response != nil {
		var resp ReflectionChatResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ReflectionChatResponse{
				Response: string(response), FollowUpQuestions: []string{}, Resources: []Resource{},
			}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// ReflectionSuggestExercisesHandler - POST /reflection/suggest-exercises
func ReflectionSuggestExercisesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ReflectionSuggestExercisesHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ReflectionSuggestExercisesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ReflectionSuggestExercisesHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "reflectionExercisesPrompt", "/reflection/suggest-exercises")
	if response != nil {
		var resp ReflectionSuggestExercisesResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ReflectionSuggestExercisesResponse{Exercises: []ReflectionExercise{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}
