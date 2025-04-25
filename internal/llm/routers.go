// Here we implement the files that map an endpoint to its Bedrock configurations
// For every endpoint, we can grab quick configs and also call Bedrock API here.
package llm

import (
	"net/http"
)

// RegisterHandlers returns a ServeMux with all Bedrock endpoints mapped.
func RegisterHandlers() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/generate-lesson", GenerateLessonHandler)
	mux.HandleFunc("/generate-course", GenerateCourseHandler)
	mux.HandleFunc("/generate-notes", GenerateNotesHandler)
	mux.HandleFunc("/generate-exams", GenerateExamsHandler)
	mux.HandleFunc("/generate-resources", GenerateResourcesHandler)
	mux.HandleFunc("/chat", ChatHandler)
	return mux
}