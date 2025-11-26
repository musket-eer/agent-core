# Persona-Aware AI Agent Integration

## Overview

The backend now supports **persona-aware AI agents** that adapt their behavior, tone, and communication style based on user-customized persona settings sent from the frontend.

**Key Principle:** The frontend sends `agentId` and `persona` with every request. The backend uses this to build persona-aware system prompts that customize the LLM's behavior.

---

## Architecture

```
Frontend → Request with persona → Backend → Persona-aware prompt → LLM → Customized response
```

1. Frontend sends request with `agentId` and `persona` object
2. Backend extracts persona information
3. Backend builds persona-aware system prompt
4. LLM generates response following persona instructions
5. Backend returns response to frontend

---

## Request Format

All agent endpoints now accept:

```json
{
  "agentId": "tasks",
  "persona": {
    "agentId": "tasks",
    "name": "Tasks Assistant",
    "description": "Helps with task planning...",
    "icon": "fa-list-check",
    "defaultTraits": { ... },
    "userTraits": { ... },
    "mergedTraits": { ... },
    "instructions": "Persona Customization:\nCommunication Tone: friendly\nVerbosity: concise\n..."
  },
  // ... other endpoint-specific fields
}
```

**Important:**
- `agentId` identifies which agent is being used
- `persona` contains the full persona configuration including user customizations
- `persona.instructions` is the formatted string of persona traits that guide the LLM
- Backend trusts this data as authoritative (no DB queries needed)

---

## Implementation Details

### 1. Core Files

#### `internal/llm/persona.go`
Contains persona extraction and system prompt building logic:

```go
// Persona represents the agent's personality and behavior settings
type Persona struct {
    AgentID       string
    Name          string
    Description   string
    Icon          string
    DefaultTraits map[string]interface{}
    UserTraits    map[string]interface{}
    MergedTraits  map[string]interface{}
    Instructions  string
}

// PersonaRequest is embedded in all agent requests
type PersonaRequest struct {
    AgentID string   `json:"agentId,omitempty"`
    Persona *Persona `json:"persona,omitempty"`
}

// ExtractPersona extracts agentId and persona from a request
func ExtractPersona(req interface{}) (string, *Persona)

// BuildPersonaSystemPrompt creates a persona-aware system prompt
func BuildPersonaSystemPrompt(agentID string, persona *Persona, basePrompt string) string
```

### 2. Updated Request Types

All chat and agent request types now embed `PersonaRequest`:

**Updated Types:**
- `ProfessorChatRequest`
- `GenerateNotesRequest`
- `GenerateLessonRequest`
- `TasksChatRequest`
- `DashboardChatRequest`
- `ReflectionChatRequest`
- `CourseChatRequest`
- `OrchestratorRequest`

Each type implements `GetPersona()` method:

```go
type TasksChatRequest struct {
    PersonaRequest
    Message      string
    UserID       string
    // ... other fields
}

func (r TasksChatRequest) GetPersona() (string, *Persona) {
    return r.AgentID, r.Persona
}
```

### 3. Updated Handler Logic

The core `handleLLMRequest` function in `internal/llm/courses_handlers.go` now:

1. Loads the base prompt from file
2. Extracts persona information from request
3. Builds persona-aware system prompt
4. Calls LLM with customized prompt

```go
func handleLLMRequest(w http.ResponseWriter, req interface{}, promptKey string, endpoint string) []byte {
    // ... cache check ...
    
    basePrompt, err := prompt.GetPrompt(promptKey)
    // ... error handling ...
    
    // Extract persona information if available
    agentID, persona := ExtractPersona(req)
    
    // Build persona-aware system prompt
    systemPrompt := BuildPersonaSystemPrompt(agentID, persona, basePrompt)
    
    // ... rest of LLM call ...
}
```

---

## Persona System Prompt Format

When a persona is provided, the system prompt is structured as:

```
You are the [Agent Name] agent in the Most Learned platform.

Follow these persona instructions exactly when you write responses:

[Persona Instructions from frontend]

If traits indicate specific tone, verbosity, strictness, or style,
reflect that consistently in your responses.

When relevant, explain your reasoning clearly but stay aligned
with the persona settings.

---

[Original base prompt from .txt file]
```

**Example:**

If the frontend sends:
```json
{
  "agentId": "tasks",
  "persona": {
    "name": "Tasks Assistant",
    "instructions": "Communication Tone: friendly\nVerbosity: concise\nStrictness: flexible"
  }
}
```

