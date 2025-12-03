# Generate Tasks API - Milestone Task Generation

## Overview

The `/generate-tasks` endpoint breaks a single milestone into actionable tasks. It's designed to support the frontend's "Generate Tasks" button on milestone cards.

**Purpose:** Given a milestone title, description, and project context, generate 3-7 concrete tasks that a learner can complete.

---

## Endpoint

**URL:** `POST /generate-tasks`

**CORS:** Must allow frontend origin (e.g., `http://localhost:3000`)

---

## Request Contract

```json
{
  "milestoneId": "milestone-uuid-or-string",
  "milestoneTitle": "Write blog post",
  "milestoneDescription": "Draft and publish a blog post about the project",
  "projectContext": "Project description or short context string",
  "taskCount": 5
}
```

### Field Descriptions

- `milestoneId` (optional) - Milestone identifier for logging/correlation
- `milestoneTitle` (required) - Milestone name
- `milestoneDescription` (optional) - What the milestone accomplishes
- `projectContext` (optional) - Free-form project description
- `taskCount` (optional) - Desired number of tasks (3-10, defaults to 5)

---

## Response Contract

```json
{
  "tasks": [
    {
      "title": "Outline the blog post",
      "description": "Create a bullet-point outline with intro, body, and conclusion.",
      "estimated_minutes": 45,
      "priority": "medium",
      "tags": ["writing", "planning"]
    },
    {
      "title": "Write the introduction",
      "description": "Draft an engaging introduction that hooks the reader and explains the project's purpose.",
      "estimated_minutes": 60,
      "priority": "high",
      "tags": ["writing", "content"]
    },
    {
      "title": "Draft the main content",
      "description": "Write the body sections covering technical details, challenges, and solutions.",
      "estimated_minutes": 90,
      "priority": "high",
      "tags": ["writing", "content"]
    },
    {
      "title": "Add code examples and screenshots",
      "description": "Include relevant code snippets and visual examples to illustrate key points.",
      "estimated_minutes": 60,
      "priority": "medium",
      "tags": ["documentation", "visuals"]
    },
    {
      "title": "Edit and proofread",
      "description": "Review for clarity, grammar, and technical accuracy. Fix any errors.",
      "estimated_minutes": 45,
      "priority": "low",
      "tags": ["editing", "quality"]
    }
  ]
}
```

### Task Fields

- `title` (required) - Short, action-oriented task name
- `description` (required) - What to do and what the deliverable is
- `estimated_minutes` (required) - Time estimate (15-120 minutes)
- `priority` (optional) - "low", "medium", "high", or "critical"
- `tags` (optional) - Array of categories for filtering

---

## Frontend Integration

From `ProjectGenerationService.js`:

```javascript
async generateTasks(milestoneId, milestoneTitle, milestoneDescription, projectContext, taskCount = 5) {
  try {
    const response = await fetch(`${API_BASE_URL}/generate-tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        milestoneId,
        milestoneTitle,
        milestoneDescription,
        projectContext,
        taskCount
      })
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    
    return {
      success: true,
      tasks: data.tasks || []
    };
  } catch (error) {
    console.error('Error generating tasks:', error);
    return {
      success: false,
      error: error.message,
      tasks: []
    };
  }
}
```

The frontend then:
1. Calls this method when "Generate Tasks" button is clicked
2. Receives the tasks array
3. Inserts tasks into `ml_tasks_staging` table
4. Updates the milestone checklist UI

---

## Backend Implementation

### Handler Flow

```go
func GenerateTasksHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Decode request
    var req GenerateTasksRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 2. Validate
    if req.MilestoneTitle == "" {
        return error
    }
    
    // 3. Ensure safe task count (3-10)
    if req.TaskCount <= 0 || req.TaskCount > 10 {
        req.TaskCount = 5
    }
    
    // 4. Call LLM
    response := handleLLMRequest(w, req, "generateTasksPrompt", "/generate-tasks")
    
    // 5. Parse
    var tasksResp GenerateTasksResponse
    json.Unmarshal(response, &tasksResp)
    
    // 6. Normalize tasks
    for each task {
        - Ensure title is present
        - Default estimated_minutes to 45 if missing
        - Validate priority (default to "medium")
        - Ensure tags is array (default to [])
    }
    
    // 7. Return
    json.Marshal(normalizedTasks)
}
```

### Prompt Construction

**System Prompt:**
```text
You are an assistant that breaks project milestones into small, actionable tasks.

Given:
- A milestone (title + description)
- A short project context
- A desired number of tasks

