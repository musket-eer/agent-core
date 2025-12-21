# V2 Memory Architecture - Complete Implementation Package

## 🎯 Overview

This package contains everything needed to migrate from V1 to V2 endpoints with full memory architecture integration. The migration is **gradual, parallel, and safe** - V1 continues working while V2 is built and tested.

---

## 📦 What's Included

### Documentation
1. **`V2_MIGRATION_PLAN.md`** - Complete migration strategy with phases, timelines, and examples
2. **`V2_IMPLEMENTATION_GUIDE.md`** - Step-by-step implementation guide with code examples
3. **`V2_SUMMARY.md`** - Quick reference and current status
4. **`V2_README.md`** (this file) - Getting started guide

### Code Files
5. **`internal/llm/v2/types/common_v2.go`** - Common V2 types
6. **`internal/llm/v2/types/tasks_v2.go`** - Tasks agent types
7. **`internal/llm/v2/types/reflection_v2.go`** - Reflection agent types
8. **`internal/llm/v2/middleware/context_processor.go`** - Context utilities
9. **`internal/llm/v2/middleware/feature_flags.go`** - Feature flag system
10. **`internal/llm/v2/handlers/tasks_v2.go`** - Tasks handlers (template)
11. **`internal/llm/routers_v2.go`** - V2 router

---

## 🚀 Quick Start

### Step 1: Review the Plan
Read `V2_MIGRATION_PLAN.md` to understand:
- Overall architecture
- Request/response structure
- Migration phases
- Timeline

### Step 2: Review the Implementation Guide
Read `V2_IMPLEMENTATION_GUIDE.md` for:
- Detailed implementation steps
- Code examples
- Testing strategy
- Troubleshooting

### Step 3: Fix Import Issues
The template handlers have some import issues that need to be fixed:

```bash
# Create prompt loader helper
# See V2_IMPLEMENTATION_GUIDE.md Step 2
```

### Step 4: Complete Handlers
Finish implementing the handler functions:
- Tasks V2 handlers (5 endpoints)
- Reflection V2 handlers (5 endpoints)
- Feature flags handlers (2 endpoints)

### Step 5: Test Locally
```bash
# Start the server
go run cmd/main.go

# Test V2 endpoint
curl -X POST http://localhost:8080/tasks/v2/generate-suggestions \
  -H "Content-Type: application/json" \
  -d @test_payload.json
```

### Step 6: Deploy and Rollout
Follow the gradual rollout strategy:
- Week 1: 1% of users
- Week 2: 10% of users
- Week 4: 50% of users
- Week 6: 100% of users

---

## 🔑 Key Concepts

### V2 Request Structure

```json
{
  "agentId": "tasks",
  "userId": "user-123",
  "sessionId": "session-456",
  "componentType": "task-generator",
  "context": {
    "userInstructions": "Be concise",
    "sessionMetadata": "Device: Desktop",
    "userMemory": "Name: Sarah | Learning Style: Visual",
    "conversationSummaries": "Recent: Completed React hooks",
    "currentSession": "",
    "componentContext": "Goal: Build portfolio"
  },
  "messages": [...],
  "requestData": {...}
}
```

### V2 Response Structure

```json
{
  "success": true,
  "data": {...},
  "metadata": {
    "tokensUsed": 1450,
    "processingTime": 2.1,
    "cacheHit": false,
    "contextSources": ["userMemory", "conversationSummaries"]
  }
}
```

---

## 📊 V1 vs V2 Comparison

### V1 (Current)
- ❌ No memory integration
- ❌ No user preferences
- ❌ No conversation history
- ❌ No session awareness
- ❌ No personalization tracking
- ✅ Simple request/response
- ✅ Fast implementation

### V2 (New)
- ✅ Full memory integration
- ✅ User preferences included
- ✅ Conversation history
- ✅ Session awareness
- ✅ Personalization tracking
- ✅ Metadata for monitoring
- ✅ Gradual rollout support

---

## 🔄 Migration Flow

```
1. Frontend compiles context from memory system
   ↓
2. Frontend sends V2 request with full context
   ↓
3. Backend checks feature flags (should user use V2?)
   ↓
4. Backend extracts context from request
   ↓
5. Backend loads base prompt template
   ↓
6. Backend builds enriched prompt (base + context)
   ↓
7. Backend calls LLM with enriched prompt
   ↓
8. Backend formats response with V2 metadata
   ↓
9. Backend caches response
   ↓
10. Frontend receives V2 response
```

---

## 📋 Implementation Checklist

### Foundation ✅
- [x] V2 directory structure
- [x] Common V2 types
- [x] Tasks V2 types
- [x] Reflection V2 types
- [x] Context processor
- [x] Feature flags system
- [x] Documentation

### Implementation ⏳
- [ ] Fix import issues
- [ ] Create prompt loader helper
- [ ] Complete Tasks V2 handlers
- [ ] Complete Reflection V2 handlers
- [ ] Add V2 configs
- [ ] Initialize feature flags
- [ ] Update main router

