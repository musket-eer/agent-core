package llm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"
)

const generateProjectPromptEndpoint = "/api/prompts/generate"

var (
	errPromptAssetNotFound    = errors.New("prompt asset not found")
	allowedProjectPromptModes = map[string]struct{}{
		"implementation": {},
		"planning":       {},
		"assignment":     {},
	}
	allowedProjectPromptTypes = map[string]struct{}{
		"project":    {},
		"assignment": {},
		"planning":   {},
	}
	projectPromptAssetByKey = map[string]projectPromptAssetRef{
		"project:implementation": {
			Key:        "prompt_generator_project_implementation",
			PromptName: "project_prompt_generator_implementation",
			Version:    1,
		},
		"project:planning": {
			Key:        "prompt_generator_project_planning",
			PromptName: "project_prompt_generator_planning",
			Version:    1,
		},
		"assignment:assignment": {
			Key:        "prompt_generator_assignment",
			PromptName: "project_prompt_generator_assignment",
			Version:    1,
		},
		"planning:planning": {
			Key:        "prompt_generator_planning",
			PromptName: "project_prompt_generator_planning",
			Version:    1,
		},
		"graph:implementation": {
			Key:        "prompt_graph_generator_implementation",
			PromptName: "prompt_graph_generator_implementation",
			Version:    1,
		},
		"graph:planning": {
			Key:        "prompt_graph_generator_planning",
			PromptName: "prompt_graph_generator_planning",
			Version:    1,
		},
		"graph:assignment": {
			Key:        "prompt_graph_generator_assignment",
			PromptName: "prompt_graph_generator_assignment",
			Version:    1,
		},
	}
)

type GenerateProjectPromptRequest struct {
	UserID         string                      `json:"userId"`
	Type           string                      `json:"type"`
	GenerationMode string                      `json:"generationMode"`
	Context        ProjectPromptRequestContext `json:"context"`
	Input          ProjectPromptInput          `json:"input"`
	SourcePrompts  []ProjectPromptSourceNode   `json:"sourcePrompts"`
	Snapshots      ProjectPromptSnapshots      `json:"snapshots"`
}

type ProjectPromptRequestContext struct {
	CourseID           string `json:"courseId"`
	ProjectID          string `json:"projectId"`
	GraphSelectionMode string `json:"graphSelectionMode"`
	RootPromptID       string `json:"rootPromptId"`
}

type ProjectPromptInput struct {
	Title           string   `json:"title"`
	Goal            string   `json:"goal"`
	TargetOutcome   string   `json:"targetOutcome"`
	Constraints     []string `json:"constraints"`
	PreferredStack  string   `json:"preferredStack"`
	DesiredOutput   string   `json:"desiredOutput"`
	DifficultyScope string   `json:"difficultyScope"`
}

type ProjectPromptSnapshots struct {
	Course         ProjectPromptCourseSnapshot  `json:"course"`
	LearningEvents []ProjectPromptLearningEvent `json:"learningEvents"`
}

type ProjectPromptCourseSnapshot struct {
	ID                 string                    `json:"id"`
	Title              string                    `json:"title"`
	Description        string                    `json:"description"`
	DifficultyLevel    string                    `json:"difficultyLevel"`
	CurrentLessonIndex int                       `json:"currentLessonIndex"`
	Objectives         []string                  `json:"objectives"`
	PriorKnowledge     []string                  `json:"priorKnowledge"`
	Units              []ProjectPromptCourseUnit `json:"units"`
	Notes              []ProjectPromptCourseNote `json:"notes"`
}

type ProjectPromptCourseUnit struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Position     int      `json:"position"`
	Objective    string   `json:"objective"`
	Assessment   string   `json:"assessment"`
	Constraints  []string `json:"constraints"`
	FailureModes []string `json:"failureModes"`
}

type ProjectPromptCourseNote struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Summary      string   `json:"summary"`
	Constraints  []string `json:"constraints"`
	FailureModes []string `json:"failureModes"`
}

