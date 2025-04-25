package cache

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env"); err != nil {
		println("[DEBUG] No .env file found. Falling back to system env vars.")
	}
	os.Exit(m.Run())
}

func TestCacheEmbeddingAndCheck_Integration(t *testing.T) {
	apiUrl := os.Getenv("SUPABASE_API_URL")
	apiKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	table := os.Getenv("SUPABASE_TABLE_NAME")

	if apiUrl == "" || apiKey == "" || table == "" {
		t.Skip("SUPABASE_API_URL, SUPABASE_SERVICE_ROLE_KEY, SUPABASE_TABLE_NAME must be set for integration test")
	}

	t.Logf("[DEBUG] SUPABASE_API_URL: %s", apiUrl)
	t.Logf("[DEBUG] SUPABASE_TABLE_NAME: %s", table)

	client := &SupabasePgvectorClient{ApiUrl: apiUrl, ApiKey: apiKey, TableName: table}

	key := "test-key-123"
	embedding := []float64{0.42, 0.17, 0.33, 0.11, 0.05}
	response := "test-response"

	if err := client.CacheEmbedding(key, embedding, response); err != nil {
		t.Fatalf("CacheEmbedding() error: %v", err)
	}

	id, score, hit, err := client.Check(embedding)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	t.Logf("id: %s, score: %f, hit: %v", id, score, hit)
	if !hit {
		t.Errorf("Expected cache hit after insertion, got miss")
	}
}

func TestCheck_Integration(t *testing.T) {
	apiUrl := os.Getenv("SUPABASE_API_URL")
	apiKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	table := os.Getenv("SUPABASE_TABLE_NAME")

	if apiUrl == "" || apiKey == "" || table == "" {
		t.Skip("SUPABASE_API_URL, SUPABASE_SERVICE_ROLE_KEY, SUPABASE_TABLE_NAME must be set for integration test")
	}

	t.Logf("[DEBUG] SUPABASE_API_URL: %s", apiUrl)
	t.Logf("[DEBUG] SUPABASE_TABLE_NAME: %s", table)

	client := &SupabasePgvectorClient{ApiUrl: apiUrl, ApiKey: apiKey, TableName: table}

	embedding := []float64{0.7, 0.2, 0.1, 0.0, 0.0}
	id, score, hit, err := client.Check(embedding)
	if err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	t.Logf("id: %s, score: %f, hit: %v", id, score, hit)
}
