package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/fx"
)

type Handler struct {
	cache   *fx.Cache
	history *fx.HistoryCache
}

func New(cache *fx.Cache, history *fx.HistoryCache) *Handler {
	return &Handler{cache: cache, history: history}
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