type ProjectPromptLearningEvent struct {
	EventFamily string                 `json:"event_family"`
	EventType   string                 `json:"event_type"`
	Success     bool                   `json:"success"`
	Payload     map[string]interface{} `json:"payload"`
}

type ProjectPromptSourceNode struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Type           string   `json:"type"`
	GenerationMode string   `json:"generationMode"`
	Prompt         string   `json:"prompt"`
	Role           string   `json:"role"`
	Position       int      `json:"position"`
	Tags           []string `json:"tags"`
}

type GenerateProjectPromptResponse struct {
	PromptID        string                        `json:"promptId"`
	GeneratedPrompt string                        `json:"generatedPrompt"`
	Metadata        GenerateProjectPromptMetadata `json:"metadata"`
}

type GenerateProjectPromptMetadata struct {
	Type                 string   `json:"type"`
	Mode                 string   `json:"mode"`
	AssetKey             string   `json:"assetKey"`
	AssetVersion         int      `json:"assetVersion"`
	ConceptCount         int      `json:"conceptCount"`
	EventCount           int      `json:"eventCount"`
	LLMModel             string   `json:"llmModel"`
	SourcePromptCount    int      `json:"sourcePromptCount,omitempty"`
	TraversalMode        string   `json:"traversalMode,omitempty"`
	DerivedFromPromptIDs []string `json:"derivedFromPromptIds,omitempty"`
}

type projectPromptAssetRef struct {
	Key        string
	PromptName string
	Version    int
}

type projectPromptAsset struct {
	Key     string `json:"assetKey"`
	Version int    `json:"assetVersion"`
	Body    string `json:"templateBody"`
}

type compiledProjectPromptContext struct {
	Request        compiledProjectPromptRequest        `json:"request"`
	Course         compiledProjectPromptCourse         `json:"course"`
	Project        compiledProjectPromptProject        `json:"project"`
	LearnerHistory compiledProjectPromptLearnerHistory `json:"learnerHistory"`
	SourcePrompts  []compiledProjectPromptSource       `json:"sourcePrompts,omitempty"`
	Instructions   compiledProjectPromptInstructions   `json:"instructions"`
}

type compiledProjectPromptRequest struct {
	UserID             string `json:"userId"`
	Type               string `json:"type"`
	GenerationMode     string `json:"generationMode"`
	GraphSelectionMode string `json:"graphSelectionMode,omitempty"`
	RootPromptID       string `json:"rootPromptId,omitempty"`
}

type compiledProjectPromptCourse struct {
	ID              string                               `json:"id,omitempty"`
	Title           string                               `json:"title"`
	Description     string                               `json:"description"`
	DifficultyLevel string                               `json:"difficultyLevel,omitempty"`
	ActiveUnit      string                               `json:"activeUnit,omitempty"`
	CompletedUnits  []compiledProjectPromptCourseSummary `json:"completedUnits"`
	Concepts        []string                             `json:"concepts"`
	Constraints     []string                             `json:"constraints"`
	FailureModes    []string                             `json:"failureModes"`
	RelevantNotes   []compiledProjectPromptNoteSummary   `json:"relevantNotes"`
}

type compiledProjectPromptCourseSummary struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type compiledProjectPromptNoteSummary struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type compiledProjectPromptProject struct {
	ID              string   `json:"id,omitempty"`
	Title           string   `json:"title"`
	Goal            string   `json:"goal"`
	TargetOutcome   string   `json:"targetOutcome"`
	Constraints     []string `json:"constraints"`
	PreferredStack  string   `json:"preferredStack,omitempty"`
	DesiredOutput   string   `json:"desiredOutput,omitempty"`
	DifficultyScope string   `json:"difficultyScope,omitempty"`
}

type compiledProjectPromptLearnerHistory struct {
	RecentEvents  []compiledProjectPromptEvent `json:"recentEvents"`
	Struggles     []string                     `json:"struggles"`
	Strengths     []string                     `json:"strengths"`
	PriorAttempts []string                     `json:"priorAttempts"`
}

