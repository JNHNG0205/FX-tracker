package fx

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type RangeContext struct {
	Current    float64   `json:"current_myr_usd"`
	Min        float64   `json:"min_myr_usd"`
	Max        float64   `json:"max_myr_usd"`
	Assessment string    `json:"assessment"`
	Stale      bool      `json:"stale"`
	FetchedAt  time.Time `json:"fetched_at"`
}

type Cache struct {
	client  *http.Client
	baseURL string

	mu       sync.RWMutex
	current  Rate
	min, max float64
	hasData  bool
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
	if !c.hasData || rate.MyrUsd < c.min {
		c.min = rate.MyrUsd
	}
	if !c.hasData || rate.MyrUsd > c.max {
		c.max = rate.MyrUsd
	}
	c.hasData = true
	return nil
}

func (c *Cache) Get() Rate {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current
}

func (c *Cache) Context() RangeContext {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return RangeContext{
		Current:    c.current.MyrUsd,
		Min:        c.min,
		Max:        c.max,
		Assessment: Assess(c.current.MyrUsd, c.min, c.max),
		Stale:      c.current.Stale,
		FetchedAt:  c.current.FetchedAt,
	}
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
