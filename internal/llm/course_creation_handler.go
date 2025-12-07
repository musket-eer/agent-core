package llm

import (
	"agent-core/internal/llm/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// CourseCreationChatHandler - POST /course-creation/chat
// Handles dialogue-based course creation with state machine
func CourseCreationChatHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CourseCreationChatHandler] Incoming request: %s %s", r.Method, r.URL.Path)

	var req CourseCreationChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CourseCreationChatHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" {
		log.Printf("[CourseCreationChatHandler] Missing required field: userId")
		utils.WriteJSONError(w, "userId is required", http.StatusBadRequest)
		return
	}

	if req.CurrentStep == "" {
		req.CurrentStep = "greeting"
	}

	// Validate step
	if !ValidCourseCreationSteps[req.CurrentStep] {
		log.Printf("[CourseCreationChatHandler] Invalid step: %s", req.CurrentStep)
		utils.WriteJSONError(w, "Invalid currentStep", http.StatusBadRequest)
		return
	}

	// Initialize course draft if nil
	if req.CourseDraft == nil {
		req.CourseDraft = &CourseDraft{
			Objectives:     []string{},
			ReferenceBooks: []string{},
			Units:          []DraftUnit{},
		}
	}

	// Truncate chat history to last 10 turns to manage token usage
	if len(req.ChatHistory) > 10 {
		req.ChatHistory = req.ChatHistory[len(req.ChatHistory)-10:]
	}

	log.Printf("[CourseCreationChatHandler] Step: %s, User: %s", req.CurrentStep, req.UserID)

	// Call LLM with course creation prompt
	response := handleLLMRequest(w, req, "courseCreationPrompt", "/course-creation/chat")
	if response == nil {
		return
	}

	// Parse LLM response
	var chatResp CourseCreationChatResponse
	if err := json.Unmarshal(response, &chatResp); err != nil {
		log.Printf("[CourseCreationChatHandler] Failed to parse LLM response: %v", err)
		// Return fallback response
		chatResp = CourseCreationChatResponse{
			Reply:        "I'm having trouble processing that. Could you rephrase?",
			UpdatedDraft: map[string]interface{}{},
			NextStep:     req.CurrentStep, // Stay on current step
			Suggestions:  []string{},
		}
	}

	// Validate and normalize response
	if chatResp.Reply == "" {
		chatResp.Reply = "Let's continue building your course."
	}

	if chatResp.UpdatedDraft == nil {
		chatResp.UpdatedDraft = map[string]interface{}{}
	}

	// Validate nextStep
	if !ValidCourseCreationSteps[chatResp.NextStep] {
		log.Printf("[CourseCreationChatHandler] Invalid nextStep from LLM: %s, keeping current", chatResp.NextStep)
		chatResp.NextStep = req.CurrentStep
	}

	if chatResp.Suggestions == nil {
		chatResp.Suggestions = []string{}
	}

	// CRITICAL: Validate required fields before allowing ready_to_create
	if chatResp.NextStep == "ready_to_create" || chatResp.NextStep == "final_review" {
		// Merge updatedDraft into current draft to check completeness
		mergedDraft := mergeDrafts(req.CourseDraft, chatResp.UpdatedDraft)

		if !isFinalDraftComplete(mergedDraft) {
			// Draft is incomplete - force back to clarification
			missingFields := getMissingFields(mergedDraft)
			log.Printf("[CourseCreationChatHandler] Draft incomplete, missing: %v", missingFields)

			// Override LLM response to ask for missing fields
			chatResp.Reply = buildMissingFieldsPrompt(missingFields)
			chatResp.NextStep = "final_review" // Stay in review to gather missing info
			chatResp.Suggestions = buildMissingFieldsSuggestions(missingFields)
		}
	}

	finalBytes, _ := json.Marshal(chatResp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(finalBytes)

	log.Printf("[CourseCreationChatHandler] Response sent, nextStep: %s", chatResp.NextStep)
}

// isFinalDraftComplete checks if all required fields are present and valid
func isFinalDraftComplete(draft *CourseDraft) bool {
	if draft == nil {
		return false
	}

	// Required: title (non-empty string)
	hasTitle := draft.Title != "" && strings.TrimSpace(draft.Title) != ""

	// Required: description (non-empty string)
	hasDescription := draft.Description != "" && strings.TrimSpace(draft.Description) != ""

	// Required: weeklyHours (numeric)
	weeklyOk := draft.WeeklyHours > 0

	// Required: lessons (we'll check if units exist as proxy)
	// The frontend expects a "lessons" field, which maps to number of units
	lessonsOk := len(draft.Units) > 0

	return hasTitle && hasDescription && weeklyOk && lessonsOk
}

// getMissingFields returns a list of missing required fields
func getMissingFields(draft *CourseDraft) []string {
	missing := []string{}

	if draft == nil {
		return []string{"title", "description", "weeklyHours", "lessons"}
	}

	if draft.Title == "" || strings.TrimSpace(draft.Title) == "" {
		missing = append(missing, "title")
	}

	if draft.Description == "" || strings.TrimSpace(draft.Description) == "" {
		missing = append(missing, "description")
	}

	if draft.WeeklyHours <= 0 {
		missing = append(missing, "weeklyHours")
	}

	if len(draft.Units) == 0 {
		missing = append(missing, "lessons")
	}

	return missing
}

