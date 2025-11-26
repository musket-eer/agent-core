# Orchestrator Agent Implementation

## Overview

The Orchestrator Agent is the central command interpreter for the AI University platform. It allows the frontend to control the entire platform through natural language by translating user instructions into structured commands.

**Key Principle:** The orchestrator is a **planner, not an executor**. It emits high-level commands; the frontend executes them.

---

## Architecture

```
User Message → Frontend → POST /ai/orchestrator → LLM → Commands → Frontend Execution
```

1. User sends natural language instruction
2. Frontend calls `/ai/orchestrator` with message + app context
3. Backend LLM interprets intent and generates commands
4. Frontend receives `reply` + `commands`
5. Frontend displays reply and executes commands using existing services

---

## Endpoint

### `POST /ai/orchestrator`

**Request:**
```json
{
  "userId": "user-uuid",
  "sessionId": "optional-session-id",
  "message": "Create a task to study linear algebra tomorrow and show my completed tasks",
  "appContext": {
    "currentPage": "tasks",
    "selectedTaskId": "task-123",
    "filters": {
      "status": "todo",
      "priority": "high"
    },
    "recentTasks": [
      { "id": "task-1", "title": "Study calculus", "status": "completed" },
      { "id": "task-2", "title": "Watch ML lecture", "status": "todo" }
    ]
  }
}
```

**Response:**
```json
{
  "reply": "I've created a linear algebra study task for tomorrow and will show your completed tasks.",
  "commands": [
    {
      "type": "create_task",
      "payload": {
        "title": "Study linear algebra",
        "when": "2025-11-26T10:00:00Z",
        "priority": "medium"
      }
    },
    {
      "type": "navigate",
      "payload": {
        "view": "completed_tasks"
      }
    }
  ],
  "meta": {
    "sessionId": "orchestrator-session-xyz",
    "usedAgents": ["TasksAgent"]
  }
}
```

---

## Command Schema

### Task Commands

#### 1. `create_task`
Create a new task/activity.

**Payload:**
```typescript
{
  title: string;           // required
  when?: string;           // ISO datetime
  priority?: 'low' | 'medium' | 'high' | 'critical';
  projectId?: string | null;
  courseId?: string | null;
  description?: string;
}
```

**Example:**
```json
{
  "type": "create_task",
  "payload": {
    "title": "Study calculus",
    "when": "2025-11-26T14:00:00Z",
    "priority": "high",
    "description": "Review chapters 1-3"
  }
}
```

#### 2. `complete_task`
Mark a task as completed.

**Payload:**
```typescript
{
  taskId: string;
}
```

**Example:**
```json
{
  "type": "complete_task",
  "payload": {
    "taskId": "task-123"
  }
}
```

#### 3. `update_task_notes`
Add or update notes on a task.

**Payload:**
```typescript
{
  taskId: string;
  notes: string;
}
```

**Example:**
```json
{
  "type": "update_task_notes",
  "payload": {
    "taskId": "task-456",
    "notes": "Need to review chapter 3 before the exam"
  }
}
```

#### 4. `generate_task_reflection`
Generate AI reflection for a completed task.

**Payload:**
```typescript
{
  taskId: string;
  userReflection?: string | null;
}
```

**Example:**
```json
{
  "type": "generate_task_reflection",
  "payload": {
    "taskId": "task-789",
    "userReflection": "Task went well, learned a lot"
  }
}
```

### Navigation Commands

#### 5. `navigate`
Change the current view/page.

**Payload:**
```typescript
{
  view: 'tasks' | 'completed_tasks' | 'courses' | 'dashboard';
}
```

**Example:**
```json
{
  "type": "navigate",
  "payload": {
    "view": "completed_tasks"
  }
}
```

#### 6. `focus_task`
Focus/highlight a specific task in the UI.

**Payload:**
```typescript
{
  taskId: string;
}
```

**Example:**
```json
{
  "type": "focus_task",
  "payload": {
    "taskId": "task-999"
  }
}
```

---

## Implementation Details

### Files Created

1. **`internal/llm/types_orchestrator.go`**
   - Request/response types
   - Command schema types
   - Command type constants

2. **`internal/llm/orchestrator_handler.go`**
   - HTTP handler for `/ai/orchestrator`
   - Command validation logic
   - LLM integration

3. **`configs/prompts/orchestratorPrompt.txt`**
   - Comprehensive system prompt
   - Command schema documentation
   - Examples and guidelines

4. **`internal/llm/routers.go`** (modified)
   - Registered `/ai/orchestrator` route

5. **`internal/llm/configs/deepseek_configs.go`** (modified)
   - Added DeepSeek config for orchestrator

### Handler Flow

```go
1. Decode OrchestratorRequest from JSON
2. Validate userId and message are present
3. Call handleLLMRequest with orchestratorPrompt
4. Parse LLM response into reply + commands
5. Validate commands against schema
6. Return OrchestratorResponse with validated commands
```

### Command Validation

The handler includes `validateCommands()` which:
- Filters out unknown command types
- Only allows commands from the defined schema
- Logs dropped invalid commands
- Prevents malicious or malformed commands from reaching frontend

Valid command types:
- `create_task`
- `complete_task`
- `update_task_notes`
- `generate_task_reflection`
- `navigate`
- `focus_task`

---

## Frontend Integration

### Service Usage

```javascript
import { OrchestratorAgentService } from './services/OrchestratorAgentService';

// Send instruction
const result = await OrchestratorAgentService.sendInstruction({
  userId: currentUser.id,
  message: "Create a task to study math and show my dashboard",
  appContext: {
    currentPage: "tasks",
    selectedTaskId: selectedTask?.id,
    filters: currentFilters
  }
});

// Display reply
showNotification(result.reply);

// Execute commands
for (const command of result.commands) {
  switch (command.type) {
    case 'create_task':
      await createActivity(command.payload);
      break;
    case 'complete_task':
      await toggleActivityCompletion(command.payload.taskId);
      break;
    case 'navigate':
      navigate(`/${command.payload.view}`);
      break;
    case 'focus_task':
      setSelectedTask(command.payload.taskId);
      break;
    // ... handle other commands
  }
}
```