The system prompt becomes:
```
You are the Tasks Assistant agent in the Most Learned platform.

Follow these persona instructions exactly when you write responses:

Communication Tone: friendly
Verbosity: concise
Strictness: flexible

If traits indicate specific tone, verbosity, strictness, or style,
reflect that consistently in your responses.

When relevant, explain your reasoning clearly but stay aligned
with the persona settings.

---

[tasksChatPrompt.txt content...]
```

---

## Persona Traits

The frontend can customize these traits (examples):

### Communication Tone
- `professional` - Formal, business-like
- `friendly` - Warm, approachable
- `motivational` - Encouraging, energetic
- `casual` - Relaxed, conversational

### Verbosity
- `concise` - Brief, to-the-point responses
- `balanced` - Moderate detail
- `detailed` - Comprehensive explanations

### Strictness
- `strict` - Firm guidance, clear rules
- `flexible` - Adaptable, understanding
- `balanced` - Mix of structure and flexibility

### Formality
- `formal` - Professional language
- `informal` - Casual language
- `balanced` - Context-appropriate

---

## Testing

### Test Persona-Aware Request

```bash
curl -X POST http://localhost:8080/tasks/chat \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "tasks",
    "persona": {
      "agentId": "tasks",
      "name": "Friendly Task Helper",
      "instructions": "Communication Tone: friendly\nVerbosity: concise\nStrictness: flexible\n\nAlways use encouraging language and keep responses brief."
    },
    "message": "Help me organize my tasks",
    "userId": "test-user",
    "userName": "Test User",
    "userProfile": {},
    "chatHistory": [],
    "tasksContext": {}
  }'
```

### Test Without Persona (Fallback)

```bash
curl -X POST http://localhost:8080/tasks/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Help me organize my tasks",
    "userId": "test-user",
    "userName": "Test User",
    "userProfile": {},
    "chatHistory": [],
    "tasksContext": {}
  }'
```

The backend will use the base prompt without persona customization.

---

## Affected Endpoints

All these endpoints now support persona-aware prompts:

### Professor Agent
- `POST /professor-chat`
- `POST /generate-notes`
- `POST /generate-lesson`

### Tasks Agent
- `POST /tasks/chat`

### Dashboard Agent
- `POST /dashboard/chat`

### Reflection Agent
- `POST /reflection/chat`

### Course Browse Agent
- `POST /courses/chat`

### Orchestrator Agent
- `POST /ai/orchestrator`

---

## Logging

The persona system logs useful information:

```
[Persona] Using persona: Friendly Task Helper with custom instructions: true
```

This helps debug persona application and verify customizations are being used.

---

## Backward Compatibility

✅ **Fully backward compatible**

If frontend doesn't send `agentId` or `persona`:
- `ExtractPersona()` returns `("unknown", nil)`
- `BuildPersonaSystemPrompt()` uses just the base prompt
- Behavior is identical to pre-persona implementation

---

## Frontend Integration

### Sending Persona

```javascript
// In your agent service
const response = await apiCall('/tasks/chat', {
  agentId: 'tasks',
  persona: {
    agentId: 'tasks',
    name: agent.name,
    instructions: formatPersonaInstructions(agent.mergedTraits),
    // ... other persona fields
  },
  message: userMessage,
  // ... other fields
});
```

### Formatting Persona Instructions

```javascript
function formatPersonaInstructions(traits) {
  return Object.entries(traits)
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n');
}
```

---

## Benefits

✅ **User Customization** - Users can personalize how agents communicate  
✅ **Consistent Behavior** - Persona traits applied across all interactions  
✅ **No DB Queries** - Frontend sends all needed data  
✅ **Flexible** - Easy to add new traits or customize behavior  
✅ **Backward Compatible** - Works with or without persona data  
✅ **Centralized Logic** - Single place to handle persona prompts  

---

## Future Enhancements

### Potential Additions
- **Emoji Usage** - Control emoji frequency in responses
- **Humor Level** - Adjust joke/wit in responses
- **Technical Depth** - Control complexity of explanations
- **Proactivity** - How much the agent suggests vs waits
- **Memory** - How much context the agent references

### Advanced Features
- **Persona Templates** - Pre-built persona configurations
- **Persona Learning** - Adapt based on user interactions
- **Multi-Persona** - Different personas for different contexts
- **Persona Validation** - Backend validation of persona settings

---

## Summary

The persona-aware system allows:

1. ✅ Frontend sends `agentId` and `persona` with every request
2. ✅ Backend extracts persona information
3. ✅ Backend builds customized system prompts
4. ✅ LLM adapts behavior based on persona traits
5. ✅ Users get personalized AI interactions
6. ✅ No database queries needed
7. ✅ Fully backward compatible

All agent endpoints now support persona customization, enabling a truly personalized AI experience for each user.
