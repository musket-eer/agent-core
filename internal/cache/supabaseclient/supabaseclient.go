package supabaseclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SupabasePgvectorClient struct {
	ApiUrl     string
	ApiKey     string
	TableName  string
}

type VectorRecord struct {
	Key       string    `json:"key"`
	Embedding []float64 `json:"embedding"`
	Response  string    `json:"response"`
}

// Check queries Supabase for the most similar embedding to the provided key and returns the similarity score.
func (c *SupabasePgvectorClient) Check(key string, embedding []float64) (float64, string, bool, error) {
	// This is a placeholder for the actual HTTP request to Supabase pgvector extension.
	// You would use the embedding to run a similarity search (e.g. cosine similarity) in SQL.
	// Here, we just return not found.
	return 0.0, "", false, nil
}

// CacheEmbedding inserts a new embedding/response pair into Supabase.
func (c *SupabasePgvectorClient) CacheEmbedding(key string, embedding []float64, response string) error {
	record := VectorRecord{Key: key, Embedding: embedding, Response: response}
	body, _ := json.Marshal(record)
	url := fmt.Sprintf("%s/rest/v1/%s", c.ApiUrl, c.TableName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed to cache embedding: %s", resp.Status)
	}
	return nil
}
