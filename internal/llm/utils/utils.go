package utils

import (
	"fmt"
	"encoding/json"
	"net/http"
)

// PromptWithVars substitutes variables into the prompt template.
// You can adjust the formatting logic as needed for your prompt templates.
func PromptWithVars(template string, title, objective, educationLevel, difficulty string) string {
	return fmt.Sprintf(template, title, objective, educationLevel, difficulty)
}

// WriteJSONError writes a standardized JSON error response
func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
