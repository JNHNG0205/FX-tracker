package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		status   int
		wantErr  bool
		wantRate float64
	}{
		{
			name:     "ok",
			body:     `{"amount":1.0,"base":"MYR","date":"2026-07-08","rates":{"USD":0.21}}`,
			status:   http.StatusOK,
			wantRate: 0.21,
		},
		{
			name:    "upstream 500",
			body:    `{}`,
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:    "missing rate",
			body:    `{"amount":1.0,"base":"MYR","date":"2026-07-08","rates":{}}`,
			status:  http.StatusOK,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			got, err := Fetch(context.Background(), srv.Client(), srv.URL, "MYR", "USD")
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Rate != tt.wantRate {
				t.Fatalf("Rate = %v, want %v", got.Rate, tt.wantRate)
			}
			wantInverse := 1.0 / tt.wantRate
			if got.Inverse != wantInverse {
				t.Fatalf("Inverse = %v, want %v", got.Inverse, wantInverse)
			}
			if got.From != "MYR" || got.To != "USD" {
				t.Fatalf("From/To = %q/%q, want MYR/USD", got.From, got.To)
			}
		})
	}
}
