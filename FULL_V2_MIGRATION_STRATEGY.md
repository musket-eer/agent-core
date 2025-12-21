# Full V2 Migration Strategy - Deprecating V1 Architecture

## 🎯 Objective

**Fully migrate all endpoints to V2 memory architecture and deprecate the V1 system.**

This is a complete replacement strategy where:
- All V1 handlers are updated to use V2 context extraction
- All V1 endpoints accept V2 payload format
- The old persona-based system is deprecated
- Feature flags are removed (V2 becomes the default)

---

## 📊 Current State Analysis

### V1 Endpoints (To Be Migrated)

**Tasks Agent** (V1)
- POST `/tasks/generate-suggestions`
- POST `/tasks/prioritize`
- POST `/tasks/break-down`
- POST `/tasks/insights`
- POST `/tasks/chat`
- POST `/tasks/motivation`
- POST `/tasks/productivity`
- POST `/tasks/schedule`

**Reflection Agent** (V1)
- POST `/reflection/generate-prompts`
- POST `/reflection/generate-task-reflection`
- POST `/reflection/task-prompts`
- POST `/reflection/analyze-task`
- POST `/reflection/analyze`
- POST `/reflection/learning-insights`
- POST `/reflection/period-summary`
- POST `/reflection/chat`
- POST `/reflection/suggest-exercises`

**Projects Agent** (V1)
- POST `/projects/chat`
- POST `/projects/suggestions`
- POST `/projects/help`
- POST `/projects/break-down`
- POST `/projects/resources`
- POST `/projects/create`
- POST `/projects/assist`
- POST `/projects/translate-milestones`

**Courses Agent** (V1)
- POST `/courses/intelligent-search`
- POST `/courses/explain-relevance`
- POST `/courses/compare`
- POST `/courses/suggest-path`
- POST `/courses/chat`

**Librarian Agent** (V1)
- POST `/chat-video`
- POST `/recommend-videos`
- POST `/summarize-video`
- POST `/extract-concepts`
- POST `/generate-quiz`
- POST `/find-topic`

**Dashboard Agent** (V1)
- POST `/dashboard/chat`

**Course Creation Agent** (V1)
- POST `/course-creation/chat`

**Milestone Generation** (V1)
- POST `/generate-project`
- POST `/ai/projects/milestone-to-tasks`

### V2 Endpoints (Already Implemented)

**Tasks Agent V2**
- POST `/tasks/v2/generate-suggestions` ✅
- POST `/tasks/v2/prioritize` ✅
- POST `/tasks/v2/break-down` ✅
- POST `/tasks/v2/insights` ✅
- POST `/tasks/v2/chat` ✅

**Reflection Agent V2**
- POST `/reflection/v2/generate-questions` ✅
- POST `/reflection/v2/generate-journal` ✅
- POST `/reflection/v2/insights` ✅
- POST `/reflection/v2/quick-reflection` ✅
- POST `/reflection/v2/chat` ✅

**Professor Agent V2**
- POST `/professor/v2/chat` ✅
- POST `/professor/v2/generate-notes` ✅
- POST `/professor/v2/insights` ✅
- POST `/professor/v2/explain` ✅
- POST `/professor/v2/practice-questions` ✅

---

## 🔄 Migration Strategy

### Option 1: In-Place Migration (Recommended)

**Update V1 endpoints to accept V2 format while maintaining backward compatibility**

**Pros:**
- No URL changes for clients
- Gradual frontend migration possible
- Simpler deployment

**Cons:**
- More complex handler logic initially
- Need to support both formats temporarily

### Option 2: Parallel Deprecation

**Keep V1 and V2 separate, deprecate V1 over time**

**Pros:**
- Clean separation
- Easy rollback

**Cons:**
- Duplicate code maintenance
- Requires frontend URL changes
- More complex routing

**Decision: Use Option 1 - In-Place Migration**

---

## 📋 Migration Steps

### Phase 1: Core Infrastructure Update

#### 1.1 Update `handleLLMRequest` Function

