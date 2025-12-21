# V2 Implementation Guide - Complete Reference

## 🎯 Quick Start

This guide provides complete implementation details for V2 endpoints with memory architecture integration.

---

## 📋 Implementation Checklist

### Phase 1: Foundation Setup ✅

- [x] Create V2 directory structure
- [x] Define common V2 types (`common_v2.go`)
- [x] Define Tasks V2 types (`tasks_v2.go`)
- [x] Define Reflection V2 types (`reflection_v2.go`)
- [x] Build context processor (`context_processor.go`)
- [x] Build feature flags system (`feature_flags.go`)
- [ ] Create V2 router integration
- [ ] Add V2 config entries

### Phase 2: Handler Implementation

- [ ] Implement Tasks V2 handlers
- [ ] Implement Reflection V2 handlers
- [ ] Add proper error handling
- [ ] Add logging and monitoring
- [ ] Add caching support

### Phase 3: Testing

- [ ] Unit tests for context processor
- [ ] Unit tests for feature flags
- [ ] Integration tests for Tasks V2
- [ ] Integration tests for Reflection V2
- [ ] Load testing

### Phase 4: Deployment

- [ ] Update docker-compose.yml
- [ ] Add V2 environment variables
- [ ] Deploy to staging
- [ ] Enable for 1% of users
- [ ] Monitor and iterate

---

## 🏗️ Architecture Overview

```
Request Flow:
1. Frontend compiles context from memory system
2. Frontend sends V2 request with full context
3. Backend receives V2 request
4. Feature flags check if user should use V2
5. Context processor extracts and formats context
6. Base prompt + context = enriched system prompt
7. LLM called with enriched prompt
8. Response formatted with V2 metadata
9. Response cached for future requests
10. Frontend receives V2 response
```

---

## 📦 Request/Response Patterns

### Pattern 1: Generate Suggestions (Tasks)

**Frontend Request:**
```javascript
const response = await tasksAgentServiceV2.generateTaskSuggestions({
  userId: 'user-123',
  learningGoals: ['Master React'],
  availableTime: '2 hours',
});
```

**Backend Receives:**
```json
{
  "agentId": "tasks",
  "userId": "user-123",
  "sessionId": "session-456",
  "componentType": "task-generator",
  "context": {
    "userInstructions": "Be concise. Focus on practical projects.",
    "sessionMetadata": "Device: Desktop | Browser: Chrome",
    "userMemory": "Name: Sarah | Learning Style: Visual",
    "conversationSummaries": "Last week: Completed React hooks intro",
    "currentSession": "",
    "componentContext": "Goal: Build portfolio projects"
  },
  "messages": [...],
  "requestData": {
    "learningGoals": ["Master React"],
    "availableTime": "2 hours"
  }
}
```

**Backend Processing:**
1. Check feature flag for user-123
2. Extract context from payload
3. Load base prompt: `tasksSuggestionsPrompt`
4. Build enriched prompt: `basePrompt + context`
5. Call DeepSeek with enriched prompt
6. Parse response
7. Add metadata (tokens, processing time, context sources)
8. Return V2 response

**Backend Returns:**
```json
{
  "success": true,
  "data": {
    "suggestedTasks": [
      {
        "id": "task-1",
        "title": "Build Todo App with Hooks",
        "description": "Practice useEffect with a todo app",
        "estimatedTime": 120,
        "priority": "high",
        "reasoning": "Builds on your recent useEffect learning",
        "visualResources": ["React hooks diagram"]
      }
    ],
    "personalizationNotes": "Tasks tailored for visual learners"
  },
  "metadata": {
    "tokensUsed": 1450,
    "processingTime": 2.1,
    "cacheHit": false,
    "contextSources": ["userMemory", "conversationSummaries", "componentContext"]
  }
}
```

---

### Pattern 2: Quick Reflection (Reflection)

**Frontend Request:**
```javascript
const response = await reflectionAgentServiceV2.generateQuickReflection({
  userId: 'user-123',
  taskId: 'task-789',
  taskInfo: {...},
  timeData: {...},
  completionData: {...},
  userResponses: {...}
});
```

**Backend Receives:**
```json
{
  "agentId": "reflection",
  "userId": "user-123",
  "sessionId": "session-789",
  "componentType": "task-reflection",
  "componentId": "task-789",
  "context": {
    "userInstructions": "Keep reflections under 50 words",
    "sessionMetadata": "Device: Mobile | Time: Evening",
    "userMemory": "Name: Mike | Reflection Style: Analytical",
    "conversationSummaries": "Pattern: Underestimates UI work by 20-30%",
    "currentSession": "",
    "componentContext": "Task: Build React Dashboard | Complexity: High"
  },
  "messages": [...],
  "requestData": {
    "taskInfo": {...},
    "timeData": {...},
    "completionData": {...}
  }
}
```

