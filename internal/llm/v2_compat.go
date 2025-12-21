package llm

import (
	"log"
	"net/http"
	"reflect"

	"agent-core/internal/llm/v2/middleware"
	"agent-core/internal/llm/v2/types"
)

// V2CompatibleRequest interface for requests that support V2 context
type V2CompatibleRequest interface {
	GetV2Context() *types.V2Context
	GetUserID() string
	GetAgentID() string
}

// extractV2ContextCompat extracts V2 context from any request type
// Supports both new V2 format and legacy V1 format for backward compatibility
func extractV2ContextCompat(req interface{}) middleware.ContextData {
	// Try to extract V2 context directly
	if v2Req, ok := req.(V2CompatibleRequest); ok {
		ctx := v2Req.GetV2Context()
		if ctx != nil {
			return middleware.ContextData{
				UserInstructions:      ctx.UserInstructions,
				SessionMetadata:       ctx.SessionMetadata,
				UserMemory:            ctx.UserMemory,
				ConversationSummaries: ctx.ConversationSummaries,
				CurrentSession:        ctx.CurrentSession,
				ComponentContext:      ctx.ComponentContext,
			}
		}
	}

	// Try reflection to find Context field
	val := reflect.ValueOf(req)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		// Look for Context field
		contextField := val.FieldByName("Context")
		if contextField.IsValid() && !contextField.IsZero() {
			// Try to convert to V2Context
			if ctx, ok := contextField.Interface().(types.V2Context); ok {
				return middleware.ContextData{
					UserInstructions:      ctx.UserInstructions,
					SessionMetadata:       ctx.SessionMetadata,
					UserMemory:            ctx.UserMemory,
					ConversationSummaries: ctx.ConversationSummaries,
					CurrentSession:        ctx.CurrentSession,
					ComponentContext:      ctx.ComponentContext,
				}
			}
			// Try pointer to V2Context
			if ctxPtr, ok := contextField.Interface().(*types.V2Context); ok && ctxPtr != nil {
				return middleware.ContextData{
					UserInstructions:      ctxPtr.UserInstructions,
					SessionMetadata:       ctxPtr.SessionMetadata,
					UserMemory:            ctxPtr.UserMemory,
					ConversationSummaries: ctxPtr.ConversationSummaries,
					CurrentSession:        ctxPtr.CurrentSession,
					ComponentContext:      ctxPtr.ComponentContext,
				}
			}
		}
	}

	// Fallback: return empty context for legacy V1 requests
	log.Printf("[extractV2ContextCompat] No V2 context found in request, using empty context")
	return middleware.ContextData{}
}

// buildV2SystemPrompt builds a system prompt with V2 context integration
// This replaces the old persona-based system
func buildV2SystemPrompt(basePrompt string, contextData middleware.ContextData) string {
	return middleware.BuildSystemPrompt(basePrompt, contextData)
}

// handleLLMRequestV2 is the new V2-compatible LLM request handler
// It replaces handleLLMRequest and uses V2 context extraction
func handleLLMRequestV2(w http.ResponseWriter, req interface{}, promptKey string, endpoint string) []byte {
	log.Printf("[handleLLMRequestV2] Processing request for endpoint: %s", endpoint)

	// Use the existing handleLLMRequest but with V2 context extraction
	// This is a transitional approach - we'll fully replace handleLLMRequest later
	return handleLLMRequest(w, req, promptKey, endpoint)
}

// isV2Request checks if a request contains V2 context
func isV2Request(req interface{}) bool {
	if v2Req, ok := req.(V2CompatibleRequest); ok {
		return v2Req.GetV2Context() != nil
	}

	// Check via reflection
	val := reflect.ValueOf(req)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		contextField := val.FieldByName("Context")
		if contextField.IsValid() && !contextField.IsZero() {
			// Check if it's a V2Context type
			_, ok := contextField.Interface().(types.V2Context)
			if ok {
				return true
			}
			_, ok = contextField.Interface().(*types.V2Context)
			return ok
		}
	}

	return false
}

// logRequestType logs whether request is V1 or V2 format
func logRequestType(req interface{}, endpoint string) {
	if isV2Request(req) {
		log.Printf("[%s] V2 request detected (with memory context)", endpoint)
	} else {
		log.Printf("[%s] V1 request detected (legacy format)", endpoint)
	}
}
