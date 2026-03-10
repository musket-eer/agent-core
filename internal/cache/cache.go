// here we will implement the wrapper to a semantic cache service eg. pgvector
// TODO:-  implement a service that receives a request and check for semantically similar requests,
//
//	Must query the cache and immediately return, or else pass request to llm and store response
//	in cache using embedding: response -> embedding -> store in cache
//	returns response to caller
package cache

// TODO: Implement cache service wrapper
// Stays in the middle of a request that has passed the rate limit test, receives and checks
// the request for semantic similarity with existing requests in the cache.
// If a similar request is found, returns the cached response, else passes the request to llm
// and stores the response in the cache using embedding.
// THis will essentially wrap our llm service and add semantic caching to it.

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"sync"
	"time"
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

	mu       sync.RWMutex
	memCache map[string]*memEntry
	memList  *list.List

	maxEntries int
	memTTL     time.Duration
}

type memEntry struct {
	key       string
	value     string
	expiresAt time.Time
	element   *list.Element
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
	maxEntries := getenvInt("CACHE_MEM_MAX_ENTRIES", 2000)
	memTTL := getenvDuration("CACHE_MEM_TTL_SECONDS", 900*time.Second) // 15m
	janitorInterval := getenvDuration("CACHE_MEM_JANITOR_SECONDS", 60*time.Second)

	if maxEntries < 1 {
		maxEntries = 2000
	}
	if memTTL <= 0 {
		memTTL = 900 * time.Second
	}
	if janitorInterval <= 0 {
		janitorInterval = 60 * time.Second
	}

	c := &Cache{
		Upstash:    upstash,
		memCache:   make(map[string]*memEntry),
		memList:    list.New(),
		maxEntries: maxEntries,
		memTTL:     memTTL,
	}
	go c.startJanitor(janitorInterval)
	return c
}

// CheckByPayload checks cache by payload hash (in-memory first, then upstash)
func (c *Cache) CheckByPayload(payload string) ([]byte, bool, error) {
	hash := HashPayload(payload)

	if resp, found := c.getMem(hash); found {
		return []byte(resp), true, nil
	}

	if c.Upstash == nil {
		return nil, false, nil
	}

	hit, err := c.Upstash.Check(hash)
	if hit && err == nil {
		resp, err := c.Upstash.Retrieve(hash)
		if err == nil {
			c.setMem(hash, string(resp))
			return resp, true, nil
		}
		return nil, false, err
	}
	return nil, false, err
}

// CacheByPayload stores the response for a payload hash in both Supabase and memory
func (c *Cache) CacheByPayload(payload string, response []byte) error {
	hash := HashPayload(payload)

	c.setMem(hash, string(response))

	if c.Upstash == nil {
		return nil
	}

	if err := c.Upstash.Store(hash, response); err != nil {
		return err
	}
	return nil
}

func (c *Cache) getMem(key string) (string, bool) {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.memCache[key]
	if !ok {
		return "", false
	}
	if now.After(entry.expiresAt) {
		c.removeEntry(entry)
		return "", false
	}
	c.memList.MoveToFront(entry.element)
	return entry.value, true
}

func (c *Cache) setMem(key, value string) {
	now := time.Now()
	expires := now.Add(c.memTTL)

	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.memCache[key]; ok {
		entry.value = value
		entry.expiresAt = expires
		c.memList.MoveToFront(entry.element)
		return
	}

	elem := c.memList.PushFront(key)
	c.memCache[key] = &memEntry{
		key:       key,
		value:     value,
		expiresAt: expires,
		element:   elem,
	}
	c.evictIfNeededLocked()
}

func (c *Cache) evictIfNeededLocked() {
	for len(c.memCache) > c.maxEntries {
		back := c.memList.Back()
		if back == nil {
			return
		}
		key, _ := back.Value.(string)
		entry := c.memCache[key]
		if entry != nil {
			c.removeEntry(entry)
		} else {
			c.memList.Remove(back)
		}
	}
}

func (c *Cache) removeEntry(entry *memEntry) {
	delete(c.memCache, entry.key)
	if entry.element != nil {
		c.memList.Remove(entry.element)
	}
}

func (c *Cache) startJanitor(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		c.pruneExpired()
	}
}

func (c *Cache) pruneExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, entry := range c.memCache {
		if now.After(entry.expiresAt) {
			c.removeEntry(entry)
		}
	}
	c.evictIfNeededLocked()
}

func getenvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return time.Duration(v) * time.Second
}