**Current:** Uses persona-based system
```go
func handleLLMRequest(w http.ResponseWriter, req interface{}, promptKey string, endpoint string) []byte {
    // Extract persona
    agentID, persona := ExtractPersona(req)
    systemPrompt := BuildPersonaSystemPrompt(agentID, persona, basePrompt)
    // ...
}
```

**New:** Use V2 context extraction
```go
func handleLLMRequestV2(w http.ResponseWriter, req interface{}, promptKey string, endpoint string) []byte {
    // Extract V2 context
    contextData := extractV2Context(req)
    systemPrompt := middleware.BuildSystemPrompt(basePrompt, contextData)
    // ...
}
```

#### 1.2 Create Context Extraction Helper

```go
// extractV2Context extracts context from any request type
func extractV2Context(req interface{}) middleware.ContextData {
    // Use reflection or type assertion to extract context fields
    // Support both V2 format and legacy format for backward compatibility
}
```

#### 1.3 Update All Request Types

Add V2 context fields to all existing request structs:

```go
type TasksGenerateSuggestionsRequest struct {
    // Legacy fields (keep for backward compatibility)
    UserID    string `json:"userId"`
    Message   string `json:"message"`
    // ... other legacy fields
    
    // V2 fields (new)
    AgentID       string                 `json:"agentId,omitempty"`
    SessionID     string                 `json:"sessionId,omitempty"`
    ComponentType string                 `json:"componentType,omitempty"`
    ComponentID   string                 `json:"componentId,omitempty"`
    Context       *V2Context             `json:"context,omitempty"`
    Messages      []V2Message            `json:"messages,omitempty"`
}
```

### Phase 2: Handler Migration

#### 2.1 Update All Handlers

For each handler file:
1. Import V2 middleware
2. Update request types to include V2 fields
3. Replace `handleLLMRequest` with `handleLLMRequestV2`
4. Update response format to include V2 metadata

**Example:**
```go
func TasksGenerateSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
    var req TasksGenerateSuggestionsRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Use V2 handler
    response := handleLLMRequestV2(w, req, "tasksSuggestionsPrompt", "/tasks/generate-suggestions")
    if response != nil {
        w.Write(response)
    }
}
```

#### 2.2 Files to Update

- `tasks_handlers.go` (8 handlers)
- `reflection_handlers.go` (9 handlers)
- `projects_handlers.go` (5 handlers)
- `project_handlers.go` (3 handlers)
- `courses_handlers.go` (5 handlers)
- `librarian_handlers.go` (6 handlers)
- `dashboard_handlers.go` (1 handler)
- `course_creation_handler.go` (1 handler)
- `milestone_generation_handlers.go` (2 handlers)

**Total: ~40 handlers to migrate**

### Phase 3: Prompt Updates

#### 3.1 Update All Prompts

Each prompt needs to reference V2 context variables:

**Old format:**
```
You are {{userName}}'s assistant.
User profile: {{userProfile}}
```

**New format:**
```
You are the assistant.

## User Context
{{context.userMemory}}

## User Preferences
{{context.userInstructions}}

## Recent Activity
{{context.conversationSummaries}}

## Current Session
{{context.currentSession}}

## Component Context
{{context.componentContext}}
```

#### 3.2 Prompts to Update

