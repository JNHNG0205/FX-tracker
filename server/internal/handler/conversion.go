package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/model"
)

type createConversionReq struct {
	Date       *time.Time `json:"date"`
	MyrAmount  float64    `json:"myr_amount"`
	RateMyrUsd float64    `json:"rate_myr_usd"`
	Note       string     `json:"note"`
}

func (h *Handler) CreateConversion(c *gin.Context) {
	var req createConversionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.MyrAmount <= 0 || req.RateMyrUsd <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "myr_amount and rate_myr_usd must be > 0"})
		return
	}
	date := time.Now()
	if req.Date != nil {
		date = *req.Date
	}
	conv := &model.Conversion{
		Date:       date,
		MyrAmount:  req.MyrAmount,
		RateMyrUsd: req.RateMyrUsd,
		Note:       req.Note,
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
	liveRate := h.cache.Get().MyrUsd
	status, err := h.conv.Status(c.Request.Context(), liveRate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not compute status"})
		return
	}
	c.JSON(http.StatusOK, status)
}
