package supabaseclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

type SupabaseClient struct {
	SupabaseURL string
	SupabaseKey string
	TableName   string
	Client      *http.Client
}

func NewSupabaseClient(apiURL, apiKey, tableName string) *SupabaseClient {
	return &SupabaseClient{
		SupabaseURL: apiURL,
		SupabaseKey: apiKey,
		TableName:   tableName,
		Client:      &http.Client{Timeout: 10 * time.Second},
	}
}


// APICacheRecord represents a cached row in the database
type APICacheRecord struct {
	Hash      string `json:"hash"`
	Response  string `json:"response"`
}

// CheckByHash returns cached response by hash if it's not older than 20 minutes
func (c *SupabaseClient) CheckByHash(hash string) (string, bool, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?hash=eq.%s&select=response", c.SupabaseURL, c.TableName, hash)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("apikey", c.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+c.SupabaseKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var records []APICacheRecord
		json.Unmarshal(body, &records)
		if len(records) == 0 {
			return "", false, nil
		}

		return records[0].Response, true, nil
	}
	return "", false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
}

// CacheByHash inserts a new record into the Supabase table
func (c *SupabaseClient) CacheByHash(hash string, response string) error {
	url := fmt.Sprintf("%s/rest/v1/%s", c.SupabaseURL, c.TableName)

	payload := APICacheRecord{
		Hash:     hash,
		Response: response,
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("apikey", c.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+c.SupabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("cache insert failed: status=%d, body=%s", resp.StatusCode, string(body))
}