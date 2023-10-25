package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Object Storage

// FROM Naver Cloud Storage
func MigrationNCSToLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-NCS-Linux",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationNCSToLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-NCS-Linux",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationNCSToWindowsGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-NCS-Windows",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationNCSToWindowsPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-NCS-Windows",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationNCSToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-NCS-S3",
			"NCPRegions": GetNCPRegions(),
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationNCSToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-NCS-S3",
			"NCPRegions": GetNCPRegions(),
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationNCSToGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-NCS-GCS",
			"NCPRegions": GetNCPRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationNCSToGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-NCS-GCS",
			"NCPRegions": GetNCPRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}
