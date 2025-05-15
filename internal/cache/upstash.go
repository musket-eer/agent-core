package cache

import (
	"bytes"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

// UpstashClient wraps the Upstash Redis REST API
// See: https://docs.upstash.com/redis/restapi/
type UpstashClient struct {
	RestUrl   string
	RestToken string
}

// Global singleton Upstash client
var Upstash *UpstashClient

// NewUpstashClient creates a new UpstashClient from parameters
func NewUpstashClient(restUrl, restToken string) *UpstashClient {
	return &UpstashClient{
		RestUrl:   restUrl,
		RestToken: restToken,
	}
}

// Store sets a key to a JSON value
func (u *UpstashClient) Store(key string, value []byte) error {
	url := fmt.Sprintf("%s/set/%s", u.RestUrl, key)
	req, err := http.NewRequest("POST", url, bytes.NewReader(value))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", u.RestToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upstash store failed: %s", string(body))
	}
	return nil
}


// Retrieve gets a value by key (as raw JSON string)
func (u *UpstashClient) Retrieve(key string) ([]byte, error) {
	url := fmt.Sprintf("%s/get/%s", u.RestUrl, key)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", u.RestToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upstash retrieve failed: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respObj struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(body, &respObj); err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(respObj.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	return decoded, nil
}

// Check returns true if key exists in Upstash
func (u *UpstashClient) Check(key string) (bool, error) {
	url := fmt.Sprintf("%s/exists/%s", u.RestUrl, key)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", u.RestToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("upstash check failed: %s", string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	var respObj struct {
		Result int `json:"result"`
	}
	if err := json.Unmarshal(body, &respObj); err != nil {
		return false, err
	}
	return respObj.Result == 1, nil
}
