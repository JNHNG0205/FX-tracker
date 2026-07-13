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

// memRepo is an in-memory ConversionRepository for handler tests.
type memRepo struct{ items []model.Conversion }

func (m *memRepo) Create(ctx context.Context, c *model.Conversion) error {
	c.ID = uint(len(m.items) + 1)
	m.items = append([]model.Conversion{*c}, m.items...)
	return nil
}
func (m *memRepo) List(ctx context.Context) ([]model.Conversion, error) { return m.items, nil }
func (m *memRepo) BlendedRate(ctx context.Context) (float64, float64, float64, error) {
	var tu, tm float64
	for _, c := range m.items {
		tu += c.MyrAmount * c.RateMyrUsd
		tm += c.MyrAmount
	}
	if tm == 0 {
		return 0, 0, 0, nil
	}
	return tu / tm, tm, tu, nil
}

func (m *memRepo) Update(ctx context.Context, c *model.Conversion) error {
	for i := range m.items {
		if m.items[i].ID == c.ID {
			m.items[i].MyrAmount = c.MyrAmount
			m.items[i].RateMyrUsd = c.RateMyrUsd
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

func newConvHandler(t *testing.T) *Handler {
	t.Helper()
	svc := service.NewConversionService(&memRepo{})
	return New(newLiveCache(t), newHistory(t), svc)
}

func TestCreateConversionValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/conversions", newConvHandler(t).CreateConversion)

	w := httptest.NewRecorder()
	body := `{"myr_amount":0,"rate_myr_usd":0.24}`
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/conversions", bytes.NewBufferString(body)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	var e map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &e)
	if e["error"] == "" {
		t.Fatalf("expected {\"error\":...}, got %s", w.Body.String())
	}
}

func TestCreateAndListConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newConvHandler(t)
	r := gin.New()
	r.POST("/api/conversions", h.CreateConversion)
	r.GET("/api/conversions", h.ListConversions)

	w := httptest.NewRecorder()
	body := `{"myr_amount":1000,"rate_myr_usd":0.24,"note":"first"}`
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/conversions", bytes.NewBufferString(body)))
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", w.Code)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/conversions", nil))
	var list []model.Conversion
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].MyrAmount != 1000 {
		t.Fatalf("list = %+v", list)
	}
}

func TestUpdateConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &memRepo{}
	repo.items = []model.Conversion{{ID: 1, MyrAmount: 100, RateMyrUsd: 0.2, Note: "a"}}
	h := New(newLiveCache(t), newHistory(t), service.NewConversionService(repo))
	r := gin.New()
	r.PUT("/api/conversions/:id", h.UpdateConversion)

	// success
	w := httptest.NewRecorder()
	body := `{"myr_amount":150,"rate_myr_usd":0.25,"note":"fixed"}`
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/conversions/1", bytes.NewBufferString(body)))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	// validation failure
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/conversions/1", bytes.NewBufferString(`{"myr_amount":0,"rate_myr_usd":0.25}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad-value status = %d, want 400", w.Code)
	}

	// missing id
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/conversions/999", bytes.NewBufferString(`{"myr_amount":10,"rate_myr_usd":0.2}`)))
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d, want 404", w.Code)
	}

	// non-numeric id
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/conversions/abc", bytes.NewBufferString(`{"myr_amount":10,"rate_myr_usd":0.2}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad-id status = %d, want 400", w.Code)
	}
}

func TestDeleteConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &memRepo{}
	repo.items = []model.Conversion{{ID: 1, MyrAmount: 100, RateMyrUsd: 0.2}}
	h := New(newLiveCache(t), newHistory(t), service.NewConversionService(repo))
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
