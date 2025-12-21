package middleware

import (
	"fmt"
	"strings"
	"time"
)

// ContextData contains all memory-derived context
type ContextData struct {
	UserInstructions      string
	SessionMetadata       string
	UserMemory            string
	ConversationSummaries string
	CurrentSession        string
	ComponentContext      string
}

// BuildSystemPrompt creates a context-enriched system prompt
func BuildSystemPrompt(basePrompt string, context ContextData) string {
	var sections []string

	// Start with base prompt
	sections = append(sections, basePrompt)

	// Add context sections in priority order
	if context.UserMemory != "" {
		sections = append(sections, fmt.Sprintf("\n## User Profile\n%s", context.UserMemory))
	}

	if context.UserInstructions != "" {
		sections = append(sections, fmt.Sprintf("\n## User Preferences\n%s", context.UserInstructions))
	}

	if context.ConversationSummaries != "" {
		sections = append(sections, fmt.Sprintf("\n## Recent Context\n%s", context.ConversationSummaries))
	}

	if context.ComponentContext != "" {
		sections = append(sections, fmt.Sprintf("\n## Current Context\n%s", context.ComponentContext))
	}

	if context.CurrentSession != "" {
		sections = append(sections, fmt.Sprintf("\n## Conversation History\n%s", context.CurrentSession))
	}

	if context.SessionMetadata != "" {
		sections = append(sections, fmt.Sprintf("\n## Session Info\n%s", context.SessionMetadata))
	}

	return strings.Join(sections, "\n")
}

// ExtractContext extracts context from V2 payload
func ExtractContext(payload map[string]interface{}) ContextData {
	context := ContextData{}

	if ctx, ok := payload["context"].(map[string]interface{}); ok {
		if val, ok := ctx["userInstructions"].(string); ok {
			context.UserInstructions = val
		}
		if val, ok := ctx["sessionMetadata"].(string); ok {
			context.SessionMetadata = val
		}
		if val, ok := ctx["userMemory"].(string); ok {
			context.UserMemory = val
		}
		if val, ok := ctx["conversationSummaries"].(string); ok {
			context.ConversationSummaries = val
		}
		if val, ok := ctx["currentSession"].(string); ok {
			context.CurrentSession = val
		}
		if val, ok := ctx["componentContext"].(string); ok {
			context.ComponentContext = val
		}
	}

	return context
}

// GetContextSources returns which context sources were used
func GetContextSources(context ContextData) []string {
	sources := []string{}

	if context.UserMemory != "" {
		sources = append(sources, "userMemory")
	}
	if context.UserInstructions != "" {
		sources = append(sources, "userInstructions")
	}
	if context.ConversationSummaries != "" {
		sources = append(sources, "conversationSummaries")
	}
	if context.ComponentContext != "" {
		sources = append(sources, "componentContext")
	}
	if context.CurrentSession != "" {
		sources = append(sources, "currentSession")
	}
	if context.SessionMetadata != "" {
		sources = append(sources, "sessionMetadata")
	}

	return sources
}

// EstimateTokens provides a rough token estimate for context
func EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token
	return len(text) / 4
}

// TruncateContext truncates context to fit token limits
func TruncateContext(context ContextData, maxTokens int) ContextData {
	currentTokens := EstimateTokens(context.UserMemory) +
		EstimateTokens(context.UserInstructions) +
		EstimateTokens(context.ConversationSummaries) +
		EstimateTokens(context.ComponentContext) +
		EstimateTokens(context.CurrentSession) +
		EstimateTokens(context.SessionMetadata)

	if currentTokens <= maxTokens {
		return context
	}

	// Priority order for truncation (keep most important)
	// 1. User instructions (always keep)
	// 2. Component context (current task/activity)
	// 3. User memory (profile)
	// 4. Conversation summaries
	// 5. Current session (truncate oldest messages)
	// 6. Session metadata

	truncated := context

	// Truncate session metadata first (least important)
	if currentTokens > maxTokens {
		truncated.SessionMetadata = truncateString(truncated.SessionMetadata, 200)
		currentTokens = recalculateTokens(truncated)
	}

	// Truncate current session (keep recent messages)
	if currentTokens > maxTokens {
		truncated.CurrentSession = truncateString(truncated.CurrentSession, 1000)
		currentTokens = recalculateTokens(truncated)
	}

	// Truncate conversation summaries
	if currentTokens > maxTokens {
		truncated.ConversationSummaries = truncateString(truncated.ConversationSummaries, 500)
		currentTokens = recalculateTokens(truncated)
	}

	// Truncate user memory
	if currentTokens > maxTokens {
		truncated.UserMemory = truncateString(truncated.UserMemory, 400)
	}

	return truncated
}

// truncateString truncates a string to approximately N characters
func truncateString(s string, maxChars int) string {
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "..."
}

// recalculateTokens recalculates total tokens for context
func recalculateTokens(context ContextData) int {
	return EstimateTokens(context.UserMemory) +
		EstimateTokens(context.UserInstructions) +
		EstimateTokens(context.ConversationSummaries) +
		EstimateTokens(context.ComponentContext) +
		EstimateTokens(context.CurrentSession) +
		EstimateTokens(context.SessionMetadata)
}

// FormatTimestamp formats a timestamp for display
func FormatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

// ParseMessages extracts messages from V2 payload
func ParseMessages(payload map[string]interface{}) []map[string]string {
	messages := []map[string]string{}

	if msgs, ok := payload["messages"].([]interface{}); ok {
		for _, msg := range msgs {
			if m, ok := msg.(map[string]interface{}); ok {
				message := map[string]string{}
				if role, ok := m["role"].(string); ok {
					message["role"] = role
				}
				if content, ok := m["content"].(string); ok {
					message["content"] = content
				}
				messages = append(messages, message)
			}
		}
	}

	return messages
}
