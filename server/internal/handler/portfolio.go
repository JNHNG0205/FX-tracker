package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/fx"
)

// Portfolio resolves the home currency from ?home= (validated) or falls
// back to the user's saved settings, then returns the full per-holding and
// aggregate return breakdown.
func (h *Handler) Portfolio(c *gin.Context) {
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

	resp, err := h.portfolio.Compute(c.Request.Context(), home)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not compute portfolio"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
