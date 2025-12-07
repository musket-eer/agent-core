# Course Creation Chat API - Dialogue-Based Course Design

## Overview

The `/course-creation/chat` endpoint implements a conversational AI that guides humans through creating a personalized course. It uses a state machine to manage the dialogue flow and progressively builds a course draft through back-and-forth conversation.

**Key Principle:** Human-AI collaboration to co-create courses that fit the learner's goals, schedule, and learning style.

---

## Endpoint

**URL:** `POST /course-creation/chat`

**Auth:** User session / JWT (same as other agents)

---

## Request Contract

```json
{
  "agentId": "course_creation",
  "userId": "uuid-of-user",
  "currentStep": "greeting",
  "courseDraft": {
    "id": null,
    "title": "Intro to Machine Learning",
    "description": "I want a gentle intro to ML.",
    "objectives": ["Understand basic ML concepts"],
    "priorKnowledge": "Basic Python",
    "difficultyLevel": "beginner",
    "weeklyHours": 5,
    "totalDurationWeeks": 4,
    "referenceBooks": [],
    "units": [
      {
        "id": "unit-1",
        "title": "Foundations of ML",
        "objective": "Understand supervised vs unsupervised",
        "assessment": "Short quiz",
        "position": 1
      }
    ]
  },
  "userMessage": "I only know basic Python and stats.",
  "chatHistory": [
    { "sender": "assistant", "text": "What do you want to learn?" },
    { "sender": "user", "text": "A beginner ML course." }
  ]
}
```

### Field Descriptions

- `agentId` - Agent identifier ("course_creation")
- `userId` (required) - User UUID
- `currentStep` (required) - Current state in the flow
- `courseDraft` - Current course draft (can be null or partial)
- `userMessage` - Latest message from user
- `chatHistory` - Recent conversation turns (last 10 kept)

---

## Response Contract

```json
{
  "reply": "Got it. A beginner ML course for someone with basic Python. Next, how many hours per week can you study?",
  "updatedDraft": {
    "difficultyLevel": "beginner",
    "priorKnowledge": "Basic Python and stats"
  },
  "nextStep": "ask_time_commitment",
  "suggestions": [
    "I can study 3 hours per week",
    "Let's do something very light"
  ]
}
```

### Response Fields

- `reply` - Message to show in chat UI
- `updatedDraft` - Partial update (frontend merges into local draft)
- `nextStep` - Next state in the flow
- `suggestions` - Optional quick reply buttons
- `createdCourseId` - (Optional) UUID if course was created server-side

---

## State Machine Flow

### 1. **greeting / ask_topic**

**Goal:** Understand what they want to learn

**AI asks:** "What would you like to learn about?"

**Extracts:**
- Topic
- Working title
- One-sentence description

**Updates:**
- `courseDraft.title`
- `courseDraft.description`

**Next:** `ask_level`

**Example:**
```json
{
  "reply": "Great! Let's create a course on Machine Learning. I'm thinking we could call it 'Intro to Machine Learning'. Sound good?",
  "updatedDraft": {
    "title": "Intro to Machine Learning",
    "description": "A gentle introduction to ML fundamentals"
  },
  "nextStep": "ask_level",
  "suggestions": ["Sounds perfect!", "Can we adjust the title?"]
}
```

---

### 2. **ask_level**

**Goal:** Determine difficulty level

**AI asks:** "What's your experience level with this topic?"

**Extracts:**
- `difficultyLevel`: "beginner" | "intermediate" | "advanced"

**Next:** `ask_prior_knowledge`

**Example:**
```json
{
  "reply": "What's your experience level with Machine Learning?",
  "updatedDraft": {},
  "nextStep": "ask_level",
  "suggestions": [
    "I'm a complete beginner",
    "I have some experience",
    "I'm fairly advanced"
  ]
}
```

---

### 3. **ask_prior_knowledge**

**Goal:** Understand what they already know

**AI asks:** "What do you already know about [topic]?"

**Extracts:**
- `priorKnowledge`: Free-form text

**Next:** `ask_time_commitment`

**Example:**
```json
{
  "reply": "What do you already know about Machine Learning?",
  "updatedDraft": {},
  "nextStep": "ask_prior_knowledge",
  "suggestions": [
    "Just basic Python",
    "I know some statistics",
    "Nothing yet"
  ]
}
```

