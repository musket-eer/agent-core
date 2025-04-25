package memory

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type PromptBuffer struct {
	Name    string
	Content string
}

type PromptMemory struct {
	Buffers map[string]*PromptBuffer
	mu      sync.RWMutex
}

// NewPromptMemory loads all prompt files from the given directory into memory
func NewPromptMemory(promptsDir string) (*PromptMemory, error) {
	buffers := make(map[string]*PromptBuffer)
	files, err := ioutil.ReadDir(promptsDir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(promptsDir, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		buffers[file.Name()] = &PromptBuffer{
			Name:    file.Name(),
			Content: string(content),
		}
	}
	return &PromptMemory{Buffers: buffers}, nil
}

// GetBuffer returns the prompt buffer by filename
func (pm *PromptMemory) GetBuffer(name string) (string, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	buf, ok := pm.Buffers[name]
	if !ok {
		return "", false
	}
	return buf.Content, true
}

// CombineBuffers combines multiple prompt buffers by filenames
func (pm *PromptMemory) CombineBuffers(names ...string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	result := ""
	for _, name := range names {
		if buf, ok := pm.Buffers[name]; ok {
			result += buf.Content + "\n"
		}
	}
	return result
}
