// Here we implement the files that map an endpoint to its Bedrock configurations
// For every endpoint, we can grab quick configs and also call Bedrock API here.
package llm

import (
	"log"
	"net/http"
)

// RegisterHandlers returns a ServeMux with all Bedrock endpoints mapped.

func logRequest(endpoint string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[logRequest] Incoming request: %s %s", r.Method, endpoint)
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

	// Existing endpoints
	mux.HandleFunc("/generate-lesson", logRequest("/generate-lesson", GenerateLessonHandler))
	mux.HandleFunc("/generate-course", logRequest("/generate-course", GenerateCourseHandler))
	mux.HandleFunc("/generate-notes", logRequest("/generate-notes", GenerateNotesHandler))
	mux.HandleFunc("/generate-exams", logRequest("/generate-exams", GenerateExamsHandler))
	mux.HandleFunc("/generate-resources", logRequest("/generate-resources", GenerateResourcesHandler))
	mux.HandleFunc("/chat", logRequest("/chat", ChatHandler))
	mux.HandleFunc("/professor-chat", logRequest("/professor-chat", ProfessorChatHandler))
	mux.HandleFunc("/librarian-chat", logRequest("/librarian-chat", LibrarianChatHandler))
	mux.HandleFunc("/examiner-chat", logRequest("/examiner-chat", ExaminerChatHandler))

	// Orchestrator Agent endpoint
	mux.HandleFunc("/ai/orchestrator", logRequest("/ai/orchestrator", OrchestratorHandler))

	// Project Generation endpoint (stateless, aligned with frontend)
	mux.HandleFunc("/generate-project", logRequest("/generate-project", GenerateProjectHandler))

	// Generate Tasks endpoint (for milestone task generation)
	mux.HandleFunc("/generate-tasks", logRequest("/generate-tasks", GenerateTasksHandler))

	// Course Creation Chat endpoint (dialogue-based course creation)
	mux.HandleFunc("/course-creation/chat", logRequest("/course-creation/chat", CourseCreationChatHandler))

	// Optional: Milestone task refinement endpoint
	mux.HandleFunc("/ai/projects/milestone-to-tasks", logRequest("/ai/projects/milestone-to-tasks", MilestoneToTasksHandler))

	// Dashboard Agent endpoints
	mux.HandleFunc("/dashboard/insights", logRequest("/dashboard/insights", DashboardInsightsHandler))
	mux.HandleFunc("/dashboard/recommendations", logRequest("/dashboard/recommendations", DashboardRecommendationsHandler))
	mux.HandleFunc("/dashboard/analyze-progress", logRequest("/dashboard/analyze-progress", DashboardAnalyzeProgressHandler))
	mux.HandleFunc("/dashboard/chat", logRequest("/dashboard/chat", DashboardChatHandler))

	// Course Browse Agent endpoints
	mux.HandleFunc("/courses/intelligent-search", logRequest("/courses/intelligent-search", CourseIntelligentSearchHandler))
	mux.HandleFunc("/courses/explain-relevance", logRequest("/courses/explain-relevance", CourseExplainRelevanceHandler))
	mux.HandleFunc("/courses/compare", logRequest("/courses/compare", CourseCompareHandler))
	mux.HandleFunc("/courses/suggest-path", logRequest("/courses/suggest-path", CourseSuggestPathHandler))
	mux.HandleFunc("/courses/chat", logRequest("/courses/chat", CourseChatHandler))

	// Reflection Agent endpoints
	mux.HandleFunc("/reflection/generate-prompts", logRequest("/reflection/generate-prompts", ReflectionGeneratePromptsHandler))
	mux.HandleFunc("/reflection/analyze", logRequest("/reflection/analyze", ReflectionAnalyzeHandler))
	mux.HandleFunc("/reflection/learning-insights", logRequest("/reflection/learning-insights", ReflectionLearningInsightsHandler))
	mux.HandleFunc("/reflection/period-summary", logRequest("/reflection/period-summary", ReflectionPeriodSummaryHandler))
	mux.HandleFunc("/reflection/chat", logRequest("/reflection/chat", ReflectionChatHandler))
	mux.HandleFunc("/reflection/suggest-exercises", logRequest("/reflection/suggest-exercises", ReflectionSuggestExercisesHandler))
	mux.HandleFunc("/reflection/generate-task-reflection", logRequest("/reflection/generate-task-reflection", GenerateTaskReflectionHandler))
	mux.HandleFunc("/reflection/task-prompts", logRequest("/reflection/task-prompts", TaskPromptsHandler))
	mux.HandleFunc("/reflection/analyze-task", logRequest("/reflection/analyze-task", AnalyzeTaskHandler))

	// Tasks Agent endpoints
	mux.HandleFunc("/tasks/generate-suggestions", logRequest("/tasks/generate-suggestions", TasksGenerateSuggestionsHandler))
	mux.HandleFunc("/tasks/prioritize", logRequest("/tasks/prioritize", TasksPrioritizeHandler))
	mux.HandleFunc("/tasks/break-down", logRequest("/tasks/break-down", TasksBreakDownHandler))
	mux.HandleFunc("/tasks/productivity-insights", logRequest("/tasks/productivity-insights", TasksProductivityInsightsHandler))
	mux.HandleFunc("/tasks/generate-schedule", logRequest("/tasks/generate-schedule", TasksGenerateScheduleHandler))
	mux.HandleFunc("/tasks/chat", logRequest("/tasks/chat", TasksChatHandler))
	mux.HandleFunc("/tasks/motivation", logRequest("/tasks/motivation", TasksMotivationHandler))

	// Librarian Agent endpoints
	mux.HandleFunc("/chat-video", logRequest("/chat-video", ChatVideoHandler))
	mux.HandleFunc("/recommend-videos", logRequest("/recommend-videos", RecommendVideosHandler))
	mux.HandleFunc("/summarize-video", logRequest("/summarize-video", SummarizeVideoHandler))
	mux.HandleFunc("/extract-concepts", logRequest("/extract-concepts", ExtractConceptsHandler))
	mux.HandleFunc("/generate-quiz", logRequest("/generate-quiz", GenerateQuizHandler))
	mux.HandleFunc("/find-topic", logRequest("/find-topic", FindTopicHandler))

	// Project Assistant Agent endpoints
	mux.HandleFunc("/projects/chat", logRequest("/projects/chat", ProjectsChatHandler))
	mux.HandleFunc("/projects/suggestions", logRequest("/projects/suggestions", ProjectsSuggestionsHandler))
	mux.HandleFunc("/projects/help", logRequest("/projects/help", ProjectsHelpHandler))
	mux.HandleFunc("/projects/break-down", logRequest("/projects/break-down", ProjectsBreakDownHandler))
	mux.HandleFunc("/projects/resources", logRequest("/projects/resources", ProjectsResourcesHandler))

	// Project AI Agent endpoints (stateless)
	mux.HandleFunc("/projects/create", logRequest("/projects/create", ProjectCreatorHandler))
	mux.HandleFunc("/projects/assist", logRequest("/projects/assist", ProjectAssistantHandler))
	mux.HandleFunc("/projects/translate-milestones", logRequest("/projects/translate-milestones", MilestoneTranslatorHandler))

	return mux
}
