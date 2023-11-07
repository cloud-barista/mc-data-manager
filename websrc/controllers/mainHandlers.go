package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MainGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "main",
			"error":   nil,
		})
	}
}
