# V2 Memory Architecture Migration Plan

## Overview

This document outlines the complete migration strategy for implementing V2 endpoints that integrate the new memory architecture. The migration will be **gradual and parallel** - V1 endpoints remain functional while V2 endpoints are built and tested.

---

## 🎯 Core Principles

1. **Parallel Implementation**: V2 endpoints run alongside V1 (no breaking changes)
2. **Same Functionality**: V2 performs identical operations as V1, just with richer context
3. **Memory Integration**: Frontend compiles context from memory system, backend uses it
4. **Gradual Rollout**: Feature flags control which users get V2
5. **Easy Rollback**: Can switch back to V1 instantly if issues arise

---

## 📦 V2 Payload Structure

### Standard V2 Request Format

```json
{
  "agentId": "tasks",
  "userId": "user-uuid",
  "sessionId": "session-uuid",
  "componentType": "task",
  "componentId": "task-123",
  
  "context": {
    "userInstructions": "Be concise and action-oriented. Focus on practical steps.",
    "sessionMetadata": "Device: Desktop | Browser: Chrome | Time: Morning",
    "userMemory": "Name: John | Learning Style: Visual | Prefers: Step-by-step guides",
    "conversationSummaries": "Recent: Discussed task prioritization strategies",
    "currentSession": "User: How do I prioritize? | Assistant: Use Eisenhower Matrix...",
    "componentContext": "Current Task: Build React dashboard | Priority: High | Due: Tomorrow"
  },
  
  "messages": [
    {
      "role": "system",
      "content": "You are a task management assistant. [Context compiled from above]"
    },
    {
      "role": "user",
      "content": "Help me break down this task"
    }
  ],
  
  "requestData": {
    "learningGoals": ["Master React", "Build production apps"],
    "availableTime": "2 hours",
    "currentTasks": [...]
  }
}
```

### Standard V2 Response Format

```json
{
  "success": true,
  "data": {
    "suggestedTasks": [...],
    "reasoning": "Based on your visual learning style...",
    "nextSteps": [...]
  },
  "metadata": {
    "tokensUsed": 1250,
    "processingTime": 2.3,
    "cacheHit": false,
    "contextSources": ["userMemory", "sessionMetadata", "conversationSummaries"]
  }
}
```

---

## 🏗️ Implementation Architecture

### Directory Structure

```
agent-core/
├── internal/llm/
│   ├── v2/
│   │   ├── handlers/
│   │   │   ├── tasks_v2.go
│   │   │   ├── reflection_v2.go
│   │   │   ├── projects_v2.go
│   │   │   └── ...
│   │   ├── types/
│   │   │   ├── common_v2.go
│   │   │   ├── tasks_v2.go
│   │   │   └── ...
│   │   ├── middleware/
│   │   │   ├── context_processor.go
│   │   │   ├── session_manager.go
│   │   │   └── feature_flags.go
│   │   └── utils/
│   │       ├── prompt_builder.go
│   │       └── response_formatter.go
│   ├── routers_v2.go
│   └── ...existing files...
```

---

## 📋 Migration Phases

### Phase 1: Foundation (Week 1)

**Goal**: Set up V2 infrastructure without touching V1

#### Tasks:
1. ✅ Create V2 directory structure
2. ✅ Define V2 common types
3. ✅ Build context processor utility
4. ✅ Create feature flag system
5. ✅ Set up V2 router

#### Deliverables:
- `internal/llm/v2/types/common_v2.go`
- `internal/llm/v2/middleware/context_processor.go`
- `internal/llm/v2/middleware/feature_flags.go`
- `internal/llm/routers_v2.go`

---

### Phase 2: Tasks Agent V2 (Week 2)

**Goal**: Implement first V2 agent as reference implementation

#### Tasks:
1. ✅ Create Tasks V2 types
2. ✅ Implement Tasks V2 handlers
3. ✅ Add V2 routes for Tasks
4. ✅ Test with sample data
5. ✅ Compare V1 vs V2 responses

#### Endpoints:
- `POST /tasks/v2/generate-suggestions`
- `POST /tasks/v2/prioritize`
- `POST /tasks/v2/break-down`
- `POST /tasks/v2/insights`
- `POST /tasks/v2/chat`

