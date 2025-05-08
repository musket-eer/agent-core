package utils

import 
(
	"strings"
)

func ExtractNonEmptyLines(input string) []string {
	lines := strings.Split(input, "\n")
	out := []string{}
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func SplitByPipe(line string) []string {
	parts := strings.SplitN(line, "|", 2)
	if len(parts) != 2 {
		return nil
	}
	return []string{
		strings.TrimSpace(parts[0]),
		strings.TrimSpace(parts[1]),
	}
}