---

### 4. **ask_time_commitment**

**Goal:** Understand time availability

**AI asks:** "How many hours per week can you dedicate? And for about how many weeks?"

**Extracts:**
- `weeklyHours`: Integer
- `totalDurationWeeks`: Integer

**Next:** `ask_goals`

**Example:**
```json
{
  "reply": "How many hours per week can you dedicate to this? And for about how many weeks?",
  "updatedDraft": {},
  "nextStep": "ask_time_commitment",
  "suggestions": [
    "3-5 hours per week",
    "I have about 4 weeks",
    "Let's keep it flexible"
  ]
}
```

---

### 5. **ask_goals**

**Goal:** Understand success criteria

**AI asks:** "What would make this course feel successful for you?"

**Extracts:**
- `objectives`: Array of 2-5 learning goals

**Next:** `draft_outline`

**Example:**
```json
{
  "reply": "What would make this course feel successful for you?",
  "updatedDraft": {},
  "nextStep": "ask_goals",
  "suggestions": [
    "Build a real ML project",
    "Understand the fundamentals",
    "Get job-ready skills"
  ]
}
```

---

### 6. **draft_outline**

**Goal:** Generate initial course structure

**AI generates:** 3-8 units based on all gathered info

**Each unit has:**
- `id`: "unit-1", "unit-2", etc.
- `title`: Clear, specific title
- `objective`: What they'll learn
- `assessment`: Quiz, project, exercise, etc.
- `position`: 1, 2, 3...

**Next:** `refine_outline`

**Example:**
```json
{
  "reply": "Here's what I'm thinking:\n\nUnit 1: ML Fundamentals\nUnit 2: Data Preprocessing\nUnit 3: Supervised Learning\nUnit 4: Model Evaluation\n\nWhat do you think?",
  "updatedDraft": {
    "units": [
      {
        "id": "unit-1",
        "title": "ML Fundamentals",
        "objective": "Understand what ML is and key concepts",
        "assessment": "Concept quiz",
        "position": 1
      },
      {
        "id": "unit-2",
        "title": "Data Preprocessing",
        "objective": "Learn to clean and prepare data",
        "assessment": "Hands-on exercise",
        "position": 2
      }
    ]
  },
  "nextStep": "refine_outline",
  "suggestions": [
    "Looks great!",
    "Can we add a unit on neural networks?",
    "Remove unit 3"
  ]
}
```

---

### 7. **refine_outline**

**Goal:** Let user adjust units

**AI listens for:**
- "Add a unit on X" → Insert new unit
- "Remove unit 3" → Delete unit
- "Change unit 2 to focus on Y" → Update unit
- "Swap units 1 and 2" → Reorder

**Updates:** `units` array

**Next:** `final_review` (when user is satisfied)

**Example:**
```json
{
  "reply": "I've added a unit on Neural Networks as unit 5. Anything else you'd like to adjust?",
  "updatedDraft": {
    "units": [
      {
        "id": "unit-5",
        "title": "Introduction to Neural Networks",
        "objective": "Understand basic neural network architecture",
        "assessment": "Build a simple neural network",
        "position": 5
      }
    ]
  },
  "nextStep": "refine_outline",
  "suggestions": [
    "Perfect, let's move on",
    "Can we adjust unit 3?",
    "Remove the last unit"
  ]
}
```

---

### 8. **final_review**

**Goal:** Show complete summary and get confirmation

**AI shows:**
- Title and description
- Difficulty level
- Time commitment
- Learning objectives
- List of units

**AI asks:** "Ready to create this course?"

**Next:** 
- If yes → `ready_to_create`
- If no → `refine_outline`

**Example:**
```json
{
  "reply": "Here's your complete course:\n\n📚 Intro to Machine Learning\n🎯 Beginner level\n⏰ 5 hours/week for 4 weeks\n\nYou'll learn:\n- Understand ML fundamentals\n- Build real ML models\n\nUnits:\n1. ML Fundamentals\n2. Data Preprocessing\n3. Supervised Learning\n4. Model Evaluation\n\nReady to create this course?",
  "updatedDraft": {},
  "nextStep": "final_review",
  "suggestions": [
    "Yes, let's create it!",
    "Wait, let me adjust something"
  ]
}
```

---

