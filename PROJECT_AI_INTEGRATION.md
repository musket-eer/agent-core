# Project AI Integration - Stateless LLM Agents

## Overview

The Project AI system provides three stateless LLM agents that help users plan, manage, and execute learning projects. These agents follow a **pure function model**: JSON in → JSON out, with no database operations or side effects.

**Key Principle:** The backend is a stateless orchestrator that receives complete context from the frontend, applies prompts, and returns structured JSON responses.

---

## Architecture

```
Frontend (state + storage) → Backend (stateless LLM) → Frontend (applies changes)
```

1. Frontend sends complete context in JSON payload
2. Backend applies persona + prompt to LLM
3. LLM returns structured JSON response
4. Frontend decides how to apply/store the response

**What the backend does:**
- ✅ Receive JSON payload
- ✅ Apply persona-aware prompts
- ✅ Call LLM with full context
- ✅ Return structured JSON

**What the backend does NOT do:**
- ❌ Query database
- ❌ Fetch additional data
- ❌ Persist anything
- ❌ Make assumptions about storage

---

## Three Project AI Agents

### 1. Project Creator (`agentId: "project_creator"`)

**Purpose:** Turn a learning goal into a structured project plan.

**Endpoint:** `POST /projects/create`

**Input:**
```json
{
  "agentId": "project_creator",
  "persona": {
    "bio": "...",
    "instructions": "..."
  },
  "userId": "user-123",
  "goal": "Build a personal spaced-repetition system",
  "timeframeWeeks": 4,
  "difficulty": "intermediate",
  "constraints": ["2 hours/day max", "Prefer Python"],
  "userProfile": { /* optional */ },
  "activeCourses": [ /* optional */ ],
  "recentTasks": [ /* optional */ ]
}
```

**Output:**
```json
{
  "project": {
    "title": "Personal Spaced-Repetition System",
    "description": "Build a CLI spaced repetition app using Python...",
    "status": "planned",
    "priority": "medium",
    "domain": "Software Development",
    "estimatedDurationWeeks": 4
  },
  "milestones": [
    {
      "title": "Research & Algorithm Design",
      "description": "Study spaced repetition algorithms...",
      "milestone_type": "milestone",
      "estimated_hours": 6,
      "due_in_days": 5,
      "children": [
        {
          "title": "Study SM-2 algorithm",
          "description": "Read documentation...",
          "milestone_type": "task_like",
          "estimated_hours": 2
        }
      ]
    }
  ],
  "courses": [
    {
      "courseId": "python-data-structures",
      "role": "supporting"
    }
  ]
}
```

**What it does:**
- Analyzes goal and constraints
- Designs realistic project plan
- Creates 3-7 top-level milestones with optional children
- Suggests relevant courses
- Respects timeframe and difficulty

---

### 2. Project Assistant (`agentId: "project_assistant"`)

**Purpose:** Act as an assistant on existing projects - answer questions and propose changes.

**Endpoint:** `POST /projects/assist`

**Input:**
```json
{
  "agentId": "project_assistant",
  "persona": {
    "bio": "...",
    "instructions": "..."
  },
  "userId": "user-123",
  "message": "I'm stuck on the API design. What should I do next?",
  "requestType": "assist",
  "project": { /* project object */ },
  "milestones": [ /* array of milestones */ ],
  "courses": [ /* optional course links */ ],
  "userProfile": { /* optional */ },
  "recentTasks": [ /* optional */ ],
  "timeLogs": [ /* optional */ ]
}
```

**Request Types:**
- `assist` - General help, answer questions
- `replan` - Suggest adjustments to timeline/milestones
- `prioritize` - Recommend what to focus on next
- `summarize` - Provide project status summary
- `risks` - Identify potential blockers

**Output:**
```json
{
  "assistantMessage": "For the API design, start by defining the core operations: add card, get due cards, submit review...",
  "suggestedChanges": {
    "milestoneUpdates": [
      {
        "milestoneId": "m3",
        "estimated_hours": 6,
        "description": "Create minimal CLI (reduced scope)",
        "notes": "Simplified to meet timeline"
      }
    ],
    "newMilestones": [
      {
        "parentMilestoneId": "m1",
        "title": "Define API operations",
        "description": "List core functions...",
        "milestone_type": "task_like",
        "estimated_hours": 1,
        "due_in_days": 2
      }
    ]
  }
}
```

**What it does:**
- Reads project snapshot
- Answers user questions with context
- Proposes structured changes when appropriate
- References specific milestones/courses

---

### 3. Milestone Translator (`agentId: "milestone_translator"`)

**Purpose:** Break milestones into concrete, actionable tasks.

**Endpoint:** `POST /projects/translate-milestones`

