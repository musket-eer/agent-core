package llm

import (
	"agent-core/internal/llm/utils"
	"encoding/json"
	"log"
	"net/http"
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
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "taskReflectionPrompt", "/reflection/generate-task-reflection")
	if response != nil {
		var resp TaskReflectionResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = TaskReflectionResponse{
				Reflection:       string(response),
				Insights:         []string{},
				Patterns:         []string{},
				Recommendations:  []string{},
				DepthScore:       0,
				EmotionalTone:    "neutral",
				KeyLearnings:     []string{},
				ChallengesFaced:  []string{},
				SolutionsApplied: []string{},
				SkillsPracticed:  []string{},
			}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
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
