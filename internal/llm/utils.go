package llm

import "fmt"

// promptWithVars substitutes variables into the prompt template.
// You can adjust the formatting logic as needed for your prompt templates.
func promptWithVars(template string, title, objective, educationLevel, difficulty string) string {
	return fmt.Sprintf(template, title, objective, educationLevel, difficulty)
}