### 9. **ready_to_create**

**Goal:** Course is ready for creation

**CRITICAL:** This step is ONLY reached if all required fields are present:
- `title` - Non-empty string
- `description` - Non-empty string
- `weeklyHours` - Numeric value > 0
- `units` - Array with at least 1 unit

**Backend Validation:** The handler validates these fields before allowing `ready_to_create`. If any are missing, it forces back to `final_review` and asks for the missing field.

**AI confirms:** "Awesome! Your course is ready."

**Frontend:** Calls `saveCourseWithUnits(courseDraft)` to create in Supabase

**Example:**
```json
{
  "reply": "Awesome! Your course is ready. Click 'Create Course' to get started.",
  "updatedDraft": {},
  "nextStep": "ready_to_create",
  "suggestions": []
}
```

---

## Frontend Integration

### Initial Request (Greeting)

```javascript
const response = await fetch(`${API_BASE_URL}/course-creation/chat`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    agentId: 'course_creation',
    userId: currentUser.id,
    currentStep: 'greeting',
    courseDraft: null,
    userMessage: '',
    chatHistory: []
  })
});

const data = await response.json();
// data.reply: "Hi! What would you like to learn about?"
// data.nextStep: "ask_topic"
```

### Subsequent Requests

```javascript
// User sends message
const response = await fetch(`${API_BASE_URL}/course-creation/chat`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    agentId: 'course_creation',
    userId: currentUser.id,
    currentStep: currentStep, // From previous response
    courseDraft: courseDraft, // Merged with previous updatedDraft
    userMessage: userInput,
    chatHistory: [
      ...chatHistory,
      { sender: 'user', text: userInput }
    ]
  })
});

const data = await response.json();

// Update UI
setChatHistory([
  ...chatHistory,
  { sender: 'user', text: userInput },
  { sender: 'assistant', text: data.reply }
]);

// Merge draft updates
setCourseDraft({
  ...courseDraft,
  ...data.updatedDraft
});

// Update step
setCurrentStep(data.nextStep);
```

### Creating the Course

**IMPORTANT:** The course draft from the AI contains basic unit structure (title, objective, assessment) but NOT full lesson content. The frontend must:

1. Take the draft when `nextStep === 'ready_to_create'`
2. Call `generateCourseContentThunk` to generate full lesson content
3. Then save to database

```javascript
// When nextStep === 'ready_to_create' and user confirms
if (data.nextStep === 'ready_to_create' && userConfirmed) {
  // Prepare course data for content generation
  const courseData = {
    title: courseDraft.title,
    description: courseDraft.description,
    weeklyHours: courseDraft.weeklyHours,
    lessons: courseDraft.units.length, // Number of units
    objectives: courseDraft.objectives,
    priorKnowledge: courseDraft.priorKnowledge,
    difficultyLevel: courseDraft.difficultyLevel,
    referenceBooks: courseDraft.referenceBooks,
  };
  
  // Generate full course content (units with lessons)
  const result = await dispatch(generateCourseContentThunk({
    courseData,
    userId: currentUser.id
  }));
  
  if (result.success) {
    // Navigate to new course
    navigate(`/courses/${result.courseId}`);
  }
}
```

**Why this flow?**
- The AI chat generates the **course structure** (units with titles/objectives)
- `generateCourseContentThunk` generates **full lesson content** for each unit
- This separation allows the AI to focus on course design while the content generation service handles detailed lesson creation

---

## Backend Implementation

### Handler Flow

