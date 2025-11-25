# API Implementation Guide

## Overview

This document describes the comprehensive API implementation for the AI University agent-core backend. The system implements **43 endpoints** across **8 specialized AI agents**.

## Architecture

### Project Structure

```
agent-core/
├── internal/llm/
│   ├── types_common.go          # Shared types (UserProfile, Activity, etc.)
│   ├── types_dashboard.go       # Dashboard agent types
│   ├── types_courses.go         # Course browse agent types
│   ├── types_reflection.go      # Reflection agent types
│   ├── types_tasks.go           # Tasks agent types
│   ├── types_agents.go          # Professor, Examiner, Librarian, Projects types
│   ├── dashboard_handlers.go    # Dashboard endpoints (4 handlers)
│   ├── courses_handlers.go      # Course browse endpoints (5 handlers)
│   ├── reflection_handlers.go   # Reflection endpoints (6 handlers)
│   ├── tasks_handlers.go        # Tasks endpoints (7 handlers)
│   ├── librarian_handlers.go    # Librarian endpoints (7 handlers)
│   ├── projects_handlers.go     # Projects endpoints (5 handlers)
│   ├── routers.go               # Route registration
│   └── utils/
│       └── errors.go            # Enhanced error handling
├── configs/prompts/             # AI prompts for each endpoint (33 files)
└── docs/
    ├── llms-payloads.md         # API specification
    └── API_IMPLEMENTATION.md    # This file
```

## Implemented Agents

### 1. Dashboard Agent (4 endpoints)
- `POST /dashboard/insights` - Analyze learning data and provide insights
- `POST /dashboard/recommendations` - Generate course recommendations
- `POST /dashboard/analyze-progress` - Analyze progress over time
- `POST /dashboard/chat` - Conversational dashboard assistant

### 2. Course Browse Agent (5 endpoints)
- `POST /courses/intelligent-search` - AI-powered course search
- `POST /courses/explain-relevance` - Explain why a course is relevant
- `POST /courses/compare` - Compare multiple courses
- `POST /courses/suggest-path` - Generate learning paths
- `POST /courses/chat` - Course discovery assistant

### 3. Reflection Agent (6 endpoints)
- `POST /reflection/generate-prompts` - Generate reflection questions
- `POST /reflection/analyze` - Analyze reflection depth and quality
- `POST /reflection/learning-insights` - Identify learning patterns
- `POST /reflection/period-summary` - Summarize learning periods
- `POST /reflection/chat` - Reflection coaching
- `POST /reflection/suggest-exercises` - Suggest reflection exercises

### 4. Tasks Agent (7 endpoints)
- `POST /tasks/generate-suggestions` - Generate task suggestions
- `POST /tasks/prioritize` - Prioritize tasks intelligently
- `POST /tasks/break-down` - Break tasks into subtasks
- `POST /tasks/productivity-insights` - Analyze productivity patterns
- `POST /tasks/generate-schedule` - Generate optimized schedules
- `POST /tasks/chat` - Task management assistant
- `POST /tasks/motivation` - Generate motivational messages

### 5. Professor Agent (3 endpoints - existing)
- `POST /professor-chat` - Interactive lesson chat
- `POST /generate-notes` - Generate lesson notes
- `POST /generate-lesson` - Generate lesson plans

### 6. Examiner Agent (4 endpoints - existing)
- `POST /examiner-chat` - Exam guidance chat
- `POST /generate-exam-questions` - Generate exam questions
- `POST /submit-exam` - Submit and grade exams
- `GET /exam-details/{examId}` - Get exam details

### 7. Librarian Agent (7 endpoints)
- `POST /generate-resources` - Generate learning resources (existing)
- `POST /chat-video` - Video content assistant
- `POST /recommend-videos` - Recommend educational videos
- `POST /summarize-video` - Summarize video content
- `POST /extract-concepts` - Extract key concepts from videos
- `POST /generate-quiz` - Generate quizzes from videos
- `POST /find-topic` - Find topics in video transcripts

### 8. Project Assistant Agent (5 endpoints)
- `POST /projects/chat` - Project guidance chat
- `POST /projects/suggestions` - Suggest learning projects
- `POST /projects/help` - Get project help and resources
- `POST /projects/break-down` - Break projects into milestones
- `POST /projects/resources` - Get milestone-specific resources

## Key Features

