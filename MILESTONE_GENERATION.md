# Milestone Generation - Stateless LLM Endpoints

## Overview

The Milestone Generation system provides two stateless endpoints that generate project milestones with embedded tasks. These endpoints follow the **pure function model**: JSON in → JSON out, with no database operations.

**Key Principle:** The backend receives a complete project definition from the frontend, applies prompts, and returns structured milestones with tasks. The frontend decides how to store and display them.

---

## Architecture

```
Frontend → Backend (stateless LLM) → Frontend
   |            |                        |
   |            |                        |
Project      Generate                Store in
Definition   Milestones              Projects UI
             with Tasks              & Tasks UI
```

**Separation of Concerns:**
- **Projects UI:** Displays milestones as project phases
- **Tasks UI:** Displays individual tasks (milestone context hidden)
- **Backend:** Stateless generator, no knowledge of UI or storage

---

## Endpoints

### 1. Generate Project Milestones

**Endpoint:** `POST /ai/projects/generate-milestones`

**Purpose:** Generate a complete project plan with milestones and embedded tasks.

**Request:**
```json
{
  "user": {
    "id": "user-uuid",
    "name": "Optional User Name"
  },
  "project": {
    "title": "Build a Sentiment Analysis Model",
    "description": "Create a machine learning model to classify movie reviews as positive or negative.",
    "project_type": "application",
    "technologies": ["Python", "scikit-learn", "pandas"],
    "learning_objectives": [
      "Understand sentiment analysis",
      "Practice text preprocessing",
      "Evaluate ML models"
    ],
    "prerequisites": [
      "Basic Python",
      "Intro to machine learning"
    ],
    "deliverables": [
      "Trained model",
      "Jupyter notebook",
      "GitHub repository",
      "Short write-up"
    ]
  },
  "constraints": {
    "maxMilestones": 6,
    "maxTasksPerMilestone": 6,
    "targetDifficulty": "intermediate"
  }
}
```

**Response:**
```json
{
  "milestones": [
    {
      "title": "Set up project environment",
      "description": "Create repository, environment, and base notebook.",
      "estimated_hours": 2,
      "due_offset_days": 2,
      "status": "pending",
      "tasks": [
        {
          "title": "Create GitHub repository",
          "description": "Initialize repository with README and .gitignore for Python projects.",
          "estimated_minutes": 30,
          "priority": "high",
          "tags": ["setup", "github"]
        },
        {
          "title": "Set up Python environment",
          "description": "Create virtualenv or conda environment and install scikit-learn, pandas, numpy, and jupyter.",
          "estimated_minutes": 60,
          "priority": "high",
          "tags": ["setup", "environment"]
        },
        {
          "title": "Create base Jupyter notebook",
          "description": "Set up notebook structure with sections for data loading, preprocessing, modeling, and evaluation.",
          "estimated_minutes": 30,
          "priority": "medium",
          "tags": ["setup", "notebook"]
        }
      ],
      "resources": [
        "https://docs.github.com/en/get-started/quickstart/create-a-repo",
        "https://scikit-learn.org/stable/install.html"
      ],
      "success_criteria": [
        "Repository created and pushed to GitHub",
        "Environment activates without errors",
        "Base notebook runs successfully"
      ]
    }
  ]
}
```

**Field Descriptions:**

**Milestone Fields:**
- `title` - Short, descriptive name for the milestone
- `description` - What this milestone accomplishes
- `estimated_hours` - Total time estimate for the milestone
- `due_offset_days` - Days from project start (e.g., 2, 5, 10, 15)
- `status` - Always "pending" from backend (frontend manages state)
- `tasks` - Array of embedded tasks (2-6 tasks per milestone)
- `resources` - Helpful links (docs, tutorials, examples)
- `success_criteria` - Measurable outcomes to verify completion

**Task Fields:**
- `title` - Action-oriented task name (starts with verb)
- `description` - What to do and what the deliverable is
- `estimated_minutes` - Realistic time estimate (15-120 minutes)
- `priority` - "low", "medium", or "high"
- `tags` - Categories for filtering (e.g., "setup", "coding", "testing")

---

### 2. Refine Tasks for Single Milestone

**Endpoint:** `POST /ai/projects/milestone-to-tasks`

**Purpose:** Re-generate or refine tasks for a specific milestone.

**Request:**
```json
{
  "milestone": {
    "title": "Model training",
    "description": "Train and evaluate a baseline sentiment classifier."
  },
  "project": {
    "title": "Build a Sentiment Analysis Model",
    "technologies": ["Python", "scikit-learn"]
  },
  "constraints": {
    "maxTasks": 6
  }
}
```

