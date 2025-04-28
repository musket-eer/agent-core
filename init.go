package main

import (
	"log"
	"agent-core/configs/memory"
)

var PromptBuffers *memory.PromptMemory

func init() {
	promptsDir := "./configs/prompts"
	bufs, err := memory.NewPromptMemory(promptsDir)
	if err != nil {
		log.Fatalf("Failed to initialize prompt buffers: %v", err)
	}
	PromptBuffers = bufs
}
