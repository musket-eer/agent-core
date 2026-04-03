package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"agent-core/internal/llm/configs"
	"agent-core/internal/llm/utils"
	"agent-core/internal/prompt"

	"github.com/google/uuid"
)

const courseCreationGenerateEndpoint = "/course-creation/generate"

type OneShotCourseCreationRequest struct {
	Prompt         string                      `json:"prompt"`
	LearnerProfile OneShotCourseLearnerProfile `json:"learnerProfile"`
}

type OneShotCourseLearnerProfile struct {
	Name           string `json:"name"`
	EducationLevel string `json:"educationLevel"`
	CareerGoals    string `json:"careerGoals"`
	Interests      string `json:"interests"`
	Certifications string `json:"certifications"`
}

type OneShotCourseCreationResponse struct {
	Course CourseGraphCourse `json:"course"`
}

func CourseCreationGenerateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CourseCreationGenerateHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OneShotCourseCreationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CourseCreationGenerateHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		utils.WriteJSONError(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	course, err := generateOneShotCourse(r.Context(), req)
	if err != nil {
		log.Printf("[CourseCreationGenerateHandler] Failed to generate course: %v", err)
		utils.WriteJSONError(w, "Failed to generate course", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(OneShotCourseCreationResponse{Course: course}); err != nil {
		log.Printf("[CourseCreationGenerateHandler] Failed to encode response: %v", err)
	}
}

func generateOneShotCourse(ctx context.Context, req OneShotCourseCreationRequest) (CourseGraphCourse, error) {
	systemPrompt, err := prompt.GetPrompt("oneShotCourseCreationPrompt")
	if err != nil {
		return CourseGraphCourse{}, err
	}

	cfg, ok := configs.GetDeepseekConfig(courseCreationGenerateEndpoint)
	if !ok {
		return CourseGraphCourse{}, fmt.Errorf("no config found for endpoint: %s", courseCreationGenerateEndpoint)
	}

	clientName := cfg.ClientName
	if clientName == "" {
		clientName = "general"
	}
	client, err := GetLLMClient(clientName)
	if err != nil {
		return CourseGraphCourse{}, err
	}

	userPayload, err := json.MarshalIndent(map[string]interface{}{
		"prompt":         strings.TrimSpace(req.Prompt),
		"learnerProfile": req.LearnerProfile,
	}, "", "  ")
	if err != nil {
		return CourseGraphCourse{}, err
	}

	payload := map[string]interface{}{
		"model": cfg.ModelId,
		"messages": []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: string(userPayload)},
		},
	}

	rawResponse, err := getLLMCallQueue().submit(ctx, func() (string, error) {
		return client.Call(courseCreationGenerateEndpoint, payload)
	})
	if err != nil {
		return CourseGraphCourse{}, err
	}

	cleaned := utils.CleanJSONResponse(rawResponse)

	var wrapped OneShotCourseCreationResponse
	if err := json.Unmarshal([]byte(cleaned), &wrapped); err == nil && strings.TrimSpace(wrapped.Course.Title) != "" {
		normalizeGeneratedCourse(&wrapped.Course)
		return wrapped.Course, nil
	}

	var course CourseGraphCourse
	if err := json.Unmarshal([]byte(cleaned), &course); err != nil {
		return CourseGraphCourse{}, err
	}
	normalizeGeneratedCourse(&course)
	return course, nil
}

func normalizeGeneratedCourse(course *CourseGraphCourse) {
	if course == nil {
		return
	}

	if strings.TrimSpace(course.ID) == "" {
		course.ID = "course_generated_" + uuid.New().String()
	}
	course.Title = strings.TrimSpace(course.Title)
	course.Description = strings.TrimSpace(course.Description)
	course.Objectives = trimStringSlice(course.Objectives)
	course.PriorKnowledge = trimStringSlice(course.PriorKnowledge)
	course.DifficultyLevel = strings.TrimSpace(course.DifficultyLevel)
	course.ReferenceBooks = trimStringSlice(course.ReferenceBooks)
	if course.WeeklyHours <= 0 {
		course.WeeklyHours = 4
	}

	for i := range course.Units {
		if strings.TrimSpace(course.Units[i].ID) == "" {
			course.Units[i].ID = "unit_" + uuid.New().String()
		}
		course.Units[i].Title = strings.TrimSpace(course.Units[i].Title)
		course.Units[i].Objective = strings.TrimSpace(course.Units[i].Objective)
		course.Units[i].Assessment = strings.TrimSpace(course.Units[i].Assessment)
		course.Units[i].DeepDivePrompt = strings.TrimSpace(course.Units[i].DeepDivePrompt)
		course.Units[i].Position = i
		if strings.TrimSpace(course.Units[i].KGTopicID) == "" {
			if i < len(course.KnowledgeGraph) && strings.TrimSpace(course.KnowledgeGraph[i].ID) != "" {
				course.Units[i].KGTopicID = course.KnowledgeGraph[i].ID
			} else {
				course.Units[i].KGTopicID = fmt.Sprintf("topic_%d", i+1)
			}
		}
	}

	for i := range course.KnowledgeGraph {
		if strings.TrimSpace(course.KnowledgeGraph[i].ID) == "" {
			course.KnowledgeGraph[i].ID = fmt.Sprintf("topic_%d", i+1)
		}
		course.KnowledgeGraph[i].Title = strings.TrimSpace(course.KnowledgeGraph[i].Title)
		course.KnowledgeGraph[i].Summary = strings.TrimSpace(course.KnowledgeGraph[i].Summary)
		course.KnowledgeGraph[i].KeyPoints = trimStringSlice(course.KnowledgeGraph[i].KeyPoints)
		course.KnowledgeGraph[i].Examples = trimStringSlice(course.KnowledgeGraph[i].Examples)
		course.KnowledgeGraph[i].Prerequisites = trimStringSlice(course.KnowledgeGraph[i].Prerequisites)
		course.KnowledgeGraph[i].DeepDivePrompt = strings.TrimSpace(course.KnowledgeGraph[i].DeepDivePrompt)
		course.KnowledgeGraph[i].Order = i
	}

	if course.Lessons <= 0 {
		course.Lessons = len(course.Units)
	}
	if strings.TrimSpace(course.Metadata.GenerationMode) == "" {
		course.Metadata.GenerationMode = "one_shot_prompt"
	}
}

func trimStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
