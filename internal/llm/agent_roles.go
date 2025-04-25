package llm

// AgentRole defines the type for chat agent roles.
type AgentRole string

const (
	ProfessorAgent AgentRole = "professor"
	LibrarianAgent AgentRole = "librarian"
	ExaminerAgent  AgentRole = "examiner"
)

// ValidAgentRoles lists all supported agent roles.
var ValidAgentRoles = []AgentRole{
	ProfessorAgent,
	LibrarianAgent,
	ExaminerAgent,
}
