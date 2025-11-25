# Quick Start Guide

## Running the Server

```bash
cd /Users/_musket.eer/Documents/ai-university/agent-core
go run cmd/main.go
```

Expected output:
```
🚀 Server running at http://localhost:8080
```

## Testing Endpoints

### 1. Dashboard Insights

```bash
curl -X POST http://localhost:8080/dashboard/insights \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-001",
    "userName": "Alice Johnson",
    "educationLevel": "Undergraduate",
    "careerGoals": "Software Engineer",
    "interests": ["AI", "Web Development", "Python"],
    "recentActivities": [
      {"id": "act-1", "title": "Completed Python Basics", "completedAt": "2024-01-15T10:00:00Z", "type": "lesson"}
    ],
    "completedCourses": ["intro-to-python"],
    "currentCourses": ["data-structures", "web-dev-101"],
    "learningStreak": 7,
    "totalStudyTime": 3600
  }'
```

### 2. Course Search

```bash
curl -X POST http://localhost:8080/courses/intelligent-search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning for beginners",
    "userId": "user-001",
    "userProfile": {
      "name": "Alice Johnson",
      "education_level": "Undergraduate",
      "careerGoals": "Data Scientist",
      "interests": ["AI", "Python", "Statistics"]
    },
    "filters": {
      "difficulty": "beginner",
      "duration": 40
    },
    "userContext": {
      "completedCourses": ["intro-to-python"],
      "interests": ["AI"]
    }
  }'
```

### 3. Reflection Analysis

```bash
curl -X POST http://localhost:8080/reflection/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "reflection": "Today I learned about recursion. At first it was confusing, but after working through examples, I understand how functions can call themselves. I need to practice more with base cases.",
    "activity": {
      "id": "lesson-recursion",
      "title": "Understanding Recursion",
      "type": "lesson",
      "difficulty": "medium"
    },
    "userProfile": {
      "name": "Alice Johnson",
      "education_level": "Undergraduate",
      "learningStyle": "visual"
    },
    "learningGoals": ["Master algorithms", "Improve problem-solving"]
  }'
```

### 4. Task Generation

```bash
curl -X POST http://localhost:8080/tasks/generate-suggestions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-001",
    "learningGoals": ["Master Python", "Build a portfolio project"],
    "currentCourses": [
      {"id": "ds-101", "title": "Data Structures", "progress": 60}
    ],
    "availableTime": "flexible",
    "preferredDifficulty": "medium",
    "interests": ["Web Development", "APIs"],
    "upcomingDeadlines": [
      {"courseId": "ds-101", "deadline": "2024-02-01T00:00:00Z", "type": "assignment"}
    ]
  }'
```

### 5. Project Suggestions

```bash
curl -X POST http://localhost:8080/projects/suggestions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-001",
    "interests": ["Web Development", "AI", "Python"],
    "careerGoals": "Full Stack Developer",
    "completedCourses": ["intro-to-python", "html-css-basics"],
    "skillLevel": "intermediate"
  }'
```

### 6. Video Chat

```bash
curl -X POST http://localhost:8080/chat-video \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Can you explain the concept mentioned at 5:30?",
    "context": {
      "transcript": "In this video, we discuss object-oriented programming...",
      "currentTime": 330,
      "videoTitle": "Introduction to OOP",
      "videoDescription": "Learn the basics of object-oriented programming",
      "courseName": "Python Programming"
    }
  }'
```

### 7. Learning Path

```bash
curl -X POST http://localhost:8080/courses/suggest-path \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-001",
    "targetGoal": "Become a Machine Learning Engineer",
    "currentSkills": ["Python", "Basic Statistics"],
    "timeCommitment": "part-time",
    "preferredDifficulty": "progressive",
    "interests": ["AI", "Data Science", "Mathematics"]
  }'
```

### 8. Task Prioritization

