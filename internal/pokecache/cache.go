// Package pokecache implements a bounded, concurrent TTL cache of byte slices.
package pokecache

import (
	"sync"
	"time"
)

const maxEntries = 256
const maxBytes = 16 << 20
const maxKeyBytes = 4096

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}
type Cache struct {
	mu       sync.Mutex
	entries  map[string]cacheEntry
	interval time.Duration
	bytes    int
	done     chan struct{}
	stopped  chan struct{}
	once     sync.Once
}

// NewCache starts an expiry worker. Call Close when the cache is no longer used.
// Nonpositive intervals use the default five-minute TTL.
func NewCache(interval time.Duration) *Cache {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	c := &Cache{entries: make(map[string]cacheEntry), interval: interval, done: make(chan struct{}), stopped: make(chan struct{})}
	go c.reapLoop()
	return c
}

// Add copies val so callers cannot mutate a cached response.
func (c *Cache) Add(key string, val []byte) {
	if len(val) > maxBytes || len(key) > maxKeyBytes {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if old, ok := c.entries[key]; ok {
		c.bytes -= len(old.val)
		delete(c.entries, key)
	}
	for len(c.entries) >= maxEntries || c.bytes+len(val) > maxBytes {
		var oldest string
		var earliest time.Time
		for k, v := range c.entries {
			if earliest.IsZero() || v.createdAt.Before(earliest) {
				oldest, earliest = k, v.createdAt
			}
		}
		c.bytes -= len(c.entries[oldest].val)
		delete(c.entries, oldest)
	}
	c.entries[key] = cacheEntry{createdAt: time.Now(), val: append([]byte(nil), val...)}
	c.bytes += len(val)
}

// Get checks expiry immediately and returns an independent copy.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Since(entry.createdAt) >= c.interval {
		c.bytes -= len(entry.val)
		delete(c.entries, key)
		return nil, false
	}
	return append([]byte(nil), entry.val...), true
}
func (c *Cache) reap(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, entry := range c.entries {
		if now.Sub(entry.createdAt) >= c.interval {
			c.bytes -= len(entry.val)
			delete(c.entries, key)
		}
	}
}
func (c *Cache) reapLoop() {
	defer close(c.stopped)
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			c.reap(now)
		case <-c.done:
			return
		}
	}
}

// Close stops the expiry worker and is safe to call more than once.
func (c *Cache) Close() { c.once.Do(func() { close(c.done) }); <-c.stopped }