**Input:**
```json
{
  "agentId": "milestone_translator",
  "persona": {
    "bio": "...",
    "instructions": "..."
  },
  "userId": "user-123",
  "projectId": "proj-456",
  "milestones": [
    {
      "id": "m1",
      "title": "Research & Algorithm Design",
      "description": "Study spaced repetition algorithms...",
      "estimated_hours": 6,
      "courseContext": {
        "courseId": "python-advanced",
        "lessonIds": ["lesson-oop", "lesson-sqlite"]
      }
    }
  ],
  "taskSystemConstraints": {
    "maxTasksPerMilestone": 5,
    "maxDurationPerTaskHours": 2
  },
  "userProfile": { /* optional */ },
  "activeCourses": [ /* optional */ ]
}
```

**Output:**
```json
{
  "tasks": [
    {
      "milestoneId": "m1",
      "title": "Read SM-2 algorithm documentation",
      "description": "Study the SuperMemo 2 algorithm, understand the formula...",
      "estimated_minutes": 60,
      "difficulty": "easy",
      "courseId": "python-advanced",
      "lessonId": "lesson-oop"
    },
    {
      "milestoneId": "m1",
      "title": "Design database schema for cards",
      "description": "Create ERD with tables for flashcards...",
      "estimated_minutes": 90,
      "difficulty": "medium"
    }
  ]
}
```

**What it does:**
- Breaks each milestone into 1-N small tasks
- Respects task count and duration constraints
- Aligns tasks with courses/lessons if provided
- Makes tasks atomic and actionable
- Estimates realistic durations

---

## Implementation Details

### Files Created

1. **`internal/llm/types_projects.go`**
   - Request/response types for all three agents
   - Embedded `PersonaRequest` for persona support
   - Implements `GetPersona()` for each request type

2. **`internal/llm/project_handlers.go`**
   - `ProjectCreatorHandler` - Creates project plans
   - `ProjectAssistantHandler` - Provides assistance
   - `MilestoneTranslatorHandler` - Translates to tasks
   - All use existing `handleLLMRequest` pattern

3. **`configs/prompts/projectCreatorPrompt.txt`**
   - Comprehensive prompt for project creation
   - Examples and output format
   - Design principles for milestones

4. **`configs/prompts/projectAssistantPrompt.txt`**
   - Prompt for project assistance
   - Request type handling
   - Change suggestion guidelines

5. **`configs/prompts/milestoneTranslatorPrompt.txt`**
   - Prompt for milestone-to-task translation
   - Task design principles
   - Course alignment logic

6. **`internal/llm/routers.go`** (modified)
   - Registered three new endpoints

7. **`internal/llm/configs/deepseek_configs.go`** (modified)
   - Added configs with appropriate token limits

### Handler Flow

```go
1. Decode request JSON (with agentId and persona)
2. Validate required fields
3. Call handleLLMRequest with appropriate prompt
4. Parse LLM response into typed struct
5. Return JSON response (with fallback on parse error)
```

### Persona Integration

All three agents support persona customization:

```go
type ProjectCreatorRequest struct {
    PersonaRequest  // Embedded
    UserID string
    Goal string
    // ... other fields
}

func (r ProjectCreatorRequest) GetPersona() (string, *Persona) {
    return r.AgentID, r.Persona
}
```

The `handleLLMRequest` function automatically:
1. Extracts persona from request
2. Builds persona-aware system prompt
3. Applies to LLM call

---

## Testing

### Test Project Creator

```bash
curl -X POST http://localhost:8080/projects/create \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "project_creator",
    "userId": "test-user",
    "goal": "Build a personal spaced-repetition system",
    "timeframeWeeks": 4,
    "difficulty": "intermediate",
    "constraints": ["2 hours/day max", "Prefer Python"]
  }'
```

### Test Project Assistant

```bash
curl -X POST http://localhost:8080/projects/assist \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "project_assistant",
    "userId": "test-user",
    "message": "I'\''m stuck on the API design. What should I do next?",
    "requestType": "assist",
    "project": {
      "title": "Personal Spaced-Repetition System",
      "status": "in_progress"
    },
    "milestones": [
      {
        "id": "m1",
        "title": "Core Implementation",
        "status": "in_progress"
      }
    ]
  }'
```

### Test Milestone Translator

```bash
curl -X POST http://localhost:8080/projects/translate-milestones \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "milestone_translator",
    "userId": "test-user",
    "projectId": "proj-123",
    "milestones": [
      {
        "id": "m1",
        "title": "Research & Algorithm Design",
        "description": "Study spaced repetition algorithms",
        "estimated_hours": 6
      }
    ],
    "taskSystemConstraints": {
      "maxTasksPerMilestone": 5,
      "maxDurationPerTaskHours": 2
    }
  }'
```

---

## Frontend Integration

### Creating a Project

```javascript
import { apiCall } from './apiClient';

const createProject = async (goal, timeframe, difficulty, constraints) => {
  const response = await apiCall('/projects/create', {
    agentId: 'project_creator',
    persona: getPersona('project_creator'),
    userId: currentUser.id,
    goal,
    timeframeWeeks: timeframe,
    difficulty,
    constraints,
    userProfile: currentUser.profile,
    activeCourses: currentUser.courses,
    recentTasks: getRecentTasks()
  });

  // Frontend stores the project, milestones, and courses
  await saveProject(response.project);
  await saveMilestones(response.milestones);
  await linkCourses(response.courses);

  return response;
};
```