```bash
curl -X POST http://localhost:8080/tasks/prioritize \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-001",
    "tasks": [
      {
        "id": "task-1",
        "what": "Complete Python assignment",
        "urgencyLevel": "high",
        "scheduledTime": "2024-01-20T14:00:00Z",
        "estimatedDuration": 120,
        "learningGoal": "Master Python",
        "dependencies": []
      },
      {
        "id": "task-2",
        "what": "Watch ML lecture",
        "urgencyLevel": "medium",
        "scheduledTime": "2024-01-20T16:00:00Z",
        "estimatedDuration": 60,
        "learningGoal": "Learn ML basics",
        "dependencies": []
      }
    ],
    "userContext": {
      "availableTimeToday": 240,
      "energyLevel": "high",
      "deadlines": [
        {"taskId": "task-1", "deadline": "2024-01-21T00:00:00Z"}
      ]
    }
  }'
```

## Testing All Endpoints

Create a test script:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "Testing Dashboard Insights..."
curl -X POST $BASE_URL/dashboard/insights -H "Content-Type: application/json" -d '{"userId":"test","userName":"Test","educationLevel":"Undergraduate","careerGoals":"Engineer","interests":[],"recentActivities":[],"completedCourses":[],"currentCourses":[],"learningStreak":0,"totalStudyTime":0}'

echo -e "\n\nTesting Course Search..."
curl -X POST $BASE_URL/courses/intelligent-search -H "Content-Type: application/json" -d '{"query":"python","userId":"test","userProfile":{"name":"Test","education_level":"Undergraduate","careerGoals":"Engineer","interests":[]},"filters":{},"userContext":{}}'

echo -e "\n\nTesting Reflection Analysis..."
curl -X POST $BASE_URL/reflection/analyze -H "Content-Type: application/json" -d '{"reflection":"Test reflection","activity":{"id":"1","title":"Test"},"userProfile":{"name":"Test","education_level":"Undergraduate"},"learningGoals":[]}'

echo -e "\n\nTesting Task Suggestions..."
curl -X POST $BASE_URL/tasks/generate-suggestions -H "Content-Type: application/json" -d '{"userId":"test","learningGoals":[],"currentCourses":[],"availableTime":"flexible","preferredDifficulty":"medium","interests":[],"upcomingDeadlines":[]}'

echo -e "\n\nTesting Project Suggestions..."
curl -X POST $BASE_URL/projects/suggestions -H "Content-Type: application/json" -d '{"userId":"test","interests":[],"careerGoals":"Engineer","completedCourses":[],"skillLevel":"intermediate"}'

echo -e "\n\nAll tests completed!"
```

Save as `test_new_endpoints.sh` and run:
```bash
chmod +x test_new_endpoints.sh
./test_new_endpoints.sh
```

## Checking Logs

```bash
tail -f api.log
```

## Common Issues

### 1. Server won't start
- Check if port 8080 is already in use
- Verify .env file exists with required credentials
- Check Redis connection (Upstash)

### 2. Endpoints return errors
- Verify request body matches expected schema
- Check prompt files exist in `configs/prompts/`
- Review logs for detailed error messages

### 3. Empty responses
- LLM may not be returning valid JSON
- Check prompt templates
- Verify LLM client configuration

### 4. Slow responses
- First request will be slower (no cache)
- Subsequent identical requests should be fast (cached)
- LLM calls typically take 2-5 seconds

## Monitoring

Check server health:
```bash
curl http://localhost:8080/health
```

View all registered routes:
```bash
# Routes are logged on server startup
grep "HandleFunc" api.log
```

## Next Steps

1. ✅ Test basic connectivity
2. ✅ Test each agent's endpoints
3. ✅ Verify error handling
4. ✅ Check caching behavior
5. ✅ Review LLM response quality
6. ✅ Adjust prompts as needed
7. ✅ Add authentication
8. ✅ Deploy to production

## Support

- API Spec: `docs/llms-payloads.md`
- Implementation Guide: `docs/API_IMPLEMENTATION.md`
- Summary: `docs/SUMMARY.md`
- Logs: `api.log`
