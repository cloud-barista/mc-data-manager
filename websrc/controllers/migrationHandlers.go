package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MigrationLinuxToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-S3",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationLinuxToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-S3",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationLinuxToGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationLinuxToGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationLinuxToNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationLinuxToNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationWindowsToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-S3",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationWindowsToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-S3",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationWindowsToGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationWindowsToGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationWindowsToNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationWindowsToNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

// SQL Database

func MigrationMySQLGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MySQL",
			"error":   nil,
		})
	}
}

func MigrationMySQLPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MySQL",
			"error":   nil,
		})
	}
}
