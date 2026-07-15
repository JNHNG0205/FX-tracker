package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func mustDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestFetchHistory(t *testing.T) {
	body := `{"amount":1.0,"base":"MYR","start_date":"2026-06-09","end_date":"2026-06-12",
	  "rates":{"2026-06-09":{"USD":0.25},"2026-06-10":{"USD":0.20},"2026-06-12":{"USD":0.50}}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	pts, err := FetchHistory(context.Background(), srv.Client(), srv.URL, "MYR", "USD", mustDate("2026-06-09"), mustDate("2026-06-12"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pts) != 3 {
		t.Fatalf("len = %d, want 3 (business-day gap on 06-11 preserved as absent)", len(pts))
	}
	// Ascending by date, Value = rates[date][to].
	if !pts[0].Date.Equal(mustDate("2026-06-09")) || pts[0].Value != 0.25 {
		t.Fatalf("pts[0] = %+v", pts[0])
	}
	if !pts[2].Date.Equal(mustDate("2026-06-12")) || pts[2].Value != 0.50 {
		t.Fatalf("pts[2] = %+v", pts[2])
	}
}

func TestFetchHistoryFollowsRedirect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/final", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"rates":{"2026-06-09":{"USD":0.25}}}`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusMovedPermanently)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	pts, err := FetchHistory(context.Background(), srv.Client(), srv.URL, "MYR", "USD", mustDate("2026-06-09"), mustDate("2026-06-09"))
	if err != nil {
		t.Fatalf("expected redirect to be followed, got error: %v", err)
	}
	if len(pts) != 1 || pts[0].Value != 0.25 {
		t.Fatalf("pts = %+v", pts)
	}
}

func TestFetchHistoryErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	if _, err := FetchHistory(context.Background(), srv.Client(), srv.URL, "MYR", "USD", mustDate("2026-06-09"), mustDate("2026-06-10")); err == nil {
		t.Fatalf("expected error on 500")
	}
}