// buildMissingFieldsPrompt creates a prompt asking for missing fields
func buildMissingFieldsPrompt(missingFields []string) string {
	if len(missingFields) == 0 {
		return "Your course looks complete! Ready to create it?"
	}

	prompts := map[string]string{
		"title":       "We still need a course title. What should we call this course?",
		"description": "Please give me 1-2 sentences describing what this course will cover.",
		"weeklyHours": "About how many hours per week can you commit to this course? Please give a whole number like 3, 5, or 10.",
		"lessons":     "How many lessons (units) should this course have? Please give a number like 4, 6, or 8.",
	}

	// Ask about the first missing field
	firstMissing := missingFields[0]
	if prompt, ok := prompts[firstMissing]; ok {
		return prompt
	}

	return "We need a bit more information to create your course. Can you provide more details?"
}

// buildMissingFieldsSuggestions creates suggestion chips for missing fields
func buildMissingFieldsSuggestions(missingFields []string) []string {
	if len(missingFields) == 0 {
		return []string{"Yes, create it!", "Let me review again"}
	}

	suggestions := map[string][]string{
		"title":       {"Introduction to [Topic]", "Mastering [Topic]", "Complete [Topic] Course"},
		"description": {"A comprehensive guide to...", "Learn the fundamentals of...", "Master the skills of..."},
		"weeklyHours": {"3 hours per week", "5 hours per week", "10 hours per week"},
		"lessons":     {"4 lessons", "6 lessons", "8 lessons"},
	}

	firstMissing := missingFields[0]
	if sugg, ok := suggestions[firstMissing]; ok {
		return sugg
	}

	return []string{}
}

// mergeDrafts merges updatedDraft map into the existing CourseDraft
func mergeDrafts(current *CourseDraft, updates map[string]interface{}) *CourseDraft {
	if current == nil {
		current = &CourseDraft{
			Objectives:     []string{},
			ReferenceBooks: []string{},
			Units:          []DraftUnit{},
		}
	}

	// Create a copy to avoid mutating the original
	merged := &CourseDraft{
		ID:                 current.ID,
		Title:              current.Title,
		Description:        current.Description,
		Objectives:         current.Objectives,
		PriorKnowledge:     current.PriorKnowledge,
		DifficultyLevel:    current.DifficultyLevel,
		WeeklyHours:        current.WeeklyHours,
		TotalDurationWeeks: current.TotalDurationWeeks,
		ReferenceBooks:     current.ReferenceBooks,
		Units:              current.Units,
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok {
		merged.Title = title
	}
	if desc, ok := updates["description"].(string); ok {
		merged.Description = desc
	}
	if obj, ok := updates["objectives"].([]interface{}); ok {
		merged.Objectives = interfaceSliceToStringSlice(obj)
	}
	if prior, ok := updates["priorKnowledge"].(string); ok {
		merged.PriorKnowledge = prior
	}
	if diff, ok := updates["difficultyLevel"].(string); ok {
		merged.DifficultyLevel = diff
	}
	if weekly, ok := updates["weeklyHours"].(float64); ok {
		merged.WeeklyHours = int(weekly)
	} else if weeklyStr, ok := updates["weeklyHours"].(string); ok {
		if val, err := strconv.Atoi(weeklyStr); err == nil {
			merged.WeeklyHours = val
		}
	}
	if duration, ok := updates["totalDurationWeeks"].(float64); ok {
		merged.TotalDurationWeeks = int(duration)
	}
	if refs, ok := updates["referenceBooks"].([]interface{}); ok {
		merged.ReferenceBooks = interfaceSliceToStringSlice(refs)
	}
	if units, ok := updates["units"].([]interface{}); ok {
		merged.Units = interfaceSliceToDraftUnits(units)
	}

	// Also check for "lessons" field which might be a number
	if lessons, ok := updates["lessons"].(float64); ok {
		// If lessons is provided but units is empty, we might need to generate units
		// For now, just log it
		log.Printf("[mergeDrafts] lessons field provided: %d", int(lessons))
	}

	return merged
}

// interfaceSliceToStringSlice converts []interface{} to []string
func interfaceSliceToStringSlice(slice []interface{}) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// interfaceSliceToDraftUnits converts []interface{} to []DraftUnit
func interfaceSliceToDraftUnits(slice []interface{}) []DraftUnit {
	result := make([]DraftUnit, 0, len(slice))
	for _, item := range slice {
		if unitMap, ok := item.(map[string]interface{}); ok {
			unit := DraftUnit{}
			if id, ok := unitMap["id"].(string); ok {
				unit.ID = id
			}
			if title, ok := unitMap["title"].(string); ok {
				unit.Title = title
			}
			if obj, ok := unitMap["objective"].(string); ok {
				unit.Objective = obj
			}
			if assess, ok := unitMap["assessment"].(string); ok {
				unit.Assessment = assess
			}
			if pos, ok := unitMap["position"].(float64); ok {
				unit.Position = int(pos)
			}
			result = append(result, unit)
		}
	}
	return result
}