**Backend Returns:**
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
    "trendAnalysis": {
      "estimationAccuracy": "Improving: 70% → 80% over 30 days"
    }
  },
  "metadata": {
    "tokensUsed": 1820,
    "processingTime": 2.8,
    "cacheHit": false,
    "contextSources": ["userMemory", "conversationSummaries", "componentContext"],
    "personalizationApplied": ["analytical-tone", "data-driven-insights"]
  }
}
```

---

## 🔧 Implementation Steps

### Step 1: Fix Import Issues

The handlers need to import from the correct package. Update imports:

```go
// internal/llm/v2/handlers/tasks_v2.go
import (
	"agent-core/internal/llm"        // For GetLLMClient
	"agent-core/internal/llm/configs" // For GetDeepseekConfig
	"agent-core/internal/llm/utils"   // For WriteJSONError, CleanJSONResponse
	"agent-core/internal/cache"       // For cache
)

// Load prompts using the existing prompt package
basePrompt, err := llm.LoadPrompt("tasksSuggestionsPrompt")
```

### Step 2: Create Prompt Loader Helper

Add to `internal/llm/helpers.go`:

```go
package llm

import (
	"os"
	"path/filepath"
)

// LoadPrompt loads a prompt from the configs/prompts directory
func LoadPrompt(promptName string) (string, error) {
	promptPath := filepath.Join("configs", "prompts", promptName + ".txt")
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
```

### Step 3: Update Main Router

Modify `internal/llm/routers.go` to include V2 routes:

```go
func RegisterHandlers() http.Handler {
	mux := http.NewServeMux()
	
	// ... existing V1 routes ...
	
	// Register V2 routes
	RegisterV2Handlers(mux)
	
	return mux
}
```

### Step 4: Add V2 Configs

Add to `internal/llm/configs/deepseek_configs.go`:

```go
// V2 Tasks endpoints
"/tasks/v2/generate-suggestions": {
	ModelId:     "deepseek-chat",
	Temperature: 0.7,
	MaxTokens:   2048,
	ClientName:  "general",
},
"/tasks/v2/prioritize": {
	ModelId:     "deepseek-chat",
	Temperature: 0.7,
	MaxTokens:   2048,
	ClientName:  "general",
},
// ... add all V2 endpoints ...
```

### Step 5: Initialize Feature Flags

Add to `cmd/main.go`:

```go
import (
	"agent-core/internal/llm/v2/middleware"
)

func main() {
	// ... existing setup ...
	
	// Initialize V2 feature flags
	featureFlags := middleware.InitFeatureFlags()
	
	// Start with V2 disabled (enable gradually)
	featureFlags.SetV2Enabled(false)
	featureFlags.SetV2RolloutPercentage(0.0)
	
	// Enable specific agents for testing
	// featureFlags.EnableAgentV2("tasks")
	// featureFlags.EnableAgentV2("reflection")
	
	// ... rest of setup ...
}
```

### Step 6: Add Environment Variables

Update `.env`:

```bash
# V2 Feature Flags
V2_ENABLED=false
V2_ROLLOUT_PERCENTAGE=0.0
V2_ENABLED_AGENTS=tasks,reflection
```

Update `docker-compose.yml`:

```yaml
environment:
  - V2_ENABLED=${V2_ENABLED:-false}
  - V2_ROLLOUT_PERCENTAGE=${V2_ROLLOUT_PERCENTAGE:-0.0}
  - V2_ENABLED_AGENTS=${V2_ENABLED_AGENTS:-}
```

---

## 🧪 Testing Strategy

### Unit Tests

```go
// internal/llm/v2/middleware/context_processor_test.go
package middleware

import (
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	basePrompt := "You are a helpful assistant."
	context := ContextData{
		UserMemory: "Name: John | Learning Style: Visual",
		UserInstructions: "Be concise",
	}
	
	result := BuildSystemPrompt(basePrompt, context)
	
	if !strings.Contains(result, "Name: John") {
		t.Error("Expected user memory in prompt")
	}
	if !strings.Contains(result, "Be concise") {
		t.Error("Expected user instructions in prompt")
	}
}

func TestFeatureFlags(t *testing.T) {
	ff := InitFeatureFlags()
	ff.SetV2Enabled(true)
	ff.SetV2RolloutPercentage(50.0)
	
	// Test rollout distribution
	v2Count := 0
	totalUsers := 1000
	
	for i := 0; i < totalUsers; i++ {
		userId := fmt.Sprintf("user-%d", i)
		if ff.ShouldUseV2(userId, "tasks") {
			v2Count++
		}
	}
	
	percentage := float64(v2Count) / float64(totalUsers) * 100
	if percentage < 45 || percentage > 55 {
		t.Errorf("Expected ~50%% rollout, got %.1f%%", percentage)
	}
}
```

### Integration Tests

```bash
# Test V2 endpoint
curl -X POST http://localhost:8080/tasks/v2/generate-suggestions \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "tasks",
    "userId": "test-user",
    "context": {
      "userInstructions": "Be concise",
      "userMemory": "Name: Test | Learning Style: Visual"
    },
    "messages": [],
    "requestData": {
      "learningGoals": ["Master React"],
      "availableTime": "2 hours"
    }
  }'