### Testing ⏳
- [ ] Unit tests for context processor
- [ ] Unit tests for feature flags
- [ ] Integration tests for Tasks V2
- [ ] Integration tests for Reflection V2
- [ ] Load testing

### Deployment ⏳
- [ ] Deploy to staging
- [ ] Enable for 1% of users
- [ ] Monitor metrics
- [ ] Gradual rollout to 100%
- [ ] Deprecate V1

---

## 🎯 Success Metrics

### Performance
- V2 latency within 10% of V1
- Cache hit rate >60%
- Error rate <0.5%

### Quality
- Context utilized in >80% of responses
- User satisfaction >4.5/5
- Personalization visible

### Adoption
- Smooth rollout without incidents
- <5% rollback rate
- 100% adoption by Week 8

---

## 🔧 Common Tasks

### Enable V2 for Testing
```go
featureFlags := middleware.GetFeatureFlags()
featureFlags.SetV2Enabled(true)
featureFlags.SetV2RolloutPercentage(1.0)
featureFlags.EnableAgentV2("tasks")
```

### Test V2 Endpoint
```bash
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

### Check Feature Flag Stats
```bash
curl http://localhost:8080/v2/feature-flags/stats
```

### Rollback to V1
```bash
curl -X POST http://localhost:8080/v2/feature-flags \
  -d '{"v2Enabled": false}'
```

---

## 🆘 Troubleshooting

### Import Errors
**Problem**: `could not import agent-core/internal/llm/prompt`  
**Solution**: Create `internal/llm/helpers.go` with `LoadPrompt()` function

### Import Cycle
**Problem**: `import cycle not allowed` in routers_v2.go  
**Solution**: Move router to separate package or use interface

### Undefined Handlers
**Problem**: Reflection handlers not found  
**Solution**: Create stub implementations in `reflection_v2.go`

### Feature Flags Not Working
**Problem**: V2 not being used  
**Solution**: Ensure `InitFeatureFlags()` is called in `main.go`

---

## 📚 Documentation Structure

```
V2_README.md (this file)
├── Quick start guide
├── Key concepts
└── Common tasks

V2_MIGRATION_PLAN.md
├── Overall strategy
├── Request/response examples
├── Migration phases
└── Timeline

V2_IMPLEMENTATION_GUIDE.md
├── Step-by-step implementation
├── Code examples
├── Testing strategy
└── Deployment checklist

V2_SUMMARY.md
├── Current status
├── What's done
├── What's next
└── Known issues
```

---

## 🚀 Next Steps

1. **Read `V2_MIGRATION_PLAN.md`** - Understand the overall strategy
2. **Read `V2_IMPLEMENTATION_GUIDE.md`** - Follow implementation steps
3. **Fix import issues** - Create helpers.go
4. **Complete handlers** - Implement all endpoints
5. **Test locally** - Verify functionality
6. **Deploy to staging** - Test with real data
7. **Gradual rollout** - Enable for increasing % of users
8. **Monitor and iterate** - Track metrics and adjust

---

## 💡 Key Insights

### Why V2?
- **Richer Context**: Full memory architecture integration
- **Better Personalization**: User preferences and learning style
- **Conversation Continuity**: History and summaries
- **Session Awareness**: Device, time, usage patterns
- **Better Monitoring**: Metadata for tracking and optimization

### Why Gradual Migration?
- **Safety**: V1 continues working during migration
- **Testing**: Validate V2 with small user groups first
- **Rollback**: Easy to revert if issues arise
- **Learning**: Iterate and improve based on feedback
- **Confidence**: Build confidence before full rollout

### What Makes V2 Better?
- Same functionality as V1
- Plus memory integration
- Plus personalization
- Plus conversation continuity
- Plus session awareness
- Plus better monitoring

---

## 📞 Support

For questions or issues:
1. Check `V2_IMPLEMENTATION_GUIDE.md` troubleshooting section
2. Review `V2_SUMMARY.md` known issues
3. Check logs for error messages
4. Test with curl to isolate issues

---

## ✅ Summary

**Status**: Foundation Complete, Ready for Implementation

**What's Ready**:
- ✅ Complete architecture design
- ✅ All types defined
- ✅ Context processor built
- ✅ Feature flags system built
- ✅ Comprehensive documentation
- ✅ Migration plan
- ✅ Implementation guide

**What's Needed**:
- ⏳ Fix import issues
- ⏳ Complete handler implementations
- ⏳ Add V2 configs
- ⏳ Initialize feature flags
- ⏳ Test and deploy

**Timeline**: 6-8 weeks for full migration

**Risk**: Low (parallel implementation, gradual rollout, easy rollback)

---

**Ready to start? Begin with `V2_MIGRATION_PLAN.md` to understand the overall strategy, then move to `V2_IMPLEMENTATION_GUIDE.md` for step-by-step implementation.**