### Getting Project Assistance

```javascript
const getProjectHelp = async (message, requestType = 'assist') => {
  const response = await apiCall('/projects/assist', {
    agentId: 'project_assistant',
    persona: getPersona('project_assistant'),
    userId: currentUser.id,
    message,
    requestType,
    project: currentProject,
    milestones: currentMilestones,
    courses: linkedCourses,
    recentTasks: getRecentTasks(),
    timeLogs: getTimeLogs()
  });

  // Display assistant message
  showMessage(response.assistantMessage);

  // Optionally apply suggested changes
  if (response.suggestedChanges) {
    if (confirm('Apply suggested changes?')) {
      await applyMilestoneUpdates(response.suggestedChanges.milestoneUpdates);
      await createNewMilestones(response.suggestedChanges.newMilestones);
    }
  }

  return response;
};
```

### Translating Milestones to Tasks

```javascript
const translateMilestonesToTasks = async (milestones) => {
  const response = await apiCall('/projects/translate-milestones', {
    agentId: 'milestone_translator',
    persona: getPersona('milestone_translator'),
    userId: currentUser.id,
    projectId: currentProject.id,
    milestones: milestones.map(m => ({
      id: m.id,
      title: m.title,
      description: m.description,
      estimated_hours: m.estimatedHours,
      courseContext: m.courseContext
    })),
    taskSystemConstraints: {
      maxTasksPerMilestone: 5,
      maxDurationPerTaskHours: 2
    },
    activeCourses: currentUser.courses
  });

  // Frontend creates tasks in its system
  for (const task of response.tasks) {
    await createTask({
      ...task,
      projectId: currentProject.id,
      milestoneId: task.milestoneId
    });
  }

  return response.tasks;
};
```

---

## Configuration

### DeepSeek Configs

```go
"/projects/create": {
    ModelId:     "deepseek-chat",
    Temperature: 0.7,
    MaxTokens:   3072,  // Higher for complex project plans
    ClientName:  "general",
}

"/projects/assist": {
    ModelId:     "deepseek-chat",
    Temperature: 0.7,
    MaxTokens:   2048,
    ClientName:  "general",
}

"/projects/translate-milestones": {
    ModelId:     "deepseek-chat",
    Temperature: 0.6,  // Lower for more consistent task generation
    MaxTokens:   2048,
    ClientName:  "general",
}
```

---

## Stateless Design Benefits

✅ **No database coupling** - Backend doesn't need project schema knowledge  
✅ **Frontend control** - Frontend decides what to store and how  
✅ **Scalable** - Pure functions scale horizontally  
✅ **Testable** - Easy to test with mock inputs  
✅ **Flexible** - Frontend can evolve schema without backend changes  
✅ **Cacheable** - Responses can be cached by payload  
✅ **Portable** - Can swap LLM providers easily  

---

## Error Handling

All handlers include fallback responses:

```go
// If LLM response parsing fails, return minimal valid response
if err := json.Unmarshal(response, &creatorResp); err != nil {
    creatorResp = ProjectCreatorResponse{
        Project: ProjectPlan{
            Title: "Project Plan",
            Description: "Unable to generate detailed plan",
            Status: "planned",
            Priority: "medium",
        },
        Milestones: []MilestonePlan{},
        Courses: []CourseLink{},
    }
}
```

Frontend should handle:
- Empty milestones/tasks arrays
- Missing optional fields
- Network errors
- Invalid JSON (though backend validates)

---

## Future Enhancements

### Potential Features
- **Project Templates** - Pre-built project structures
- **Progress Tracking** - Analyze completion patterns
- **Risk Assessment** - Identify blockers proactively
- **Resource Recommendations** - Suggest articles, videos, tools
- **Collaboration** - Multi-user project planning
- **Time Estimation Learning** - Improve estimates based on history

### Advanced Capabilities
- **Dependency Analysis** - Identify milestone dependencies
- **Critical Path** - Calculate project critical path
- **What-If Scenarios** - Simulate timeline changes
- **Skill Gap Analysis** - Identify missing prerequisites
- **Adaptive Planning** - Adjust plans based on progress

---

## Summary

The Project AI system provides three stateless agents:

1. **Project Creator** - Turns goals into structured plans
2. **Project Assistant** - Provides guidance and suggests changes
3. **Milestone Translator** - Breaks milestones into tasks

All agents:
- ✅ Follow pure function model (JSON in → JSON out)
- ✅ Support persona customization
- ✅ Respect constraints and context
- ✅ Return structured, validated JSON
- ✅ Include comprehensive prompts with examples
- ✅ Integrate with existing LLM infrastructure

The backend remains stateless while providing intelligent project planning capabilities.
