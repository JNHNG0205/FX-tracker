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
		{"ok", `{"c":560.12,"h":562,"l":499,"o":500,"pc":555,"t":1753500000}`, true, 560.12, false},
		{"no data", `{"c":0,"h":0,"l":0,"o":0,"pc":0,"t":0}`, false, 0, false},
		{"malformed", "garbage", false, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.name == "ok" && !strings.Contains(r.URL.RawQuery, "symbol=VOO") {
					t.Errorf("symbol not uppercased: %s", r.URL.RawQuery)
				}
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			q, err := Fetch(context.Background(), srv.Client(), srv.URL, "testtoken", "voo")
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

	_, err := Fetch(context.Background(), srv.Client(), srv.URL, "testtoken", "VOO")
	if err == nil {
		t.Fatal("want err for non-200 response")
	}
}
