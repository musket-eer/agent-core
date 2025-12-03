# Generate Project API - Frontend Integration

## Overview

The `/generate-project` endpoint is a stateless LLM service that generates complete project plans with milestones and embedded tasks. It's designed to integrate directly with the frontend's `projectService.createProject()` method.

**Key Principle:** The backend receives a project definition from the frontend, generates a complete project structure via LLM, and returns JSON that can be directly passed to `projectService.createProject()`.

---

## Endpoint

**URL:** `POST /generate-project`

**Purpose:** Generate a complete project plan with milestones and tasks for:
- Custom projects (user-defined)
- Course-based projects (generated from course/unit context)

---

## Request Contract

### Full Request Schema

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
  "course_context": {
    "courseId": "ml-course-id-or-null",
    "unitId": "unit-id-or-null",
    "courseTitle": "Intro to Machine Learning",
    "courseDescription": "What the course covers",
    "unitTitle": "Text classification unit",
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

### Field Descriptions

**user:**
- `id` (required) - User UUID
- `name` (optional) - User's display name

**project:**
- `title` (required) - Project title
- `description` (required) - Project description/goal
- `project_type` - "application", "research", "learning", etc.
- `technologies` - Array of tech stack items
- `learning_objectives` - What the user will learn
- `prerequisites` - Required knowledge/skills
- `deliverables` - Expected outputs

**course_context** (optional, omit for custom projects):
- `courseId` - Course identifier (or null)
- `unitId` - Unit identifier (or null)
- `courseTitle` - Course name
- `courseDescription` - Course overview
- `unitTitle` - Unit name
- `unitObjective` - Unit learning goal
- `topicsCovered` - Array of topics in the unit

**constraints:**
- `maxMilestones` - Maximum number of milestones (typically 6)
- `maxTasksPerMilestone` - Maximum tasks per milestone (typically 6)
- `estimatedHours` - Total project time budget
- `targetDifficulty` - "beginner", "intermediate", or "advanced"

---

## Response Contract

### Full Response Schema

```json
{
  "title": "Build a Sentiment Analysis Model",
  "description": "You will build and evaluate a sentiment classifier using Python and scikit-learn. This project covers data preprocessing, feature engineering, model training, and evaluation.",
  "project_type": "application",
  "difficulty_level": "intermediate",
  "estimated_hours": 20,
  "learning_objectives": [
    "Understand sentiment analysis fundamentals",
    "Learn to preprocess text data for ML",
    "Practice training and evaluating classification models",
    "Gain experience with scikit-learn and pandas"
  ],
  "prerequisites": [
    "Basic Python programming",
    "Introduction to machine learning concepts",
    "Familiarity with Jupyter notebooks"
  ],
  "technologies": ["Python", "scikit-learn", "pandas", "numpy", "matplotlib"],
  "deliverables": [
    "Trained sentiment classification model",
    "Jupyter notebook with full pipeline",
    "GitHub repository with code and documentation",
    "Short report or blog post explaining methodology and results"
  ],
  "sharing_platforms": ["github", "blog"],
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
    },
    {
      "title": "Data collection and exploration",
      "description": "Find and explore a movie review dataset for sentiment analysis.",
      "estimated_hours": 3,
      "due_offset_days": 5,
      "status": "pending",
      "tasks": [
        {
          "title": "Download IMDB movie review dataset",
          "description": "Get the IMDB dataset from Kaggle or use scikit-learn's built-in dataset.",
          "estimated_minutes": 45,
          "priority": "high",
          "tags": ["data", "download"]
        },
        {
          "title": "Load and inspect dataset",
          "description": "Load data into pandas DataFrame and examine structure, size, and sample reviews.",
          "estimated_minutes": 60,
          "priority": "high",
          "tags": ["data", "exploration"]
        },
        {
          "title": "Analyze class distribution",
          "description": "Check balance between positive and negative reviews, visualize distribution.",
          "estimated_minutes": 45,
          "priority": "medium",
          "tags": ["data", "analysis"]
        }
      ],
      "resources": [
        "https://www.kaggle.com/datasets/lakshmi25npathi/imdb-dataset-of-50k-movie-reviews"
      ],
      "success_criteria": [
        "Dataset loaded successfully",
        "Class distribution documented",
        "Data quality report completed"
      ]
    }
  ],
  "generation_prompt": null
}
```

### Response Field Descriptions

**Top-level fields:**
- `title` - Project title (may be refined from input)
- `description` - Expanded project description
- `project_type` - Project category
- `difficulty_level` - Matches `targetDifficulty` from constraints
- `estimated_hours` - Total time estimate
- `learning_objectives` - Array of learning goals
- `prerequisites` - Array of required knowledge
- `technologies` - Array of tech stack items
- `deliverables` - Array of expected outputs
- `sharing_platforms` - Array of platforms (e.g., ["github", "blog", "portfolio"])
- `milestones` - Array of milestone objects
- `generation_prompt` - Optional debug info (usually null)

**Milestone fields:**
- `title` - Milestone name
- `description` - What this milestone accomplishes
- `estimated_hours` - Time estimate for milestone
- `due_offset_days` - Days from project start (e.g., 2, 5, 10, 15)
- `status` - Always "pending" from backend
- `tasks` - Array of embedded task objects
- `resources` - Array of helpful URLs
- `success_criteria` - Array of measurable outcomes

**Task fields:**
- `title` - Action-oriented task name
- `description` - What to do and deliverable
- `estimated_minutes` - Time estimate (15-120 minutes)
- `priority` - "low", "medium", or "high"
- `tags` - Array of categories (e.g., ["setup", "coding"])

---

## Frontend Integration

### Example: Generate Custom Project

```javascript
import { apiCall } from './apiClient';
import { projectService } from './supabaseClient/projectsService';

const generateCustomProject = async (projectData) => {
  // Call backend to generate project structure
  const response = await apiCall('/generate-project', {
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
    // No course_context for custom projects
    constraints: {
      maxMilestones: 6,
      maxTasksPerMilestone: 6,
      estimatedHours: projectData.estimatedHours || 20,
      targetDifficulty: projectData.difficulty || 'intermediate'
    }
  });

  // Response can be directly passed to projectService.createProject
  const project = await projectService.createProject({
    ...response,
    user_id: currentUser.id,
    milestones: response.milestones
  });

  return project;
};
```

### Example: Generate Course-Based Project

```javascript
const generateCourseProject = async (courseData, unitData) => {
  const response = await apiCall('/generate-project', {
    user: {
      id: currentUser.id,
      name: currentUser.name
    },
    project: {
      title: `${unitData.title} Project`,
      description: unitData.description,
      project_type: 'application',
      technologies: courseData.technologies || [],
      learning_objectives: unitData.objectives || [],
      prerequisites: courseData.prerequisites || [],
      deliverables: ['Working application', 'Documentation', 'GitHub repository']
    },
    course_context: {
      courseId: courseData.id,
      unitId: unitData.id,
      courseTitle: courseData.title,
      courseDescription: courseData.description,
      unitTitle: unitData.title,
      unitObjective: unitData.objective,
      topicsCovered: unitData.topics || []
    },
    constraints: {
      maxMilestones: 6,
      maxTasksPerMilestone: 6,
      estimatedHours: unitData.estimatedHours || 15,
      targetDifficulty: courseData.difficulty || 'intermediate'
    }
  });

  // Create project linked to course/unit
  const project = await projectService.createProject({
    ...response,
    user_id: currentUser.id,
    course_id: courseData.id,
    unit_id: unitData.id,
    milestones: response.milestones
  });

  return project;
};
```

---

## Backend Implementation

### Handler Flow

```go
func GenerateProjectHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Decode request
    var req GenerateProjectRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 2. Validate required fields
    if req.User.ID == "" || req.Project.Title == "" {
        return error
    }
    
    // 3. Call LLM with prompt
    response := handleLLMRequest(w, req, "generateProjectPrompt", "/generate-project")
    
    // 4. Parse LLM response
    var projectResp GenerateProjectResponse
    json.Unmarshal(response, &projectResp)
    
    // 5. Ensure safe defaults (no nil arrays)
    if projectResp.LearningObjectives == nil {
        projectResp.LearningObjectives = []string{}
    }
    // ... (repeat for all array fields)
    
    // 6. Return JSON
    json.Marshal(projectResp)
}
```

### Prompt Construction

The backend builds a dynamic prompt from the request:

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

Course / unit context (optional):
- Course ID: {{course_context.courseId}}
- Course title: {{course_context.courseTitle}}
- Unit title: {{course_context.unitTitle}}
- Topics covered: {{course_context.topicsCovered}}

Constraints:
- Max milestones: {{constraints.maxMilestones}}
- Max tasks per milestone: {{constraints.maxTasksPerMilestone}}
- Target estimated hours: {{constraints.estimatedHours}}
- Target difficulty: {{constraints.targetDifficulty}}

Please design a project plan and return ONLY the JSON object.
```

---

## Testing

### Test Custom Project

```bash
curl -X POST http://localhost:8080/generate-project \
  -H "Content-Type: application/json" \
  -d '{
    "user": {
      "id": "test-user-123",
      "name": "Test User"
    },
    "project": {
      "title": "Build a Sentiment Analysis Model",
      "description": "Create an ML model to classify movie reviews",
      "project_type": "application",
      "technologies": ["Python", "scikit-learn", "pandas"],
      "learning_objectives": ["Understand sentiment analysis"],
      "prerequisites": ["Basic Python"],
      "deliverables": ["Trained model", "Notebook"]
    },
    "constraints": {
      "maxMilestones": 6,
      "maxTasksPerMilestone": 6,
      "estimatedHours": 20,
      "targetDifficulty": "intermediate"
    }
  }'
