package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// V2 fields template to add to each request struct
const v2FieldsTemplate = `
	// V2 fields (optional for backward compatibility)
	AgentID       string               ` + "`json:\"agentId,omitempty\"`" + `
	SessionID     string               ` + "`json:\"sessionId,omitempty\"`" + `
	ComponentType string               ` + "`json:\"componentType,omitempty\"`" + `
	ComponentID   string               ` + "`json:\"componentId,omitempty\"`" + `
	Context       *types.V2Context     ` + "`json:\"context,omitempty\"`" + `
	Messages      []types.V2Message    ` + "`json:\"messages,omitempty\"`" + `
}`

func main() {
	fmt.Println("V2 Migration Type Updater")
	fmt.Println("=========================")
	fmt.Println()

	// Files to process
	typeFiles := []string{
		"internal/llm/types_reflection.go",
		"internal/llm/types_projects.go",
		"internal/llm/types_courses.go",
		"internal/llm/types_agents.go",
		"internal/llm/types_dashboard.go",
		"internal/llm/types_course_creation.go",
		"internal/llm/types_milestone_generation.go",
		"internal/llm/types_generate_tasks.go",
	}

	for _, file := range typeFiles {
		fmt.Printf("Processing %s...\n", file)
		if err := processFile(file); err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  ✓ Complete\n")
		}
	}

	fmt.Println()
	fmt.Println("Migration complete!")
	fmt.Println("Please review the changes and test the endpoints.")
}

func processFile(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	text := string(content)

	// Check if import is needed
	if !strings.Contains(text, "\"agent-core/internal/llm/v2/types\"") {
		// Add import after package declaration
		packageRe := regexp.MustCompile(`(package llm\n)`)
		text = packageRe.ReplaceAllString(text, "$1\nimport (\n\t\"agent-core/internal/llm/v2/types\"\n)\n")
	}

	// Find all Request structs
	structRe := regexp.MustCompile(`type\s+(\w+Request)\s+struct\s+{([^}]+)}`)
	matches := structRe.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		structName := match[1]
		structBody := match[2]

		// Skip if already has V2 fields
		if strings.Contains(structBody, "V2 fields") || strings.Contains(structBody, "Context       *types.V2Context") {
			continue
		}

		// Skip if it's PersonaRequest (base type)
		if structName == "PersonaRequest" {
			continue
		}

		fmt.Printf("  Adding V2 fields to %s\n", structName)

		// Replace the struct
		oldStruct := fmt.Sprintf("type %s struct {%s}", structName, structBody)
		newStruct := fmt.Sprintf("type %s struct {%s%s", structName, structBody, v2FieldsTemplate)

		text = strings.Replace(text, oldStruct, newStruct, 1)

		// Add interface methods after the struct
		methods := fmt.Sprintf(`

func (r %s) GetV2Context() *types.V2Context { return r.Context }
func (r %s) GetUserID() string { return r.UserID }
func (r %s) GetAgentID() string { return r.AgentID }
`, structName, structName, structName)

		// Insert methods after the struct definition
		text = strings.Replace(text, newStruct, newStruct+methods, 1)
	}

	// Write back
	return ioutil.WriteFile(filename, []byte(text), 0644)
}
