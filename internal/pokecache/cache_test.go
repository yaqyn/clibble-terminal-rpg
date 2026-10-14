package pokecache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAddGetCopiesAndExpiry(t *testing.T) {
	c := NewCache(time.Hour)
	defer c.Close()
	input := []byte("pokemon")
	c.Add("key", input)
	input[0] = 'X'
	got, ok := c.Get("key")
	if !ok || string(got) != "pokemon" {
		t.Fatalf("got %q, %v", got, ok)
	}
	got[0] = 'Y'
	if again, _ := c.Get("key"); string(again) != "pokemon" {
		t.Fatal("Get exposed mutable storage")
	}
	c.mu.Lock()
	entry := c.entries["key"]
	entry.createdAt = time.Now().Add(-2 * time.Hour)
	c.entries["key"] = entry
	c.mu.Unlock()
	if _, ok := c.Get("key"); ok {
		t.Fatal("expired entry returned")
	}
	if _, ok := c.Get("missing"); ok {
		t.Fatal("missing key found")
	}
}
func TestReapAndBounds(t *testing.T) {
	c := NewCache(time.Hour)
	defer c.Close()
	for i := 0; i < maxEntries+20; i++ {
		c.Add(fmt.Sprint(i), []byte("data"))
	}
	if len(c.entries) != maxEntries {
		t.Fatalf("entries=%d", len(c.entries))
	}
	c.reap(time.Now().Add(2 * time.Hour))
	if len(c.entries) != 0 || c.bytes != 0 {
		t.Fatal("reap did not clear expired entries")
	}
	c.Add("large", make([]byte, maxBytes+1))
	if _, ok := c.Get("large"); ok {
		t.Fatal("oversize entry accepted")
	}
}
func TestConcurrentAccessAndClose(t *testing.T) {
	c := NewCache(time.Millisecond)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.Add("key", []byte("value"))
				c.Get("key")
			}
		}()
	}
	wg.Wait()
	c.Close()
	c.Close()
	select {
	case <-c.stopped:
	default:
		t.Fatal("worker still running")
	}
}
