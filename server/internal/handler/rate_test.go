package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/fx"
	"fx-tracker/internal/service"
)

// newLiveCache returns a fx.Cache pointed at an httptest upstream that
// answers both /latest and /<start>..<end> (history) requests for any pair.
func newLiveCache(t *testing.T) *fx.Cache {
	t.Helper()
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		to := r.URL.Query().Get("to")
		if to == "" {
			to = "USD"
		}
		if r.URL.Path == "/latest" {
			_, _ = w.Write([]byte(`{"rates":{"` + to + `":0.21}}`))
			return
		}
		_, _ = w.Write([]byte(`{"rates":{"2026-07-07":{"` + to + `":0.20},"2026-07-08":{"` + to + `":0.21}}}`))
	}))
	t.Cleanup(up.Close)
	return fx.NewCache(up.Client(), up.URL)
}

func newTestSettings() *fakeSettingsRepo {
	return &fakeSettingsRepo{code: "MYR"}
}

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(newLiveCache(t), service.NewConversionService(&memRepo{}), newTestSettings())
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
	h := New(newLiveCache(t), service.NewConversionService(&memRepo{}), newTestSettings())
	r.GET("/api/rate", h.Rate)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rate?from=MYR&to=USD", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var rate fx.Rate
	if err := json.Unmarshal(w.Body.Bytes(), &rate); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rate.Rate != 0.21 {
		t.Fatalf("rate = %v", rate.Rate)
	}

	// from == to
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rate?from=MYR&to=MYR", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("same-pair status = %d, want 400", w.Code)
	}

	// unsupported currency
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rate?from=MYR&to=XXX", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("unsupported status = %d, want 400", w.Code)
	}
}

func TestRateContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(newLiveCache(t), service.NewConversionService(&memRepo{}), newTestSettings())
	r.GET("/api/rate/context", h.RateContext)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rate/context?from=MYR&to=USD", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var ctx fx.RateContext
	if err := json.Unmarshal(w.Body.Bytes(), &ctx); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(ctx.Timeframes) != 5 {
		t.Fatalf("timeframes = %d, want 5", len(ctx.Timeframes))
	}
	if ctx.From != "MYR" || ctx.To != "USD" {
		t.Fatalf("from/to = %s/%s", ctx.From, ctx.To)
	}
	if ctx.Rate != 0.21 {
		t.Fatalf("rate = %v", ctx.Rate)
	}
}

func TestRateHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := New(newLiveCache(t), service.NewConversionService(&memRepo{}), newTestSettings())
	r.GET("/api/rate/history", h.RateHistory)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/rate/history?from=MYR&to=USD", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body fx.RateHistory
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.Points) != 2 {
		t.Fatalf("points = %d, want 2", len(body.Points))
	}
	if body.Points[0].Date == "" || body.Points[0].Value == 0 {
		t.Fatalf("point not populated: %+v", body.Points[0])
	}
}
