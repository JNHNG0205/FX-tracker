package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		status     int
		wantErr    bool
		wantUsdMyr float64
	}{
		{
			name:       "ok",
			body:       `{"amount":1.0,"base":"USD","date":"2026-07-08","rates":{"MYR":4.72}}`,
			status:     http.StatusOK,
			wantUsdMyr: 4.72,
		},
		{
			name:    "upstream 500",
			body:    `{}`,
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:    "missing MYR rate",
			body:    `{"amount":1.0,"base":"USD","date":"2026-07-08","rates":{}}`,
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

			got, err := Fetch(context.Background(), srv.Client(), srv.URL)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.UsdMyr != tt.wantUsdMyr {
				t.Fatalf("UsdMyr = %v, want %v", got.UsdMyr, tt.wantUsdMyr)
			}
			wantMyrUsd := 1.0 / tt.wantUsdMyr
			if got.MyrUsd != wantMyrUsd {
				t.Fatalf("MyrUsd = %v, want %v", got.MyrUsd, wantMyrUsd)
			}
		})
	}
}
