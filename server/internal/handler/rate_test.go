package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/fx"
)

func newTestCache(t *testing.T) *fx.Cache {
	t.Helper()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"rates":{"MYR":4.72}}`))
	}))
	t.Cleanup(up.Close)
	c := fx.NewCache(up.Client(), up.URL)
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatalf("seed cache: %v", err)
	}
	return c
}

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(newTestCache(t))
	r.GET("/health", h.Health)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status = %q, want ok", body["status"])
	}
}

func TestRate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(newTestCache(t))
	r.GET("/api/rate", h.Rate)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rate", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var got fx.Rate
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.UsdMyr != 4.72 {
		t.Fatalf("UsdMyr = %v, want 4.72", got.UsdMyr)
	}
}
