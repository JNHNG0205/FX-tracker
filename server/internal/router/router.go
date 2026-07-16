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
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type"},
	}))

	r.GET("/health", h.Health)
	r.GET("/api/rate", h.Rate)
	r.GET("/api/rate/context", h.RateContext)
	r.GET("/api/rate/history", h.RateHistory)
	r.GET("/api/currencies", h.Currencies)

	r.POST("/api/conversions", h.CreateConversion)
	r.GET("/api/conversions", h.ListConversions)
	r.GET("/api/conversions/status", h.ConversionStatus)
	r.PUT("/api/conversions/:id", h.UpdateConversion)
	r.DELETE("/api/conversions/:id", h.DeleteConversion)

	r.GET("/api/settings", h.GetSettings)
	r.PUT("/api/settings", h.UpdateSettings)
	return r
}