---

### Phase 3: Reflection Agent V2 (Week 3)

**Goal**: Implement second agent, refine patterns

#### Tasks:
1. ✅ Create Reflection V2 types
2. ✅ Implement Reflection V2 handlers
3. ✅ Add V2 routes for Reflection
4. ✅ Test with real user data
5. ✅ Document patterns learned

#### Endpoints:
- `POST /reflection/v2/generate-questions`
- `POST /reflection/v2/generate-journal`
- `POST /reflection/v2/insights`
- `POST /reflection/v2/quick-reflection`
- `POST /reflection/v2/chat`

---

### Phase 4: Remaining Agents (Weeks 4-5)

**Goal**: Complete all agent migrations

#### Agents to Migrate:
- Projects Agent
- Dashboard Agent
- Courses Agent
- Librarian Agent
- Professor Agent
- Examiner Agent

---

### Phase 5: Testing & Rollout (Week 6)

**Goal**: Validate and gradually enable V2

#### Tasks:
1. ✅ A/B testing framework
2. ✅ Enable for 10% of users
3. ✅ Monitor metrics (response quality, latency, errors)
4. ✅ Expand to 50% of users
5. ✅ Full rollout to 100%

---

### Phase 6: Deprecation (Week 7-8)

**Goal**: Remove V1 code

#### Tasks:
1. ✅ Announce V1 deprecation (2 week notice)
2. ✅ Redirect V1 to V2 internally
3. ✅ Remove V1 handlers
4. ✅ Clean up V1 types
5. ✅ Update documentation

---

## 🔄 Request/Response Flow Examples

### Example 1: Tasks - Generate Suggestions

#### V1 Request (Current)
```json
POST /tasks/generate-suggestions

{
  "userId": "user-123",
  "learningGoals": ["Master React"],
  "availableTime": "2 hours",
  "currentCourses": [...]
}
```

#### V1 Response (Current)
```json
{
  "suggestedTasks": [
    {
      "id": "task-1",
      "title": "Build Todo App",
      "description": "Create a React todo application",
      "estimatedTime": 120,
      "priority": "high"
    }
  ]
}
```

#### V2 Request (New)
```json
POST /tasks/v2/generate-suggestions

{
  "agentId": "tasks",
  "userId": "user-123",
  "sessionId": "session-456",
  "componentType": "task-generator",
  
  "context": {
    "userInstructions": "Be concise. Focus on practical projects.",
    "sessionMetadata": "Device: Desktop | Browser: Chrome | Time: 09:30 AM",
    "userMemory": "Name: Sarah | Learning Style: Visual, hands-on | Experience: Beginner React",
    "conversationSummaries": "Last week: Completed intro to React hooks. Struggled with useEffect.",
    "currentSession": "",
    "componentContext": "Goal: Build portfolio projects | Timeline: 2 weeks"
  },
  
  "messages": [
    {
      "role": "system",
      "content": "You are a task management assistant for Most Learned platform.\n\nUser Context:\n- Name: Sarah\n- Learning Style: Visual, hands-on\n- Recent Activity: Completed React hooks intro\n- Current Goal: Build portfolio projects\n\nInstructions: Be concise. Focus on practical projects."
    },
    {
      "role": "user",
      "content": "Suggest tasks for learning React"
    }
  ],
  
  "requestData": {
    "learningGoals": ["Master React"],
    "availableTime": "2 hours",
    "currentCourses": [...]
  }
}
```

#### V2 Response (New)
```json
{
  "success": true,
  "data": {
    "suggestedTasks": [
      {
        "id": "task-1",
        "title": "Build Todo App with Hooks",
        "description": "Since you recently learned hooks, practice useEffect with a todo app",
        "estimatedTime": 120,
        "priority": "high",
        "reasoning": "Builds on your recent useEffect learning",
        "visualResources": ["React hooks diagram", "Component lifecycle chart"]
      }
    ],
    "personalizationNotes": "Tasks tailored for visual learners with hands-on practice focus"
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

### Example 2: Reflection - Generate Task Reflection

#### V1 Request (Current)
```json
POST /reflection/generate-task-reflection

