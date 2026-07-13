package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/fx"
	"fx-tracker/internal/service"
)

func newLiveCache(t *testing.T) *fx.Cache {
	t.Helper()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"rates":{"MYR":4.72}}`))
	}))
	t.Cleanup(up.Close)
	c := fx.NewCache(up.Client(), up.URL)
	if err := c.Refresh(context.Background()); err != nil {
		t.Fatalf("seed live cache: %v", err)
	}
	return c
}

func newHistory(t *testing.T) *fx.HistoryCache {
	t.Helper()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"rates":{"2026-07-07":{"MYR":4.80},"2026-07-08":{"MYR":4.90}}}`))
	}))
	t.Cleanup(up.Close)
	h := fx.NewHistoryCache(up.Client(), up.URL)
	if err := h.Refresh(context.Background()); err != nil {
		t.Fatalf("seed history: %v", err)
	}
	return h
}

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(newLiveCache(t), newHistory(t), service.NewConversionService(&memRepo{}))
	r.GET("/health", h.Health)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestRate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(newLiveCache(t), newHistory(t), service.NewConversionService(&memRepo{}))
	r.GET("/api/rate", h.Rate)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rate", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var rate fx.Rate
	if err := json.Unmarshal(w.Body.Bytes(), &rate); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rate.UsdMyr != 4.72 {
		t.Fatalf("usd_myr = %v", rate.UsdMyr)
	}
}

func TestRateContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(newLiveCache(t), newHistory(t), service.NewConversionService(&memRepo{}))
	r.GET("/api/rate/context", h.RateContext)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rate/context", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var ctx fx.RateContext
	if err := json.Unmarshal(w.Body.Bytes(), &ctx); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(ctx.Timeframes) != 5 {
		t.Fatalf("timeframes = %d, want 5", len(ctx.Timeframes))
	}
	if ctx.CurrentUsdMyr != 4.72 {
		t.Fatalf("current_usd_myr = %v", ctx.CurrentUsdMyr)
	}
	if ctx.Timeframes[0].Label != "7d" {
		t.Fatalf("first timeframe = %q", ctx.Timeframes[0].Label)
	}
}

func TestRateHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(newLiveCache(t), newHistory(t), service.NewConversionService(&memRepo{}))
	r.GET("/api/rate/history", h.RateHistory)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rate/history", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body fx.RateHistory
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// newHistory seeds two business days.
	if len(body.Points) != 2 {
		t.Fatalf("points = %d, want 2", len(body.Points))
	}
	if body.Points[0].Date == "" || body.Points[0].MyrUsd == 0 {
		t.Fatalf("point not populated: %+v", body.Points[0])
	}
}
