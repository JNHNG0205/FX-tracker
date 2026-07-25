package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"fx-tracker/internal/fx"
	"fx-tracker/internal/model"
)

type createDividendReq struct {
	Ticker   string `json:"ticker"`
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
	Date     string `json:"date,omitempty"`
	Note     string `json:"note"`
}

// validate parses and checks the request body, returning the parsed amount
// and either the resolved date (or now, if omitted) or a validation
// message. Mirrors createHoldingReq.validate.
func (r createDividendReq) validate() (amount decimal.Decimal, date time.Time, msg string) {
	if r.Ticker == "" {
		return decimal.Decimal{}, time.Time{}, "ticker is required"
	}
	if !fx.IsSupported(r.Currency) {
		return decimal.Decimal{}, time.Time{}, "unsupported currency"
	}
	amount, err := decimal.NewFromString(r.Amount)
	if err != nil {
		return decimal.Decimal{}, time.Time{}, "amount must be a valid number"
	}
	if !amount.GreaterThan(decimal.Zero) {
		return decimal.Decimal{}, time.Time{}, "amount must be > 0"
	}
	date = time.Now()
	if r.Date != "" {
		parsed, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			return decimal.Decimal{}, time.Time{}, "date must be YYYY-MM-DD"
		}
		date = parsed
	}
	return amount, date, ""
}

func (h *Handler) CreateDividend(c *gin.Context) {
	var req createDividendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	amount, date, msg := req.validate()
	if msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	dividend := &model.Dividend{
		Ticker:   req.Ticker,
		Currency: req.Currency,
		Amount:   amount,
		Date:     date,
		Note:     req.Note,
	}
	if err := h.dividends.Create(c.Request.Context(), dividend); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save dividend"})
		return
	}
	c.JSON(http.StatusCreated, dividend)
}

// ListDividends resolves the home currency from ?home= (validated) or falls
// back to the user's saved settings, then returns the dividend + withholding
// summary. Mirrors Portfolio's home-resolution pattern.
func (h *Handler) ListDividends(c *gin.Context) {
	home := c.Query("home")
	if home != "" {
		if !fx.IsSupported(home) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported currency"})
			return
		}
	} else {
		s, err := h.settings.Get(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load settings"})
			return
		}
		home = s.HomeCurrency
	}

	summary, err := h.dividends.Summary(c.Request.Context(), home)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not compute dividend summary"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) UpdateDividend(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req createDividendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	amount, date, msg := req.validate()
	if msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	dividend := &model.Dividend{
		ID:       id,
		Ticker:   req.Ticker,
		Currency: req.Currency,
		Amount:   amount,
		Date:     date,
		Note:     req.Note,
	}
	if err := h.dividends.Update(c.Request.Context(), dividend); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "dividend not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update dividend"})
		return
	}
	c.JSON(http.StatusOK, dividend)
}

func (h *Handler) DeleteDividend(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.dividends.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "dividend not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete dividend"})
		return
	}
	c.Status(http.StatusNoContent)
}
