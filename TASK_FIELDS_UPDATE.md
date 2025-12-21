# Task Fields Update - IDs to Titles

## 🎯 Changes Made

Updated task-related types to use **titles instead of IDs** for courses and projects, making the data more human-readable and user-friendly.

---

## 📝 Updated Types

### 1. **CreateTaskPayload** (types_orchestrator.go) ✅

**Before:**
```go
type CreateTaskPayload struct {
    Title       string  `json:"title"`
    When        string  `json:"when,omitempty"`
    Priority    string  `json:"priority,omitempty"`
    ProjectID   *string `json:"projectId,omitempty"`    // ❌ ID
    CourseID    *string `json:"courseId,omitempty"`     // ❌ ID
    Description string  `json:"description,omitempty"`
}
```

**After:**
```go
type CreateTaskPayload struct {
    Title        string  `json:"title"`
    When         string  `json:"when,omitempty"`
    Priority     string  `json:"priority,omitempty"`
    ProjectTitle *string `json:"projectTitle,omitempty"` // ✅ Title
    CourseTitle  *string `json:"courseTitle,omitempty"`  // ✅ Title
    Description  string  `json:"description,omitempty"`
}
```

**Impact:**
- Frontend sends project/course titles instead of IDs
- More readable in logs and debugging
- User-friendly display

---

### 2. **TranslatedTask** (types_projects.go) ✅

**Before:**
```go
type TranslatedTask struct {
    MilestoneID      string `json:"milestoneId"`
    Title            string `json:"title"`
    Description      string `json:"description"`
    EstimatedMinutes int    `json:"estimated_minutes"`
    Difficulty       string `json:"difficulty"`
    CourseID         string `json:"courseId,omitempty"`  // ❌ ID
    LessonID         string `json:"lessonId,omitempty"`  // ❌ ID
}
```

**After:**
```go
type TranslatedTask struct {
    MilestoneID      string `json:"milestoneId"`
    Title            string `json:"title"`
    Description      string `json:"description"`
    EstimatedMinutes int    `json:"estimated_minutes"`
    Difficulty       string `json:"difficulty"`
    CourseTitle      string `json:"courseTitle,omitempty"` // ✅ Title
    LessonTitle      string `json:"lessonTitle,omitempty"` // ✅ Title
}
```

**Impact:**
- Tasks show course/lesson titles directly
- No need to lookup IDs to display names
- Better UX in task lists

---

## 🔄 Migration Guide

### Frontend Changes Required

**1. Creating Tasks (Orchestrator)**

**Old:**
```json
{
  "type": "create_task",
  "payload": {
    "title": "Complete React tutorial",
    "projectId": "proj-123",
    "courseId": "course-456"
  }
}
```

**New:**
```json
{
  "type": "create_task",
  "payload": {
    "title": "Complete React tutorial",
    "projectTitle": "Build E-commerce App",
    "courseTitle": "React Fundamentals"
  }
}
```

**2. Milestone Translation**

**Old Response:**
```json
{
  "tasks": [{
    "title": "Learn useState",
    "courseId": "course-456",
    "lessonId": "lesson-789"
  }]
}
```

**New Response:**
```json
{
  "tasks": [{
    "title": "Learn useState",
    "courseTitle": "React Fundamentals",
    "lessonTitle": "State Management Basics"
  }]
}
```

---

## ✅ Benefits

1. **Human-Readable**
   - Titles are immediately understandable
   - No need to lookup IDs in database

2. **Better UX**
   - Users see meaningful names
   - Easier to scan task lists

3. **Simplified Display**
   - Frontend can display titles directly
   - No additional API calls needed

4. **Better Logging**
   - Logs show actual course/project names
   - Easier debugging

5. **Markdown Friendly**
   - Titles work better in markdown responses
   - Can create readable task descriptions

---

## 📋 Example Usage

### Creating a Task with Course and Project

```json
{
  "commands": [{
    "type": "create_task",
    "payload": {
      "title": "Build authentication system",
      "when": "today",
      "priority": "high",
      "projectTitle": "E-commerce Platform",
      "courseTitle": "Node.js Security",
      "description": "Implement **JWT authentication** with refresh tokens. See [Auth Guide](https://jwt.io/introduction)"
    }
  }]
}
```

### Displaying Tasks

**Before (with IDs):**
```
Task: Build auth system
Course: course-abc-123  ❌ Not helpful
Project: proj-xyz-789   ❌ Not helpful
```

**After (with Titles):**
```
Task: Build authentication system
Course: Node.js Security ✅ Clear
Project: E-commerce Platform ✅ Clear
```

---

## 🔍 Related Types (Not Changed)

These types still use IDs where appropriate:

- **MilestoneTranslatorRequest.ProjectID** - Still uses ID (needed for backend lookup)
- **TaskID, MilestoneID** - Still use IDs (internal references)
- **SearchResult.CourseID** - Still uses ID (for navigation/linking)

**Why?** These are for backend processing or navigation, not user display.

---

## 📚 Files Modified

1. `/internal/llm/types_orchestrator.go`
   - Updated `CreateTaskPayload`
   - Changed `ProjectID` → `ProjectTitle`
   - Changed `CourseID` → `CourseTitle`

2. `/internal/llm/types_projects.go`
   - Updated `TranslatedTask`
   - Changed `CourseID` → `CourseTitle`
   - Changed `LessonID` → `LessonTitle`

---

## 🚀 Next Steps

### Backend
- ✅ Types updated
- ⏳ Update handlers to use titles
- ⏳ Update prompts to generate titles

### Frontend
- ⏳ Update task creation to send titles
- ⏳ Update task display to show titles
- ⏳ Remove ID lookup logic

---

**Status:** Types updated ✅ | Ready for handler updates
