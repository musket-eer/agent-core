# Generate Project API - Simplified Implementation

## Overview

Stateless LLM endpoint that generates project plans with milestones and string array tasks. Designed for direct integration with frontend's `projectService.createProject()`.

---

## Endpoint

**URL:** `POST /generate-project`

---

## Request

```json
{
  "user": {
    "id": "user-uuid",
    "name": "Optional Name"
  },
  "project": {
    "title": "Build a Sentiment Analysis App",
    "description": "Short description from form",
    "project_type": "application",
    "technologies": ["Python", "FastAPI"],
    "learning_objectives": ["Practice NLP", "Deploy a web app"],
    "prerequisites": ["Basic Python"],
    "deliverables": ["GitHub repo", "Deployed app URL"]
  },
  "course_context": {
    "courseId": "ml-course-id-or-null",
    "unitId": "unit-id-or-null",
    "courseTitle": "Intro to ML",
    "courseDescription": "What the course covers",
    "unitTitle": "NLP unit",
    "unitObjective": "Apply ML to text data",
    "topicsCovered": ["NLP", "logistic regression"]
  },
  "constraints": {
    "maxMilestones": 6,
    "maxTasksPerMilestone": 6,
    "estimatedHours": 20,
    "targetDifficulty": "intermediate"
  }
}
```

---

## Response

```json
{
  "title": "Build a Sentiment Analysis App",
  "description": "You will build and deploy a sentiment classifier...",
  "project_type": "application",
  "estimated_hours": 20,
  "learning_objectives": [
    "Understand sentiment analysis",
    "Practice text preprocessing"
  ],
  "prerequisites": [
    "Basic Python",
    "Intro to machine learning"
  ],
  "technologies": ["Python", "FastAPI", "scikit-learn"],
  "deliverables": [
    "GitHub repository",
    "Deployed web app URL"
  ],
  "milestones": [
    {
      "title": "Set up project",
      "description": "Create repo, environment, and base app.",
      "estimated_hours": 2,
      "status": "not_started",
      "tasks": [
        "Create GitHub repository",
        "Set up Python environment with FastAPI and scikit-learn",
        "Create basic FastAPI app skeleton"
      ]
    },
    {
      "title": "Implement sentiment model",
      "description": "Train and evaluate sentiment classifier.",
      "estimated_hours": 6,
      "status": "not_started",
      "tasks": [
        "Collect or load sentiment dataset",
        "Preprocess text",
        "Train baseline model",
        "Evaluate and log metrics"
      ]
    }
  ],
  "generation_prompt": null
}
```

---

## Key Features

✅ **Simple tasks** - String array instead of complex objects  
✅ **Light normalization** - Ensures arrays are never nil  
✅ **Nullable fields** - `estimated_hours` and `generation_prompt` can be null  
✅ **Status field** - Always "not_started" from backend  
✅ **Course-aware** - Uses course context when provided  
✅ **Direct integration** - Maps to `projectService.createProject()`  

---

## Backend Implementation

### Handler Flow

```go
func GenerateProjectHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Decode request
    var req GenerateProjectRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 2. Validate
    if req.User.ID == "" || req.Project.Title == "" {
        return error
    }
    
    // 3. Call LLM
    response := handleLLMRequest(w, req, "generateProjectPrompt", "/generate-project")
    
    // 4. Parse
    var projectResp GenerateProjectResponse
    json.Unmarshal(response, &projectResp)
    
    // 5. Light normalization (ensure arrays never nil)
    if projectResp.LearningObjectives == nil {
        projectResp.LearningObjectives = []string{}
    }
    // ... (repeat for all arrays)
    
    for i := range projectResp.Milestones {
        if projectResp.Milestones[i].Tasks == nil {
            projectResp.Milestones[i].Tasks = []string{}
        }
    }
    
    // 6. Return
    json.Marshal(projectResp)
}
```

---

## Testing

### Custom Project

```bash
curl -X POST http://localhost:8080/generate-project \
  -H "Content-Type: application/json" \
  -d '{
    "user": {"id": "test-user", "name": "Test"},
    "project": {
      "title": "Build a Sentiment Analysis App",
      "description": "ML model for movie reviews",
      "project_type": "application",
      "technologies": ["Python", "FastAPI"],
      "learning_objectives": ["Practice NLP"],
      "prerequisites": ["Basic Python"],
      "deliverables": ["GitHub repo", "Deployed app"]
    },
    "constraints": {
      "maxMilestones": 6,
      "maxTasksPerMilestone": 6,
      "estimatedHours": 20,
      "targetDifficulty": "intermediate"
    }
  }'
```

### Course-Based Project

```bash
curl -X POST http://localhost:8080/generate-project \
  -H "Content-Type: application/json" \
  -d '{
    "user": {"id": "test-user", "name": "Test"},
    "project": {
      "title": "NLP Project",
      "description": "Apply ML to text",
      "project_type": "application",
      "technologies": ["Python"],
      "learning_objectives": ["Practice NLP"],
      "prerequisites": ["Intro to ML"],
      "deliverables": ["Working model"]
    },
    "course_context": {
      "courseId": "ml-101",
      "unitId": "unit-5",
      "courseTitle": "Intro to ML",
      "unitTitle": "Text Classification",
      "topicsCovered": ["NLP", "TF-IDF"]
    },
    "constraints": {
      "maxMilestones": 5,
      "maxTasksPerMilestone": 5,
      "estimatedHours": 15,
      "targetDifficulty": "beginner"
    }
  }'
```

---

## Frontend Integration

```javascript
const response = await apiCall('/generate-project', requestData);

// Response can be directly passed to projectService
const project = await projectService.createProject({
  ...response,
  user_id: currentUser.id,
  milestones: response.milestones
});
```

---

## Differences from Previous Version

| Aspect | Previous | Current |
|--------|----------|---------|
| Tasks | Complex objects with priority, tags, etc. | Simple string array |
| Fields | `difficulty_level`, `sharing_platforms`, `due_offset_days`, `resources`, `success_criteria` | Removed for simplicity |
| Status | "pending" | "not_started" |
| Nullable | Few fields | `estimated_hours`, `generation_prompt` |
| Normalization | Extensive fallback logic | Light array nil checks only |

---

## Summary

The simplified `/generate-project` endpoint:

1. ✅ Receives project definition from frontend
2. ✅ Generates project structure via LLM
3. ✅ Returns JSON with **string array tasks**
4. ✅ Applies **light normalization** (arrays never nil)
5. ✅ Supports both custom and course-based projects
6. ✅ Maps directly to `projectService.createProject()`

Tasks are now simple strings like `"Create GitHub repository"` instead of complex objects, making the contract simpler and easier to work with.