All prompts in `/configs/prompts/`:
- tasksSuggestionsPrompt.txt ✅ (already updated)
- tasksPrioritizePrompt.txt
- tasksBreakDownPrompt.txt
- tasksChatPrompt.txt ✅ (already updated)
- tasksMotivationPrompt.txt
- tasksProductivityPrompt.txt
- tasksSchedulePrompt.txt
- reflectionPromptsPrompt.txt
- taskReflectionPrompt.txt
- reflectionAnalyzePrompt.txt
- reflectionInsightsPrompt.txt
- reflectionSummaryPrompt.txt
- reflectionChatPrompt.txt
- reflectionExercisesPrompt.txt
- projectsChatPrompt.txt
- projectsSuggestionsPrompt.txt
- projectsHelpPrompt.txt
- projectsBreakDownPrompt.txt
- projectsResourcesPrompt.txt
- projectCreatorPrompt.txt
- projectAssistantPrompt.txt
- milestoneTranslatorPrompt.txt
- courseIntelligentSearchPrompt.txt
- courseExplainRelevancePrompt.txt
- courseComparePrompt.txt
- courseSuggestPathPrompt.txt
- coursesChatPrompt.txt
- librarianVideoChatPrompt.txt
- librarianRecommendVideosPrompt.txt
- librarianSummarizeVideoPrompt.txt
- librarianExtractConceptsPrompt.txt
- librarianGenerateQuizPrompt.txt
- librarianFindTopicPrompt.txt
- dashboardChatPrompt.txt
- courseCreationPrompt.txt
- generateProjectPrompt.txt
- milestoneToTasksPrompt.txt

**Total: ~38 prompts to update**

### Phase 4: Remove V1 Infrastructure

#### 4.1 Deprecate Persona System

- Remove `ExtractPersona` function
- Remove `BuildPersonaSystemPrompt` function
- Remove persona-related types

#### 4.2 Remove Feature Flags

- Remove V2 feature flag system
- Remove `/v2/feature-flags` endpoints
- Update main.go to remove flag initialization

#### 4.3 Clean Up V2 Routes

- Remove `/v2/` prefix from routes
- Update V2 handlers to replace V1 handlers
- Remove duplicate V1 handler files

### Phase 5: Testing & Validation

#### 5.1 Create Test Suite

- Test all 40+ endpoints with V2 payloads
- Verify context extraction works correctly
- Validate prompt enrichment
- Check response format

#### 5.2 Update Test Payloads

- Update all test payloads to V2 format
- Create comprehensive test coverage

#### 5.3 Integration Testing

- Test with frontend
- Verify memory integration
- Check performance

---

## 🗓️ Implementation Timeline

### Week 1: Infrastructure
- ✅ Day 1-2: Update `handleLLMRequest` to `handleLLMRequestV2`
- ✅ Day 3-4: Create context extraction helpers
- ✅ Day 5: Update all request types

### Week 2: Handler Migration (Part 1)
- Day 1: Tasks handlers (8)
- Day 2: Reflection handlers (9)
- Day 3: Projects handlers (8)
- Day 4: Courses handlers (5)
- Day 5: Remaining handlers (10)

### Week 3: Prompt Updates
- Day 1-2: Update Tasks prompts
- Day 2-3: Update Reflection prompts
- Day 3-4: Update Projects prompts
- Day 4-5: Update remaining prompts

### Week 4: Cleanup & Testing
- Day 1-2: Remove V1 infrastructure
- Day 3-4: Comprehensive testing
- Day 5: Documentation updates

---

## ✅ Success Criteria

- [ ] All 40+ handlers use V2 context extraction
- [ ] All 38+ prompts use V2 context variables
- [ ] All test payloads updated to V2 format
- [ ] Persona system completely removed
- [ ] Feature flags removed
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Frontend successfully integrated

---

## 🚨 Risks & Mitigation

### Risk 1: Breaking Changes
**Mitigation:** Support both V1 and V2 formats during transition

### Risk 2: Performance Impact
**Mitigation:** Monitor response times, optimize context extraction

### Risk 3: Frontend Not Ready
**Mitigation:** Maintain backward compatibility until frontend migrates

### Risk 4: Prompt Quality
**Mitigation:** Test prompts thoroughly, iterate based on results

---

## 📝 Next Steps

1. **Approve this migration plan**
2. **Start with Phase 1: Core Infrastructure Update**
3. **Create backward-compatible context extraction**
4. **Migrate handlers one agent at a time**
5. **Update prompts incrementally**
6. **Test thoroughly before removing V1 code**

---

**Status:** Ready to begin migration
**Estimated Effort:** 4 weeks
**Risk Level:** Medium (mitigated by backward compatibility)
