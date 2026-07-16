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
	"fx-tracker/internal/repository"
	"fx-tracker/internal/service"
)

// memRepo is an in-memory ConversionRepository for handler tests.
type memRepo struct{ items []model.Conversion }

func (m *memRepo) Create(ctx context.Context, c *model.Conversion) error {
	c.ID = uint(len(m.items) + 1)
	m.items = append([]model.Conversion{*c}, m.items...)
	return nil
}
func (m *memRepo) List(ctx context.Context) ([]model.Conversion, error) { return m.items, nil }

func (m *memRepo) BlendedRate(ctx context.Context, from, to string) (rate, totalHome, totalTarget float64, err error) {
	for _, c := range m.items {
		if c.FromCurrency != from || c.ToCurrency != to {
			continue
		}
		totalTarget += c.FromAmount * c.Rate
		totalHome += c.FromAmount
	}
	if totalHome == 0 {
		return 0, 0, 0, nil
	}
	return totalTarget / totalHome, totalHome, totalTarget, nil
}

func (m *memRepo) Pairs(ctx context.Context) ([]repository.Pair, error) {
	seen := map[repository.Pair]bool{}
	var out []repository.Pair
	for _, c := range m.items {
		p := repository.Pair{From: c.FromCurrency, To: c.ToCurrency}
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out, nil
}

func (m *memRepo) Update(ctx context.Context, c *model.Conversion) error {
	for i := range m.items {
		if m.items[i].ID == c.ID {
			m.items[i].FromCurrency = c.FromCurrency
			m.items[i].ToCurrency = c.ToCurrency
			m.items[i].FromAmount = c.FromAmount
			m.items[i].Rate = c.Rate
			m.items[i].Note = c.Note
			m.items[i].Date = c.Date
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (m *memRepo) Delete(ctx context.Context, id uint) error {
	for i := range m.items {
		if m.items[i].ID == id {
			m.items = append(m.items[:i], m.items[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

// fakeSettingsRepo is an in-memory SettingsRepository for handler tests.
type fakeSettingsRepo struct{ code string }

func (f *fakeSettingsRepo) Get(ctx context.Context) (model.Setting, error) {
	return model.Setting{ID: 1, HomeCurrency: f.code}, nil
}

func (f *fakeSettingsRepo) SetHomeCurrency(ctx context.Context, code string) error {
	f.code = code
	return nil
}

func newConvHandler(t *testing.T) *Handler {
	t.Helper()
	svc := service.NewConversionService(&memRepo{})
	return New(newLiveCache(t), svc, newTestSettings())
}

func TestCreateConversionValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/conversions", newConvHandler(t).CreateConversion)

	cases := []struct {
		name string
		body string
	}{
		{"amount zero", `{"from_currency":"MYR","to_currency":"USD","from_amount":0,"rate":0.24}`},
		{"same currency", `{"from_currency":"MYR","to_currency":"MYR","from_amount":100,"rate":0.24}`},
		{"unsupported currency", `{"from_currency":"MYR","to_currency":"XXX","from_amount":100,"rate":0.24}`},
		{"rate zero", `{"from_currency":"MYR","to_currency":"USD","from_amount":100,"rate":0}`},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/conversions", bytes.NewBufferString(tc.body)))
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

func TestCreateAndListConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newConvHandler(t)
	r := gin.New()
	r.POST("/api/conversions", h.CreateConversion)
	r.GET("/api/conversions", h.ListConversions)

	w := httptest.NewRecorder()
	body := `{"from_currency":"MYR","to_currency":"USD","from_amount":1000,"rate":0.24,"note":"first"}`
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/conversions", bytes.NewBufferString(body)))
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201, body = %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/conversions", nil))
	var list []model.Conversion
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].FromAmount != 1000 || list[0].FromCurrency != "MYR" || list[0].ToCurrency != "USD" {
		t.Fatalf("list = %+v", list)
	}
}

func TestConversionStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &memRepo{}
	repo.items = []model.Conversion{{ID: 1, FromCurrency: "MYR", ToCurrency: "USD", FromAmount: 1000, Rate: 0.2}}
	h := New(newLiveCache(t), service.NewConversionService(repo), newTestSettings())
	r := gin.New()
	r.GET("/api/conversions/status", h.ConversionStatus)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/conversions/status?from=MYR&to=USD", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var status service.DCAStatus
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !status.HasData || status.From != "MYR" || status.To != "USD" {
		t.Fatalf("status = %+v", status)
	}

	// invalid pair
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/conversions/status?from=MYR&to=MYR", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("same-pair status = %d, want 400", w.Code)
	}
}

func TestUpdateConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &memRepo{}
	repo.items = []model.Conversion{{ID: 1, FromCurrency: "MYR", ToCurrency: "USD", FromAmount: 100, Rate: 0.2, Note: "a"}}
	h := New(newLiveCache(t), service.NewConversionService(repo), newTestSettings())
	r := gin.New()
	r.PUT("/api/conversions/:id", h.UpdateConversion)

	// success
	w := httptest.NewRecorder()
	body := `{"from_currency":"MYR","to_currency":"USD","from_amount":150,"rate":0.25,"note":"fixed"}`
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/conversions/1", bytes.NewBufferString(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body = %s", w.Code, w.Body.String())
	}

	// validation failure
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/conversions/1", bytes.NewBufferString(`{"from_currency":"MYR","to_currency":"USD","from_amount":0,"rate":0.25}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad-value status = %d, want 400", w.Code)
	}

	// missing id
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/conversions/999", bytes.NewBufferString(`{"from_currency":"MYR","to_currency":"USD","from_amount":10,"rate":0.2}`)))
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d, want 404", w.Code)
	}

	// non-numeric id
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/conversions/abc", bytes.NewBufferString(`{"from_currency":"MYR","to_currency":"USD","from_amount":10,"rate":0.2}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad-id status = %d, want 400", w.Code)
	}
}

func TestDeleteConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &memRepo{}
	repo.items = []model.Conversion{{ID: 1, FromCurrency: "MYR", ToCurrency: "USD", FromAmount: 100, Rate: 0.2}}
	h := New(newLiveCache(t), service.NewConversionService(repo), newTestSettings())
	r := gin.New()
	r.DELETE("/api/conversions/:id", h.DeleteConversion)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/conversions/1", nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/conversions/1", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("second delete status = %d, want 404", w.Code)
	}
}
