// Here we implement the files that map an endpoint to its Bedrock configurations
// For every endpoint, we can grab quick configs and also call Bedrock API here.
package llm

import (
	"net/http"
)

// RegisterHandlers returns a ServeMux with all Bedrock endpoints mapped.
func logRequest(endpoint string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		handler(w, r)
	}
}

func RegisterHandlers() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/generate-lesson", logRequest("/generate-lesson", GenerateLessonHandler))
	mux.HandleFunc("/generate-course", logRequest("/generate-course", GenerateCourseHandler))
	mux.HandleFunc("/generate-notes", logRequest("/generate-notes", GenerateNotesHandler))
	mux.HandleFunc("/generate-exams", logRequest("/generate-exams", GenerateExamsHandler))
	mux.HandleFunc("/generate-resources", logRequest("/generate-resources", GenerateResourcesHandler))
	mux.HandleFunc("/chat", logRequest("/chat", ChatHandler))
	mux.HandleFunc("/professor-chat", logRequest("/professor-chat", ProfessorChatHandler))
	mux.HandleFunc("/librarian-chat", logRequest("/librarian-chat", LibrarianChatHandler))
	mux.HandleFunc("/examiner-chat", logRequest("/examiner-chat", ExaminerChatHandler))
	return mux
}