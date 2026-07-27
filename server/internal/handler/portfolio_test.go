package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/model"
	"fx-tracker/internal/portfolio"
	"fx-tracker/internal/service"
)

func TestPortfolioHomeFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeHoldingRepo{items: []model.Holding{
		{ID: 1, Ticker: "VOO", Shares: 10, AvgCost: 500, Currency: "USD"},
	}}
	holdings := service.NewHoldingService(repo)
	portfolioSvc := service.NewPortfolioService(repo, fakePriceSource{}, &memRepo{}, newLiveCache(t))
	h := New(newLiveCache(t), service.NewConversionService(&memRepo{}), newTestSettings(), holdings, portfolioSvc, dividendSvcForTest(t))

	r := gin.New()
	r.GET("/api/portfolio", h.Portfolio)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/portfolio?home=MYR", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp portfolio.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.HomeCurrency != "MYR" {
		t.Fatalf("home_currency = %s, want MYR", resp.HomeCurrency)
	}
	if len(resp.Holdings) != 1 {
		t.Fatalf("holdings = %d, want 1", len(resp.Holdings))
	}

	// unsupported home currency
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/portfolio?home=XXX", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad home status = %d, want 400", w.Code)
	}
}

func TestPortfolioHomeFromSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &fakeHoldingRepo{}
	holdings := service.NewHoldingService(repo)
	portfolioSvc := service.NewPortfolioService(repo, fakePriceSource{}, &memRepo{}, newLiveCache(t))
	settings := &fakeSettingsRepo{code: "SGD"}
	h := New(newLiveCache(t), service.NewConversionService(&memRepo{}), settings, holdings, portfolioSvc, dividendSvcForTest(t))

	r := gin.New()
	r.GET("/api/portfolio", h.Portfolio)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/portfolio", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var resp portfolio.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.HomeCurrency != "SGD" {
		t.Fatalf("home_currency = %s, want SGD (from settings default)", resp.HomeCurrency)
	}
}
