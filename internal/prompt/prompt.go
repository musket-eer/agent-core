// Reads the prompts and persists them in memory as well as have the memory addresses of buffered prompts.
package prompt


import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// promptStore holds prompt contents in memory and maps endpoint -> pointer to prompt string
var (
	promptStore      = make(map[string]*string)
	promptStoreMutex sync.RWMutex
	loadOnce        sync.Once
)

// LoadPrompts loads all prompt files from configs/prompts/ into memory
func LoadPrompts() error {
	var loadErr error
	loadOnce.Do(func() {
		promptDir := filepath.Join("configs", "prompts")
		files, err := ioutil.ReadDir(promptDir)
		if err != nil {
			loadErr = fmt.Errorf("failed to read prompts directory: %w", err)
			return
		}
		promptStoreMutex.Lock()
		defer promptStoreMutex.Unlock()
		for _, file := range files {
			if file.IsDir() { continue }
			name := file.Name()
			path := filepath.Join(promptDir, name)
			contentBytes, err := os.ReadFile(path)
			if err != nil {
				loadErr = fmt.Errorf("failed to read prompt file %s: %w", name, err)
				return
			}
			// Use filename (without extension) as endpoint key
			endpoint := strings.TrimSuffix(name, filepath.Ext(name))
			content := string(contentBytes)
			promptStore[endpoint] = &content
		}
	})
	return loadErr
}

// GetPromptPtr returns the pointer to the prompt string for a given endpoint
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

// GetPrompt returns the prompt string for a given endpoint
func GetPrompt(endpoint string) (string, error) {
	ptr, err := GetPromptPtr(endpoint)
	if err != nil {
		return "", err
	}
	return *ptr, nil
}
