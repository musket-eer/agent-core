package embedding

import (
	"fmt"
	"os"
)

// EmbeddingConfig holds global embedding settings.
type EmbeddingConfig struct {
	Dimension int
}

// PrintEmbeddingConfig prints the current embedding configuration.
func PrintEmbeddingConfig() {
	fmt.Printf("[Embedding] Using embedding dimension: %d\n", GlobalEmbeddingConfig.Dimension)
}

var GlobalEmbeddingConfig = EmbeddingConfig{
	Dimension: getEmbeddingDimensionFromEnv(),
}

func getEmbeddingDimensionFromEnv() int {
	dim := os.Getenv("EMBEDDING_DIMENSION")
	if dim == "" {
		return 5 // Default for testing
	}
	var d int
	_, err := fmt.Sscanf(dim, "%d", &d)
	if err != nil || d <= 0 {
		return 5
	}
	return d
}

// EmbeddableRequest is a generic interface for things that can be embedded.
type EmbeddableRequest interface {
	ToEmbeddingText() string
}

// GetEmbedding takes an EmbeddableRequest and returns an embedding of the specified dimension.
// Replace this dummy implementation with a real embedding model or API call.
func GetEmbedding(req EmbeddableRequest) ([]float64, error) {
	text := req.ToEmbeddingText()
	dim := GlobalEmbeddingConfig.Dimension
	embedding := make([]float64, dim)
	for i, b := range []byte(text) {
		embedding[i%dim] += float64(b)
	}
	return embedding, nil
}

// JSONStringWrapper implements EmbeddableRequest for arbitrary JSON strings.
type JSONStringWrapper string

func (j JSONStringWrapper) ToEmbeddingText() string {
	return string(j)
}
