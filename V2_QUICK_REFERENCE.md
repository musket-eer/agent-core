# V2 Quick Reference Card

## 🎯 Current Status: Phase 1 Complete ✅

---

## 📦 What You Have Now

### Working V2 Endpoints (13 total)
```
Tasks V2:
  POST /tasks/v2/generate-suggestions
  POST /tasks/v2/prioritize
  POST /tasks/v2/break-down
  POST /tasks/v2/insights
  POST /tasks/v2/chat

Reflection V2:
  POST /reflection/v2/generate-questions
  POST /reflection/v2/generate-journal
  POST /reflection/v2/insights
  POST /reflection/v2/quick-reflection
  POST /reflection/v2/chat

Feature Flags:
  GET/POST /v2/feature-flags
  GET /v2/feature-flags/stats
  POST /v2/feature-flags/simulate
```

### Key Files
```
Types:           internal/llm/v2/types/*.go
Middleware:      internal/llm/v2/middleware/*.go
Handlers:        internal/llm/v2/handlers/*.go
Config:          internal/llm/configs/deepseek_configs.go
Router:          internal/llm/routers.go
Main:            cmd/main.go
Test Payloads:   test-payloads/*.json
Documentation:   V2_*.md
```

---

## 🚀 Quick Start Commands

### Start Server
```bash
cd /Users/_musket.eer/Documents/ai-university/agent-core
go run cmd/main.go
```

### Enable V2 (Testing Mode - 1%)
```bash
curl -X POST http://localhost:8080/v2/feature-flags \
  -H "Content-Type: application/json" \
  -d '{"quickRolloutStage": "testing"}'
```

### Test Tasks Endpoint
```bash
curl -X POST http://localhost:8080/tasks/v2/generate-suggestions \
  -H "Content-Type: application/json" \
  -d @test-payloads/tasks-generate-suggestions.json | jq
```

### Test Reflection Endpoint
```bash
curl -X POST http://localhost:8080/reflection/v2/quick-reflection \
  -H "Content-Type: application/json" \
  -d @test-payloads/reflection-quick-reflection.json | jq
```

### Check Feature Flags
```bash
curl http://localhost:8080/v2/feature-flags/stats | jq
```

---

## 🎛️ Feature Flag Control

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

### Manual Control
```bash
# Enable V2 with custom percentage
curl -X POST http://localhost:8080/v2/feature-flags \
  -H "Content-Type: application/json" \
  -d '{
    "v2Enabled": true,
    "rolloutPercentage": 25.0,
    "enabledAgents": ["tasks", "reflection"]
  }'
```

---

## 📊 V2 Request Structure

```json
{
  "agentId": "tasks",
  "userId": "user-123",
  "sessionId": "session-456",
  "componentType": "task-generator",
  "context": {
    "userInstructions": "User preferences",
    "sessionMetadata": "Device, browser, time",
    "userMemory": "User profile, learning style",
    "conversationSummaries": "Recent conversations",
    "currentSession": "Current conversation",
    "componentContext": "Current task/activity"
  },
  "messages": [],
  "requestData": {
    // Endpoint-specific data
  }
}
```

---

## 📊 V2 Response Structure

```json
{
  "success": true,
  "data": {
    // Endpoint-specific response
  },
  "metadata": {
    "tokensUsed": 150,
    "processingTime": 0.5,
    "cacheHit": false,
    "contextSources": ["userMemory", "conversationSummaries"],
    "personalizationApplied": ["visual-learner", "concise-format"]
  }
}
```

---

## ⏳ Next Steps (Phase 2)

### 1. Create V2 Prompts
```bash
# Create prompt files in configs/prompts/
tasksSuggestionsPromptV2.txt
tasksPrioritizePromptV2.txt
tasksBreakDownPromptV2.txt
tasksInsightsPromptV2.txt
reflectionQuestionsPromptV2.txt
reflectionJournalPromptV2.txt
reflectionInsightsPromptV2.txt
reflectionQuickReflectionPromptV2.txt
```

### 2. Implement LLM Calls
Replace placeholder responses in handlers with actual LLM calls:
- Load prompt with `prompt.GetPrompt()`
- Build enriched prompt with `middleware.BuildSystemPrompt()`
- Call LLM with `client.Call()`
- Parse and return response

### 3. Test Thoroughly
- Unit tests for context processor
- Integration tests for all endpoints
- Compare V1 vs V2 responses
- Load testing

### 4. Deploy
- Deploy to staging
- Enable for 1% of users
- Monitor metrics
- Gradual rollout

---

## 🐛 Troubleshooting

### Server won't start
```bash
# Check Go installation
go version

# Install dependencies
go mod tidy

# Build to check for errors
go build ./cmd/main.go
```

### V2 endpoints return 404
Check server logs for "V2 endpoints registered successfully"

### Feature flags not working
Check server logs for "Initializing V2 feature flags..."

### Context not being used
Verify request payload has complete `context` object

---

## 📚 Documentation Files

- **`V2_README.md`** - Start here (overview & getting started)
- **`V2_TESTING_GUIDE.md`** - Complete testing instructions
- **`V2_IMPLEMENTATION_STATUS.md`** - Current status & progress
- **`V2_MIGRATION_PLAN.md`** - Detailed migration strategy
- **`V2_IMPLEMENTATION_GUIDE.md`** - Step-by-step implementation
- **`V2_SUMMARY.md`** - Quick summary & next actions
- **`V2_QUICK_REFERENCE.md`** - This file

---

## 💡 Key Points

✅ **Phase 1 Complete** - Foundation is ready
⏳ **Phase 2 Next** - Implement actual LLM calls
🔒 **Safe Migration** - V1 continues working
🎚️ **Gradual Rollout** - Feature flags control
📊 **Rich Context** - Full memory integration
🔄 **Easy Rollback** - Can disable V2 instantly

---

## 🎯 Success Metrics

**Foundation (Phase 1)** ✅
- [x] 13 V2 endpoints registered
- [x] Feature flags system working
- [x] Context processor implemented
- [x] Documentation complete

**LLM Integration (Phase 2)** ⏳
- [ ] All handlers call real LLM
- [ ] All prompts created
- [ ] Responses properly formatted

**Testing (Phase 3)** ⏳
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] Performance acceptable

**Deployment (Phase 4)** ⏳
- [ ] Deployed to staging
- [ ] Gradual rollout complete
- [ ] V1 deprecated

---

## 🚀 Ready to Continue?

**Current**: Phase 1 Complete ✅
**Next**: Phase 2 - LLM Integration

Start by creating V2 prompt templates, then implement actual LLM calls in the handlers.

---

**Quick Tip**: Keep this file open as a reference while working on Phase 2!