```

### Test Course-Based Project

```bash
curl -X POST http://localhost:8080/generate-project \
  -H "Content-Type: application/json" \
  -d '{
    "user": {
      "id": "test-user-123",
      "name": "Test User"
    },
    "project": {
      "title": "Text Classification Project",
      "description": "Apply ML to text data",
      "project_type": "application",
      "technologies": ["Python", "scikit-learn"],
      "learning_objectives": ["Practice NLP"],
      "prerequisites": ["Intro to ML"],
      "deliverables": ["Working model"]
    },
    "course_context": {
      "courseId": "ml-101",
      "unitId": "unit-5",
      "courseTitle": "Intro to Machine Learning",
      "courseDescription": "Learn ML fundamentals",
      "unitTitle": "Text Classification",
      "unitObjective": "Apply ML to text",
      "topicsCovered": ["NLP", "TF-IDF", "Logistic Regression"]
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

## Error Handling

### Backend Fallback

If LLM response parsing fails, backend returns safe defaults:

```go
projectResp = GenerateProjectResponse{
    Title:              req.Project.Title,
    Description:        req.Project.Description,
    ProjectType:        req.Project.ProjectType,
    DifficultyLevel:    req.Constraints.TargetDifficulty,
    EstimatedHours:     req.Constraints.EstimatedHours,
    LearningObjectives: req.Project.LearningObjectives,
    Prerequisites:      req.Project.Prerequisites,
    Technologies:       req.Project.Technologies,
    Deliverables:       req.Project.Deliverables,
    SharingPlatforms:   []string{},
    Milestones:         []MilestoneWithTasks{},
}
```

### Frontend Handling

```javascript
try {
  const response = await apiCall('/generate-project', requestData);
  
  // Check for empty milestones
  if (!response.milestones || response.milestones.length === 0) {
    throw new Error('No milestones generated');
  }
  
  // Create project
  const project = await projectService.createProject({
    ...response,
    user_id: currentUser.id,
    milestones: response.milestones
  });
  
  return project;
} catch (error) {
  console.error('Failed to generate project:', error);
  // Show error to user
}
```

---

## Database Integration

The response structure maps directly to Supabase tables:

### `ml_projects` table
- `title`, `description`, `project_type`, `difficulty_level`
- `estimated_hours`, `learning_objectives`, `prerequisites`
- `technologies`, `deliverables`, `sharing_platforms`

### `ml_project_milestones` table
- `title`, `description`, `estimated_hours`, `due_offset_days`
- `status`, `resources`, `success_criteria`
- `project_id` (foreign key)

### `ml_project_tasks` table
- `title`, `description`, `estimated_minutes`, `priority`, `tags`
- `milestone_id` (foreign key)

The frontend's `projectService.createProject()` handles the insertion into all three tables.

---

## Benefits

✅ **Direct Integration** - Response maps to `projectService.createProject()`  
✅ **Stateless** - No backend database operations  
✅ **Flexible** - Works for custom and course-based projects  
✅ **Complete** - Returns all fields needed by frontend  
✅ **Safe Defaults** - Never returns nil arrays  
✅ **Course-Aware** - Uses course context when provided  
✅ **Realistic** - Generates actionable milestones and tasks  

---

## Summary

The `/generate-project` endpoint:

1. ✅ Receives project definition from frontend
2. ✅ Generates complete project structure via LLM
3. ✅ Returns JSON matching `projectService.createProject()` contract
4. ✅ Supports both custom and course-based projects
5. ✅ Includes milestones with embedded tasks
6. ✅ Provides all metadata (resources, success criteria, tags)
7. ✅ Returns safe defaults on errors

Frontend can directly pass the response to `projectService.createProject()` for database insertion.
