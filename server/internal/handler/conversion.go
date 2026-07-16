package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"fx-tracker/internal/fx"
	"fx-tracker/internal/model"
)

type createConversionReq struct {
	Date         *time.Time `json:"date"`
	FromCurrency string     `json:"from_currency"`
	ToCurrency   string     `json:"to_currency"`
	FromAmount   float64    `json:"from_amount"`
	Rate         float64    `json:"rate"`
	Note         string     `json:"note"`
}

// validate reports the first validation error for the request body, or ""
// if it's valid.
func (r createConversionReq) validate() string {
	if !fx.IsSupported(r.FromCurrency) || !fx.IsSupported(r.ToCurrency) {
		return "unsupported currency"
	}
	if r.FromCurrency == r.ToCurrency {
		return "from_currency and to_currency must differ"
	}
	if r.FromAmount <= 0 {
		return "from_amount must be > 0"
	}
	if r.Rate <= 0 {
		return "rate must be > 0"
	}
	return ""
}

func (h *Handler) CreateConversion(c *gin.Context) {
	var req createConversionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if msg := req.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	date := time.Now()
	if req.Date != nil {
		date = *req.Date
	}
	conv := &model.Conversion{
		Date:         date,
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		FromAmount:   req.FromAmount,
		Rate:         req.Rate,
		Note:         req.Note,
	}
	if err := h.conv.Create(c.Request.Context(), conv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save conversion"})
		return
	}
	c.JSON(http.StatusCreated, conv)
}

func (h *Handler) ListConversions(c *gin.Context) {
	list, err := h.conv.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list conversions"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) ConversionStatus(c *gin.Context) {
	from, to, ok := pairParams(c)
	if !ok {
		return
	}
	rate, err := h.cache.Rate(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "rate source unavailable"})
		return
	}
	status, err := h.conv.Status(c.Request.Context(), from, to, rate.Rate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not compute status"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func parseID(c *gin.Context) (uint, bool) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id64), true
}

func (h *Handler) UpdateConversion(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req createConversionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if msg := req.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	date := time.Now()
	if req.Date != nil {
		date = *req.Date
	}
	conv := &model.Conversion{
		ID:           id,
		Date:         date,
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		FromAmount:   req.FromAmount,
		Rate:         req.Rate,
		Note:         req.Note,
	}
	if err := h.conv.Update(c.Request.Context(), conv); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "conversion not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update conversion"})
		return
	}
	c.JSON(http.StatusOK, conv)
}

func (h *Handler) DeleteConversion(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.conv.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "conversion not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete conversion"})
		return
	}
	c.Status(http.StatusNoContent)
}
