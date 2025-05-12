package supabaseclient_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"

	cachepkg "agent-core/internal/cache"
	supabaseclient "agent-core/internal/cache/supabaseclient"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../../.env"); err != nil {
		println("[DEBUG] No .env file found. Falling back to system env vars.")
	}
	os.Exit(m.Run())
}

func TestCacheByHashAndCheckByHash_Integration(t *testing.T) {
	apiUrl := os.Getenv("SUPABASE_API_URL")
	apiKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	table := os.Getenv("SUPABASE_TABLE_NAME")

	if apiUrl == "" || apiKey == "" || table == "" {
		t.Skip(" Missing SUPABASE_API_URL, SUPABASE_SERVICE_ROLE_KEY, or SUPABASE_TABLE_NAME for integration test.")
	}

	client := supabaseclient.NewSupabaseClient(apiUrl, apiKey, table)

	cache := cachepkg.NewCache(client)

	payload := "test-hash-123"
	expectedResponse := "test-response-hash"

	// Cache the response by payload
	if err := cache.CacheByPayload(payload, expectedResponse); err != nil {
		t.Fatalf("CacheByPayload() failed: %v", err)
	}

	// Check cache hit by payload
	cachedResp, hit, err := cache.CheckByPayload(payload)
	if err != nil {
		t.Fatalf("CheckByPayload() failed: %v", err)
	}
	if !hit {
		t.Errorf("Expected cache hit after insertion, got miss")
	}
	if cachedResp != expectedResponse {
		t.Errorf("Expected cached response %q, got %q", expectedResponse, cachedResp)
	}
}

func TestCheck_Integration(t *testing.T) {
	expectedResponse := "test-response-hash"
	apiUrl := os.Getenv("SUPABASE_API_URL")
	apiKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	table := os.Getenv("SUPABASE_TABLE_NAME")

	if apiUrl == "" || apiKey == "" || table == "" {
		t.Skip(" Missing SUPABASE_API_URL, SUPABASE_SERVICE_ROLE_KEY, or SUPABASE_TABLE_NAME for integration test.")
	}

	client := supabaseclient.NewSupabaseClient(apiUrl, apiKey, table)

	hash := "test-hash-123"
	// Cache the response by payload
	if err := client.CacheByHash(hash, expectedResponse); err != nil {
		t.Fatalf(" CacheByHash() failed: %v", err)
	}

	// Check cache hit by payload
	cachedResp, hit, err := client.CheckByHash(hash)
	if err != nil {
		t.Fatalf("❌ CheckByHash() failed: %v", err)
	}
	if !hit {
		t.Errorf("❌ Expected cache hit after insertion, got miss")
	}
	if cachedResp != expectedResponse {
		t.Errorf("❌ Expected cached response '%s', got '%s'", expectedResponse, cachedResp)
	}
}
