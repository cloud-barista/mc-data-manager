package controllers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

func MainGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "main",
			"os":      runtime.GOOS,
			"error":   nil,
		})
	}
}
