# LLM API Endpoints & Payloads

**Base URL**: `{REACT_APP_BACKEND_URL}`   
**Auth**: Bearer token | **Content-Type**: `application/json` 

---

## Dashboard Agent

### POST /dashboard/insights
**Request**: `{ userId, userName, educationLevel, careerGoals, interests[], recentActivities[], completedCourses[], currentCourses[], learningStreak, totalStudyTime }`   
**Response**: `{ insights: { summary, strengths[], recommendations[], nextSteps[] }, motivationalMessage, progressHighlights[] }` 

### POST /dashboard/recommendations
**Request**: `{ userId, interests[], careerGoals, completedCourses[], skillLevel, learningStyle }`   
**Response**: `{ recommendations: [{ courseId, title, relevanceScore, reasoning, estimatedDuration, difficulty }] }` 

### POST /dashboard/analyze-progress
**Request**: `{ userId, timeframe, completedActivities[], coursesInProgress[], assessmentScores[], studyTimeData }`   
**Response**: `{ analysis: { summary, trends[], achievements[], areasForImprovement[], recommendations[] } }` 

### POST /dashboard/chat
**Request**: `{ message, userId, userName, userProfile, chatHistory[], dashboardContext }`   
**Response**: `{ response, suggestions[] }` 

---

## Course Browse Agent

### POST /courses/intelligent-search
**Request**: `{ query, userId, userProfile, filters, userContext }`   
**Response**: `{ results: [{ courseId, title, description, relevanceScore, matchReason, highlights[] }], searchInsights }` 

### POST /courses/explain-relevance
**Request**: `{ course: { id, title, description, difficulty, topics[], skills[] }, userProfile }`   
**Response**: `{ explanation }` 

### POST /courses/compare
**Request**: `{ courses[], userContext }`   
**Response**: `{ comparison: { summary, recommendations, comparisonTable[], bestFor } }` 

### POST /courses/suggest-path
**Request**: `{ userId, targetGoal, currentSkills[], timeCommitment, preferredDifficulty, interests[] }`   
**Response**: `{ learningPath: { title, description, estimatedDuration, phases[], milestones[], reasoning } }` 

### POST /courses/chat
**Request**: `{ message, userId, userProfile, chatHistory[], browseContext }`   
**Response**: `{ response, suggestedCourses[], suggestedFilters[] }` 

---

## Reflection Agent

### POST /reflection/generate-prompts
**Request**: `{ activity, userContext, courseContext }`   
**Response**: `{ prompts: [{ id, question, category, followUp }] }` 

### POST /reflection/analyze
**Request**: `{ reflection, activity, userProfile, learningGoals[] }`   
**Response**: `{ insights, strengths[], areasForGrowth[], suggestions[], depthScore }` 

### POST /reflection/learning-insights
**Request**: `{ userId, timeframe, reflections[], activities[], learningGoals[] }`   
**Response**: `{ patterns[], growthAreas[], recommendations[], milestones[], metacognitiveScore }` 

### POST /reflection/period-summary
**Request**: `{ userId, startDate, endDate, completedActivities[], reflections[], achievements[], challenges[] }`   
**Response**: `{ summary }` 

### POST /reflection/chat
**Request**: `{ message, userId, userName, userProfile, chatHistory[], reflectionContext }`   
**Response**: `{ response, followUpQuestions[], resources[] }` 

### POST /reflection/suggest-exercises
**Request**: `{ userId, learningStyle, currentChallenges[], learningGoals[], recentActivities[] }`   
**Response**: `{ exercises: [{ id, title, description, duration, type, instructions[] }] }` 

---

## Tasks Agent

### POST /tasks/generate-suggestions
**Request**: `{ userId, learningGoals[], currentCourses[], availableTime, preferredDifficulty, interests[], upcomingDeadlines[] }`   
**Response**: `{ suggestedTasks: [{ who, what, when, where, why, how, urgencyLevel, learningGoal, estimatedDuration, reasoning }] }` 

### POST /tasks/prioritize
**Request**: `{ userId, tasks[], userContext }`   
**Response**: `{ prioritizedTasks[], reasoning, timeAllocation: { morning[], afternoon[], evening[] }, suggestions[] }` 

### POST /tasks/break-down
**Request**: `{ task, userContext }`   
**Response**: `{ subtasks: [{ sequence, what, estimatedDuration, difficulty, resources[], checkpoints[] }] }` 

### POST /tasks/productivity-insights
**Request**: `{ userId, timeframe, completedTasks[], pendingTasks[], taskCompletionTimes[], urgencyDistribution }`   
**Response**: `{ completionRate, patterns[], strengths[], improvements[], recommendations[], optimalWorkTimes[] }` 

