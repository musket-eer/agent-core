package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"agent-core/internal/llm/utils"
)

// TasksGenerateSuggestionsHandler - POST /tasks/generate-suggestions
func TasksGenerateSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TasksGenerateSuggestionsHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req TasksGenerateSuggestionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksGenerateSuggestionsHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "tasksSuggestionsPrompt", "/tasks/generate-suggestions")
	if response != nil {
		var resp TasksGenerateSuggestionsResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = TasksGenerateSuggestionsResponse{SuggestedTasks: []SuggestedTask{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// TasksPrioritizeHandler - POST /tasks/prioritize
func TasksPrioritizeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TasksPrioritizeHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req TasksPrioritizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksPrioritizeHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "tasksPrioritizePrompt", "/tasks/prioritize")
	if response != nil {
		var resp TasksPrioritizeResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = TasksPrioritizeResponse{
				PrioritizedTasks: []PrioritizedTask{}, Reasoning: string(response),
				TimeAllocation: TimeAllocation{Morning: []string{}, Afternoon: []string{}, Evening: []string{}},
				Suggestions: []string{},
			}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// TasksBreakDownHandler - POST /tasks/break-down
func TasksBreakDownHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TasksBreakDownHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req TasksBreakDownRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksBreakDownHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "tasksBreakDownPrompt", "/tasks/break-down")
	if response != nil {
		var resp TasksBreakDownResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = TasksBreakDownResponse{Subtasks: []Subtask{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// TasksProductivityInsightsHandler - POST /tasks/productivity-insights
func TasksProductivityInsightsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TasksProductivityInsightsHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req TasksProductivityInsightsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksProductivityInsightsHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "tasksProductivityPrompt", "/tasks/productivity-insights")
	if response != nil {
		var resp TasksProductivityInsightsResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = TasksProductivityInsightsResponse{
				CompletionRate: 0, Patterns: []ProductivityPattern{}, Strengths: []string{},
				Improvements: []string{}, Recommendations: []string{}, OptimalWorkTimes: []OptimalWorkTime{},
			}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// TasksGenerateScheduleHandler - POST /tasks/generate-schedule
func TasksGenerateScheduleHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TasksGenerateScheduleHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req TasksGenerateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksGenerateScheduleHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "tasksSchedulePrompt", "/tasks/generate-schedule")
	if response != nil {
		var resp TasksGenerateScheduleResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = TasksGenerateScheduleResponse{
				Schedule: []ScheduleItem{}, Reasoning: string(response), Alternatives: []ScheduleAlternative{},
			}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// TasksChatHandler - POST /tasks/chat
func TasksChatHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TasksChatHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req TasksChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksChatHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "tasksChatPrompt", "/tasks/chat")
	if response != nil {
		var resp TasksChatResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = TasksChatResponse{
				Response: string(response), SuggestedActions: []SuggestedAction{}, 
				TaskRecommendations: []TaskRecommendation{},
			}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// TasksMotivationHandler - POST /tasks/motivation
func TasksMotivationHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[TasksMotivationHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req TasksMotivationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[TasksMotivationHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "tasksMotivationPrompt", "/tasks/motivation")
	if response != nil {
		var resp TasksMotivationResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = TasksMotivationResponse{Message: string(response)}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}
