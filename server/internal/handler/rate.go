package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/fx"
)

type Handler struct {
	cache *fx.Cache
}

func New(cache *fx.Cache) *Handler {
	return &Handler{cache: cache}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Rate(c *gin.Context) {
	c.JSON(http.StatusOK, h.cache.Get())
}

func (h *Handler) RateContext(c *gin.Context) {
	c.JSON(http.StatusOK, h.cache.Context())
}