---

## Testing

### Test with cURL

```bash
# Basic task creation
curl -X POST http://localhost:8000/ai/orchestrator \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "test-user",
    "message": "Create a task to study calculus tomorrow at 2pm"
  }'

# Multi-step instruction
curl -X POST http://localhost:8000/ai/orchestrator \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "test-user",
    "message": "Create a high priority task for linear algebra and then show my completed tasks",
    "appContext": {
      "currentPage": "tasks"
    }
  }'

# Context-aware command
curl -X POST http://localhost:8000/ai/orchestrator \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "test-user",
    "message": "Mark this task as done",
    "appContext": {
      "selectedTaskId": "task-123"
    }
  }'

# Navigation
curl -X POST http://localhost:8000/ai/orchestrator \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "test-user",
    "message": "Show me my dashboard"
  }'
```

### Expected Response Format

```json
{
  "reply": "I've created a high priority task for linear algebra and will show your completed tasks.",
  "commands": [
    {
      "type": "create_task",
      "payload": {
        "title": "Linear algebra",
        "priority": "high"
      }
    },
    {
      "type": "navigate",
      "payload": {
        "view": "completed_tasks"
      }
    }
  ],
  "meta": {
    "sessionId": "orchestrator-session",
    "usedAgents": []
  }
}
```

---

## Security & Guardrails

### Backend Validation
- ✅ Command type validation (only known types allowed)
- ✅ Payload structure validation
- ✅ No direct database access from orchestrator
- ✅ Logging of dropped invalid commands

### Frontend Safeguards
- Frontend should validate task IDs exist before executing commands
- Destructive actions should require user confirmation
- Rate limiting on orchestrator calls
- Sanitize user input before sending to backend

### What the Orchestrator CANNOT Do
- ❌ Directly modify database
- ❌ Execute commands itself
- ❌ Access external APIs
- ❌ Perform authentication/authorization
- ❌ Make up task IDs or other identifiers

---

## Extending the System

### Adding New Commands

1. **Define the command type in `types_orchestrator.go`:**
```go
const (
    // ... existing commands
    CommandDeleteTask = "delete_task"
)

type DeleteTaskPayload struct {
    TaskID string `json:"taskId"`
}
```

2. **Update `isValidCommand()` in `orchestrator_handler.go`:**
```go
validTypes := map[string]bool{
    // ... existing types
    CommandDeleteTask: true,
}
```

3. **Document in `orchestratorPrompt.txt`:**
```
7. delete_task
   Purpose: Delete a task permanently
   Payload: { "taskId": string }
   Example: User says "Delete task-123"
   Command: { "type": "delete_task", "payload": { "taskId": "task-123" } }
```

4. **Implement frontend handler:**
```javascript
case 'delete_task':
  if (confirm('Are you sure you want to delete this task?')) {
    await deleteActivity(command.payload.taskId);
  }
  break;
```

---

## Prompt Engineering Tips

The orchestrator prompt is designed to:
- **Be explicit about constraints** (planner not executor)
- **Provide concrete examples** for each command type
- **Use the app context** for contextual awareness
- **Validate output format** (strict JSON schema)
- **Handle edge cases** (ambiguous requests, missing IDs)

Key sections:
- Command schema with examples
- Context usage guidelines
- Reply formatting rules
- Multi-step instruction handling
- Error handling strategies

---

## Troubleshooting

### Issue: Commands not being generated
- Check LLM response parsing in logs
- Verify prompt template is loaded correctly
- Ensure JSON mode is working in LLM client

### Issue: Commands being dropped
- Check validation logs for dropped command types
- Verify command type matches schema exactly (case-sensitive)
- Ensure payload structure is correct

### Issue: Reply contains technical details
- Update prompt to emphasize user-friendly language
- Add more examples of good vs bad replies
- Adjust temperature if responses are too technical

### Issue: Hallucinated task IDs
- Emphasize in prompt: only use IDs from appContext
- Add validation in frontend before executing commands
- Log and drop commands with invalid IDs

---

## Performance Considerations

- **Caching:** Orchestrator responses are cached by payload (via existing cache layer)
- **Token usage:** 2048 max tokens allows for complex multi-command responses
- **Temperature:** 0.7 balances creativity with consistency
- **Latency:** Typical response time 1-3 seconds depending on LLM

---

## Future Enhancements

### Potential New Commands
- `create_course` - Enroll in or create a course
- `schedule_task` - Add task to calendar
- `set_reminder` - Create a reminder for a task
- `bulk_complete` - Complete multiple tasks at once
- `filter_tasks` - Apply filters to task view
- `export_data` - Export tasks or progress data

### Advanced Features
- **Session memory:** Track conversation history for context
- **Undo/redo:** Reverse previous commands
- **Batch operations:** Execute multiple similar commands efficiently
- **Conditional commands:** "If X then Y" logic
- **Scheduled commands:** Execute commands at specific times

---

## Summary

The Orchestrator Agent provides a natural language interface to the entire platform by:
1. ✅ Interpreting user intent from natural language
2. ✅ Generating structured commands from a defined schema
3. ✅ Validating commands before sending to frontend
4. ✅ Providing clear, user-friendly replies
5. ✅ Maintaining security through validation and constraints

The frontend remains the "central brain" that executes commands, while the orchestrator acts as a smart translator between natural language and structured actions.
