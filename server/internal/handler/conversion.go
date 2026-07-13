package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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
	if req.MyrAmount <= 0 || req.RateMyrUsd <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "myr_amount and rate_myr_usd must be > 0"})
		return
	}
	date := time.Now()
	if req.Date != nil {
		date = *req.Date
	}
	conv := &model.Conversion{ID: id, Date: date, MyrAmount: req.MyrAmount, RateMyrUsd: req.RateMyrUsd, Note: req.Note}
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
