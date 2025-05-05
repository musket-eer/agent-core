package prompt

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	promptStore      = make(map[string]*string)
	promptStoreMutex sync.RWMutex
	loadOnce         sync.Once
	requiredPrompts  = []string{
		"lessonSequencePrompt",
		// Add other critical prompts here if needed
	}
)

// LoadPrompts loads all prompt files into memory
func LoadPrompts() error {
	var loadErr error
	loadOnce.Do(func() {
		promptDir := os.Getenv("PROMPT_DIR")
		if promptDir == "" {
			promptDir = filepath.Join("configs", "prompts")
		}

		files, err := os.ReadDir(promptDir)
		if err != nil {
			loadErr = fmt.Errorf("failed to read prompts directory: %w", err)
			return
		}

		promptStoreMutex.Lock()
		defer promptStoreMutex.Unlock()

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			name := file.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if ext != ".txt" && ext != ".md" && ext != ".yaml" && ext != ".yml" {
				continue // skip non-prompt files
			}

			endpoint := strings.TrimSuffix(name, filepath.Ext(name))
			path := filepath.Join(promptDir, name)
			contentBytes, err := os.ReadFile(path)
			if err != nil {
				loadErr = err
				return
			}
			content := string(contentBytes)
			promptStore[endpoint] = &content
		}

		// Ensure required prompts are loaded
		for _, required := range requiredPrompts {
			if _, ok := promptStore[required]; !ok {
				loadErr = fmt.Errorf("🚨 required prompt '%s' not found in %s", required, promptDir)
				log.Println(loadErr)
				return
			}
		}
	})
	return loadErr
}

func GetPromptPtr(endpoint string) (*string, error) {
	if err := LoadPrompts(); err != nil {
		return nil, err
	}
	promptStoreMutex.RLock()
	defer promptStoreMutex.RUnlock()
	ptr, ok := promptStore[endpoint]
	if !ok {
		return nil, fmt.Errorf("prompt for endpoint '%s' not found", endpoint)
	}
	return ptr, nil
}

func GetPrompt(endpoint string) (string, error) {
	ptr, err := GetPromptPtr(endpoint)
	if err != nil {
		return "", err
	}
	return *ptr, nil
}