{
  "userId": "user-123",
  "taskId": "task-789",
  "taskInfo": {
    "title": "Build React Dashboard",
    "description": "Create analytics dashboard"
  },
  "timeData": {
    "estimatedMinutes": 120,
    "timeSpent": 9000
  },
  "completionData": {
    "completed": true,
    "productivityRating": 8
  },
  "userResponses": {
    "userReflection": "Took longer than expected but learned a lot"
  }
}
```

#### V1 Response (Current)
```json
{
  "reflection": "I've documented a 30-minute overrun. Task completed successfully with high productivity rating.",
  "insights": ["Time estimation needs improvement"],
  "patterns": ["Underestimates complex UI work"],
  "recommendations": ["Add 25% buffer for UI tasks"],
  "depthScore": 7.5,
  "emotionalTone": "positive",
  "keyLearnings": ["React state management"],
  "challengesFaced": ["Complex data visualization"],
  "solutionsApplied": ["Used Chart.js library"],
  "skillsPracticed": ["React", "Data visualization"]
}
```

#### V2 Request (New)
```json
POST /reflection/v2/generate-task-reflection

{
  "agentId": "reflection",
  "userId": "user-123",
  "sessionId": "session-789",
  "componentType": "task-reflection",
  "componentId": "task-789",
  
  "context": {
    "userInstructions": "Keep reflections under 50 words. Focus on actionable insights.",
    "sessionMetadata": "Device: Mobile | Time: Evening | Location: Home",
    "userMemory": "Name: Mike | Reflection Style: Analytical | Prefers: Data-driven insights",
    "conversationSummaries": "Pattern: Consistently underestimates UI work by 20-30%",
    "currentSession": "",
    "componentContext": "Task: Build React Dashboard | Type: UI Development | Complexity: High"
  },
  
  "messages": [
    {
      "role": "system",
      "content": "You are a reflection coach. User prefers analytical, data-driven insights under 50 words.\n\nUser Pattern: Consistently underestimates UI work by 20-30%"
    },
    {
      "role": "user",
      "content": "Reflect on task completion"
    }
  ],
  
  "requestData": {
    "taskInfo": {
      "title": "Build React Dashboard",
      "description": "Create analytics dashboard"
    },
    "timeData": {
      "estimatedMinutes": 120,
      "timeSpent": 9000
    },
    "completionData": {
      "completed": true,
      "productivityRating": 8
    },
    "userResponses": {
      "userReflection": "Took longer than expected but learned a lot"
    }
  }
}
```

#### V2 Response (New)
```json
{
  "success": true,
  "data": {
    "reflection": "30-min overrun on UI work—consistent with your 25% underestimation pattern. High productivity (8/10) despite complexity.",
    "insights": [
      "Time estimation accuracy: 80% (improving from 70% last month)",
      "UI tasks consistently take 25% longer than estimated",
      "Productivity remains high even with overruns"
    ],
    "patterns": [
      "Underestimates UI complexity (4th occurrence this month)",
      "Works best in evening sessions (productivity +15%)",
      "High engagement with data visualization tasks"
    ],
    "recommendations": [
      "Apply 30% buffer to all UI estimation (data-backed)",
      "Schedule complex UI work for evening slots",
      "Track estimation accuracy weekly for improvement"
    ],
    "depthScore": 8.2,
    "emotionalTone": "positive",
    "keyLearnings": [
      "React state management patterns",
      "Chart.js integration techniques",
      "Component composition for dashboards"
    ],
    "challengesFaced": [
      "Complex data visualization requirements",
      "State management across multiple charts"
    ],
    "solutionsApplied": [
      "Used Chart.js for visualization",
      "Implemented centralized state management"
    ],
    "skillsPracticed": [
      "React hooks (useEffect, useState)",
      "Data visualization",
      "Component architecture"
    ],
    "trendAnalysis": {
      "estimationAccuracy": "Improving: 70% → 80% over 30 days",
      "productivityTrend": "Stable high performance (avg 8.1/10)",
      "skillGrowth": "Data visualization +40% proficiency"
    }
  },
  "metadata": {
    "tokensUsed": 1820,
    "processingTime": 2.8,
    "cacheHit": false,
    "contextSources": ["userMemory", "conversationSummaries", "componentContext"],
    "personalizationApplied": ["analytical-tone", "data-driven-insights", "concise-format"]
  }
}
```

---

## 🔧 Technical Implementation Details

### Context Processor Utility

The context processor extracts and formats context from the V2 payload:

```go
// internal/llm/v2/middleware/context_processor.go

