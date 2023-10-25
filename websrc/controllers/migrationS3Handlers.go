package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Object Storage

// FROM AWS S3
func MigrationS3ToLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-S3-Linux",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationS3ToLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-S3-Linux",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationS3ToWindowsGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-S3-Windows",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationS3ToWindowsPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-S3-Windows",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationS3ToGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-S3-GCS",
			"AWSRegions": GetAWSRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationS3ToGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-S3-GCS",
			"AWSRegions": GetAWSRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationS3ToNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-S3-NCS",
			"AWSRegions": GetAWSRegions(),
			"NCPRegions": GetNCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationS3ToNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-S3-NCS",
			"AWSRegions": GetAWSRegions(),
			"NCPRegions": GetNCPRegions(),
			"error":      nil,
		})
	}
}
