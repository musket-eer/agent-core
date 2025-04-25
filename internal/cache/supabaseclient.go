package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

type SupabasePgvectorClient struct {
	ApiUrl    string
	ApiKey    string
	TableName string
}

type VectorRecord struct {
	Query     string    `json:"query"`
	Embedding []float64 `json:"embedding"`
	Response  string    `json:"response"`
}

// CacheEmbedding inserts a new (query, embedding, response) record into Supabase.
func (c *SupabasePgvectorClient) CacheEmbedding(query string, embedding []float64, response string) error {
	record := map[string]interface{}{
		"query":     query,
		"embedding": embedding,
		"response":  response,
	}

	body, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to encode embedding record: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/%s", c.ApiUrl, c.TableName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Supabase insert failed: %s\n%s", resp.Status, string(bodyBytes))
	}

	return nil
}

// Check uses Supabase RPC or a custom SQL endpoint to check for the nearest embedding match.
// You must create a Postgres function for vector search and expose it via RPC.
func (c *SupabasePgvectorClient) Check(embedding []float64) (string, float64, bool, error) {
	url := fmt.Sprintf("%s/rest/v1/rpc/match_embedding", c.ApiUrl)

	payload := map[string]interface{}{
		"embedding": embedding,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", 0.0, false, fmt.Errorf("failed to encode check payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", 0.0, false, fmt.Errorf("failed to create check request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0.0, false, fmt.Errorf("Supabase RPC check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", 0.0, false, fmt.Errorf("Supabase returned error: %s\n%s", resp.Status, string(bodyBytes))
	}

	var results []struct {
		Response   json.RawMessage `json:"response"`  // Flexible JSON parsing
		Similarity float64         `json:"similarity"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", 0.0, false, fmt.Errorf("failed to decode Supabase RPC response: %w", err)
	}

	if len(results) == 0 {
		return "", 0.0, false, nil
	}

	return string(results[0].Response), results[0].Similarity, true, nil
}

// READY FOR USE. WE NEED TO MODIFY THE DIMENSIONS OF THE EMBEDDING IN SUPABASE TABLE
