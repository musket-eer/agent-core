# Course Creation Frontend Integration Guide

## Problem: Units Not Populating in Dashboard

**Root Cause:** The Course Creation Chat generates course **structure** only, not full lesson content. Units need to be generated via `generateCourseContentThunk`.

---

## Two-Step Course Creation Flow

### Step 1: Course Design (AI Chat)

**Backend:** `/course-creation/chat`

**What it does:**
- Guides user through conversational course design
- Collects: title, description, weeklyHours, objectives, etc.
- Generates basic unit structure (titles, objectives, assessments)
- Validates required fields
- Returns `ready_to_create` when complete

**What it DOESN'T do:**
- Generate full lesson content
- Create database records
- Populate units with lessons

### Step 2: Content Generation (Frontend)

**Frontend:** `generateCourseContentThunk`

**What it does:**
- Takes course structure from chat
- Generates full lesson content for each unit
- Creates course in database
- Creates units with lessons in database
- Returns course ID

---

## Correct Frontend Implementation

### When User Confirms Course Creation

```javascript
// In your Course Creation component
const handleCreateCourse = async () => {
  if (currentStep !== 'ready_to_create') {
    console.error('Course not ready to create');
    return;
  }

  try {
    setIsGenerating(true);
    
    // Prepare course data from the chat draft
    const courseData = {
      title: courseDraft.title,
      description: courseDraft.description,
      weeklyHours: courseDraft.weeklyHours,
      lessons: courseDraft.units.length, // Number of units
      objectives: courseDraft.objectives || [],
      priorKnowledge: courseDraft.priorKnowledge || '',
      difficultyLevel: courseDraft.difficultyLevel || 'beginner',
      referenceBooks: courseDraft.referenceBooks || [],
    };

    // CRITICAL: Call generateCourseContentThunk to create full course
    const result = await dispatch(generateCourseContentThunk({
      courseData,
      userId: currentUser.id
    }));

    if (result.success) {
      toast.success('Course created successfully!');
      navigate(`/courses/${result.courseId}`);
    } else {
      toast.error('Failed to create course');
    }
  } catch (error) {
    console.error('Error creating course:', error);
    toast.error('Failed to create course');
  } finally {
    setIsGenerating(false);
  }
};
```

---

## What NOT To Do

### ❌ Wrong: Directly saving courseDraft

```javascript
// DON'T DO THIS - units will be empty
const course = await courseService.saveCourseWithUnits({
  ...courseDraft,
  user_id: currentUser.id
});
```

**Why this fails:**
- `courseDraft.units` only has structure (title, objective)
- No lesson content
- Database expects full units with lessons
- Dashboard shows empty units

---

## What TO Do

### ✅ Correct: Use generateCourseContentThunk

```javascript
// DO THIS - generates full content
const result = await dispatch(generateCourseContentThunk({
  courseData: {
    title: courseDraft.title,
    description: courseDraft.description,
    weeklyHours: courseDraft.weeklyHours,
    lessons: courseDraft.units.length,
    objectives: courseDraft.objectives,
    priorKnowledge: courseDraft.priorKnowledge,
    difficultyLevel: courseDraft.difficultyLevel,
    referenceBooks: courseDraft.referenceBooks,
  },
  userId: currentUser.id
}));
```

**Why this works:**
- `generateCourseContentThunk` calls backend to generate full lesson content
- Creates course in database with all fields
- Creates units with complete lesson data
- Dashboard shows populated units

---

## Data Flow Diagram

```
User Chat
    ↓
[Course Creation Chat API]
    ↓
courseDraft {
  title: "Intro to ML",
  description: "...",
  weeklyHours: 5,
  units: [
    { title: "Foundations", objective: "...", assessment: "..." },
    { title: "Algorithms", objective: "...", assessment: "..." }
  ]
}
    ↓
[Frontend: generateCourseContentThunk]
    ↓
[Backend: Generate Course Content API]
    ↓
Full Course {
  id: "uuid",
  title: "Intro to ML",
  description: "...",
  weekly_hours: 5,
  units: [
    {
      id: "uuid",
      title: "Foundations",
      objective: "...",
      lessons: [
        { title: "What is ML?", content: "...", duration: 30 },
        { title: "Types of ML", content: "...", duration: 45 }
      ]
    },
    {
      id: "uuid",
      title: "Algorithms",
      objective: "...",
      lessons: [
        { title: "Linear Regression", content: "...", duration: 60 },
        { title: "Decision Trees", content: "...", duration: 45 }
      ]
    }
  ]
}
    ↓
[Database: ml_courses_staging + ml_units_staging]
    ↓
[Dashboard: Shows course with populated units]
```

