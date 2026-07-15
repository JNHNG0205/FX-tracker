package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheRatePerPairAndTTL(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		to := r.URL.Query().Get("to")
		_, _ = w.Write([]byte(`{"rates":{"` + to + `":0.21}}`))
	}))
	defer srv.Close()
	c := NewCache(srv.Client(), srv.URL)
	fixed := mustDate("2026-07-15")
	c.Now = func() time.Time { return fixed }
	r1, err := c.Rate(context.Background(), "MYR", "USD")
	if err != nil {
		t.Fatal(err)
	}
	if r1.Rate != 0.21 || r1.From != "MYR" || r1.To != "USD" {
		t.Fatalf("%+v", r1)
	}
	// second call for same pair within TTL: no new upstream call
	_, _ = c.Rate(context.Background(), "MYR", "USD")
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 upstream call, got %d", calls)
	}
	// different pair: separate fetch
	_, _ = c.Rate(context.Background(), "MYR", "EUR")
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2, got %d", calls)
	}
}

func TestCacheRateStaleOnError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			_, _ = w.Write([]byte(`{"rates":{"USD":0.21}}`))
			return
		}
		w.WriteHeader(500)
	}))
	defer srv.Close()
	c := NewCache(srv.Client(), srv.URL)
	c.liveTTL = 0 // force re-fetch
	_, _ = c.Rate(context.Background(), "MYR", "USD")
	got, err := c.Rate(context.Background(), "MYR", "USD")
	if err != nil {
		t.Fatalf("expected stale fallback, got err %v", err)
	}
	if !got.Stale || got.Rate != 0.21 {
		t.Fatalf("expected stale last-known, got %+v", got)
	}
}

func TestCacheHistoryPerPairAndTTL(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		to := r.URL.Query().Get("to")
		_, _ = w.Write([]byte(`{"rates":{"2026-07-08":{"` + to + `":0.24},"2026-07-09":{"` + to + `":0.25}}}`))
	}))
	defer srv.Close()
	c := NewCache(srv.Client(), srv.URL)
	c.Now = func() time.Time { return mustDate("2026-07-09") }

	h1, err := c.History(context.Background(), "MYR", "USD")
	if err != nil {
		t.Fatal(err)
	}
	if h1.From != "MYR" || h1.To != "USD" || len(h1.Points) != 2 || h1.Stale {
		t.Fatalf("%+v", h1)
	}
	if h1.Points[0].Date != "2026-07-08" || h1.Points[0].Value != 0.24 {
		t.Fatalf("points[0] = %+v", h1.Points[0])
	}

	// second call for same pair within TTL: no new upstream call
	_, _ = c.History(context.Background(), "MYR", "USD")
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 upstream call, got %d", calls)
	}

	// different pair: separate fetch
	_, _ = c.History(context.Background(), "MYR", "EUR")
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2, got %d", calls)
	}
}

func TestCacheHistoryStaleOnError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			_, _ = w.Write([]byte(`{"rates":{"2026-07-09":{"USD":0.25}}}`))
			return
		}
		w.WriteHeader(500)
	}))
	defer srv.Close()
	c := NewCache(srv.Client(), srv.URL)
	c.Now = func() time.Time { return mustDate("2026-07-09") }
	c.histTTL = 0 // force re-fetch

	_, _ = c.History(context.Background(), "MYR", "USD")
	got, err := c.History(context.Background(), "MYR", "USD")
	if err != nil {
		t.Fatalf("expected stale fallback, got err %v", err)
	}
	if !got.Stale || len(got.Points) != 1 {
		t.Fatalf("expected stale last-known, got %+v", got)
	}
}

