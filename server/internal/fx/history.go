package fx

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type Assessment struct {
	Label      string  `json:"label"`
	Assessment string  `json:"assessment"`
	Percentile int     `json:"percentile"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Samples    int     `json:"samples"`
	Start      string  `json:"start"`
	End        string  `json:"end"`
}

// timeframe: Days==0 means year-to-date.
type timeframe struct {
	label string
	days  int
}

var timeframes = []timeframe{
	{"7d", 7}, {"14d", 14}, {"30d", 30}, {"90d", 90}, {"YTD", 0},
}

type HistoryCache struct {
	client  *http.Client
	baseURL string
	Now     func() time.Time // seam for tests

	mu      sync.RWMutex
	points  []HistPoint // ascending by date
	stale   bool
	hasData bool
}

func NewHistoryCache(client *http.Client, baseURL string) *HistoryCache {
	return &HistoryCache{client: client, baseURL: baseURL, Now: time.Now, stale: true}
}

// windowStart returns the earliest date included for a timeframe as of `asOf`.
func windowStart(tf timeframe, asOf time.Time) time.Time {
	if tf.days == 0 { // YTD
		return time.Date(asOf.Year(), time.January, 1, 0, 0, 0, 0, asOf.Location())
	}
	return asOf.AddDate(0, 0, -tf.days)
}

// Refresh fetches the longest span any timeframe needs (max of 90d and YTD).
func (h *HistoryCache) Refresh(ctx context.Context) error {
	asOf := h.Now()
	start := windowStart(timeframe{"90d", 90}, asOf)
	if ytd := windowStart(timeframe{"YTD", 0}, asOf); ytd.Before(start) {
		start = ytd
	}
	pts, err := FetchHistory(ctx, h.client, h.baseURL, start, asOf)
	if err != nil {
		h.mu.Lock()
		if h.hasData {
			h.stale = true
		}
		h.mu.Unlock()
		return err
	}
	h.mu.Lock()
	h.points = pts
	h.stale = false
	h.hasData = true
	h.mu.Unlock()
	return nil
}

// Run refreshes immediately, then on every tick until ctx is cancelled.
func (h *HistoryCache) Run(ctx context.Context, interval time.Duration) {
	if err := h.Refresh(ctx); err != nil {
		log.Printf("fx history: initial refresh failed: %v", err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("fx history: refresh loop stopped")
			return
		case <-ticker.C:
			if err := h.Refresh(ctx); err != nil {
				log.Printf("fx history: refresh failed (serving stale): %v", err)
			}
		}
	}
}

func (h *HistoryCache) Stale() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.stale
}

// Assess ranks current against each timeframe's window, sliced from the cached
// series. Never blocks on the network (read lock only).
func (h *HistoryCache) Assess(current float64, asOf time.Time) []Assessment {
	h.mu.RLock()
	points := h.points
	h.mu.RUnlock()

	out := make([]Assessment, 0, len(timeframes))
	for _, tf := range timeframes {
		start := windowStart(tf, asOf)
		var window []float64
		var firstDate, lastDate time.Time
		for _, p := range points {
			if p.Date.Before(start) {
				continue
			}
			if window == nil {
				firstDate = p.Date
			}
			lastDate = p.Date
			window = append(window, p.MyrUsd)
		}
		pct, verdict, min, max := Percentile(current, window)
		a := Assessment{
			Label:      tf.label,
			Assessment: verdict,
			Percentile: pct,
			Min:        min,
			Max:        max,
			Samples:    len(window),
		}
		if len(window) > 0 {
			a.Start = firstDate.Format("2006-01-02")
			a.End = lastDate.Format("2006-01-02")
		}
		out = append(out, a)
	}
	return out
}
