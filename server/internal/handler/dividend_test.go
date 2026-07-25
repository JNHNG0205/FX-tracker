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
	"fx-tracker/internal/service"
)

// fakeDividendRepo is an in-memory repository.DividendRepository for handler
// tests.
type fakeDividendRepo struct{ items []model.Dividend }

func (f *fakeDividendRepo) Create(ctx context.Context, d *model.Dividend) error {
	d.ID = uint(len(f.items) + 1)
	f.items = append([]model.Dividend{*d}, f.items...)
	return nil
}

func (f *fakeDividendRepo) List(ctx context.Context) ([]model.Dividend, error) {
	return f.items, nil
}

func (f *fakeDividendRepo) Update(ctx context.Context, d *model.Dividend) error {
	for i := range f.items {
		if f.items[i].ID == d.ID {
			f.items[i].Ticker = d.Ticker
			f.items[i].Currency = d.Currency
			f.items[i].Amount = d.Amount
			f.items[i].Date = d.Date
			f.items[i].Note = d.Note
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeDividendRepo) Delete(ctx context.Context, id uint) error {
	for i := range f.items {
		if f.items[i].ID == id {
			f.items = append(f.items[:i], f.items[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

// dividendSvcForTest builds a fresh DividendService backed by an in-memory
// fake repo and the live-cache fake spot source, for use as the extra
// New(...) arg in handler tests.
func dividendSvcForTest(t *testing.T) *service.DividendService {
	t.Helper()
	return service.NewDividendService(&fakeDividendRepo{}, newLiveCache(t))
}

func newDividendHandler(t *testing.T) *Handler {
	t.Helper()
	holdings, portfolio := newHoldingsAndPortfolio(t)
	return New(newLiveCache(t), service.NewConversionService(&memRepo{}), newTestSettings(), holdings, portfolio, dividendSvcForTest(t))
}

func TestCreateDividendValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/dividends", newDividendHandler(t).CreateDividend)

	cases := []struct {
		name string
		body string
	}{
		{"empty ticker", `{"ticker":"","currency":"USD","amount":"100"}`},
		{"unsupported currency", `{"ticker":"VOO","currency":"XXX","amount":"100"}`},
		{"amount zero", `{"ticker":"VOO","currency":"USD","amount":"0"}`},
		{"amount negative", `{"ticker":"VOO","currency":"USD","amount":"-5"}`},
		{"amount not numeric", `{"ticker":"VOO","currency":"USD","amount":"abc"}`},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/dividends", bytes.NewBufferString(tc.body)))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400, body = %s", tc.name, w.Code, w.Body.String())
		}
		var e map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &e)
		if e["error"] == "" {
			t.Fatalf("%s: expected {\"error\":...}, got %s", tc.name, w.Body.String())
		}
	}
}

func TestCreateAndListDividend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDividendHandler(t)
	r := gin.New()
	r.POST("/api/dividends", h.CreateDividend)
	r.GET("/api/dividends", h.ListDividends)

	w := httptest.NewRecorder()
	body := `{"ticker":"VOO","currency":"USD","amount":"100"}`
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/dividends", bytes.NewBufferString(body)))
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201, body = %s", w.Code, w.Body.String())
	}
	var created model.Dividend
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created: %v", err)
	}
	if created.Ticker != "VOO" || !created.Amount.Equal(created.Amount) {
		t.Fatalf("created = %+v", created)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/dividends?home=MYR", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200, body = %s", w.Code, w.Body.String())
	}
	var summary service.DividendSummary
	if err := json.Unmarshal(w.Body.Bytes(), &summary); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if summary.HomeCurrency != "MYR" {
		t.Fatalf("HomeCurrency = %s, want MYR", summary.HomeCurrency)
	}
	if len(summary.Dividends) != 1 || summary.Dividends[0].Ticker != "VOO" {
		t.Fatalf("Dividends = %+v", summary.Dividends)
	}
}

func TestUpdateAndDeleteDividendNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newDividendHandler(t)
	r := gin.New()
	r.PUT("/api/dividends/:id", h.UpdateDividend)
	r.DELETE("/api/dividends/:id", h.DeleteDividend)

	w := httptest.NewRecorder()
	body := `{"ticker":"VOO","currency":"USD","amount":"100"}`
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/dividends/999", bytes.NewBufferString(body)))
	if w.Code != http.StatusNotFound {
		t.Fatalf("update missing status = %d, want 404, body = %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/dividends/999", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("delete missing status = %d, want 404, body = %s", w.Code, w.Body.String())
	}
}