---

## Required Fields Mapping

### From Chat Draft to Content Generation

| Chat Draft Field | Content Generation Field | Required |
|-----------------|-------------------------|----------|
| `title` | `title` | ✅ Yes |
| `description` | `description` | ✅ Yes |
| `weeklyHours` | `weeklyHours` | ✅ Yes |
| `units.length` | `lessons` | ✅ Yes |
| `objectives` | `objectives` | ⚠️ Recommended |
| `priorKnowledge` | `priorKnowledge` | ⚠️ Recommended |
| `difficultyLevel` | `difficultyLevel` | ⚠️ Recommended |
| `referenceBooks` | `referenceBooks` | ⚠️ Recommended |

**Note:** The backend validates these fields before allowing `ready_to_create`.

---

## Debugging Checklist

If units aren't showing in the dashboard:

1. ✅ Check that `currentStep === 'ready_to_create'` before creating
2. ✅ Verify `courseDraft` has all required fields:
   - `title` (non-empty)
   - `description` (non-empty)
   - `weeklyHours` (> 0)
   - `units` (length > 0)
3. ✅ Confirm you're calling `generateCourseContentThunk`, NOT `saveCourseWithUnits`
4. ✅ Check that `generateCourseContentThunk` returns `success: true`
5. ✅ Verify the course ID is valid
6. ✅ Check database tables:
   - `ml_courses_staging` has the course row
   - `ml_units_staging` has unit rows with `course_id` matching
7. ✅ Check browser console for errors
8. ✅ Check backend logs for content generation errors

---

## Example Component Structure

```javascript
import { useState } from 'react';
import { useDispatch } from 'react-redux';
import { generateCourseContentThunk } from './store/courseSlice';

function CourseCreationChat() {
  const [currentStep, setCurrentStep] = useState('greeting');
  const [courseDraft, setCourseDraft] = useState(null);
  const [chatHistory, setChatHistory] = useState([]);
  const [isGenerating, setIsGenerating] = useState(false);
  const dispatch = useDispatch();

  const sendMessage = async (userMessage) => {
    const response = await fetch('/course-creation/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        agentId: 'course_creation',
        userId: currentUser.id,
        currentStep,
        courseDraft,
        userMessage,
        chatHistory
      })
    });

    const data = await response.json();

    // Update chat history
    setChatHistory([
      ...chatHistory,
      { sender: 'user', text: userMessage },
      { sender: 'assistant', text: data.reply }
    ]);

    // Merge draft updates
    setCourseDraft({
      ...courseDraft,
      ...data.updatedDraft
    });

    // Update step
    setCurrentStep(data.nextStep);
  };

  const handleCreateCourse = async () => {
    if (currentStep !== 'ready_to_create') return;

    setIsGenerating(true);
    
    const result = await dispatch(generateCourseContentThunk({
      courseData: {
        title: courseDraft.title,
        description: courseDraft.description,
        weeklyHours: courseDraft.weeklyHours,
        lessons: courseDraft.units.length,
        objectives: courseDraft.objectives,
        priorKnowledge: courseDraft.priorKnowledge,
        difficultyLevel: courseDraft.difficultyLevel,
        referenceBooks: courseDraft.referenceBooks,
      },
      userId: currentUser.id
    }));

    setIsGenerating(false);

    if (result.success) {
      navigate(`/courses/${result.courseId}`);
    }
  };

  return (
    <div>
      {/* Chat UI */}
      <ChatMessages messages={chatHistory} />
      <ChatInput onSend={sendMessage} />
      
      {/* Create button (only show when ready) */}
      {currentStep === 'ready_to_create' && (
        <button 
          onClick={handleCreateCourse}
          disabled={isGenerating}
        >
          {isGenerating ? 'Creating Course...' : 'Create Course'}
        </button>
      )}
    </div>
  );
}
```

---

## Summary

**The Issue:**
- Course Creation Chat generates structure only
- Units need full lesson content to appear in dashboard

**The Solution:**
- After `ready_to_create`, call `generateCourseContentThunk`
- This generates full lesson content and saves to database
- Dashboard will then show populated units

**Key Takeaway:**
Course Creation = **Structure (Chat)** + **Content (Generation)** = **Complete Course**