type compiledProjectPromptEvent struct {
	Type    string `json:"type"`
	Summary string `json:"summary"`
	Success bool   `json:"success"`
}

type compiledProjectPromptSource struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Type           string   `json:"type"`
	GenerationMode string   `json:"generationMode"`
	Prompt         string   `json:"prompt"`
	Role           string   `json:"role"`
	Position       int      `json:"position"`
	Tags           []string `json:"tags,omitempty"`
}

type compiledProjectPromptInstructions struct {
	Type         string `json:"type"`
	Mode         string `json:"mode"`
	AssetKey     string `json:"assetKey"`
	AssetVersion int    `json:"assetVersion"`
	TemplateBody string `json:"templateBody"`
}

type compiledProjectPromptStats struct {
	ConceptCount         int
	EventCount           int
	SourcePromptCount    int
	TraversalMode        string
	DerivedFromPromptIDs []string
}

func GenerateProjectPromptHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GenerateProjectPromptHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateProjectPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GenerateProjectPromptHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateGenerateProjectPromptRequest(r.URL.Path, req); err != nil {
		utils.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	asset, err := resolvePromptAssetForRequest(req)
	if err != nil {
		if errors.Is(err, errPromptAssetNotFound) {
			utils.WriteJSONError(w, "Prompt asset not found", http.StatusNotFound)
			return
		}
		log.Printf("[GenerateProjectPromptHandler] Failed to resolve prompt asset: %v", err)
		utils.WriteJSONError(w, "Failed to load prompt asset", http.StatusInternalServerError)
		return
	}

	contextPacket, stats := compileProjectPromptContext(req, asset)
	generatedPrompt, modelName, err := generateProjectPromptText(r.Context(), contextPacket)
	if err != nil {
		log.Printf("[GenerateProjectPromptHandler] Generation failed: %v", err)
		utils.WriteJSONError(w, "Failed to generate project prompt", http.StatusInternalServerError)
		return
	}

	resp := GenerateProjectPromptResponse{
		PromptID:        newProjectPromptID(),
		GeneratedPrompt: generatedPrompt,
		Metadata: GenerateProjectPromptMetadata{
			Type:                 strings.TrimSpace(strings.ToLower(req.Type)),
			Mode:                 strings.TrimSpace(strings.ToLower(req.GenerationMode)),
			AssetKey:             asset.Key,
			AssetVersion:         asset.Version,
			ConceptCount:         stats.ConceptCount,
			EventCount:           stats.EventCount,
			LLMModel:             modelName,
			SourcePromptCount:    stats.SourcePromptCount,
			TraversalMode:        stats.TraversalMode,
			DerivedFromPromptIDs: stats.DerivedFromPromptIDs,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[GenerateProjectPromptHandler] Failed to encode response: %v", err)
	}
}

func validateGenerateProjectPromptRequest(routePath string, req GenerateProjectPromptRequest) error {
	if strings.TrimSpace(req.UserID) == "" {
		return fmt.Errorf("userId is required")
	}
	if strings.TrimSpace(req.Type) == "" {
		return fmt.Errorf("type is required")
	}
	reqType := strings.TrimSpace(strings.ToLower(req.Type))
	if _, ok := allowedProjectPromptTypes[reqType]; !ok {
		return fmt.Errorf("unsupported type")
	}
	if strings.TrimSpace(req.GenerationMode) == "" {
		return fmt.Errorf("generationMode is required")
	}
	if strings.TrimSpace(req.Context.CourseID) == "" {
		return fmt.Errorf("courseId is required")
	}
	mode := strings.TrimSpace(strings.ToLower(req.GenerationMode))
	if _, ok := allowedProjectPromptModes[mode]; !ok {
		return fmt.Errorf("unsupported generationMode")
	}
	if strings.TrimSpace(req.Input.Title) == "" && strings.TrimSpace(req.Input.Goal) == "" {
		return fmt.Errorf("Prompt input is too thin")
	}
	if req.SourcePrompts != nil {
		for i, promptNode := range req.SourcePrompts {
			if strings.TrimSpace(promptNode.ID) == "" {
				return fmt.Errorf("sourcePrompts[%d].id is required", i)
			}
			if strings.TrimSpace(promptNode.Prompt) == "" {
				return fmt.Errorf("sourcePrompts[%d].prompt is required", i)
			}
		}
	}
	if requiresGraphSourcePrompts(routePath) && len(req.SourcePrompts) == 0 {
		return fmt.Errorf("sourcePrompts is required for graph generation")
	}
	return nil
}

func requiresGraphSourcePrompts(routePath string) bool {
	return strings.TrimSpace(routePath) == "/api/prompts/generate-from-graph"
}

func resolveProjectPromptAsset(promptType string, mode string) (projectPromptAsset, error) {
	keyType := strings.TrimSpace(strings.ToLower(promptType))
	if keyType == "" {
		keyType = "project"
	}
	key := keyType + ":" + strings.TrimSpace(strings.ToLower(mode))
	ref, ok := projectPromptAssetByKey[key]
	if !ok {
		return projectPromptAsset{}, errPromptAssetNotFound
	}

	body, err := prompt.GetPrompt(ref.PromptName)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return projectPromptAsset{}, errPromptAssetNotFound
		}
		return projectPromptAsset{}, err
	}

	return projectPromptAsset{
		Key:     ref.Key,
		Version: ref.Version,
		Body:    strings.TrimSpace(body),
	}, nil
}

