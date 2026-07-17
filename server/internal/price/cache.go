package price

import (
	"context"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// maxConcurrentFetches bounds how many Stooq requests a single Quotes batch
// issues at once, so a large holdings list can't fan out unbounded network
// calls.
const maxConcurrentFetches = 6

// Req is one ticker lookup for a Quotes batch. A non-nil Manual short-
// circuits the network entirely and is echoed back as the quote.
type Req struct {
	Ticker   string
	Currency string
	Manual   *float64
}

// entry is a cached quote for one ticker, with its fetch time.
type entry struct {
	quote Quote
	at    time.Time
}

// Cache is a keyed, on-demand TTL cache for share prices, one entry per
// ticker. Unlike a background-refreshing cache, entries are only fetched
// when requested, and then reused until they expire.
type Cache struct {
	client *http.Client
	base   string // test hook; defaults to StooqBase

	ttl time.Duration
	Now func() time.Time // seam for tests

	mu      sync.RWMutex
	entries map[string]entry
}

func NewCache(client *http.Client) *Cache {
	return &Cache{
		client:  client,
		base:    StooqBase,
		ttl:     5 * time.Minute,
		Now:     time.Now,
		entries: map[string]entry{},
	}
}

// Quote returns the latest price for ticker, fetching it if the cached
// entry is missing or expired. On a fetch error with a cached entry
// present, it returns the last-known value rather than propagating the
// error (stale-on-error).
func (c *Cache) Quote(ctx context.Context, ticker, currency string) (Quote, error) {
	c.mu.RLock()
	e, ok := c.entries[ticker]
	c.mu.RUnlock()
	if ok && c.Now().Sub(e.at) < c.ttl {
		return e.quote, nil
	}

	q, err := Fetch(ctx, c.client, c.base, ticker, currency) // outside the lock
	if err != nil {
		if ok {
			return e.quote, nil
		}
		return Quote{}, err
	}

	c.mu.Lock()
	c.entries[ticker] = entry{quote: q, at: c.Now()}
	c.mu.Unlock()
	return q, nil
}

// Quotes resolves a batch of ticker requests concurrently, bounded to
// maxConcurrentFetches in-flight fetches at a time. It is partial-failure
// tolerant: a failed or not-found ticker yields Quote{Ticker, Found:false}
// in the result rather than failing the whole batch. A request with a
// non-nil Manual price never touches the network.
func (c *Cache) Quotes(ctx context.Context, reqs []Req) map[string]Quote {
	results := make(map[string]Quote, len(reqs))
	var resMu sync.Mutex

	sem := make(chan struct{}, maxConcurrentFetches)
	g, gctx := errgroup.WithContext(ctx)

	for _, req := range reqs {
		req := req
		if req.Manual != nil {
			resMu.Lock()
			results[req.Ticker] = Quote{Ticker: req.Ticker, Price: *req.Manual, Found: true}
			resMu.Unlock()
			continue
		}

		g.Go(func() error {
			select {
			case sem <- struct{}{}:
			case <-gctx.Done():
				resMu.Lock()
				results[req.Ticker] = Quote{Ticker: req.Ticker, Found: false}
				resMu.Unlock()
				return nil
			}
			defer func() { <-sem }()

			q, err := c.Quote(gctx, req.Ticker, req.Currency)
			if err != nil || !q.Found {
				q = Quote{Ticker: req.Ticker, Found: false}
			}
			resMu.Lock()
			results[req.Ticker] = q
			resMu.Unlock()
			return nil
		})
	}

	_ = g.Wait() // partial-failure tolerant: individual errors are folded into results, never propagated
	return results
}
