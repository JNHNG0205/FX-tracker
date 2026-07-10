package fx

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type Cache struct {
	client  *http.Client
	baseURL string

	mu      sync.RWMutex
	current Rate
	hasData bool
}

func NewCache(client *http.Client, baseURL string) *Cache {
	// Before any successful refresh, current is a zero-value Rate; mark it
	// stale so readers never mistake it for a real, fresh rate.
	return &Cache{client: client, baseURL: baseURL, current: Rate{Stale: true}}
}

// Refresh fetches a fresh rate. On failure it keeps the last-known value and
// marks it stale, returning the error so callers/log can see it.
func (c *Cache) Refresh(ctx context.Context) error {
	rate, err := Fetch(ctx, c.client, c.baseURL)
	if err != nil {
		c.mu.Lock()
		if c.hasData {
			c.current.Stale = true
		}
		c.mu.Unlock()
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = rate
	c.hasData = true
	return nil
}

func (c *Cache) Get() Rate {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current
}

// Run refreshes immediately, then on every tick until ctx is cancelled.
func (c *Cache) Run(ctx context.Context, interval time.Duration) {
	if err := c.Refresh(ctx); err != nil {
		log.Printf("fx: initial refresh failed: %v", err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("fx: refresh loop stopped")
			return
		case <-ticker.C:
			if err := c.Refresh(ctx); err != nil {
				log.Printf("fx: refresh failed (serving stale): %v", err)
			}
		}
	}
}