func TestCacheContextAssessesTimeframes(t *testing.T) {
	// 10 business days, Value ascending.
	body := `{"rates":{
	  "2026-06-25":{"USD":0.2326},"2026-06-26":{"USD":0.2331},"2026-06-29":{"USD":0.2336},
	  "2026-06-30":{"USD":0.2342},"2026-07-01":{"USD":0.2347},"2026-07-02":{"USD":0.2353},
	  "2026-07-03":{"USD":0.2358},"2026-07-06":{"USD":0.2364},"2026-07-07":{"USD":0.2370},
	  "2026-07-08":{"USD":0.2375}}}`
	histSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest" {
			_, _ = w.Write([]byte(`{"rates":{"USD":0.2400}}`))
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	defer histSrv.Close()

	c := NewCache(histSrv.Client(), histSrv.URL)
	c.Now = func() time.Time { return mustDate("2026-07-09") }

	got, err := c.Context(context.Background(), "MYR", "USD", mustDate("2026-07-09"))
	if err != nil {
		t.Fatal(err)
	}
	if got.From != "MYR" || got.To != "USD" {
		t.Fatalf("From/To = %q/%q", got.From, got.To)
	}
	if got.Rate != 0.2400 {
		t.Fatalf("Rate = %v, want 0.2400", got.Rate)
	}
	if len(got.Timeframes) != 5 || got.Timeframes[0].Label != "7d" || got.Timeframes[4].Label != "YTD" {
		t.Fatalf("timeframes = %+v", got.Timeframes)
	}
	var d30 Assessment
	for _, a := range got.Timeframes {
		if a.Label == "30d" {
			d30 = a
		}
	}
	if d30.Samples != 10 || d30.Assessment != "good" || d30.Percentile != 100 {
		t.Fatalf("30d = %+v", d30)
	}
	if got.HistoryStale {
		t.Fatalf("expected fresh history, got HistoryStale=true")
	}
}

func TestCacheContextToleratesHistoryFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest" {
			_, _ = w.Write([]byte(`{"rates":{"USD":0.24}}`))
			return
		}
		w.WriteHeader(500)
	}))
	defer srv.Close()

	c := NewCache(srv.Client(), srv.URL)
	c.Now = func() time.Time { return mustDate("2026-07-09") }

	got, err := c.Context(context.Background(), "MYR", "USD", mustDate("2026-07-09"))
	if err != nil {
		t.Fatalf("expected Context to tolerate history failure, got err %v", err)
	}
	if got.Rate != 0.24 || got.Stale {
		t.Fatalf("live rate should still be fresh: %+v", got)
	}
	if !got.HistoryStale {
		t.Fatalf("expected HistoryStale=true when history fetch failed")
	}
	if len(got.Timeframes) != 5 || got.Timeframes[0].Assessment != "unknown" {
		t.Fatalf("expected empty/unknown timeframes, got %+v", got.Timeframes)
	}
}

// TestCacheConcurrentAccess exercises concurrent Rate/History/Context calls
// against the shared cache state, for use with `go test -race`.
func TestCacheConcurrentAccess(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n%5 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if r.URL.Path == "/latest" {
			to := r.URL.Query().Get("to")
			_, _ = w.Write([]byte(`{"rates":{"` + to + `":0.24}}`))
			return
		}
		to := r.URL.Query().Get("to")
		_, _ = w.Write([]byte(`{"rates":{"2026-07-08":{"` + to + `":0.24}}}`))
	}))
	defer srv.Close()

	c := NewCache(srv.Client(), srv.URL)
	c.liveTTL = time.Millisecond
	c.histTTL = time.Millisecond

	pairs := [][2]string{{"MYR", "USD"}, {"MYR", "EUR"}, {"MYR", "GBP"}}
	const iterations = 50
	var wg sync.WaitGroup

	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pair := pairs[i%len(pairs)]
			for j := 0; j < iterations; j++ {
				_, _ = c.Rate(context.Background(), pair[0], pair[1])
				_, _ = c.History(context.Background(), pair[0], pair[1])
				_, _ = c.Context(context.Background(), pair[0], pair[1], time.Now())
			}
		}(i)
	}

	wg.Wait()
}