You must return ONLY JSON with this shape:
{
  "tasks": [
    {
      "title": string,
      "description": string,
      "estimated_minutes": integer,
      "priority": "low" | "medium" | "high" | "critical",
      "tags": string[]
    }
  ]
}

Rules:
- Output STRICTLY valid JSON. No markdown, no backticks, no commentary.
- 3–7 tasks is typical; do not exceed taskCount.
- Keep tasks small and concrete (30–90 minutes each).
```

**User Prompt:**
```text
Milestone:
- ID: {milestoneId}
- Title: {milestoneTitle}
- Description: {milestoneDescription}

Project context:
{projectContext}

Target number of tasks: {taskCount}

Please return ONLY the JSON object described in the system message.
```

---

## Testing

### Example Request

```bash
curl -X POST http://localhost:8080/generate-tasks \
  -H "Content-Type: application/json" \
  -d '{
    "milestoneId": "milestone-123",
    "milestoneTitle": "Write blog post",
    "milestoneDescription": "Draft and publish a blog post about the project",
    "projectContext": "Building a sentiment analysis web app with Python and FastAPI",
    "taskCount": 5
  }'
```

### Expected Response

```json
{
  "tasks": [
    {
      "title": "Outline the blog post",
      "description": "Create a bullet-point outline with intro, body, and conclusion.",
      "estimated_minutes": 45,
      "priority": "medium",
      "tags": ["writing", "planning"]
    },
    {
      "title": "Write the introduction",
      "description": "Draft an engaging introduction that hooks the reader.",
      "estimated_minutes": 60,
      "priority": "high",
      "tags": ["writing", "content"]
    },
    {
      "title": "Draft the main content",
      "description": "Write body sections covering technical details and challenges.",
      "estimated_minutes": 90,
      "priority": "high",
      "tags": ["writing", "content"]
    },
    {
      "title": "Add code examples",
      "description": "Include relevant code snippets to illustrate key points.",
      "estimated_minutes": 60,
      "priority": "medium",
      "tags": ["documentation", "code"]
    },
    {
      "title": "Edit and proofread",
      "description": "Review for clarity, grammar, and technical accuracy.",
      "estimated_minutes": 45,
      "priority": "low",
      "tags": ["editing", "quality"]
    }
  ]
}
```

---

## Error Handling

### Missing Required Field

**Request:**
```json
{
  "milestoneId": "123",
  "taskCount": 5
}
```

**Response:** `400 Bad Request`
```json
{
  "error": "milestoneTitle is required"
}
```

### LLM Call Failed

**Response:** `502 Bad Gateway`
```json
{
  "error": "LLM call failed"
}
```

### JSON Parse Error

**Response:** `500 Internal Server Error`
```json
{
  "error": "Failed to parse LLM JSON response"
}
```

---

## CORS Configuration

The backend must allow requests from the frontend origin:

```go
// In main.go or middleware
func enableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

Without CORS, the frontend will see:
```
TypeError: Failed to fetch
```

---

## Configuration

### DeepSeek Config

```go
"/generate-tasks": {
    ModelId:     "deepseek-chat",
    Temperature: 0.6,  // Lower for more consistent task generation
    MaxTokens:   1536, // Sufficient for 3-7 tasks
    ClientName:  "general",
}
```

---

## Key Features

✅ **Milestone-specific** - Generates tasks for a single milestone  
✅ **Configurable count** - Respects taskCount parameter (3-10 range)  
✅ **Normalized output** - Ensures all required fields are present  
✅ **Safe defaults** - Falls back to reasonable values if LLM omits fields  
✅ **Priority levels** - Supports low/medium/high/critical priorities  
✅ **Tagging** - Categorizes tasks for filtering in UI  
✅ **Time estimates** - Realistic 15-120 minute estimates  

---

## Frontend Flow

1. User clicks "Generate Tasks" on milestone card
2. Frontend calls `ProjectGenerationService.generateTasks()`
3. Backend generates tasks via LLM
4. Frontend receives `{ tasks: [...] }`
5. Frontend inserts tasks into `ml_tasks_staging`
6. Milestone checklist updates with new tasks

---

## Summary

The `/generate-tasks` endpoint:

1. ✅ Receives milestone details from frontend
2. ✅ Generates 3-7 actionable tasks via LLM
3. ✅ Returns normalized JSON with required fields
4. ✅ Supports priority levels and tags
5. ✅ Provides realistic time estimates
6. ✅ Integrates with frontend's task staging system

This enables the "Generate Tasks" button to work seamlessly, creating tasks that learners can immediately start working on.
