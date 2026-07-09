package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"fx-tracker/internal/handler"
)

func New(h *handler.Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Content-Type"},
	}))

	r.GET("/health", h.Health)
	return r
}
