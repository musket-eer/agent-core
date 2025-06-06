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
)
var Global *Cache

func InitGlobal(c *Cache) {
	Global = c
}

func Get() *Cache {
	if Global == nil {
		panic("cache.Global is not initialized")
	}
	return Global
}

type Cache struct {
	Upstash *UpstashClient
	memCache map[string]string
	mu       sync.RWMutex
}

// HashPayload returns a SHA256 hash of the payload string
func HashPayload(payload string) string {
	h := sha256.New()
	h.Write([]byte(payload))
	hash := hex.EncodeToString(h.Sum(nil))
	return hash
}

// NewCache creates a new Cache backed by upstash and with optional in-memory map
func NewCache(upstash *UpstashClient) *Cache {
	return &Cache{
		Upstash: upstash,
		memCache: make(map[string]string),
	}
}

// CheckByPayload checks cache by payload hash (in-memory first, then upstash)
func (c *Cache) CheckByPayload(payload string) ([]byte, bool, error) {
	hash := HashPayload(payload)
	c.mu.RLock()
	resp, found := c.memCache[hash]
	c.mu.RUnlock()
	if found {
		return []byte(resp), true, nil
	}
	hit, err := c.Upstash.Check(hash)
	if hit && err == nil {
		resp, err := c.Upstash.Retrieve(hash)
		if err == nil {
			c.mu.Lock()
			c.memCache[hash] = string(resp)
			c.mu.Unlock()
			return resp, true, nil
		}
		return nil, false, err
	}
	return nil, false, err
}

// CacheByPayload stores the response for a payload hash in both Supabase and memory
func (c *Cache) CacheByPayload(payload string, response []byte) error {
	hash := HashPayload(payload)
	c.mu.Lock()
	c.memCache[hash] = string(response)
	c.mu.Unlock()
	err := c.Upstash.Store(hash, response)
	if err != nil {
		return err
	}
	return nil
}
