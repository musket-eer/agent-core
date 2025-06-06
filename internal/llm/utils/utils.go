package utils

import (
	"bytes"
	"strings"
	"text/template"
	"encoding/json"
	"net/http"
)

// Struct to pass into template engine
type PromptVars struct {
	Title          string
	Objective      string
	EducationLevel string
	Difficulty     string
}

// Safe substitution with fallbacks
func PromptWithVars(tmpl string, title, objective, educationLevel, difficulty string) string {
	data := PromptVars{
		Title:          fallback(title, "Untitled Lesson"),
		Objective:      fallback(objective, "Understand key concepts"),
		EducationLevel: fallback(educationLevel, "Intermediate"),
		Difficulty:     fallback(difficulty, "Medium"),
	}

	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		return "Error parsing prompt template"
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "Error executing prompt substitution"
	}

	result := buf.String()
	return result
}

// fallback assigns a default if value is empty
func fallback(value, def string) string {
	if value == "" {
		return def
	}
	return value
}

// WriteJSONError writes a standardized JSON error response
func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
func Trim(s string) string {
	return strings.TrimSpace(s)
}