```go
func CourseCreationChatHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Decode request
    var req CourseCreationChatRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 2. Validate
    if req.UserID == "" {
        return error
    }
    
    // 3. Default to greeting if no step
    if req.CurrentStep == "" {
        req.CurrentStep = "greeting"
    }
    
    // 4. Validate step
    if !ValidCourseCreationSteps[req.CurrentStep] {
        return error
    }
    
    // 5. Initialize draft if nil
    if req.CourseDraft == nil {
        req.CourseDraft = &CourseDraft{...}
    }
    
    // 6. Truncate chat history (last 10 turns)
    if len(req.ChatHistory) > 10 {
        req.ChatHistory = req.ChatHistory[len(req.ChatHistory)-10:]
    }
    
    // 7. Call LLM
    response := handleLLMRequest(w, req, "courseCreationPrompt", "/course-creation/chat")
    
    // 8. Parse and validate
    var chatResp CourseCreationChatResponse
    json.Unmarshal(response, &chatResp)
    
    // 9. Normalize response
    if chatResp.Reply == "" {
        chatResp.Reply = "Let's continue building your course."
    }
    if !ValidCourseCreationSteps[chatResp.NextStep] {
        chatResp.NextStep = req.CurrentStep
    }
    
    // 10. CRITICAL: Validate required fields before ready_to_create
    if chatResp.NextStep == "ready_to_create" || chatResp.NextStep == "final_review" {
        mergedDraft := mergeDrafts(req.CourseDraft, chatResp.UpdatedDraft)
        
        if !isFinalDraftComplete(mergedDraft) {
            // Force back to final_review and ask for missing fields
            missingFields := getMissingFields(mergedDraft)
            chatResp.Reply = buildMissingFieldsPrompt(missingFields)
            chatResp.NextStep = "final_review"
            chatResp.Suggestions = buildMissingFieldsSuggestions(missingFields)
        }
    }
    
    // 11. Return
    json.Marshal(chatResp)
}

// isFinalDraftComplete checks required fields
func isFinalDraftComplete(draft *CourseDraft) bool {
    hasTitle := draft.Title != "" && strings.TrimSpace(draft.Title) != ""
    hasDescription := draft.Description != "" && strings.TrimSpace(draft.Description) != ""
    weeklyOk := draft.WeeklyHours > 0
    lessonsOk := len(draft.Units) > 0
    
    return hasTitle && hasDescription && weeklyOk && lessonsOk
}
```

### Required Fields Validation

Before allowing `nextStep: "ready_to_create"`, the backend validates:

1. **title** - Non-empty string
2. **description** - Non-empty string (1-2 sentences)
3. **weeklyHours** - Numeric value > 0 (e.g., 3, 5, 10)
4. **units** - Array with at least 1 unit (represents "lessons")

If any field is missing, the handler:
- Overrides the LLM response
- Forces `nextStep` back to `"final_review"`
- Asks specifically for the missing field
- Provides relevant suggestion chips

This ensures the frontend's `saveCourseWithUnits` will never receive incomplete data.

---

## Testing Examples

### Test 1: Initial Greeting

**Request:**
```bash
curl -X POST http://localhost:8080/course-creation/chat \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "course_creation",
    "userId": "test-user-123",
    "currentStep": "greeting",
    "courseDraft": null,
    "userMessage": "",
    "chatHistory": []
  }'
```

**Expected Response:**
```json
{
  "reply": "Hi! I'm excited to help you create a personalized course. What would you like to learn about?",
  "updatedDraft": {},
  "nextStep": "ask_topic",
  "suggestions": [
    "Machine Learning",
    "Web Development",
    "Data Science"
  ]
}
```

---

### Test 2: Topic Selection

**Request:**
```bash
curl -X POST http://localhost:8080/course-creation/chat \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "course_creation",
    "userId": "test-user-123",
    "currentStep": "ask_topic",
    "courseDraft": {},
    "userMessage": "I want to learn machine learning",
    "chatHistory": [
      {"sender": "assistant", "text": "What would you like to learn?"},
      {"sender": "user", "text": "I want to learn machine learning"}
    ]
  }'
```

**Expected Response:**
```json
{
  "reply": "Great! Let's create a course on Machine Learning. I'm thinking we could call it 'Introduction to Machine Learning'. Sound good?",
  "updatedDraft": {
    "title": "Introduction to Machine Learning",
    "description": "Learn the fundamentals of machine learning"
  },
  "nextStep": "ask_level",
  "suggestions": ["Sounds perfect!", "Can we adjust the title?"]
}
```

---

### Test 3: Draft Outline

**Request:**
```bash
curl -X POST http://localhost:8080/course-creation/chat \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "course_creation",
    "userId": "test-user-123",
    "currentStep": "draft_outline",
    "courseDraft": {
      "title": "Intro to ML",
      "difficultyLevel": "beginner",
      "priorKnowledge": "Basic Python",
      "weeklyHours": 5,
      "totalDurationWeeks": 4,
      "objectives": ["Understand ML basics", "Build models"]
    },
    "userMessage": "Let's see the outline",
    "chatHistory": []
  }'
```

