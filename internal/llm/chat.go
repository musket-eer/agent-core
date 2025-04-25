// Here we implement a Bedrock endpoint for chat
package llm

import (
	"encoding/json"
	"log"
	"net/http"

	"agent-core/internal/prompt"
)

// Import agent_roles.go
// (Go automatically recognizes files in the same package, so types are available)



// ChatRequest extends CourseRequest to include the agent role
// (If CourseRequest is used elsewhere, this is a superset.)
type ChatRequest struct {
	Title          string `json:"title"`
	Objective      string `json:"objective"`
	Difficulty     string `json:"difficulty"`
	EducationLevel string `json:"education_level"`
	AgentRole      string `json:"agent_role"`
}

// Handler for /chat endpoint
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body
	var req ChatRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate agent role
	role := req.AgentRole
	var promptKey string
	switch role {
	case string(ProfessorAgent):
		promptKey = "professorChatPrompt"
	case string(LibrarianAgent):
		promptKey = "librarianChatPrompt"
	case string(ExaminerAgent):
		promptKey = "examinerChatPrompt"
	default:
		http.Error(w, "Invalid agent_role. Must be one of: professor, librarian, examiner", http.StatusBadRequest)
		return
	}

	// Fetch the prompt from RAM
	promptTemplate, err := prompt.GetPrompt(promptKey)
	if err != nil {
		log.Printf("Error getting chat prompt for role %s: %v", role, err)
		http.Error(w, "Failed to get chat prompt", http.StatusInternalServerError)
		return
	}

	// Format prompt with variables
	promptStr := promptWithVars(promptTemplate, req.Title, req.Objective, req.EducationLevel, req.Difficulty)

	// Create prompt input
	promptMap := map[string]interface{}{
		"input": map[string]string{
			"prompt": promptStr,
		},
	}

	// Call LLM client
	chatResp, err := llmClient.Call("/chat", promptMap)
	if err != nil {
		log.Printf("Error calling LLM client: %v", err)
		http.Error(w, "Failed to generate chat response", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(chatResp))
}

