package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/fx"
	"fx-tracker/internal/repository"
	"fx-tracker/internal/service"
)

type Handler struct {
	cache     *fx.Cache
	conv      *service.ConversionService
	settings  repository.SettingsRepository
	holdings  *service.HoldingService
	portfolio *service.PortfolioService
	dividends *service.DividendService
}

func New(cache *fx.Cache, conv *service.ConversionService, settings repository.SettingsRepository, holdings *service.HoldingService, portfolio *service.PortfolioService, dividends *service.DividendService) *Handler {
	return &Handler{cache: cache, conv: conv, settings: settings, holdings: holdings, portfolio: portfolio, dividends: dividends}
}

// pairParams reads/validates the from/to query params shared by rate and
// conversion-status endpoints. On failure it writes the 400 response itself
// and returns ok=false so the caller can just `return`.
func pairParams(c *gin.Context) (from, to string, ok bool) {
	from = c.Query("from")
	to = c.Query("to")
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to are required"})
		return "", "", false
	}
	if !fx.IsSupported(from) || !fx.IsSupported(to) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported currency"})
		return "", "", false
	}
	if from == to {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to must differ"})
		return "", "", false
	}
	return from, to, true
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Rate(c *gin.Context) {
	from, to, ok := pairParams(c)
	if !ok {
		return
	}
	rate, err := h.cache.Rate(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "rate source unavailable"})
		return
	}
	c.JSON(http.StatusOK, rate)
}

func (h *Handler) RateContext(c *gin.Context) {
	from, to, ok := pairParams(c)
	if !ok {
		return
	}
	ctx, err := h.cache.Context(c.Request.Context(), from, to, time.Now())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "rate source unavailable"})
		return
	}
	c.JSON(http.StatusOK, ctx)
}

func (h *Handler) RateHistory(c *gin.Context) {
	from, to, ok := pairParams(c)
	if !ok {
		return
	}
	hist, err := h.cache.History(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "rate source unavailable"})
		return
	}
	c.JSON(http.StatusOK, hist)
}
