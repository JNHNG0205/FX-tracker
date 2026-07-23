package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"fx-tracker/internal/fx"
	"fx-tracker/internal/model"
)

type createHoldingReq struct {
	Ticker      string   `json:"ticker"`
	Shares      float64  `json:"shares"`
	AvgCost     float64  `json:"avg_cost"`
	Currency    string   `json:"currency"`
	ManualPrice *float64 `json:"manual_price,omitempty"`
}

// validate reports the first validation error for the request body, or ""
// if it's valid.
func (r createHoldingReq) validate() string {
	if r.Ticker == "" {
		return "ticker is required"
	}
	if r.Shares <= 0 {
		return "shares must be > 0"
	}
	if r.AvgCost <= 0 {
		return "avg_cost must be > 0"
	}
	if !fx.IsSupported(r.Currency) {
		return "unsupported currency"
	}
	if r.ManualPrice != nil && *r.ManualPrice <= 0 {
		return "manual_price must be > 0"
	}
	return ""
}

func (h *Handler) CreateHolding(c *gin.Context) {
	var req createHoldingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if msg := req.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	holding := &model.Holding{
		Ticker:      req.Ticker,
		Shares:      req.Shares,
		AvgCost:     req.AvgCost,
		Currency:    req.Currency,
		ManualPrice: req.ManualPrice,
	}
	if err := h.holdings.Create(c.Request.Context(), holding); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save holding"})
		return
	}
	c.JSON(http.StatusCreated, holding)
}

func (h *Handler) ListHoldings(c *gin.Context) {
	list, err := h.holdings.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list holdings"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) UpdateHolding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req createHoldingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if msg := req.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	holding := &model.Holding{
		ID:          id,
		Ticker:      req.Ticker,
		Shares:      req.Shares,
		AvgCost:     req.AvgCost,
		Currency:    req.Currency,
		ManualPrice: req.ManualPrice,
	}
	if err := h.holdings.Update(c.Request.Context(), holding); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "holding not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update holding"})
		return
	}
	c.JSON(http.StatusOK, holding)
}

func (h *Handler) DeleteHolding(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.holdings.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "holding not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete holding"})
		return
	}
	c.Status(http.StatusNoContent)
}
