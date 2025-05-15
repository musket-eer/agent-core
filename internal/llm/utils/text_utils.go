package utils

import (
	"log"
	"strings"
)

func ExtractNonEmptyLines(input string) []string {
	log.Printf("[ExtractNonEmptyLines] Input: %q", input)
	lines := strings.Split(input, "\n")
	out := []string{}
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	log.Printf("[ExtractNonEmptyLines] Output: %+v", out)
	return out
}

func SplitByPipe(line string) []string {
	log.Printf("[SplitByPipe] Input: %q", line)
	parts := strings.SplitN(line, "|", 2)
	if len(parts) != 2 {
		log.Printf("[SplitByPipe] Output: nil (no pipe found)")
		return nil
	}
	result := []string{
		strings.TrimSpace(parts[0]),
		strings.TrimSpace(parts[1]),
	}
	log.Printf("[SplitByPipe] Output: %+v", result)
	return result
}
