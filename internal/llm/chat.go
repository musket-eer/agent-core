// Here we implement a Bedrock endpoint for chat
package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"agent-core/internal/cache"
	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"
)

// add datapoint to customize agent support chat
// StudentProfile is now UserProfile in types_common.go

type ChatAgentRequest struct {
	Title          string        `json:"lessonTitle"`
	Objective      string        `json:"objective"`
	Difficulty     string        `json:"difficulty"`
	EducationLevel string        `json:"education_level"`
	AgentRole      string        `json:"agent_role"`
	ActivityId     string        `json:"activityId"`
	StudentMessage string        `json:"studentMessage"`
	StudentProfile UserProfile   `json:"studentProfile"`
	CourseUnit     CourseUnit    `json:"courseUnit"`
	Notes          interface{}   `json:"notes"`
	ChatHistory    []interface{} `json:"chatHistory"`
	LeaderRole     string        `json:"leader"`
}

// Generic chat handler for a specific agent role and prompt
func chatHandlerWithRole(promptKey string, agentRole string, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[chatHandlerWithRole] Incoming request for role: %s, prompt: %s", agentRole, promptKey)
		if r.Method != http.MethodPost {
			log.Printf("[chatHandlerWithRole] Invalid method: %s", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read raw request body
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		rawBodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("[chatHandlerWithRole] Failed to read request body: %v", err)
			utils.WriteJSONError(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Decode request into struct
		var req ChatAgentRequest
		err = json.Unmarshal(rawBodyBytes, &req)
		if err != nil {
			log.Printf("[chatHandlerWithRole] Invalid request body: %v", err)
			utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		log.Printf("[chatHandlerWithRole] Parsed request for lesson: %s", req.Title)

		// Overwrite agent role to ensure it's correct for this endpoint
		req.AgentRole = agentRole

		promptTemplate, err := prompt.GetPrompt(promptKey)
		if err != nil {
			log.Printf("[chatHandlerWithRole] Failed to load prompt: %v", err)
			utils.WriteJSONError(w, "Failed to load chat prompt", http.StatusInternalServerError)
			return
		}

		// Construct final prompt
		promptStr := utils.PromptWithVars(promptTemplate, req.Title, req.Objective, req.EducationLevel, req.Difficulty)

		userDetailsJson, err := json.Marshal(req)
		if err != nil {
			log.Printf("[chatHandlerWithRole] Failed to serialize user message: %v", err)
			utils.WriteJSONError(w, "Failed to serialize user message", http.StatusInternalServerError)
			return
		}

		// Prepare cache key
		rand.Seed(time.Now().UnixNano())
		emotionalResponse := rand.Intn(10)
		payloadStr := string(userDetailsJson) + fmt.Sprintf("%d", emotionalResponse)
		cacheInstance := cache.Get()
		if cached, found, _ := cacheInstance.CheckByPayload(payloadStr); found {
			log.Printf("[chatHandlerWithRole] Cache hit for chat")
			w.Header().Set("Content-Type", "application/json")
			w.Write(cached)
			return
		}
		log.Printf("[chatHandlerWithRole] Cache miss, calling LLM")

		fullPrompt := promptStr + "\n" + string(userDetailsJson)

		payload := map[string]interface{}{
			"modelName": "deepseek-chat",
			"input": map[string]string{
				"prompt": fullPrompt,
			},
		}
		config, ok := configs.GetDeepseekConfig(endpoint)
		if !ok {
			log.Printf("[chatHandlerWithRole] No config found for endpoint: %s", endpoint)
			utils.WriteJSONError(w, "no config found for endpoint: "+endpoint, http.StatusInternalServerError)
			return
		}

		clientName := config.ClientName
		if clientName == "" {
			clientName = "general" // fallback
		}

		client, err := GetLLMClient(clientName)
		if err != nil {
			log.Printf("[chatHandlerWithRole] Failed to get client %s: %v", clientName, err)
			utils.WriteJSONError(w, "failed to get client "+clientName+": "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("[chatHandlerWithRole] Calling LLM client: %s", clientName)

		chatResp, err := client.Call(endpoint, payload)
		if err != nil {
			log.Printf("[chatHandlerWithRole] LLM call failed: %v", err)
			utils.WriteJSONError(w, "Failed to generate chat response", http.StatusInternalServerError)
			return
		}
		log.Printf("[chatHandlerWithRole] Received LLM response")

		finalResponse := map[string]string{
			"response": strings.TrimSpace(chatResp),
		}

		respBytes, err := json.Marshal(finalResponse)
		if err != nil {
			log.Printf("[chatHandlerWithRole] Failed to serialize chat response: %v", err)
			utils.WriteJSONError(w, "Failed to serialize chat response", http.StatusInternalServerError)
			return
		}
		// Store response in cache
		_ = cache.Get().CacheByPayload(payloadStr, respBytes)
		log.Printf("[chatHandlerWithRole] Successfully generated chat response, cached")
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}
}

// Handlers for each chat endpoint
var ProfessorChatHandler = chatHandlerWithRole("professorChatPrompt", string(ProfessorAgent), "/professor-chat")
var LibrarianChatHandler = chatHandlerWithRole("librarianChatPrompt", string(LibrarianAgent), "/librarian-chat")
var ExaminerChatHandler = chatHandlerWithRole("examinerChatPrompt", string(ExaminerAgent), "/examiner-chat")
var ChatHandler = chatHandlerWithRole("professorChatPrompt", string(ProfessorAgent), "/chat")
