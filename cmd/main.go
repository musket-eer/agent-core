package main

import (
	"fmt"
	"log"
	"net/http"

	"agent-core/internal/llm"
)

func main() {
	port := 8080
	mux := llm.RegisterHandlers()
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server running at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