**Response:**
```json
{
  "tasks": [
    {
      "title": "Implement TF-IDF vectorization",
      "description": "Convert preprocessed text to TF-IDF feature vectors using scikit-learn's TfidfVectorizer.",
      "estimated_minutes": 60,
      "priority": "high",
      "tags": ["features", "coding"]
    },
    {
      "title": "Split data into train/test sets",
      "description": "Create 80/20 train/test split with stratification to preserve class balance.",
      "estimated_minutes": 30,
      "priority": "high",
      "tags": ["data", "modeling"]
    },
    {
      "title": "Train baseline logistic regression",
      "description": "Train a simple logistic regression model using TF-IDF features as baseline classifier.",
      "estimated_minutes": 45,
      "priority": "high",
      "tags": ["modeling", "training"]
    }
  ]
}
```

**Use Cases:**
- User wants more/fewer tasks for a milestone
- User wants different task breakdown
- Milestone scope changed and needs new tasks

---

## Implementation Details

### Files Created

1. **`internal/llm/types_milestone_generation.go`**
   - `GenerateMilestonesRequest/Response`
   - `MilestoneToTasksRequest/Response`
   - `MilestoneWithTasks` - Milestone with embedded tasks
   - `EmbeddedTask` - Task structure

2. **`internal/llm/milestone_generation_handlers.go`**
   - `GenerateMilestonesHandler` - Generates full project plan
   - `MilestoneToTasksHandler` - Generates tasks for one milestone
   - Both use `handleLLMRequest` pattern

3. **`configs/prompts/generateMilestonesPrompt.txt`**
   - Comprehensive prompt for milestone generation
   - Detailed example with 6 milestones
   - Design principles and rules

4. **`configs/prompts/milestoneToTasksPrompt.txt`**
   - Focused prompt for task generation
   - Multiple examples for different contexts
   - Task design principles

5. **`internal/llm/routers.go`** (modified)
   - Registered `/ai/projects/generate-milestones`
   - Registered `/ai/projects/milestone-to-tasks`

6. **`internal/llm/configs/deepseek_configs.go`** (modified)
   - Added configs with appropriate token limits (4096, 2048)

---

## LLM Prompt Design

### System Prompt (Generate Milestones)

```text
You are an AI project planner for a learning platform.

Your job:
- Given a project description and constraints, design a small set of clear milestones.
- Each milestone has a few concrete, actionable tasks.
- In the Projects UI, these are called "milestones".
- In the Tasks UI, these tasks are just tasks.

Requirements:
- Output strictly valid JSON. Do not include any explanation or markdown.
- Use simple English and keep everything scoped to what a single learner can realistically do.
- Prefer 3–6 milestones and 2–6 tasks per milestone unless constraints say otherwise.
```

### User Prompt (Dynamic)

Built from request data:

```text
User: {{user.name}} (id={{user.id}})

Project description:
- Title: {{project.title}}
- Type: {{project.project_type}}
- Description: {{project.description}}

Technologies: {{project.technologies}}
Learning objectives: {{project.learning_objectives}}
Prerequisites: {{project.prerequisites}}
Expected deliverables: {{project.deliverables}}

Constraints:
- Max milestones: {{constraints.maxMilestones}}
- Max tasks per milestone: {{constraints.maxTasksPerMilestone}}
- Target difficulty: {{constraints.targetDifficulty}}

Design a milestone plan and return ONLY the JSON object.
```

---

## Frontend Integration

### Generate Milestones

```javascript
import { apiCall } from './apiClient';

const generateProjectMilestones = async (projectData) => {
  const response = await apiCall('/ai/projects/generate-milestones', {
    user: {
      id: currentUser.id,
      name: currentUser.name
    },
    project: {
      title: projectData.title,
      description: projectData.description,
      project_type: projectData.type,
      technologies: projectData.technologies,
      learning_objectives: projectData.objectives,
      prerequisites: projectData.prerequisites,
      deliverables: projectData.deliverables
    },
    constraints: {
      maxMilestones: 6,
      maxTasksPerMilestone: 6,
      targetDifficulty: projectData.difficulty
    }
  });

  // Frontend stores milestones in Projects UI
  for (const milestone of response.milestones) {
    const milestoneId = await saveMilestone({
      projectId: currentProject.id,
      title: milestone.title,
      description: milestone.description,
      estimatedHours: milestone.estimated_hours,
      dueDate: calculateDueDate(milestone.due_offset_days),
      status: milestone.status,
      resources: milestone.resources,
      successCriteria: milestone.success_criteria
    });

    // Frontend stores tasks in Tasks UI
    for (const task of milestone.tasks) {
      await saveTask({
        milestoneId: milestoneId,
        projectId: currentProject.id,
        title: task.title,
        description: task.description,
        estimatedMinutes: task.estimated_minutes,
        priority: task.priority,
        tags: task.tags,
        status: 'pending'
      });
    }
  }

  return response;
};
```

### Refine Milestone Tasks