**Expected Response:**
```json
{
  "reply": "Here's what I'm thinking:\n\nUnit 1: ML Fundamentals\nUnit 2: Data Preprocessing\nUnit 3: Supervised Learning\nUnit 4: Model Evaluation\n\nWhat do you think?",
  "updatedDraft": {
    "units": [
      {
        "id": "unit-1",
        "title": "ML Fundamentals",
        "objective": "Understand what ML is",
        "assessment": "Quiz",
        "position": 1
      }
    ]
  },
  "nextStep": "refine_outline",
  "suggestions": ["Looks great!", "Add a unit on neural networks"]
}
```

---

### Test 4: Validation - Missing Required Fields

**Request:**
```bash
curl -X POST http://localhost:8080/course-creation/chat \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "course_creation",
    "userId": "test-user-123",
    "currentStep": "final_review",
    "courseDraft": {
      "title": "Intro to ML",
      "description": "",
      "weeklyHours": 0,
      "units": []
    },
    "userMessage": "Yes, create it!",
    "chatHistory": []
  }'
```

**Expected Response:**
```json
{
  "reply": "Please give me 1-2 sentences describing what this course will cover.",
  "updatedDraft": {},
  "nextStep": "final_review",
  "suggestions": [
    "A comprehensive guide to...",
    "Learn the fundamentals of...",
    "Master the skills of..."
  ]
}
```

**Note:** Backend detects missing `description`, `weeklyHours`, and `units`. It asks for the first missing field (`description`) and keeps the user in `final_review` until all required fields are present.

---

## Error Handling

### Missing Required Field

**Response:** `400 Bad Request`
```json
{
  "error": "userId is required"
}
```

### Invalid Step

**Response:** `400 Bad Request`
```json
{
  "error": "Invalid currentStep"
}
```

### LLM Parse Error

**Response:** `200 OK` (with fallback)
```json
{
  "reply": "I'm having trouble processing that. Could you rephrase?",
  "updatedDraft": {},
  "nextStep": "greeting",
  "suggestions": []
}
```

---

## Configuration

### DeepSeek Config

```go
"/course-creation/chat": {
    ModelId:     "deepseek-chat",
    Temperature: 0.7,  // Slightly higher for conversational tone
    MaxTokens:   2048, // Enough for outline generation
    ClientName:  "general",
}
```

---

## Key Features

✅ **State machine** - Structured 10-step flow  
✅ **Conversational** - One question at a time  
✅ **Progressive building** - Draft evolves through dialogue  
✅ **Flexible** - Can go back and refine  
✅ **Quick replies** - Suggestion chips for common responses  
✅ **Chat history** - Maintains context (last 10 turns)  
✅ **Validation** - Ensures valid steps and responses  
✅ **Fallback** - Graceful handling of parse errors  
✅ **Frontend creation** - Delegates to existing `saveCourseWithUnits`  

---

## Summary

The `/course-creation/chat` endpoint:

1. ✅ Guides users through 10-step course creation flow
2. ✅ Maintains conversational state via `currentStep`
3. ✅ Progressively builds `courseDraft` through dialogue
4. ✅ Provides quick reply suggestions
5. ✅ Validates all steps and responses
6. ✅ Validates required fields before `ready_to_create`
7. ✅ Supports refinement and iteration
8. ✅ Maintains chat history for context

### Course Creation Flow

**Backend (Course Creation Chat):**
- Generates course **structure** (title, description, units with titles/objectives)
- Validates required fields
- Returns `ready_to_create` when complete

**Frontend (After ready_to_create):**
- Takes the course structure from `courseDraft`
- Calls `generateCourseContentThunk` to generate **full lesson content** for each unit
- Saves complete course with units to database
- Navigates to new course

**Why Two Steps?**
1. **AI Chat** focuses on course design and structure (conversational, iterative)
2. **Content Generation** creates detailed lessons (batch processing, LLM-intensive)

This separation provides:
- ✅ Better user experience (fast chat, then content generation)
- ✅ Clear separation of concerns
- ✅ Ability to regenerate content without redesigning structure

**Units Population Issue:**
If units aren't showing in the dashboard, ensure the frontend is calling `generateCourseContentThunk` after receiving `ready_to_create`. The AI chat only provides the structure; full lesson content must be generated separately.
