package utils

import (
	"bytes"
	"log"
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
	log.Printf("[PromptWithVars] Template: %q, Title: %q, Objective: %q, EducationLevel: %q, Difficulty: %q", tmpl, title, objective, educationLevel, difficulty)
	data := PromptVars{
		Title:          fallback(title, "Untitled Lesson"),
		Objective:      fallback(objective, "Understand key concepts"),
		EducationLevel: fallback(educationLevel, "Intermediate"),
		Difficulty:     fallback(difficulty, "Medium"),
	}

	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		log.Printf("[PromptWithVars] Error parsing prompt template: %v", err)
		return "Error parsing prompt template"
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		log.Printf("[PromptWithVars] Error executing prompt substitution: %v", err)
		return "Error executing prompt substitution"
	}

	result := buf.String()
	log.Printf("[PromptWithVars] Result: %q", result)
	log.Printf("[PromptWithVars] Returning result: %q", result)
	return result
}

// fallback assigns a default if value is empty
func fallback(value, def string) string {
	if value == "" {
		log.Printf("[fallback] Using default: %q", def)
		return def
	}
	log.Printf("[fallback] Using value: %q", value)
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