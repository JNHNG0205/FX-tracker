package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/fx"
)

func (h *Handler) GetSettings(c *gin.Context) {
	s, err := h.settings.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load settings"})
		return
	}
	c.JSON(http.StatusOK, s)
}

type updateSettingsReq struct {
	HomeCurrency string `json:"home_currency"`
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var req updateSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if !fx.IsSupported(req.HomeCurrency) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported currency"})
		return
	}
	if err := h.settings.SetHomeCurrency(c.Request.Context(), req.HomeCurrency); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update settings"})
		return
	}
	s, err := h.settings.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load settings"})
		return
	}
	c.JSON(http.StatusOK, s)
}