func resolvePromptAssetForRequest(req GenerateProjectPromptRequest) (projectPromptAsset, error) {
	assetType := strings.TrimSpace(strings.ToLower(req.Type))
	if len(req.SourcePrompts) > 0 {
		assetType = "graph"
	}
	return resolveProjectPromptAsset(assetType, req.GenerationMode)
}

func compileProjectPromptContext(req GenerateProjectPromptRequest, asset projectPromptAsset) (compiledProjectPromptContext, compiledProjectPromptStats) {
	course := compileProjectPromptCourseContext(req.Snapshots.Course, req.Snapshots.LearningEvents)
	history := compileProjectPromptHistory(req.Snapshots.LearningEvents, course.FailureModes)
	sourcePrompts := compileSourcePrompts(req.SourcePrompts)

	packet := compiledProjectPromptContext{
		Request: compiledProjectPromptRequest{
			UserID:             strings.TrimSpace(req.UserID),
			Type:               strings.TrimSpace(strings.ToLower(req.Type)),
			GenerationMode:     strings.TrimSpace(strings.ToLower(req.GenerationMode)),
			GraphSelectionMode: strings.TrimSpace(strings.ToLower(req.Context.GraphSelectionMode)),
			RootPromptID:       strings.TrimSpace(req.Context.RootPromptID),
		},
		Course: course,
		Project: compiledProjectPromptProject{
			ID:              strings.TrimSpace(req.Context.ProjectID),
			Title:           strings.TrimSpace(req.Input.Title),
			Goal:            strings.TrimSpace(req.Input.Goal),
			TargetOutcome:   strings.TrimSpace(req.Input.TargetOutcome),
			Constraints:     normalizeStrings(req.Input.Constraints, 10),
			PreferredStack:  strings.TrimSpace(req.Input.PreferredStack),
			DesiredOutput:   strings.TrimSpace(req.Input.DesiredOutput),
			DifficultyScope: strings.TrimSpace(req.Input.DifficultyScope),
		},
		LearnerHistory: history,
		SourcePrompts:  sourcePrompts,
		Instructions: compiledProjectPromptInstructions{
			Type:         strings.TrimSpace(strings.ToLower(req.Type)),
			Mode:         strings.TrimSpace(strings.ToLower(req.GenerationMode)),
			AssetKey:     asset.Key,
			AssetVersion: asset.Version,
			TemplateBody: asset.Body,
		},
	}

	stats := compiledProjectPromptStats{
		ConceptCount:         len(packet.Course.Concepts),
		EventCount:           len(packet.LearnerHistory.RecentEvents),
		SourcePromptCount:    len(packet.SourcePrompts),
		TraversalMode:        packet.Request.GraphSelectionMode,
		DerivedFromPromptIDs: sourcePromptIDs(packet.SourcePrompts),
	}

	return packet, stats
}

