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
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"agent-core/internal/cache/supabaseclient"
)

type Cache struct {
	Supabase *supabaseclient.SupabaseClient
	memCache map[string]string
	mu       sync.RWMutex
}

// HashPayload returns a SHA256 hash of the payload string
func HashPayload(payload string) string {
	h := sha256.New()
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// NewCache creates a new Cache backed by Supabase and with optional in-memory map
func NewCache(supabase *supabaseclient.SupabaseClient) *Cache {
	return &Cache{
		Supabase: supabase,
		memCache: make(map[string]string),
	}
}

// CheckByPayload checks cache by payload hash (in-memory first, then Supabase)
func (c *Cache) CheckByPayload(payload string) (string, bool, error) {
	hash := HashPayload(payload)
	c.mu.RLock()
	resp, found := c.memCache[hash]
	c.mu.RUnlock()
	if found {
		return resp, true, nil
	}
	resp, hit, err := c.Supabase.CheckByHash(hash)
	if hit && err == nil {
		c.mu.Lock()
		c.memCache[hash] = resp
		c.mu.Unlock()
	}
	return resp, hit, err
}

// CacheByPayload stores the response for a payload hash in both Supabase and memory
func (c *Cache) CacheByPayload(payload string, response string) error {
	hash := HashPayload(payload)
	c.mu.Lock()
	c.memCache[hash] = response
	c.mu.Unlock()
	return c.Supabase.CacheByHash(hash, response)
}