### POST /tasks/generate-schedule
**Request**: `{ userId, tasks[], constraints, preferences }`   
**Response**: `{ schedule[], reasoning, alternatives[] }` 

### POST /tasks/chat
**Request**: `{ message, userId, userName, userProfile, chatHistory[], tasksContext }`   
**Response**: `{ response, suggestedActions[], taskRecommendations[] }` 

### POST /tasks/motivation
**Request**: `{ userId, completedToday, streak, recentAchievements[], currentChallenges[] }`   
**Response**: `{ message }` 

---

## Professor Agent

### POST /professor-chat
**Request**: `{ activityId, lessonTitle, leader, studentMessage, studentProfile, courseDetails, notes, chatHistory[] }`   
**Response**: `{ response }` 

### POST /generate-notes
**Request**: `{ courseTitle, courseDescription, courseDifficultyLevel, lessonTitle, lessonObjective, lessonAssessment, profileName, educationLevel, careerGoals, interests }`   
**Response**: `{ notes }` 

### POST /generate-lesson
**Request**: `{ courseTitle, courseDescription, lessonTitle, lessonObjective, studentProfile, additionalContext }`   
**Response**: `{ lessonPlan: { lessonMetadata, activities[] } }` 

---

## Examiner Agent

### POST /examiner-chat
**Request**: `{ model: "deepseek-chat", messages: [{ role, content }], temperature: 0.7 }`   
**Response**: `{ response }` 

### POST /generate-exam-questions
**Request**: `{ model: "deepseek-chat", messages: [{ role, content }], temperature: 0.7 }`   
**Response**: `{ questions: [{ id, question, type, options[], correctAnswer, explanation, difficulty }] }` 

### POST /submit-exam
**Request**: `{ model: "deepseek-chat", messages: [{ role, content }], temperature: 0.3 }`   
**Response**: `{ score, totalQuestions, correctAnswers, feedback[], detailedResults[] }` 

### GET /exam-details/{examId}
**Response**: `{ examId, questions[], metadata }` 

---

## Librarian Agent

### POST /generate-resources
**Request**: `{ topic, studentProfile, topicInfo }`   
**Response**: `{ resources: [{ id, title, type, url, description, relevance }] }` 

### POST /chat-video
**Request**: `{ message, context: { transcript, currentTime, videoTitle, videoDescription, courseName } }`   
**Response**: `{ response, timestamp }` 

### POST /recommend-videos
**Request**: `{ courseId, courseTopic, userProgress }`   
**Response**: `{ recommendations: [{ videoId, title, relevance, reasoning }] }` 

### POST /summarize-video
**Request**: `{ transcript, duration }`   
**Response**: `{ summary: { overview, keyPoints[], timestamps[] } }` 

### POST /extract-concepts
**Request**: `{ transcript }`   
**Response**: `{ concepts: [{ concept, explanation, timestamp }] }` 

### POST /generate-quiz
**Request**: `{ transcript, difficulty }`   
**Response**: `{ questions: [{ id, question, type, options[], correctAnswer }] }` 

### POST /find-topic
**Request**: `{ transcript, topic }`   
**Response**: `{ results: { timestamps[], context[] } }` 

---

## Project Assistant Agent

### POST /projects/chat
**Request**: `{ message, userId, userName, chatHistory[], projectContext }`   
**Response**: `{ response, suggestions[] }` 

### POST /projects/suggestions
**Request**: `{ userId, interests[], careerGoals, completedCourses[], skillLevel }`   
**Response**: `{ suggestions: [{ id, title, description, difficulty, technologies[], estimatedHours, reasoning }] }` 

### POST /projects/help
**Request**: `{ userId, project, question, chatHistory[] }`   
**Response**: `{ response, resources[], nextSteps[] }` 

### POST /projects/break-down
**Request**: `{ userId, project }`   
**Response**: `{ milestones: [{ id, title, description, estimatedHours, tasks[], deliverables[] }] }` 

### POST /projects/resources
**Request**: `{ userId, project, milestone }`   
**Response**: `{ resources: [{ title, url, type, relevance, description }] }` 

---

## Course Generation

### POST /generate-course
**Request**: `{ courseTitle, courseDescription, difficulty, targetAudience, learningObjectives[], duration }`   
**Response**: `{ course: { metadata, units[], assessments[] } }` 

### POST /generate-exam
**Request**: `{ courseId, unitId, difficulty, questionCount, topics[] }`   
**Response**: `{ exam: { questions[], timeLimit, passingScore } }` 

---

## Error Format
```json
{ "error": { "code": "string", "message": "string", "details": "string" } }
```

## Parsing Notes
- All responses are JSON objects
- Arrays return `[]`  when empty, never `null` 
- Strings return `""`  when empty, never `null` 
- Numbers are numeric types, not strings
- ISO 8601 timestamps for all dates
