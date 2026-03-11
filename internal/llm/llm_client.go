package llm

import (
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"agent-core/internal/llm/configs"
	"fmt"
)

// LLMClient defines the interface all clients must implement
type LLMClient interface {
	Call(endpoint string, params map[string]interface{}) (string, error)
}

type routedLLMClient struct {
	route string
	pool  []string
	start int
}

// global client registry
var llmClients = make(map[string]LLMClient)
var llmClientsMu sync.RWMutex
var llmRouteMu sync.Mutex
var llmRouteRoundRobin = make(map[string]uint64)

// RegisterLLMClient stores a client with a given name
func RegisterLLMClient(name string, client LLMClient) {
	llmClientsMu.Lock()
	defer llmClientsMu.Unlock()
	llmClients[name] = client
}

// GetLLMClient retrieves a client by name
func GetLLMClient(name string) (LLMClient, error) {
	// Virtual routed aliases: these are dynamically load-balanced across
	// all registered provider-specific aliases by role.
	if name == "general" || name == "chat" {
		if routedClient, selectedAlias, pool := getRoutedClient(name); routedClient != nil {
			log.Printf("[LLMRouter] route=%s selected=%s pool=%v", name, selectedAlias, pool)
			return routedClient, nil
		}
	}

	llmClientsMu.RLock()
	defer llmClientsMu.RUnlock()
	client, ok := llmClients[name]
	if !ok {
		return nil, fmt.Errorf("no LLM client registered with name: %s", name)
	}
	return client, nil
}

// Optional: List all registered clients
func ListLLMClients() []string {
	llmClientsMu.RLock()
	defer llmClientsMu.RUnlock()
	keys := make([]string, 0, len(llmClients))
	for k := range llmClients {
		keys = append(keys, k)
	}
	return keys
}

func getRoutedClient(route string) (LLMClient, string, []string) {
	pool := buildRoutePool(route)
	if len(pool) == 0 {
		return nil, "", nil
	}

	llmClientsMu.RLock()
	available := make([]string, 0, len(pool))
	for _, alias := range pool {
		if _, ok := llmClients[alias]; ok {
			available = append(available, alias)
		}
	}
	if len(available) == 0 {
		llmClientsMu.RUnlock()
		return nil, "", pool
	}
	idx := nextRouteIndex(route, len(available))
	llmClientsMu.RUnlock()

	return &routedLLMClient{
		route: route,
		pool:  available,
		start: idx,
	}, available[idx], available
}

func nextRouteIndex(route string, n int) int {
	llmRouteMu.Lock()
	defer llmRouteMu.Unlock()
	cur := llmRouteRoundRobin[route]
	idx := int(cur % uint64(n))
	llmRouteRoundRobin[route] = cur + 1
	return idx
}

func buildRoutePool(route string) []string {
	envName := "LLM_ROUTE_" + strings.ToUpper(route) // LLM_ROUTE_GENERAL / LLM_ROUTE_CHAT
	if raw := strings.TrimSpace(os.Getenv(envName)); raw != "" {
		parts := strings.Split(raw, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			a := strings.TrimSpace(p)
			if a != "" {
				out = append(out, a)
			}
		}
		return dedupeStrings(out)
	}

	// Safe defaults: route only providers with known endpoint coverage/payload compatibility.
	// Bedrock is excluded unless explicitly configured through LLM_ROUTE_*.
	var preferred []string
	switch route {
	case "general":
		preferred = []string{
			"deepseek-general",
		}
	case "chat":
		preferred = []string{
			"deepseek-chat",
		}
	default:
		preferred = nil
	}
	llmClientsMu.RLock()
	out := make([]string, 0, len(preferred))
	for _, alias := range preferred {
		if _, ok := llmClients[alias]; ok {
			out = append(out, alias)
		}
	}
	llmClientsMu.RUnlock()
	sort.Strings(out)
	return dedupeStrings(out)
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func (c *routedLLMClient) Call(endpoint string, params map[string]interface{}) (string, error) {
	if len(c.pool) == 0 {
		return "", fmt.Errorf("no providers available for route: %s", c.route)
	}

	var errs []string
	for attempt := 0; attempt < len(c.pool); attempt++ {
		alias := c.pool[(c.start+attempt)%len(c.pool)]

		llmClientsMu.RLock()
		client, ok := llmClients[alias]
		llmClientsMu.RUnlock()
		if !ok || client == nil {
			errs = append(errs, fmt.Sprintf("%s:not_registered", alias))
			continue
		}

		resp, err := client.Call(endpoint, params)
		if err == nil {
			if attempt > 0 {
				log.Printf("[LLMRouter] route=%s recovered_with=%s attempt=%d", c.route, alias, attempt+1)
			}
			return resp, nil
		}

		errs = append(errs, fmt.Sprintf("%s:%v", alias, err))
		if !isRetryableProviderError(err) {
			return "", err
		}

		if attempt < len(c.pool)-1 {
			log.Printf("[LLMRouter] route=%s failover from=%s reason=retryable_provider_error next=%s",
				c.route, alias, c.pool[(c.start+attempt+1)%len(c.pool)])
		}
	}

	return "", fmt.Errorf("all providers failed for route=%s: %s", c.route, strings.Join(errs, " | "))
}

func isRetryableProviderError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "429") ||
		strings.Contains(s, "rate limit") ||
		strings.Contains(s, "quota exceeded") ||
		strings.Contains(s, "resource_exhausted") ||
		strings.Contains(s, "too many requests")
}

// Call routes to the appropriate client based on the endpoint
func Call(endpoint string, params map[string]interface{}) (string, error) {
	config, ok := configs.GetDeepseekConfig(endpoint)
	if !ok {
		return "", fmt.Errorf("no config found for endpoint: %s", endpoint)
	}

	clientName := config.ClientName
	if clientName == "" {
		clientName = "general" // fallback default
	}

	client, err := GetLLMClient(clientName)
	if err != nil {
		return "", fmt.Errorf("failed to route LLM call: %w", err)
	}

	return client.Call(endpoint, params)
}