package middleware

import (
	"fmt"
	"strings"
)

type ContextData struct {
	UserInstructions       string
	SessionMetadata        string
	UserMemory            string
	ConversationSummaries string
	CurrentSession        string
	ComponentContext      string
}

func BuildSystemPrompt(basePrompt string, context ContextData) string {
	var sections []string
	
	// Add base prompt
	sections = append(sections, basePrompt)
	
	// Add context sections if present
	if context.UserMemory != "" {
		sections = append(sections, fmt.Sprintf("\n## User Profile\n%s", context.UserMemory))
	}
	
	if context.UserInstructions != "" {
		sections = append(sections, fmt.Sprintf("\n## User Preferences\n%s", context.UserInstructions))
	}
	
	if context.ConversationSummaries != "" {
		sections = append(sections, fmt.Sprintf("\n## Recent Context\n%s", context.ConversationSummaries))
	}
	
	if context.ComponentContext != "" {
		sections = append(sections, fmt.Sprintf("\n## Current Context\n%s", context.ComponentContext))
	}
	
	if context.SessionMetadata != "" {
		sections = append(sections, fmt.Sprintf("\n## Session Info\n%s", context.SessionMetadata))
	}
	
	return strings.Join(sections, "\n")
}

func ExtractContext(payload map[string]interface{}) ContextData {
	context := ContextData{}
	
	if ctx, ok := payload["context"].(map[string]interface{}); ok {
		if val, ok := ctx["userInstructions"].(string); ok {
			context.UserInstructions = val
		}
		if val, ok := ctx["sessionMetadata"].(string); ok {
			context.SessionMetadata = val
		}
		if val, ok := ctx["userMemory"].(string); ok {
			context.UserMemory = val
		}
		if val, ok := ctx["conversationSummaries"].(string); ok {
			context.ConversationSummaries = val
		}
		if val, ok := ctx["currentSession"].(string); ok {
			context.CurrentSession = val
		}
		if val, ok := ctx["componentContext"].(string); ok {
			context.ComponentContext = val
		}
	}
	
	return context
}
```

---

## 📊 Success Metrics

### Performance Metrics
- **Response Time**: V2 should be within 10% of V1 latency
- **Cache Hit Rate**: Maintain >60% cache hit rate
- **Error Rate**: <0.5% error rate for V2 endpoints

### Quality Metrics
- **Personalization Score**: User feedback on relevance (target: >4.5/5)
- **Context Utilization**: % of responses using context (target: >80%)
- **User Satisfaction**: NPS score improvement (target: +10 points)

### Adoption Metrics
- **V2 Usage**: % of requests using V2 (target: 100% by Week 8)
- **Rollback Rate**: % of users rolled back to V1 (target: <5%)
- **Feature Flag Adoption**: Smooth rollout without incidents

---

## 🚨 Rollback Strategy

### Automatic Rollback Triggers
- Error rate >2% for 5 minutes
- Response time >5 seconds for 10 requests
- Cache failures >10% for 5 minutes

### Manual Rollback Process
1. Set feature flag `use_v2_endpoints` to `false`
2. Monitor V1 endpoint health
3. Investigate V2 issues
4. Fix and redeploy
5. Gradual re-enable V2

---

## 📝 Next Steps

1. **Review this plan** with the team
2. **Set up V2 directory structure** (Phase 1, Day 1)
3. **Create common V2 types** (Phase 1, Day 2)
4. **Build context processor** (Phase 1, Day 3)
5. **Implement Tasks V2** (Phase 2, Week 2)
6. **Test and iterate** (Ongoing)

---

## 📚 Related Documents

- `V2_TYPES_REFERENCE.md` - Complete type definitions
- `V2_HANDLER_EXAMPLES.md` - Handler implementation examples
- `V2_TESTING_GUIDE.md` - Testing strategies
- `V2_DEPLOYMENT_GUIDE.md` - Deployment procedures
