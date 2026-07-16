package fx

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// liveEntry is a cached live rate for one pair, with its fetch time.
type liveEntry struct {
	rate Rate
	at   time.Time
}

// histEntry is a cached history series for one pair, with its fetch time.
type histEntry struct {
	points []HistPoint
	at     time.Time
	stale  bool
}

// Cache is a keyed, on-demand TTL cache for live rates and history series,
// one entry per currency pair. Unlike a background-refreshing cache, entries
// are only fetched when requested, and then reused until they expire.
type Cache struct {
	client  *http.Client
	baseURL string

	liveTTL, histTTL time.Duration
	Now              func() time.Time // seam for tests

	mu   sync.RWMutex
	live map[string]liveEntry
	hist map[string]histEntry
}

func NewCache(client *http.Client, baseURL string) *Cache {
	return &Cache{
		client:  client,
		baseURL: baseURL,
		liveTTL: 60 * time.Second,
		histTTL: 24 * time.Hour,
		Now:     time.Now,
		live:    map[string]liveEntry{},
		hist:    map[string]histEntry{},
	}
}

func key(from, to string) string { return from + ">" + to }

// Rate returns the live rate for from/to, fetching it if the cached entry is
// missing or expired. On a fetch error with a cached entry present, it
// returns the last-known value marked stale rather than propagating the
// error.
func (c *Cache) Rate(ctx context.Context, from, to string) (Rate, error) {
	k := key(from, to)

	c.mu.RLock()
	e, ok := c.live[k]
	c.mu.RUnlock()
	if ok && c.Now().Sub(e.at) < c.liveTTL {
		return e.rate, nil
	}

	r, err := Fetch(ctx, c.client, c.baseURL, from, to) // outside the lock
	if err != nil {
		if ok {
			stale := e.rate
			stale.Stale = true
			return stale, nil
		}
		return Rate{}, err
	}

	c.mu.Lock()
	c.live[k] = liveEntry{rate: r, at: c.Now()}
	c.mu.Unlock()
	return r, nil
}

// History returns the daily series for from/to over the longest window any
// timeframe needs (max of 90d and YTD), fetching it if the cached entry is
// missing or expired. On a fetch error with a cached entry present, it
// returns the last-known series marked stale rather than propagating the
// error.
func (c *Cache) History(ctx context.Context, from, to string) (RateHistory, error) {
	pts, stale, err := c.historyPoints(ctx, from, to)
	if err != nil {
		return RateHistory{}, err
	}
	return toRateHistory(from, to, pts, stale), nil
}

// historyPoints is the shared implementation behind History and Context: it
// returns the raw HistPoint series (rather than the string-dated RateHistory
// shape) so Context can hand them straight to assess without a round trip
// through date formatting/parsing.
func (c *Cache) historyPoints(ctx context.Context, from, to string) ([]HistPoint, bool, error) {
	k := key(from, to)

	c.mu.RLock()
	e, ok := c.hist[k]
	c.mu.RUnlock()
	if ok && c.Now().Sub(e.at) < c.histTTL {
		return e.points, e.stale, nil
	}

	asOf := c.Now()
	start := windowStart(timeframe{"90d", 90}, asOf)
	if ytd := windowStart(timeframe{"YTD", 0}, asOf); ytd.Before(start) {
		start = ytd
	}
	pts, err := FetchHistory(ctx, c.client, c.baseURL, from, to, start, asOf) // outside the lock
	if err != nil {
		if ok {
			return e.points, true, nil
		}
		return nil, false, err
	}

	c.mu.Lock()
	c.hist[k] = histEntry{points: pts, at: c.Now(), stale: false}
	c.mu.Unlock()
	return pts, false, nil
}

func toRateHistory(from, to string, points []HistPoint, stale bool) RateHistory {
	pts := make([]Point, len(points))
	for i, p := range points {
		pts[i] = Point{Date: p.Date.Format("2006-01-02"), Value: p.Value}
	}
	return RateHistory{From: from, To: to, Points: pts, Stale: stale}
}

// Context assembles the live rate and per-timeframe historical assessment
// for a pair. A history fetch failure does not fail the call: it is
// reflected as HistoryStale with empty/unknown timeframes.
func (c *Cache) Context(ctx context.Context, from, to string, asOf time.Time) (RateContext, error) {
	rate, err := c.Rate(ctx, from, to)
	if err != nil {
		return RateContext{}, err
	}

	points, historyStale, herr := c.historyPoints(ctx, from, to)
	if herr != nil {
		historyStale = true
	}

	return RateContext{
		From:         from,
		To:           to,
		Rate:         rate.Rate,
		Inverse:      rate.Inverse,
		Stale:        rate.Stale,
		FetchedAt:    rate.FetchedAt,
		HistoryStale: historyStale,
		Timeframes:   assess(rate.Rate, points, asOf),
	}, nil
}