### Type Safety
- Comprehensive type definitions for all request/response schemas
- Shared common types to avoid duplication
- Proper JSON serialization tags

### Error Handling
- Standardized error response format
- Error codes: `INVALID_REQUEST`, `UNAUTHORIZED`, `RATE_LIMIT_EXCEEDED`, `LLM_ERROR`, `NOT_FOUND`, `INTERNAL_ERROR`
- Helper functions for common error scenarios

### Caching
- Request payload-based caching using Upstash Redis
- Automatic cache key generation
- Cache-first strategy for improved performance

### LLM Integration
- Flexible client selection (Deepseek, Gemini)
- Configurable model parameters
- Prompt template system with 33+ specialized prompts

### CORS Support
- Pre-configured CORS headers
- OPTIONS method handling
- Cross-origin request support

## Usage Examples

### Dashboard Insights
```bash
curl -X POST http://localhost:8080/dashboard/insights \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "userName": "John Doe",
    "educationLevel": "Undergraduate",
    "careerGoals": "Software Engineer",
    "interests": ["AI", "Web Development"],
    "recentActivities": [],
    "completedCourses": ["intro-to-python"],
    "currentCourses": ["data-structures"],
    "learningStreak": 7,
    "totalStudyTime": 3600
  }'
```

### Course Search
```bash
curl -X POST http://localhost:8080/courses/intelligent-search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning for beginners",
    "userId": "user-123",
    "userProfile": {
      "name": "John Doe",
      "education_level": "Undergraduate",
      "careerGoals": "Data Scientist",
      "interests": ["AI", "Python"]
    },
    "filters": {
      "difficulty": "beginner"
    }
  }'
```

### Task Generation
```bash
curl -X POST http://localhost:8080/tasks/generate-suggestions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "learningGoals": ["Master Python", "Build portfolio"],
    "currentCourses": [],
    "availableTime": "flexible",
    "preferredDifficulty": "medium",
    "interests": ["Web Development"]
  }'
```

## Configuration

### Adding New Endpoints

1. **Define Types** in appropriate `types_*.go` file
2. **Create Handler** in corresponding `*_handlers.go` file
3. **Register Route** in `routers.go`
4. **Create Prompt** in `configs/prompts/`
5. **Update Documentation** in `docs/llms-payloads.md`

### Prompt Management

Prompts are stored as text files in `configs/prompts/` and loaded dynamically:
- Naming convention: `{agent}{Purpose}Prompt.txt`
- Example: `dashboardInsightsPrompt.txt`, `courseSearchPrompt.txt`

### LLM Client Configuration

Configure clients in `configs/llm/configs/`:
- Model selection
- Temperature settings
- Token limits
- Client routing (general vs chat)

## Testing

Run the server:
```bash
go run cmd/main.go
```

Test endpoints:
```bash
# Use the provided test script
./test_all_endpoints.sh

# Or test individual endpoints
curl -X POST http://localhost:8080/dashboard/insights -d @test_data.json
```

## Error Responses

All errors follow this format:
```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid request body",
    "details": "Missing required field: userId"
  }
}
```

## Performance Considerations

- **Caching**: All LLM responses are cached by request payload
- **Connection Pooling**: HTTP clients reuse connections
- **Async Operations**: Non-blocking I/O for LLM calls
- **Rate Limiting**: Implement rate limiting at the reverse proxy level

## Security

- **Input Validation**: All requests validated before processing
- **Error Sanitization**: Internal errors don't expose sensitive data
- **CORS Configuration**: Restrict origins in production
- **Authentication**: Implement Bearer token validation (TODO)

## Next Steps

1. Add authentication middleware
2. Implement rate limiting per user
3. Add request logging and monitoring
4. Create integration tests
5. Add OpenAPI/Swagger documentation
6. Implement webhook support for async operations

## Maintenance

### Adding a New Agent

1. Create `types_{agent}.go` with request/response types
2. Create `{agent}_handlers.go` with handler implementations
3. Create prompt files in `configs/prompts/`
4. Register routes in `routers.go`
5. Update this documentation

### Updating Existing Endpoints

1. Modify types in `types_*.go`
2. Update handler logic in `*_handlers.go`
3. Adjust prompts if needed
4. Test thoroughly
5. Update API documentation

## Support

For issues or questions:
- Check `docs/llms-payloads.md` for API specifications
- Review handler implementations in `internal/llm/`
- Examine prompt files for LLM behavior
- Check logs in `api.log`
