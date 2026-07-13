package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// seriesServer returns a frankfurter-shaped series for any request.
func seriesServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(s.Close)
	return s
}

func TestHistoryCacheAssess(t *testing.T) {
	// 10 business days, MyrUsd ascending because MYR descending.
	body := `{"rates":{
	  "2026-06-25":{"MYR":4.30},"2026-06-26":{"MYR":4.29},"2026-06-29":{"MYR":4.28},
	  "2026-06-30":{"MYR":4.27},"2026-07-01":{"MYR":4.26},"2026-07-02":{"MYR":4.25},
	  "2026-07-03":{"MYR":4.24},"2026-07-06":{"MYR":4.23},"2026-07-07":{"MYR":4.22},
	  "2026-07-08":{"MYR":4.21}}}`
	srv := seriesServer(t, body)

	c := NewHistoryCache(srv.Client(), srv.URL)
	c.Now = func() time.Time { return mustDate("2026-07-09") }
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	// current MyrUsd = 1/4.20 ≈ 0.2381, above every point (all MYR>=4.21) → good, 100th pct.
	got := c.Assess(1.0/4.20, mustDate("2026-07-09"))
	if len(got) != 5 {
		t.Fatalf("want 5 timeframes, got %d", len(got))
	}
	if got[0].Label != "7d" || got[4].Label != "YTD" {
		t.Fatalf("labels = %q..%q", got[0].Label, got[4].Label)
	}
	// 30d window includes all 10 points; current beats all → good, 100.
	var d30 Assessment
	for _, a := range got {
		if a.Label == "30d" {
			d30 = a
		}
	}
	if d30.Samples != 10 || d30.Assessment != "good" || d30.Percentile != 100 {
		t.Fatalf("30d = %+v", d30)
	}
	// 7d window (>= 2026-07-02) has 5 points (07-02,03,06,07,08).
	var d7 Assessment
	for _, a := range got {
		if a.Label == "7d" {
			d7 = a
		}
	}
	if d7.Samples != 5 {
		t.Fatalf("7d samples = %d, want 5", d7.Samples)
	}
}

func TestHistoryCacheStaleAndEmpty(t *testing.T) {
	// Before any refresh, assessments are empty/unknown and Stale() is true.
	c := NewHistoryCache(http.DefaultClient, "http://invalid.invalid")
	c.Now = func() time.Time { return mustDate("2026-07-09") }
	if !c.Stale() {
		t.Fatalf("expected stale before first refresh")
	}
	got := c.Assess(0.24, mustDate("2026-07-09"))
	if len(got) != 5 || got[0].Assessment != "unknown" || got[0].Samples != 0 {
		t.Fatalf("expected unknown/empty timeframes, got %+v", got[0])
	}
}

func TestHistoryCachePointsReturnsCopy(t *testing.T) {
	body := `{"rates":{"2026-07-07":{"MYR":4.80},"2026-07-08":{"MYR":4.90}}}`
	srv := seriesServer(t, body)
	c := NewHistoryCache(srv.Client(), srv.URL)
	c.Now = func() time.Time { return mustDate("2026-07-09") }
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	pts := c.Points()
	if len(pts) != 2 {
		t.Fatalf("len = %d, want 2", len(pts))
	}
	// ascending by date
	if !pts[0].Date.Before(pts[1].Date) {
		t.Fatalf("points not ascending: %+v", pts)
	}
	// mutating the returned slice must not affect the cache
	pts[0].MyrUsd = -999
	again := c.Points()
	if again[0].MyrUsd == -999 {
		t.Fatalf("Points() leaked the internal slice; mutation bled through")
	}
}
