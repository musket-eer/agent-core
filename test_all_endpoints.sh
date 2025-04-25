#!/bin/bash

# Bash script to test all LLM endpoints and log detailed debugging info
# Usage: bash test_all_endpoints.sh

BASE_URL="http://localhost:8080"
LOG_FILE="curl_test_log_$(date '+%Y%m%d_%H%M%S').txt"

# Function to run a curl request and log details
test_endpoint() {
  local endpoint="$1"
  local payload="$2"
  local description="$3"

  echo "\n==============================" | tee -a "$LOG_FILE"
  echo "Testing endpoint: $endpoint" | tee -a "$LOG_FILE"
  echo "Description: $description" | tee -a "$LOG_FILE"
  echo "Payload:" | tee -a "$LOG_FILE"
  echo "$payload" | tee -a "$LOG_FILE"
  echo "------------------------------" | tee -a "$LOG_FILE"

  curl -X POST "$BASE_URL$endpoint" \
    -H "Content-Type: application/json" \
    -w '\n[HTTP STATUS]: %{http_code}\n[TOTAL TIME]: %{time_total}s\n' \
    -o response.tmp \
    -d "$payload" 2>&1 | tee -a "$LOG_FILE"

  echo "[RESPONSE BODY]:" | tee -a "$LOG_FILE"
  cat response.tmp | tee -a "$LOG_FILE"
  rm response.tmp
}

# Test payloads (same structure for most endpoints)
COURSE_PAYLOAD='{ "title": "AI Fundamentals", "objective": "Teach the basics of AI", "difficulty": "beginner", "education_level": "undergraduate" }'
LESSON_PAYLOAD='{ "title": "Lesson 1: Introduction to AI", "objective": "Understand what AI is and its applications", "difficulty": "beginner", "education_level": "undergraduate" }'
CHAT_PAYLOAD='{ "title": "AI Fundamentals", "objective": "Discuss AI concepts", "difficulty": "beginner", "education_level": "undergraduate", "agent_role": "professor" }'
# Agent-specific chat payload (no agent_role needed, endpoint determines role)
PROFESSOR_CHAT_PAYLOAD='{ "title": "AI Fundamentals", "objective": "Discuss AI concepts", "difficulty": "beginner", "education_level": "undergraduate" }'
LIBRARIAN_CHAT_PAYLOAD='{ "title": "AI Fundamentals", "objective": "Find information resources", "difficulty": "beginner", "education_level": "undergraduate" }'
EXAMINER_CHAT_PAYLOAD='{ "title": "AI Fundamentals", "objective": "Quiz the student on AI concepts", "difficulty": "beginner", "education_level": "undergraduate" }'

# Run tests
test_endpoint "/generate-course" "$COURSE_PAYLOAD" "Generate a full course syllabus."
test_endpoint "/generate-lesson" "$LESSON_PAYLOAD" "Generate a lesson plan."
test_endpoint "/generate-notes" "$COURSE_PAYLOAD" "Generate study notes."
test_endpoint "/generate-exams" "$COURSE_PAYLOAD" "Generate exam questions."
test_endpoint "/generate-resources" "$COURSE_PAYLOAD" "Generate resource recommendations."
test_endpoint "/chat" "$CHAT_PAYLOAD" "Chat with the LLM as a professor."
test_endpoint "/professor-chat" "$PROFESSOR_CHAT_PAYLOAD" "Chat with the LLM as a professor (dedicated endpoint)."
test_endpoint "/librarian-chat" "$LIBRARIAN_CHAT_PAYLOAD" "Chat with the LLM as a librarian (dedicated endpoint)."
test_endpoint "/examiner-chat" "$EXAMINER_CHAT_PAYLOAD" "Chat with the LLM as an examiner (dedicated endpoint)."

echo "\nAll endpoint tests completed. See $LOG_FILE for details."
