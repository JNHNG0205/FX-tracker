package price

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestQuotesPartialFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "symbol=BAD") {
			w.WriteHeader(500)
			return
		}
		_, _ = w.Write([]byte(`{"c":42}`))
	}))
	defer srv.Close()
	c := NewCache(srv.Client(), "testtoken")
	c.base = srv.URL // test hook for the Finnhub base
	man := 99.0
	got := c.Quotes(context.Background(), []Req{
		{Ticker: "VOO", Currency: "USD"},
		{Ticker: "BAD", Currency: "USD"},
		{Ticker: "MANUAL", Currency: "EUR", Manual: &man},
	})
	if got["VOO"].Price != 42 || !got["VOO"].Found {
		t.Fatalf("VOO=%+v", got["VOO"])
	}
	if got["BAD"].Found {
		t.Fatalf("BAD should be not-found (partial failure), got %+v", got["BAD"])
	}
	if got["MANUAL"].Price != 99 || !got["MANUAL"].Found {
		t.Fatalf("MANUAL=%+v", got["MANUAL"])
	}
}

func TestCacheQuotePerTickerAndTTL(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		_, _ = w.Write([]byte(`{"c":42}`))
	}))
	defer srv.Close()
	c := NewCache(srv.Client(), "testtoken")
	c.base = srv.URL

	q1, err := c.Quote(context.Background(), "VOO", "USD")
	if err != nil {
		t.Fatal(err)
	}
	if q1.Price != 42 || !q1.Found {
		t.Fatalf("%+v", q1)
	}
	// second call for same ticker within TTL: no new upstream call
	_, _ = c.Quote(context.Background(), "VOO", "USD")
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 upstream call, got %d", calls)
	}
	// different ticker: separate fetch
	_, _ = c.Quote(context.Background(), "OTHER", "USD")
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2, got %d", calls)
	}
}

func TestCacheQuoteStaleOnError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			_, _ = w.Write([]byte(`{"c":42}`))
			return
		}
		w.WriteHeader(500)
	}))
	defer srv.Close()
	c := NewCache(srv.Client(), "testtoken")
	c.base = srv.URL
	c.ttl = 0 // force re-fetch

	_, _ = c.Quote(context.Background(), "VOO", "USD")
	got, err := c.Quote(context.Background(), "VOO", "USD")
	if err != nil {
		t.Fatalf("expected stale fallback, got err %v", err)
	}
	if got.Price != 42 || !got.Found {
		t.Fatalf("expected stale last-known, got %+v", got)
	}
}

func TestCacheQuotesConcurrentAccess(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n%5 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"c":42}`))
	}))
	defer srv.Close()

	c := NewCache(srv.Client(), "testtoken")
	c.base = srv.URL
	c.ttl = time.Millisecond

	tickers := []string{"VOO", "AAPL", "MSFT", "GOOG"}
	const iterations = 20
	for i := 0; i < iterations; i++ {
		reqs := make([]Req, len(tickers))
		for j, t := range tickers {
			reqs[j] = Req{Ticker: t, Currency: "USD"}
		}
		_ = c.Quotes(context.Background(), reqs)
	}
}

func TestCacheQuoteNoTokenSkipsNetwork(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("no network call expected with an empty token")
	}))
	defer srv.Close()

	c := NewCache(srv.Client(), "")
	c.base = srv.URL

	got, err := c.Quote(context.Background(), "VOO", "USD")
	if err != nil {
		t.Fatal(err)
	}
	if got.Found {
		t.Fatalf("expected Found:false with no token, got %+v", got)
	}
}
