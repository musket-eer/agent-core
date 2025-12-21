# V2 Testing Guide

## 🧪 Testing V2 Endpoints

This guide provides complete testing instructions for V2 endpoints.

---

## 🚀 Quick Start

### 1. Start the Server

```bash
cd /Users/_musket.eer/Documents/ai-university/agent-core
go run cmd/main.go
```

Expected output:
```
🚀 Initializing V2 feature flags...
✅ V2 Feature Flags initialized (disabled by default)
🚀 Server running at http://localhost:8080
📊 V2 Status: enabled=false, rollout=0.0%
```

### 2. Enable V2 for Testing

```bash
curl -X POST http://localhost:8080/v2/feature-flags \
  -H "Content-Type: application/json" \
  -d '{
    "quickRolloutStage": "testing"
  }'
```

This enables V2 for 1% of users.

### 3. Test a V2 Endpoint

```bash
curl -X POST http://localhost:8080/tasks/v2/generate-suggestions \
  -H "Content-Type: application/json" \
  -d @test-payloads/tasks-generate-suggestions.json
```

---

## 📦 Test Payloads

### Tasks V2 - Generate Suggestions

**File**: `test-payloads/tasks-generate-suggestions.json`

```json
{
  "agentId": "tasks",
  "userId": "test-user-123",
  "sessionId": "session-456",
  "componentType": "task-generator",
  "context": {
    "userInstructions": "Be concise and action-oriented. Focus on practical projects.",
    "sessionMetadata": "Device: Desktop | Browser: Chrome | Time: Morning",
    "userMemory": "Name: Sarah | Learning Style: Visual | Preference: Hands-on projects",
    "conversationSummaries": "Last week: Completed React hooks introduction. Currently learning state management.",
    "currentSession": "",
    "componentContext": "Goal: Build portfolio projects to showcase React skills"
  },
  "messages": [],
  "requestData": {
    "learningGoals": ["Master React", "Build portfolio"],
    "availableTime": "2 hours",
    "currentCourses": ["React Advanced Patterns"]
  }
}
```

**Expected Response**:
```json
{
  "success": true,
  "data": {
    "suggestedTasks": [
      {
        "id": "task-1",
        "title": "Build Todo App with React Hooks",
        "description": "Practice useState and useEffect",
        "estimatedTime": 120,
        "priority": "high",
        "reasoning": "Builds on recent learning"
      }
    ],
    "personalizationNotes": "Tasks tailored for visual learners"
  },
  "metadata": {
    "tokensUsed": 150,
    "processingTime": 0.5,
    "cacheHit": false,
    "contextSources": ["userMemory", "conversationSummaries", "componentContext"]
  }
}
```

---

### Reflection V2 - Quick Reflection

**File**: `test-payloads/reflection-quick-reflection.json`

```json
{
  "agentId": "reflection",
  "userId": "test-user-123",
  "sessionId": "session-789",
  "componentType": "task-reflection",
  "componentId": "task-456",
  "context": {
    "userInstructions": "Keep reflections under 50 words. Focus on actionable insights.",
    "sessionMetadata": "Device: Mobile | Time: Evening",
    "userMemory": "Name: Mike | Reflection Style: Analytical | Prefers data-driven insights",
    "conversationSummaries": "Pattern observed: Underestimates UI work by 20-30% consistently",
    "currentSession": "",
    "componentContext": "Task: Build React Dashboard | Complexity: High | Estimated: 2h | Actual: 2.5h"
  },
  "messages": [],
  "requestData": {
    "taskId": "task-456",
    "taskInfo": {
      "title": "Build React Dashboard",
      "description": "Create admin dashboard with charts",
      "priority": "high"
    },
    "timeData": {
      "estimatedMinutes": 120,
      "timeSpent": 9000,
      "startedAt": "2024-12-10T14:00:00Z",
      "completedAt": "2024-12-10T16:30:00Z"
    },
    "completionData": {
      "completed": true,
      "taskOutcome": "success",
      "completionPercentage": 100,
      "productivityRating": 8
    },
    "userResponses": {
      "whatWentWell": "Chart integration was smooth",
      "whatWasChallenging": "Responsive layout took longer than expected",
      "whatILearned": "CSS Grid is powerful for dashboards"
    }
  }
}
```

**Expected Response**:
```json
{
  "success": true,
  "data": {
    "reflection": "30-min overrun on UI work—consistent with your 25% underestimation pattern.",
    "insights": [
      "Time estimation accuracy: 80% (improving from 70% last month)"
    ],
    "patterns": [
      "Underestimates UI complexity (4th occurrence this month)"
    ],
    "recommendations": [
      "Apply 30% buffer to all UI estimation (data-backed)"
    ],
    "depthScore": 8.2,
    "emotionalTone": "positive",
    "keyLearnings": ["CSS Grid for dashboards"],
    "challengesFaced": ["Responsive layout complexity"],
    "solutionsApplied": ["Used CSS Grid"],
    "skillsPracticed": ["React", "CSS Grid", "Chart.js"]
  },
  "metadata": {
    "tokensUsed": 220,
    "processingTime": 0.8,
    "cacheHit": false,
    "contextSources": ["userMemory", "conversationSummaries", "componentContext"],
    "personalizationApplied": ["analytical-tone", "data-driven-insights"]
  }
}
```

---

## 🔧 Feature Flags Testing

### Check Current Status

```bash
curl http://localhost:8080/v2/feature-flags/stats
```

**Response**:
```json
{
  "success": true,
  "data": {
    "v2Enabled": true,
    "rolloutPercentage": 1.0,
    "enabledAgents": ["tasks", "reflection"],
    "userOverridesCount": 0
  }
}
```

