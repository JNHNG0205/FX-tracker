package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"fx-tracker/internal/model"
	"fx-tracker/internal/price"
	"fx-tracker/internal/service"
)

// fakeHoldingRepo is an in-memory repository.HoldingRepository for handler
// tests.
type fakeHoldingRepo struct{ items []model.Holding }

func (f *fakeHoldingRepo) Create(ctx context.Context, h *model.Holding) error {
	h.ID = uint(len(f.items) + 1)
	f.items = append([]model.Holding{*h}, f.items...)
	return nil
}

func (f *fakeHoldingRepo) List(ctx context.Context) ([]model.Holding, error) {
	return f.items, nil
}

func (f *fakeHoldingRepo) Update(ctx context.Context, h *model.Holding) error {
	for i := range f.items {
		if f.items[i].ID == h.ID {
			f.items[i].Ticker = h.Ticker
			f.items[i].Shares = h.Shares
			f.items[i].AvgCost = h.AvgCost
			f.items[i].Currency = h.Currency
			f.items[i].ManualPrice = h.ManualPrice
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeHoldingRepo) Delete(ctx context.Context, id uint) error {
	for i := range f.items {
		if f.items[i].ID == id {
			f.items = append(f.items[:i], f.items[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

// fakePriceSource is a no-op price.PriceSource stand-in for handler tests;
// it never resolves a live quote, so holdings fall back to manual price or
// "unavailable".
type fakePriceSource struct{}

func (fakePriceSource) Quotes(ctx context.Context, reqs []price.Req) map[string]price.Quote {
	out := make(map[string]price.Quote, len(reqs))
	for _, r := range reqs {
		if r.Manual != nil {
			out[r.Ticker] = price.Quote{Ticker: r.Ticker, Price: *r.Manual, Found: true}
		}
	}
	return out
}

// newHoldingsAndPortfolio builds a fresh HoldingService + PortfolioService
// backed by in-memory fakes, for use as the extra New(...) args in handler
// tests.
func newHoldingsAndPortfolio(t *testing.T) (*service.HoldingService, *service.PortfolioService) {
	t.Helper()
	repo := &fakeHoldingRepo{}
	holdings := service.NewHoldingService(repo)
	portfolio := service.NewPortfolioService(repo, fakePriceSource{}, &memRepo{}, newLiveCache(t))
	return holdings, portfolio
}

// holdingsSvcForTest and portfolioSvcForTest give the pre-existing rate/
// conversion handler tests fresh, empty holdings/portfolio services so they
// can satisfy the widened New(...) signature without caring about holdings.
func holdingsSvcForTest(t *testing.T) *service.HoldingService {
	t.Helper()
	holdings, _ := newHoldingsAndPortfolio(t)
	return holdings
}

func portfolioSvcForTest(t *testing.T) *service.PortfolioService {
	t.Helper()
	_, portfolio := newHoldingsAndPortfolio(t)
	return portfolio
}

func newHoldingHandler(t *testing.T) *Handler {
	t.Helper()
	holdings, portfolio := newHoldingsAndPortfolio(t)
	return New(newLiveCache(t), service.NewConversionService(&memRepo{}), newTestSettings(), holdings, portfolio)
}

func TestCreateHoldingValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/holdings", newHoldingHandler(t).CreateHolding)

	cases := []struct {
		name string
		body string
	}{
		{"empty ticker", `{"ticker":"","shares":10,"avg_cost":500,"currency":"USD"}`},
		{"shares zero", `{"ticker":"VOO","shares":0,"avg_cost":500,"currency":"USD"}`},
		{"avg_cost zero", `{"ticker":"VOO","shares":10,"avg_cost":0,"currency":"USD"}`},
		{"unsupported currency", `{"ticker":"VOO","shares":10,"avg_cost":500,"currency":"XXX"}`},
		{"manual_price zero", `{"ticker":"VOO","shares":10,"avg_cost":500,"currency":"USD","manual_price":0}`},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/holdings", bytes.NewBufferString(tc.body)))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400", tc.name, w.Code)
		}
		var e map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &e)
		if e["error"] == "" {
			t.Fatalf("%s: expected {\"error\":...}, got %s", tc.name, w.Body.String())
		}
	}
}

func TestCreateAndListHolding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHoldingHandler(t)
	r := gin.New()
	r.POST("/api/holdings", h.CreateHolding)
	r.GET("/api/holdings", h.ListHoldings)

	w := httptest.NewRecorder()
	body := `{"ticker":"VOO","shares":10,"avg_cost":500,"currency":"USD"}`
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/holdings", bytes.NewBufferString(body)))
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201, body = %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/holdings", nil))
	var list []model.Holding
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].Ticker != "VOO" || list[0].Shares != 10 {
		t.Fatalf("list = %+v", list)
	}
}

func TestUpdateHolding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	holdings, portfolio := newHoldingsAndPortfolio(t)
	h := New(newLiveCache(t), service.NewConversionService(&memRepo{}), newTestSettings(), holdings, portfolio)
	r := gin.New()
	r.POST("/api/holdings", h.CreateHolding)
	r.PUT("/api/holdings/:id", h.UpdateHolding)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/holdings", bytes.NewBufferString(`{"ticker":"VOO","shares":10,"avg_cost":500,"currency":"USD"}`)))
	var created model.Holding
	_ = json.Unmarshal(w.Body.Bytes(), &created)

	// success
	w = httptest.NewRecorder()
	body := `{"ticker":"VOO","shares":12,"avg_cost":510,"currency":"USD"}`
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/holdings/1", bytes.NewBufferString(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", w.Code, w.Body.String())
	}

	// validation failure
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/holdings/1", bytes.NewBufferString(`{"ticker":"VOO","shares":0,"avg_cost":500,"currency":"USD"}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad-value status = %d, want 400", w.Code)
	}

	// missing id
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/holdings/999", bytes.NewBufferString(body)))
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d, want 404", w.Code)
	}

	// non-numeric id
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/holdings/abc", bytes.NewBufferString(body)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad-id status = %d, want 400", w.Code)
	}
}

func TestDeleteHolding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHoldingHandler(t)
	r := gin.New()
	r.POST("/api/holdings", h.CreateHolding)
	r.DELETE("/api/holdings/:id", h.DeleteHolding)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/holdings", bytes.NewBufferString(`{"ticker":"VOO","shares":10,"avg_cost":500,"currency":"USD"}`)))

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/holdings/1", nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/holdings/1", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("second delete status = %d, want 404", w.Code)
	}
}
