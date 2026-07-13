package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/fx"
	"fx-tracker/internal/service"
)

type Handler struct {
	cache   *fx.Cache
	history *fx.HistoryCache
	conv    *service.ConversionService
}

func New(cache *fx.Cache, history *fx.HistoryCache, conv *service.ConversionService) *Handler {
	return &Handler{cache: cache, history: history, conv: conv}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Rate(c *gin.Context) {
	c.JSON(http.StatusOK, h.cache.Get())
}

func (h *Handler) RateContext(c *gin.Context) {
	rate := h.cache.Get()
	now := time.Now()
	c.JSON(http.StatusOK, fx.RateContext{
		CurrentMyrUsd: rate.MyrUsd,
		CurrentUsdMyr: rate.UsdMyr,
		Stale:         rate.Stale,
		FetchedAt:     rate.FetchedAt,
		HistoryStale:  h.history.Stale(),
		Timeframes:    h.history.Assess(rate.MyrUsd, now),
	})
}

func (h *Handler) RateHistory(c *gin.Context) {
	pts := h.history.Points()
	out := make([]fx.Point, 0, len(pts))
	for _, p := range pts {
		out = append(out, fx.Point{Date: p.Date.Format("2006-01-02"), MyrUsd: p.MyrUsd})
	}
	c.JSON(http.StatusOK, fx.RateHistory{Points: out, Stale: h.history.Stale()})
}
