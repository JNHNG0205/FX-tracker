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

func TestCacheRefreshAndStale(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			_, _ = w.Write([]byte(`{"rates":{"MYR":4.72}}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewCache(srv.Client(), srv.URL)

	// Before any successful refresh, reads must report stale rather than a
	// bogus fresh-looking zero rate.
	if got := c.Get(); !got.Stale {
		t.Fatalf("before first refresh: expected Stale=true, got %+v", got)
	}

	// First refresh succeeds.
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	got := c.Get()
	if got.UsdMyr != 4.72 || got.Stale {
		t.Fatalf("after success: got %+v", got)
	}

	// Second refresh fails: keep last-known value, flag stale.
	if err := c.Refresh(context.Background()); err == nil {
		t.Fatalf("second refresh: expected error")
	}
	got = c.Get()
	if got.UsdMyr != 4.72 {
		t.Fatalf("stale value changed: %+v", got)
	}
	if !got.Stale {
		t.Fatalf("expected stale=true, got %+v", got)
	}
}

// TestCacheConcurrentAccess exercises concurrent Refresh/Get calls
// against the shared cache state, for use with `go test -race`.
func TestCacheConcurrentAccess(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n%3 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"rates":{"MYR":4.72}}`))
	}))
	defer srv.Close()

	c := NewCache(srv.Client(), srv.URL)

	const iterations = 50
	var wg sync.WaitGroup

	// A few writers hammering Refresh concurrently.
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = c.Refresh(context.Background())
			}
		}()
	}

	// Many readers hammering Get concurrently.
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = c.Get()
			}
		}()
	}

	wg.Wait()
}

// TestCacheRunStopsOnCancel verifies Run returns promptly once its context
// is cancelled, rather than blocking forever on the ticker loop.
func TestCacheRunStopsOnCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"rates":{"MYR":4.72}}`))
	}))
	defer srv.Close()

	c := NewCache(srv.Client(), srv.URL)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		c.Run(ctx, time.Hour) // long interval: only the cancel should stop it
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}