### Enable V2 for Specific User

```bash
curl -X POST http://localhost:8080/v2/feature-flags \
  -H "Content-Type: application/json" \
  -d '{
    "v2Enabled": true,
    "rolloutPercentage": 0,
    "userOverride": {
      "userId": "test-user-123",
      "useV2": true
    }
  }'
```

### Simulate Rollout Distribution

```bash
curl -X POST http://localhost:8080/v2/feature-flags/simulate \
  -H "Content-Type: application/json" \
  -d '{
    "numUsers": 1000
  }'
```

**Response**:
```json
{
  "success": true,
  "data": {
    "v1Users": 990,
    "v2Users": 10,
    "totalUsers": 1000,
    "v2Percentage": 1,
    "targetPercentage": 1
  }
}
```

### Quick Rollout Stages

```bash
# Testing (1%)
curl -X POST http://localhost:8080/v2/feature-flags \
  -d '{"quickRolloutStage": "testing"}'

# Pilot (10%)
curl -X POST http://localhost:8080/v2/feature-flags \
  -d '{"quickRolloutStage": "pilot"}'

# Beta (50%)
curl -X POST http://localhost:8080/v2/feature-flags \
  -d '{"quickRolloutStage": "beta"}'

# Full (100%)
curl -X POST http://localhost:8080/v2/feature-flags \
  -d '{"quickRolloutStage": "full"}'

# Off (0%)
curl -X POST http://localhost:8080/v2/feature-flags \
  -d '{"quickRolloutStage": "off"}'
```

---

## 📊 All V2 Endpoints

### Tasks Agent V2

1. **Generate Suggestions**
   ```bash
   POST /tasks/v2/generate-suggestions
   ```

2. **Prioritize Tasks**
   ```bash
   POST /tasks/v2/prioritize
   ```

3. **Break Down Task**
   ```bash
   POST /tasks/v2/break-down
   ```

4. **Get Insights**
   ```bash
   POST /tasks/v2/insights
   ```

5. **Chat**
   ```bash
   POST /tasks/v2/chat
   ```

### Reflection Agent V2

1. **Generate Questions**
   ```bash
   POST /reflection/v2/generate-questions
   ```

2. **Generate Journal**
   ```bash
   POST /reflection/v2/generate-journal
   ```

3. **Get Insights**
   ```bash
   POST /reflection/v2/insights
   ```

4. **Quick Reflection**
   ```bash
   POST /reflection/v2/quick-reflection
   ```

5. **Chat**
   ```bash
   POST /reflection/v2/chat
   ```

### Feature Flags

1. **Get/Update Flags**
   ```bash
   GET/POST /v2/feature-flags
   ```

2. **Get Stats**
   ```bash
   GET /v2/feature-flags/stats
   ```

3. **Simulate Rollout**
   ```bash
   POST /v2/feature-flags/simulate
   ```

---

## ✅ Testing Checklist

### Basic Functionality
- [ ] Server starts without errors
- [ ] V2 endpoints are registered
- [ ] Feature flags initialize correctly
- [ ] Can enable/disable V2 via API
- [ ] Can query feature flag stats

### Tasks V2
- [ ] Generate suggestions returns placeholder response
- [ ] Prioritize returns placeholder response
- [ ] Break down returns placeholder response
- [ ] Insights returns placeholder response
- [ ] Chat returns placeholder response
- [ ] All endpoints return V2 metadata

### Reflection V2
- [ ] Generate questions returns placeholder response
- [ ] Generate journal returns placeholder response
- [ ] Insights returns placeholder response
- [ ] Quick reflection returns placeholder response
- [ ] Chat returns placeholder response
- [ ] All endpoints return V2 metadata

### Feature Flags
- [ ] Can check current status
- [ ] Can enable V2 globally
- [ ] Can set rollout percentage
- [ ] Can simulate rollout distribution
- [ ] Quick rollout stages work
- [ ] Stats endpoint returns correct data

### Context Processing
- [ ] Context is extracted from request
- [ ] Context sources are tracked in metadata
- [ ] Token estimation works
- [ ] Processing time is tracked

---

## 🐛 Troubleshooting

### Server Won't Start

**Issue**: Import errors or compilation failures

**Solution**: 
```bash
# Check Go version
go version

# Verify all dependencies
go mod tidy

# Try building
go build ./cmd/main.go
```

### V2 Endpoints Return 404

**Issue**: Routes not registered

**Solution**: Check server logs for "V2 endpoints registered successfully"

### Feature Flags Not Working

**Issue**: Feature flags not initialized

**Solution**: Check server logs for "Initializing V2 feature flags..."

### Context Not Being Used

**Issue**: Context extraction failing

**Solution**: Verify request payload has `context` object with all fields

---

## 📈 Next Steps

1. ✅ **Test all endpoints** with curl
2. ⏳ **Implement actual LLM calls** (replace placeholders)
3. ⏳ **Add prompt templates** for V2 endpoints
4. ⏳ **Test with real user data**
5. ⏳ **Monitor performance metrics**
6. ⏳ **Deploy to staging**
7. ⏳ **Gradual rollout to production**

---

## 💡 Tips

- Use `jq` to format JSON responses: `curl ... | jq`
- Save test payloads in `test-payloads/` directory
- Monitor server logs for debugging
- Test feature flags before testing endpoints
- Start with small rollout percentages
- Keep V1 endpoints working during migration

---

This testing guide covers all V2 functionality. Follow the checklist to ensure everything works before deployment.
