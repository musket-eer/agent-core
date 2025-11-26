package llm

import (
	"agent-core/internal/llm/utils"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// OrchestratorHandler - POST /ai/orchestrator
func OrchestratorHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[OrchestratorHandler] Incoming request: %s %s", r.Method, r.URL.Path)

	var req OrchestratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[OrchestratorHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Message == "" {
		log.Printf("[OrchestratorHandler] Missing required fields")
		utils.WriteJSONError(w, "userId and message are required", http.StatusBadRequest)
		return
	}

	log.Printf("[OrchestratorHandler] Processing message for user %s: %s", req.UserID, req.Message)

	// Call LLM to get orchestrator response
	response := handleLLMRequest(w, req, "orchestratorPrompt", "/ai/orchestrator")
	if response == nil {
		return
	}

	// Parse LLM response
	var llmResponse struct {
		Reply      string                `json:"reply"`
		Commands   []OrchestratorCommand `json:"commands"`
		SessionID  string                `json:"sessionId,omitempty"`
		UsedAgents []string              `json:"usedAgents,omitempty"`
	}

	if err := json.Unmarshal(response, &llmResponse); err != nil {
		log.Printf("[OrchestratorHandler] Failed to parse LLM response: %v", err)
		// Fallback: treat entire response as reply
		llmResponse.Reply = string(response)
		llmResponse.Commands = []OrchestratorCommand{}
	}

	// Validate and sanitize commands
	validatedCommands := validateCommands(llmResponse.Commands)

	// Build response
	orchestratorResponse := OrchestratorResponse{
		Reply:    llmResponse.Reply,
		Commands: validatedCommands,
		Meta: OrchestratorMeta{
			SessionID:  llmResponse.SessionID,
			UsedAgents: llmResponse.UsedAgents,
		},
	}

	// If no session ID from LLM, use the one from request or generate placeholder
	if orchestratorResponse.Meta.SessionID == "" {
		if req.SessionID != "" {
			orchestratorResponse.Meta.SessionID = req.SessionID
		} else {
			orchestratorResponse.Meta.SessionID = "orchestrator-session"
		}
	}

	finalBytes, _ := json.Marshal(orchestratorResponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)

	log.Printf("[OrchestratorHandler] Successfully processed request with %d commands", len(validatedCommands))
}

// validateCommands filters and validates commands against the known schema
func validateCommands(commands []OrchestratorCommand) []OrchestratorCommand {
	validCommands := []OrchestratorCommand{}

	for _, cmd := range commands {
		if isValidCommand(cmd) {
			validCommands = append(validCommands, cmd)
		} else {
			log.Printf("[OrchestratorHandler] Dropping invalid command type: %s", cmd.Type)
		}
	}

	return validCommands
}

// isValidCommand checks if a command type is in the allowed schema
func isValidCommand(cmd OrchestratorCommand) bool {
	validTypes := map[string]bool{
		CommandCreateTask:             true,
		CommandCompleteTask:           true,
		CommandUpdateTaskNotes:        true,
		CommandGenerateTaskReflection: true,
		CommandNavigate:               true,
		CommandFocusTask:              true,
	}

	cmdType := strings.ToLower(strings.TrimSpace(cmd.Type))
	return validTypes[cmdType]
}
