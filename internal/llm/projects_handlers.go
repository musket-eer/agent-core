package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"agent-core/internal/llm/utils"
)

// ProjectsChatHandler - POST /projects/chat
func ProjectsChatHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ProjectsChatHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ProjectsChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProjectsChatHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "projectsChatPrompt", "/projects/chat")
	if response != nil {
		var resp ProjectsChatResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ProjectsChatResponse{Response: string(response), Suggestions: []string{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// ProjectsSuggestionsHandler - POST /projects/suggestions
func ProjectsSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ProjectsSuggestionsHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ProjectsSuggestionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProjectsSuggestionsHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "projectsSuggestionsPrompt", "/projects/suggestions")
	if response != nil {
		var resp ProjectsSuggestionsResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ProjectsSuggestionsResponse{Suggestions: []ProjectSuggestion{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// ProjectsHelpHandler - POST /projects/help
func ProjectsHelpHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ProjectsHelpHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ProjectsHelpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProjectsHelpHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "projectsHelpPrompt", "/projects/help")
	if response != nil {
		var resp ProjectsHelpResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ProjectsHelpResponse{
				Response: string(response), Resources: []Resource{}, NextSteps: []string{},
			}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// ProjectsBreakDownHandler - POST /projects/break-down
func ProjectsBreakDownHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ProjectsBreakDownHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ProjectsBreakDownRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProjectsBreakDownHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "projectsBreakDownPrompt", "/projects/break-down")
	if response != nil {
		var resp ProjectsBreakDownResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ProjectsBreakDownResponse{Milestones: []ProjectMilestone{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// ProjectsResourcesHandler - POST /projects/resources
func ProjectsResourcesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ProjectsResourcesHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ProjectsResourcesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ProjectsResourcesHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "projectsResourcesPrompt", "/projects/resources")
	if response != nil {
		var resp ProjectsResourcesResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ProjectsResourcesResponse{Resources: []ProjectResource{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}
