package utils

import (
	"regexp"
	"strings"
)

// ParseLessonData parses lesson data text into structured data
func ParseLessonData(text string) map[string]interface{} {
	// Remove code blocks markers and trim
	text = regexp.MustCompile("```").ReplaceAllString(text, "")
	text = strings.TrimSpace(text)
	
	// Split text into sections and filter empty lines
	sections := []string{}
	for _, s := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			sections = append(sections, trimmed)
		}
	}
	
	// Extract metadata
	lessonMetadata := map[string]interface{}{
		"title": "Untitled",
	}
	
	if len(sections) > 0 {
		lessonMetadata["title"] = sections[0]
	}
	
	if len(sections) > 3 {
		objectives := sections[1:4]
		lessonMetadata["objectives"] = objectives
	} else {
		lessonMetadata["objectives"] = []string{}
	}
	
	if len(sections) > 4 {
		lessonMetadata["duration"] = sections[4]
	}
	
	if len(sections) > 5 {
		lessonMetadata["difficulty"] = sections[5]
	}
	
	// Parse activities
	activities := []map[string]string{}
	
	if len(sections) > 6 {
		for i, s := range sections[6:] {
			parts := strings.Split(s, "|")
			for j := range parts {
				parts[j] = strings.TrimSpace(parts[j])
			}
			
			activity := map[string]string{}
			if len(parts) > 0 && parts[0] != "" {
				activity["id"] = parts[0]
			} else {
				activity["id"] = "activity-" + string(rune('1'+i))
			}
			
			if len(parts) > 1 {
				activity["leader"] = parts[1]
			}
			
			if len(parts) > 2 {
				activity["title"] = parts[2]
			}
			
			activities = append(activities, activity)
		}
	}
	
	return map[string]interface{}{
		"lessonMetadata": lessonMetadata,
		"activities":     activities,
	}
}

// ParseResources parses resource data text into structured data
func ParseResources(text string) []map[string]string {
	blocks := strings.Split(text, "\n\n")
	resources := []map[string]string{}
	
	for _, block := range blocks {
		lines := strings.Split(block, "\n")
		resource := map[string]string{
			"title":       "",
			"description": "",
			"link":        "",
		}
		
		for _, line := range lines {
			if strings.HasPrefix(line, "Title:") {
				resource["title"] = strings.TrimSpace(strings.Replace(line, "Title:", "", 1))
			}
			if strings.HasPrefix(line, "Description:") {
				resource["description"] = strings.TrimSpace(strings.Replace(line, "Description:", "", 1))
			}
			if strings.HasPrefix(line, "Link:") {
				resource["link"] = strings.TrimSpace(strings.Replace(line, "Link:", "", 1))
			}
		}
		
		if resource["title"] != "" && resource["link"] != "" {
			resources = append(resources, resource)
		}
	}
	
	return resources
}

// ExtractJsonFromCodeBlock extracts JSON from code blocks
func ExtractJsonFromCodeBlock(text string) string {
	re := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]+?)\\s*```")
	matches := re.FindStringSubmatch(text)
	
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	
	return strings.TrimSpace(text)
}