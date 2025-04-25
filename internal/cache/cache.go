// here we will implement the wrapper to a semantic cache service eg. pgvector 
// TODO:-  implement a service that receives a request and check for semantically similar requests,
//         Must query the cache and immediately return, or else pass request to llm and store response 
//         in cache using embedding: response -> embedding -> store in cache
//         returns response to caller
package cache

// TODO: Implement cache service wrapper
// Stays in the middle of a request that has passed the rate limit test, receives and checks 
// the request for semantic similarity with existing requests in the cache.
// If a similar request is found, returns the cached response, else passes the request to llm
// and stores the response in the cache using embedding. 
// THis will essentially wrap our llm service and add semantic caching to it.

import (
	"agent-core/internal/cache/supabaseclient"
	"sync"
)

type Cache struct {
	Supabase *supabaseclient.SupabasePgvectorClient
	// Optional: in-memory cache for quick access
	memCache map[string]string
	mu       sync.RWMutex
}

// NewCache creates a new Cache backed by Supabase and with optional in-memory map
func NewCache(supabase *supabaseclient.SupabasePgvectorClient) *Cache {
	return &Cache{
		Supabase: supabase,
		memCache: make(map[string]string),
	}
}

// Check queries Supabase for the similarity score for a given query/embedding
func (c *Cache) Check(query string, embedding []float64) (float64, string, bool, error) {
	// Optionally check in-memory cache first
	c.mu.RLock()
	resp, found := c.memCache[query]
	c.mu.RUnlock()
	if found {
		return 1.0, resp, true, nil // Assume perfect similarity for in-memory
	}
	// Check persistent cache (Supabase)
	return c.Supabase.Check(query, embedding)
}

// CacheEmbedding stores the embedding and response in both Supabase and memory
func (c *Cache) CacheEmbedding(query string, embedding []float64, response string) error {
	c.mu.Lock()
	c.memCache[query] = response
	c.mu.Unlock()
	return c.Supabase.CacheEmbedding(query, embedding, response)
}