```

---

## 📊 Monitoring

### Key Metrics to Track

1. **V2 Adoption Rate**
   - % of requests using V2
   - V2 vs V1 usage over time

2. **Performance Metrics**
   - Average response time (V2 vs V1)
   - Cache hit rate
   - Token usage

3. **Quality Metrics**
   - User satisfaction scores
   - Context utilization rate
   - Personalization effectiveness

4. **Error Metrics**
   - V2 error rate
   - Rollback frequency
   - Failed requests

### Logging

Add structured logging:

```go
log.Printf("[V2Metrics] endpoint=%s userId=%s tokens=%d time=%.2fs cacheHit=%v contextSources=%v",
	endpoint, userId, tokensUsed, processingTime, cacheHit, contextSources)
```

---

## 🚀 Deployment Checklist

### Pre-Deployment

- [ ] All V2 types defined
- [ ] All V2 handlers implemented
- [ ] Context processor tested
- [ ] Feature flags tested
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] Documentation complete

### Deployment Steps

1. **Deploy to Staging**
   ```bash
   docker-compose -f docker-compose.staging.yml up -d
   ```

2. **Enable V2 for Testing (1%)**
   ```bash
   curl -X POST http://staging-api/v2/feature-flags \
     -d '{"v2Enabled": true, "rolloutPercentage": 1.0}'
   ```

3. **Monitor for 24 Hours**
   - Check error rates
   - Compare V1 vs V2 performance
   - Gather user feedback

4. **Gradual Rollout**
   - Day 1: 1% of users
   - Day 3: 10% of users
   - Week 2: 50% of users
   - Week 4: 100% of users

5. **Deprecate V1**
   - Week 6: Announce V1 deprecation
   - Week 8: Remove V1 endpoints

---

## 🔄 Rollback Procedure

If issues arise:

1. **Immediate Rollback**
   ```bash
   curl -X POST http://api/v2/feature-flags \
     -d '{"v2Enabled": false}'
   ```

2. **Investigate Issues**
   - Check logs for errors
   - Review performance metrics
   - Gather user feedback

3. **Fix and Redeploy**
   - Fix identified issues
   - Test thoroughly
   - Gradual re-enable

---

## 📚 Next Steps

1. **Review this guide** with the team
2. **Fix import issues** in handlers
3. **Add prompt loader** helper function
4. **Update main router** to include V2
5. **Add V2 configs** for all endpoints
6. **Initialize feature flags** in main.go
7. **Test locally** with curl
8. **Deploy to staging**
9. **Enable for 1% of users**
10. **Monitor and iterate**

---

## 💡 Tips & Best Practices

### Context Management
- Keep context concise (< 2000 tokens)
- Prioritize recent/relevant context
- Truncate old session messages

### Caching
- Cache V2 responses by full context
- Invalidate cache when context changes
- Monitor cache hit rates

### Feature Flags
- Use gradual rollout (1% → 10% → 50% → 100%)
- Monitor metrics at each stage
- Keep rollback option ready

### Error Handling
- Log all V2 errors with context
- Fallback to V1 on V2 failure
- Alert on high error rates

### Performance
- Optimize context compilation
- Use async processing where possible
- Monitor token usage

---

## 🆘 Troubleshooting

### Issue: Import errors in handlers
**Solution**: Update imports to use existing llm package functions

### Issue: Feature flags not working
**Solution**: Ensure InitFeatureFlags() is called in main.go

### Issue: Context not being used
**Solution**: Check context extraction in middleware

### Issue: High latency
**Solution**: Optimize context size, enable caching

### Issue: Inconsistent rollout
**Solution**: Verify user ID hashing function

---

This guide provides everything needed to implement V2 endpoints. Follow the steps sequentially and test thoroughly at each stage.
