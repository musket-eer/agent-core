# JSON Response Improvements

## Problem
LLM responses were sometimes wrapped in markdown code blocks or had extra text, causing JSON parsing failures. This resulted in responses like:
```json
{ "response": "Hi Student! I can see you're working on..." }
```
Instead of properly structured responses with all fields.

## Solution Implemented

### 1. Created `CleanJSONResponse` Utility Function
**Location:** `/internal/llm/utils/parser.go`

This function:
- Removes markdown code fences (```json or ```)
- Strips leading/trailing non-JSON text
- Extracts only the valid JSON object/array
- Handles malformed responses gracefully

```go
func CleanJSONResponse(rawResponse string) string {
    // Removes markdown, finds first { or [, finds last } or ]
    // Returns clean JSON string
}
```

### 2. Updated All Handlers to Use JSON Cleaning

#### Handlers Using `handleLLMRequest` Helper
These handlers now automatically clean JSON responses:
- **courses_handlers.go** - All 5 course endpoints
- **tasks_handlers.go** - All 7 task endpoints  
- **reflection_handlers.go** - All 6 reflection endpoints
- **projects_handlers.go** - All 5 project endpoints
- **librarian_handlers.go** - All 6 librarian endpoints

#### Direct Handler Updates
Added `utils.CleanJSONResponse()` to:
- **generate-course.go** - Course generation
- **generate-lesson.go** - Lesson generation
- **generate-exams.go** - Exam generation
- **generate-resources.go** - Resource generation
- **chat.go** - Chat handlers
- **dashboard_handlers.go** - Dashboard insights

## Response Structures Already Defined

All endpoints have properly typed response structures in:
- `types_tasks.go` - Task agent responses
- `types_courses.go` - Course agent responses
- `types_reflection.go` - Reflection agent responses
- `types_agents.go` - Project & Librarian responses
- `types_dashboard.go` - Dashboard responses
- `types_common.go` - Shared types

### Example: TasksChatResponse
```go
type TasksChatResponse struct {
    Response            string               `json:"response"`
    SuggestedActions    []SuggestedAction    `json:"suggestedActions"`
    TaskRecommendations []TaskRecommendation `json:"taskRecommendations"`
}

type SuggestedAction struct {
    Action      string `json:"action"`
    Description string `json:"description"`
    Priority    string `json:"priority"` // "low", "medium", "high"
}

type TaskRecommendation struct {
    What         string `json:"what"`
    Why          string `json:"why"`
    UrgencyLevel string `json:"urgencyLevel"`
}
```

## Benefits

1. **Consistent JSON Responses** - All endpoints now return properly structured JSON
2. **Better Error Handling** - Graceful fallbacks when JSON parsing fails
3. **Frontend-Friendly** - Clean, predictable response formats
4. **Maintainable** - Single utility function handles all JSON cleaning

## Testing Recommendations

Test each endpoint to ensure:
1. Responses match the defined struct types
2. All fields are populated correctly
3. No markdown or extra text in responses
4. Fallback behavior works when LLM returns invalid JSON

## Next Steps (If Needed)

If responses still don't match expected structures:
1. **Update Prompts** - Ensure system prompts explicitly request JSON format matching the struct
2. **Add Response Validation** - Log when responses don't match expected schema
3. **Improve Fallbacks** - Better default values when parsing fails
