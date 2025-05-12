package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// HashPayload takes any JSON-serializable object and returns a stable SHA-256 hash.
func HashPayload(payload any) (string, error) {
	// Marshal the payload to JSON
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Generate SHA-256 hash of the JSON bytes
	hash := sha256.Sum256(jsonBytes)

	// Return hex-encoded string
	return fmt.Sprintf("%x", hash), nil
}
