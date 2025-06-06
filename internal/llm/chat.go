// Here we implement a Bedrock endpoint for chat
package llm

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"
	"agent-core/internal/llm/configs"
	"agent-core/internal/cache"
)

type StudentProfile struct {
	Name           string   `json:"name"`
	EducationLevel string   `json:"educationLevel"`
	Interests      []string `json:"interests"`
	CareerGoals    string   `json:"careerGoals"`
}

type ChatAgentRequest struct {
	Title          string `json:"lessonTitle"`
	Objective      string `json:"objective"`
	Difficulty     string `json:"difficulty"`
	EducationLevel string `json:"education_level"`
	AgentRole      string `json:"agent_role"`
	ActivityId     string `json:"activityId"`
	StudentMessage string `json:"studentMessage"`
	StudentProfile StudentProfile `json:"studentProfile"`
	CourseUnit     CourseUnit `json:"courseUnit"`
	Notes          interface{} `json:"notes"`
	ChatHistory    []interface{} `json:"chatHistory"`
	LeaderRole     string `json:"leader"`
}

// Generic chat handler for a specific agent role and prompt
func chatHandlerWithRole(promptKey string, agentRole string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read raw request body
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		rawBodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			utils.WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Decode request into struct
		var req ChatAgentRequest
		err = json.Unmarshal(rawBodyBytes, &req)
		if err != nil {
			utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Overwrite agent role to ensure it's correct for this endpoint
		req.AgentRole = agentRole

		promptTemplate, err := prompt.GetPrompt(promptKey)
		if err != nil {
			utils.WriteJSONError(w, "Failed to load chat prompt", http.StatusInternalServerError)
			return
		}

		// Construct final prompt
		promptStr := utils.PromptWithVars(promptTemplate, req.Title, req.Objective, req.EducationLevel, req.Difficulty)

		userDetailsJson, err := json.Marshal(req)
		if err != nil {
			utils.WriteJSONError(w, "Failed to serialize user message", http.StatusInternalServerError)
			return
		}

		// Prepare cache key
		payloadStr := string(userDetailsJson)
		cacheInstance := cache.Get()
		if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
			w.Header().Set("Content-Type", "application/json")
			w.Write(cached)
			return
		}

		fullPrompt := promptStr + "\n" + string(userDetailsJson)

		payload := map[string]interface{}{
			"modelName": "deepseek-chat",
			"input": map[string]string{
				"prompt": fullPrompt,
			},
		}
		endpoint := "/chat"
		config, ok := configs.GetDeepseekConfig(endpoint)
		if !ok {
			utils.WriteJSONError(w, "no config found for endpoint: "+endpoint, http.StatusInternalServerError)
			return
		}

		clientName := config.ClientName
		if clientName == "" {
			clientName = "general" // fallback
		}

		client, err := GetLLMClient(clientName)
		if err != nil {
			utils.WriteJSONError(w, "failed to get client "+clientName+": "+err.Error(), http.StatusInternalServerError)
			return
		}

		chatResp, err := client.Call(endpoint, payload)
		if err != nil {
			utils.WriteJSONError(w, "Failed to generate chat response", http.StatusInternalServerError)
			return
		}

		cleaned := strings.TrimSpace(chatResp)
		if strings.HasPrefix(cleaned, "```") {
			cleaned = strings.TrimPrefix(cleaned, "```")
			cleaned = strings.TrimSuffix(cleaned, "```")
			cleaned = strings.TrimSpace(cleaned)
		}

		finalResponse := map[string]string{
			"response": cleaned,
		}

		respBytes, err := json.Marshal(finalResponse)
		if err != nil {
			utils.WriteJSONError(w, "Failed to serialize chat response", http.StatusInternalServerError)
			return
		}
		// Store response in cache
		_ = cache.Get().CacheByPayload(payloadStr, respBytes)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}
}

// Handlers for each chat endpoint
var ProfessorChatHandler = chatHandlerWithRole("professorChatPrompt", string(ProfessorAgent))
var LibrarianChatHandler = chatHandlerWithRole("librarianChatPrompt", string(LibrarianAgent))
var ExaminerChatHandler = chatHandlerWithRole("examinerChatPrompt", string(ExaminerAgent))
var ChatHandler = chatHandlerWithRole("professorChatPrompt", string(ProfessorAgent))
