package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Object Storage

// FROM Google Cloud Storage
func MigrationGCSToLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCS-Linux",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationGCSToLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCS-Linux",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationGCSToWindowsGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCS-Windows",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationGCSToWindowsPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCS-Windows",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationGCSToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCS-S3",
			"GCPRegions": GetGCPRegions(),
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationGCSToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCS-S3",
			"GCPRegions": GetGCPRegions(),
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationGCSToNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCS-NCS",
			"GCPRegions": GetGCPRegions(),
			"NCPRegions": GetNCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationGCSToNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCS-NCS",
			"GCPRegions": GetGCPRegions(),
			"NCPRegions": GetNCPRegions(),
			"error":      nil,
		})
	}
}
