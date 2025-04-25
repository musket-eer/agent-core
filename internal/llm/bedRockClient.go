package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"io"

	"agent-core/internal/llm/configs"
)

type BedrockClient struct {
	BaseURL   string
	APIKey    string
	Model     string
	Headers   map[string]string // Optional custom headers
	HttpClient *http.Client
}

// CallBedrock sends a POST request to the given endpoint with the provided parameters
func (c *BedrockClient) CallBedrock(endpoint string, params map[string]interface{}) (string, error) {
	return c.Call(endpoint, params)
}

// Call implements the LLMClient interface
func (c *BedrockClient) Call(endpoint string, params map[string]interface{}) (string, error) {
	// MOCK: Return a bogus response for testing
	mockResp := map[string]interface{}{
		"endpoint": endpoint,
		"params": params,
		"result": "This is a mock Bedrock response.",
	}
	respBytes, _ := json.Marshal(mockResp)
	return string(respBytes), nil

	// Fetch config for the endpoint
	config, ok := configs.GetConfig(endpoint)
	if ok {
		// Set model and other config params if not already present
		if _, exists := params["model"]; !exists {
			params["model"] = config.ModelName
		}
		if _, exists := params["temperature"]; !exists {
			params["temperature"] = config.Temperature
		}
		if _, exists := params["max_tokens"]; !exists {
			params["max_tokens"] = config.MaxTokens
		}
		// Optionally merge other params
		for k, v := range config.OtherParams {
			if _, exists := params[k]; !exists {
				params[k] = v
			}
		}
		// If you want to support custom headers from config, add a Headers field to BedrockConfig and set them here.
	}

	fullURL := fmt.Sprintf("%s/%s", c.BaseURL, endpoint)

	// Encode parameters to JSON
	bodyBytes, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("error encoding parameters: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	// Send request
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to Bedrock: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response: %d\n%s", resp.StatusCode, respBody)
	}

	return string(respBody), nil
}
