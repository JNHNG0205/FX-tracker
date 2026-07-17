package price

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantFound bool
		wantPrice float64
		wantErr   bool
	}{
		{"ok", "Symbol,Date,Time,Open,High,Low,Close,Volume\nVOO.US,2026-07-16,22:00:00,500,562,499,560.12,1000\n", true, 560.12, false},
		{"no data", "Symbol,Date,Time,Open,High,Low,Close,Volume\nZZZ.US,N/D,N/D,N/D,N/D,N/D,N/D,N/D\n", false, 0, false},
		{"malformed", "garbage", false, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.RawQuery, "s=voo.us") && tt.name == "ok" {
					t.Errorf("symbol not lowercased+.us: %s", r.URL.RawQuery)
				}
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			q, err := Fetch(context.Background(), srv.Client(), srv.URL, "VOO", "USD")
			if tt.wantErr {
				if err == nil {
					t.Fatal("want err")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if q.Found != tt.wantFound || (tt.wantFound && q.Price != tt.wantPrice) {
				t.Fatalf("q=%+v", q)
			}
		})
	}
}

func TestFetchNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := Fetch(context.Background(), srv.Client(), srv.URL, "VOO", "USD")
	if err == nil {
		t.Fatal("want err for non-200 response")
	}
}