```javascript
const refineMilestoneTasks = async (milestone) => {
  const response = await apiCall('/ai/projects/milestone-to-tasks', {
    milestone: {
      title: milestone.title,
      description: milestone.description
    },
    project: {
      title: currentProject.title,
      technologies: currentProject.technologies
    },
    constraints: {
      maxTasks: 6
    }
  });

  // Replace existing tasks with new ones
  await deleteTasksForMilestone(milestone.id);
  
  for (const task of response.tasks) {
    await saveTask({
      milestoneId: milestone.id,
      projectId: currentProject.id,
      ...task,
      status: 'pending'
    });
  }

  return response.tasks;
};
```

---

## UI Separation

### Projects UI
- Displays milestones as project phases
- Shows milestone title, description, progress
- Shows resources and success criteria
- Allows expanding to see tasks
- Tracks milestone completion

### Tasks UI
- Displays individual tasks from all projects
- No mention of "milestones" in the UI
- Tasks are filterable by tags, priority, project
- Users check off tasks as they complete them
- Time tracking per task

**Backend doesn't care about this separation** - it just generates the data structure.

---

## Testing

### Test Generate Milestones

```bash
curl -X POST http://localhost:8080/ai/projects/generate-milestones \
  -H "Content-Type: application/json" \
  -d '{
    "user": {
      "id": "test-user-123",
      "name": "Test User"
    },
    "project": {
      "title": "Build a Sentiment Analysis Model",
      "description": "Create a machine learning model to classify movie reviews.",
      "project_type": "application",
      "technologies": ["Python", "scikit-learn", "pandas"],
      "learning_objectives": [
        "Understand sentiment analysis",
        "Practice text preprocessing"
      ],
      "prerequisites": ["Basic Python"],
      "deliverables": ["Trained model", "Jupyter notebook"]
    },
    "constraints": {
      "maxMilestones": 6,
      "maxTasksPerMilestone": 6,
      "targetDifficulty": "intermediate"
    }
  }'
```

### Test Milestone to Tasks

```bash
curl -X POST http://localhost:8080/ai/projects/milestone-to-tasks \
  -H "Content-Type: application/json" \
  -d '{
    "milestone": {
      "title": "Model training",
      "description": "Train and evaluate a baseline sentiment classifier."
    },
    "project": {
      "title": "Build a Sentiment Analysis Model",
      "technologies": ["Python", "scikit-learn"]
    },
    "constraints": {
      "maxTasks": 6
    }
  }'
```

---

## Configuration

### DeepSeek Configs

```go
"/ai/projects/generate-milestones": {
    ModelId:     "deepseek-chat",
    Temperature: 0.7,
    MaxTokens:   4096,  // Higher for complex project plans with multiple milestones
    ClientName:  "general",
}

"/ai/projects/milestone-to-tasks": {
    ModelId:     "deepseek-chat",
    Temperature: 0.6,  // Lower for more consistent task generation
    MaxTokens:   2048,
    ClientName:  "general",
}
```

---

## Design Principles

### Milestone Design
- **3-6 milestones** - Enough structure, not overwhelming
- **Logical progression** - Setup → Core work → Polish
- **Clear deliverables** - Each milestone has measurable outcomes
- **Realistic timelines** - `due_offset_days` distributes work evenly
- **Resource links** - Helpful documentation and tutorials

### Task Design
- **Atomic tasks** - One clear deliverable per task
- **15-120 minutes** - Small enough to complete in one session
- **Action-oriented** - Titles start with verbs
- **Specific** - "Implement TF-IDF vectorization" not "Work on features"
- **Prioritized** - Foundation = high, polish = low
- **Tagged** - Easy filtering in Tasks UI

---

## Error Handling

Both handlers include fallback responses:

```go
// If LLM response parsing fails, return empty array
if err := json.Unmarshal(response, &milestonesResp); err != nil {
    milestonesResp = GenerateMilestonesResponse{
        Milestones: []MilestoneWithTasks{},
    }
}
```

Frontend should handle:
- Empty milestones/tasks arrays
- Missing optional fields (resources, success_criteria)
- Network errors
- Invalid JSON (though backend validates)

---

## Benefits

✅ **Stateless** - No database coupling, pure functions  
✅ **Flexible** - Frontend controls storage and display  
✅ **Scalable** - Easy to cache and horizontally scale  
✅ **Testable** - Simple input/output testing  
✅ **UI-agnostic** - Backend doesn't know about Projects vs Tasks UI  
✅ **Comprehensive** - Generates complete project structure  
✅ **Realistic** - Time estimates and difficulty-appropriate tasks  

---

## Summary

The Milestone Generation system provides:

1. **Generate Milestones** - Complete project plan with embedded tasks
2. **Refine Tasks** - Re-generate tasks for specific milestone

Both endpoints:
- ✅ Follow stateless model (JSON in → JSON out)
- ✅ Generate realistic, actionable plans
- ✅ Respect constraints (max milestones, max tasks, difficulty)
- ✅ Include helpful metadata (resources, success criteria, tags)
- ✅ Return structured JSON for easy frontend integration
- ✅ Support separation between Projects UI and Tasks UI

The backend remains stateless while providing intelligent project planning capabilities.
