package llm

import (
	"encoding/json"
	"log"
	"net/http"
	"agent-core/internal/llm/utils"
)

// LibrarianGenerateResourcesHandler - POST /generate-resources (already exists, keeping for compatibility)
// Handled by generate-resources.go

// ChatVideoHandler - POST /chat-video
func ChatVideoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ChatVideoHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ChatVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ChatVideoHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "librarianVideoChatPrompt", "/chat-video")
	if response != nil {
		var resp ChatVideoResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ChatVideoResponse{Response: string(response), Timestamp: 0}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// RecommendVideosHandler - POST /recommend-videos
func RecommendVideosHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[RecommendVideosHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req RecommendVideosRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[RecommendVideosHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "librarianRecommendVideosPrompt", "/recommend-videos")
	if response != nil {
		var resp RecommendVideosResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = RecommendVideosResponse{Recommendations: []VideoRecommendation{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// SummarizeVideoHandler - POST /summarize-video
func SummarizeVideoHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[SummarizeVideoHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req SummarizeVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[SummarizeVideoHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "librarianSummarizeVideoPrompt", "/summarize-video")
	if response != nil {
		var resp SummarizeVideoResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = SummarizeVideoResponse{
				Summary: VideoSummary{Overview: string(response), KeyPoints: []string{}, Timestamps: []TimestampTopic{}},
			}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// ExtractConceptsHandler - POST /extract-concepts
func ExtractConceptsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ExtractConceptsHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req ExtractConceptsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ExtractConceptsHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "librarianExtractConceptsPrompt", "/extract-concepts")
	if response != nil {
		var resp ExtractConceptsResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = ExtractConceptsResponse{Concepts: []Concept{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// GenerateQuizHandler - POST /generate-quiz
func GenerateQuizHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GenerateQuizHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req GenerateQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[GenerateQuizHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "librarianGenerateQuizPrompt", "/generate-quiz")
	if response != nil {
		var resp GenerateQuizResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = GenerateQuizResponse{Questions: []QuizQuestion{}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}

// FindTopicHandler - POST /find-topic
func FindTopicHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[FindTopicHandler] Incoming request: %s %s", r.Method, r.URL.Path)
	var req FindTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[FindTopicHandler] Failed to decode request: %v", err)
		utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response := handleLLMRequest(w, req, "librarianFindTopicPrompt", "/find-topic")
	if response != nil {
		var resp FindTopicResponse
		if err := json.Unmarshal(response, &resp); err != nil {
			resp = FindTopicResponse{Results: TopicResults{Timestamps: []int{}, Context: []string{}}}
		}
		finalBytes, _ := json.Marshal(resp)
		w.Write(finalBytes)
	}
}