func compileProjectPromptCourseContext(course ProjectPromptCourseSnapshot, events []ProjectPromptLearningEvent) compiledProjectPromptCourse {
	activeUnitTitle := ""
	completedUnits := make([]compiledProjectPromptCourseSummary, 0, len(course.Units))
	concepts := newOrderedSet()
	constraints := newOrderedSet()
	failureModes := newOrderedSet()
	relevantNotes := make([]compiledProjectPromptNoteSummary, 0, minInt(len(course.Notes), 5))

	for _, item := range normalizeStrings(course.Objectives, 8) {
		concepts.add(item)
	}
	for _, item := range normalizeStrings(course.PriorKnowledge, 8) {
		concepts.add(item)
	}

	units := append([]ProjectPromptCourseUnit(nil), course.Units...)
	sort.SliceStable(units, func(i, j int) bool { return units[i].Position < units[j].Position })

	for _, unit := range units {
		title := strings.TrimSpace(unit.Title)
		summary := firstNonEmpty(strings.TrimSpace(unit.Objective), strings.TrimSpace(unit.Assessment))
		if title == "" && summary == "" {
			continue
		}
		if unit.Position == course.CurrentLessonIndex && activeUnitTitle == "" {
			activeUnitTitle = title
		}
		completedUnits = append(completedUnits, compiledProjectPromptCourseSummary{
			Title:   title,
			Summary: summary,
		})
		for _, item := range normalizeStrings(unit.Constraints, 4) {
			constraints.add(item)
		}
		for _, item := range normalizeStrings(unit.FailureModes, 4) {
			failureModes.add(item)
		}
		for _, item := range normalizeStrings([]string{title, unit.Objective, unit.Assessment}, 6) {
			concepts.add(item)
		}
	}
	if len(completedUnits) > 5 {
		completedUnits = completedUnits[:5]
	}

	for _, note := range course.Notes {
		title := strings.TrimSpace(note.Title)
		summary := strings.TrimSpace(note.Summary)
		if title == "" && summary == "" {
			continue
		}
		relevantNotes = append(relevantNotes, compiledProjectPromptNoteSummary{
			Title:   title,
			Summary: summary,
		})
		for _, item := range normalizeStrings(note.Constraints, 4) {
			constraints.add(item)
		}
		for _, item := range normalizeStrings(note.FailureModes, 4) {
			failureModes.add(item)
		}
		for _, item := range normalizeStrings([]string{title, summary}, 4) {
			concepts.add(item)
		}
		if len(relevantNotes) >= 5 {
			break
		}
	}

	for _, event := range selectHighSignalEvents(events, 6) {
		for _, item := range normalizeStrings(extractEventConcepts(event), 3) {
			concepts.add(item)
		}
	}

	return compiledProjectPromptCourse{
		ID:              strings.TrimSpace(course.ID),
		Title:           strings.TrimSpace(course.Title),
		Description:     strings.TrimSpace(course.Description),
		DifficultyLevel: strings.TrimSpace(course.DifficultyLevel),
		ActiveUnit:      activeUnitTitle,
		CompletedUnits:  completedUnits,
		Concepts:        concepts.values(12),
		Constraints:     constraints.values(10),
		FailureModes:    failureModes.values(10),
		RelevantNotes:   relevantNotes,
	}
}

