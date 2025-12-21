# Tasks Agent Prompts - Updated for Dialogue & Markdown

## 🎯 Changes Made

All Tasks agent prompts have been updated to:

### 1. **Encourage Dialogue**
- End responses with follow-up questions
- Invite user to continue conversation
- Ask clarifying questions
- Examples:
  - "What aspect would you like to explore?"
  - "Need help with the implementation?"
  - "What's your next goal?"

### 2. **Keep Responses Short (~75 words max)**
- Maximum ~75 words per response
- Focus on answering direct questions
- No over-explanation
- Concise and actionable

### 3. **Use Clean Markdown Formatting**
- **Bold** for emphasis
- `code` for technical terms
- [Clickable links](url) for all URLs
- Lists for clarity
- Headings (##, ###) for structure
- ✅ Checkmarks for progress/completion

### 4. **Follow Generate-Notes Pattern**
- Clean, readable markdown
- Clickable URLs: `[Link Text](https://url.com)`
- Professional formatting
- Easy to scan and read

---

## 📝 Updated Prompts

### 1. **tasksChatPrompt.txt** ✅
**Changes:**
- Completely rewritten for conversational style
- ~75 word limit enforced
- Markdown formatting required
- Must end with follow-up questions
- Example responses provided

**Output Format:**
```json
{
  "response": "Short markdown response (~75 words) with **formatting**, [links](url), and questions"
}
```

### 2. **tasksSuggestionsPrompt.txt** ✅
**Changes:**
- Markdown in all text fields
- Clickable links in suggestions
- Brief reasoning (~75 words)
- Follow-up questions in reasoning

**Key Updates:**
- `what`: "**Task description** with [relevant links](url)"
- `reasoning`: "Short explanation with **bold** key points and follow-up question"

### 3. **tasksPrioritizePrompt.txt** ✅
**Changes:**
- Markdown formatting throughout
- ~75 word limit on reasoning
- Clickable links for methods
- Dialogue questions at end

**Key Updates:**
- `reasoning`: "**Overall strategy** in ~75 words with [helpful links](url). End with follow-up question."
- `suggestions`: "**Tip** with markdown"

### 4. **tasksBreakDownPrompt.txt** ✅
**Changes:**
- Markdown in all fields
- Clickable resource links
- ✅ Checkmarks for checkpoints
- ~50 word descriptions

**Key Updates:**
- `what`: "**Subtask description** with markdown and [helpful links](url)"
- `resources`: "[Resource Name](https://url.com) - brief description"
- `checkpoints`: "✅ **Checkpoint** - how to verify"

### 5. **tasksProductivityPrompt.txt** ✅
**Changes:**
- Markdown everywhere
- ~30-75 word limits
- Clickable links in recommendations
- Questions at end

**Key Updates:**
- All fields use **markdown**
- `recommendations`: End with questions
- `strengths/improvements`: Brief with links

### 6. **tasksMotivationPrompt.txt** ✅
**Changes:**
- ~75 word messages
- Markdown formatting
- Specific achievements mentioned
- Encouraging questions
- Emojis (1-2 max)

**Key Updates:**
- Message format: "**Great work!** You've completed **5 tasks**... What's your next goal?"

### 7. **tasksSchedulePrompt.txt** ✅
**Changes:**
- Markdown throughout
- ~75 word reasoning
- Clickable time management links
- Dialogue questions

**Key Updates:**
- `reasoning`: "**Overall strategy** in ~75 words with [time management tips](url). End with question?"

---

## 🎨 Markdown Formatting Examples

### Bold for Emphasis
```markdown
**Important point**
**Key concept**
```

### Clickable Links
```markdown
[React Documentation](https://react.dev)
[MDN Web Docs](https://developer.mozilla.org)
[Productivity Guide](https://todoist.com/productivity-methods)
```

### Lists
```markdown
- **Point 1** - description
- **Point 2** - description
```

### Checkmarks
```markdown
✅ **Completed** - description
✅ **Done** - verification
```

### Code
```markdown
Use `useState` for state management
Install with `npm install`
```

---

## 💬 Dialogue Examples

### Good Follow-up Questions
- "What aspect would you like to explore further?"
- "Need help with the implementation?"
- "Want me to explain X in more detail?"
- "What's your next goal?"
- "Which approach sounds best for your situation?"
- "What tasks are you struggling to prioritize?"

### Bad (No Dialogue)
- ❌ Just providing information without asking
- ❌ Ending abruptly without engagement
- ❌ No invitation to continue conversation

---

## 📏 Word Count Guidelines

| Field | Max Words | Notes |
|-------|-----------|-------|
| Main response | ~75 | Direct answer to question |
| Reasoning | ~75 | Overall strategy/explanation |
| Individual items | ~30-50 | Brief, focused |
| Suggestions/Tips | ~30 | Concise and actionable |

---

## ✅ Quality Checklist

Before sending a response, verify:

- [ ] Uses **markdown** formatting (bold, links, lists)
- [ ] All URLs are [clickable links](url)
- [ ] Response is ~75 words or less
- [ ] Ends with 1-2 follow-up questions
- [ ] No over-explanation
- [ ] Answers user's direct question
- [ ] Professional and readable
- [ ] Follows generate-notes pattern

---

## 🎯 Impact

**Before:**
- Long, detailed explanations
- No dialogue encouragement
- Plain text responses
- No markdown formatting
- URLs as plain text

**After:**
- Short, focused responses (~75 words)
- Every response invites dialogue
- Clean markdown formatting
- Clickable links everywhere
- Professional, scannable format

---

## 📚 Related Files

- `/configs/prompts/tasksChatPrompt.txt`
- `/configs/prompts/tasksSuggestionsPrompt.txt`
- `/configs/prompts/tasksPrioritizePrompt.txt`
- `/configs/prompts/tasksBreakDownPrompt.txt`
- `/configs/prompts/tasksProductivityPrompt.txt`
- `/configs/prompts/tasksMotivationPrompt.txt`
- `/configs/prompts/tasksSchedulePrompt.txt`

---

**Status:** All Tasks agent prompts updated ✅  
**Next:** Test with actual requests to verify formatting and dialogue
