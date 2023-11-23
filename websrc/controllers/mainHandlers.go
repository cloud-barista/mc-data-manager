package controllers

import (
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
)

func MainGetHandler(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "main",
		"os":      runtime.GOOS,
		"error":   nil,
	})
}
