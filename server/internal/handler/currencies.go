package handler

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"fx-tracker/internal/fx"
)

func (h *Handler) Currencies(c *gin.Context) {
	out := make([]fx.Currency, 0, len(fx.Supported))
	for code, name := range fx.Supported {
		out = append(out, fx.Currency{Code: code, Name: name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	c.JSON(http.StatusOK, out)
}
