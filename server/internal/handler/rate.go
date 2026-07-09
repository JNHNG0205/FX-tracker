package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