func compileProjectPromptHistory(events []ProjectPromptLearningEvent, inheritedFailureModes []string) compiledProjectPromptLearnerHistory {
	recentEvents := make([]compiledProjectPromptEvent, 0, 6)
	mistakes := newOrderedSet()
	strengths := newOrderedSet()
	priorAttempts := newOrderedSet()

	for _, item := range normalizeStrings(inheritedFailureModes, 8) {
		mistakes.add(item)
	}

	for _, event := range selectHighSignalEvents(events, 6) {
		summary := summarizeLearningEvent(event)
		recentEvents = append(recentEvents, compiledProjectPromptEvent{
			Type:    strings.TrimSpace(event.EventType),
			Summary: summary,
			Success: event.Success,
		})

		if event.Success {
			for _, item := range normalizeStrings(extractEventConcepts(event), 3) {
				strengths.add(item)
			}
			continue
		}

		for _, item := range normalizeStrings(extractEventConcepts(event), 3) {
			mistakes.add(item)
			priorAttempts.add(item)
		}
	}

	return compiledProjectPromptLearnerHistory{
		RecentEvents:  recentEvents,
		Struggles:     mistakes.values(8),
		Strengths:     strengths.values(8),
		PriorAttempts: priorAttempts.values(6),
	}
}

func selectHighSignalEvents(events []ProjectPromptLearningEvent, limit int) []ProjectPromptLearningEvent {
	if len(events) == 0 {
		return nil
	}
	out := make([]ProjectPromptLearningEvent, 0, minInt(len(events), limit))
	for i := len(events) - 1; i >= 0 && len(out) < limit; i-- {
		event := events[i]
		if strings.TrimSpace(event.EventType) == "" {
			continue
		}
		if summary := summarizeLearningEvent(event); summary == "" {
			continue
		}
		out = append(out, event)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func compileSourcePrompts(sourcePrompts []ProjectPromptSourceNode) []compiledProjectPromptSource {
	if len(sourcePrompts) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(sourcePrompts))
	out := make([]compiledProjectPromptSource, 0, len(sourcePrompts))
	for idx, source := range sourcePrompts {
		id := strings.TrimSpace(source.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, compiledProjectPromptSource{
			ID:             id,
			Title:          strings.TrimSpace(source.Title),
			Type:           strings.TrimSpace(strings.ToLower(source.Type)),
			GenerationMode: strings.TrimSpace(strings.ToLower(source.GenerationMode)),
			Prompt:         strings.TrimSpace(source.Prompt),
			Role:           firstNonEmpty(strings.TrimSpace(source.Role), "source_input"),
			Position:       source.Position,
			Tags:           normalizeStrings(source.Tags, 8),
		})
		if out[len(out)-1].Position == 0 && source.Position == 0 {
			out[len(out)-1].Position = idx
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Position == out[j].Position {
			return i < j
		}
		return out[i].Position < out[j].Position
	})

	return out
}

func sourcePromptIDs(sourcePrompts []compiledProjectPromptSource) []string {
	out := make([]string, 0, len(sourcePrompts))
	for _, source := range sourcePrompts {
		if strings.TrimSpace(source.ID) != "" {
			out = append(out, source.ID)
		}
	}
	return out
}

func summarizeLearningEvent(event ProjectPromptLearningEvent) string {
	parts := make([]string, 0, 2)
	if topic := strings.TrimSpace(asStringValue(event.Payload["topic"])); topic != "" {
		parts = append(parts, topic)
	}
	if lessonTitle := strings.TrimSpace(asStringValue(event.Payload["lessonTitle"])); lessonTitle != "" {
		parts = append(parts, lessonTitle)
	}
	if len(parts) == 0 {
		if strings.TrimSpace(event.EventFamily) != "" {
			return strings.TrimSpace(event.EventFamily)
		}
		return strings.TrimSpace(event.EventType)
	}
	return strings.Join(parts, " - ")
}

func extractEventConcepts(event ProjectPromptLearningEvent) []string {
	items := []string{
		asStringValue(event.Payload["topic"]),
		asStringValue(event.Payload["lessonTitle"]),
		asStringValue(event.Payload["concept"]),
		asStringValue(event.Payload["mistake"]),
	}
	return normalizeStrings(items, 4)
}

func generateProjectPromptText(ctx context.Context, packet compiledProjectPromptContext) (string, string, error) {
	config, ok := configs.GetDeepseekConfig(generateProjectPromptEndpoint)
	if !ok {
		return "", "", fmt.Errorf("no config found for endpoint: %s", generateProjectPromptEndpoint)
	}

	clientName := config.ClientName
	if clientName == "" {
		clientName = "general"
	}
	client, err := GetLLMClient(clientName)
	if err != nil {
		return "", config.ModelId, err
	}

	messages, err := buildPromptGenerationMessages(packet)
	if err != nil {
		return "", config.ModelId, err
	}

	payload := map[string]interface{}{
		"model":    config.ModelId,
		"messages": messages,
	}

	rawResponse, err := getLLMCallQueue().submit(ctx, func() (string, error) {
		return client.Call(generateProjectPromptEndpoint, payload)
	})
	if err != nil {
		return "", config.ModelId, err
	}

	finalText := strings.TrimSpace(rawResponse)
	if finalText == "" {
		return "", config.ModelId, fmt.Errorf("empty prompt returned from LLM")
	}
	return finalText, config.ModelId, nil
}

func buildPromptGenerationMessages(packet compiledProjectPromptContext) ([]Message, error) {
	basePacket := struct {
		Request        compiledProjectPromptRequest        `json:"request"`
		Course         compiledProjectPromptCourse         `json:"course"`
		Project        compiledProjectPromptProject        `json:"project"`
		LearnerHistory compiledProjectPromptLearnerHistory `json:"learnerHistory"`
	}{
		Request:        packet.Request,
		Course:         packet.Course,
		Project:        packet.Project,
		LearnerHistory: packet.LearnerHistory,
	}

	contextJSON, err := json.MarshalIndent(basePacket, "", "  ")
	if err != nil {
		return nil, err
	}

	messages := []Message{
		{Role: "system", Content: packet.Instructions.TemplateBody},
	}

	if len(packet.SourcePrompts) > 0 {
		for idx, source := range packet.SourcePrompts {
			messages = append(messages, Message{
				Role: "user",
				Content: fmt.Sprintf(
					"Source prompt %d (role: %s)\nID: %s\nTitle: %s\nType: %s\nMode: %s\nTags: %s\n\n%s",
					idx+1,
					firstNonEmpty(source.Role, "source_input"),
					source.ID,
					firstNonEmpty(source.Title, "Untitled Prompt"),
					firstNonEmpty(source.Type, "prompt"),
					firstNonEmpty(source.GenerationMode, "unknown"),
					strings.Join(source.Tags, ", "),
					source.Prompt,
				),
			})
		}

		messages = append(messages, Message{
			Role:    "user",
			Content: "Current generation request and grounded context:\n\n" + string(contextJSON),
		})
		return messages, nil
	}

	messages = append(messages, Message{
		Role: "user",
		Content: "Generate one high-quality prompt from the structured context below.\n" +
			"The prompt should help another AI or agent execute the learner's work using the learner's actual course knowledge.\n" +
			"Keep it concrete, structured, and constraint-aware.\n\n" +
			string(contextJSON),
	})
	return messages, nil
}

func newProjectPromptID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "prompt_fallback"
	}
	return "prompt_" + hex.EncodeToString(buf)
}

func asStringValue(v interface{}) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func normalizeStrings(values []string, maxItems int) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, minInt(len(values), maxItems))
	for _, item := range values {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
		if len(out) >= maxItems {
			break
		}
	}
	return out
}

type orderedSet struct {
	seen  map[string]struct{}
	items []string
}

func newOrderedSet() *orderedSet {
	return &orderedSet{
		seen:  make(map[string]struct{}),
		items: []string{},
	}
}

func (s *orderedSet) add(v string) {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return
	}
	key := strings.ToLower(trimmed)
	if _, ok := s.seen[key]; ok {
		return
	}
	s.seen[key] = struct{}{}
	s.items = append(s.items, trimmed)
}

func (s *orderedSet) values(limit int) []string {
	if limit <= 0 || len(s.items) <= limit {
		return append([]string(nil), s.items...)
	}
	return append([]string(nil), s.items[:limit]...)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
