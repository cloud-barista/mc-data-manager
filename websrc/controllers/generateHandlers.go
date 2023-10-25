package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GenerateLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Linux",
			"error":   nil,
		})
	}
}

func GenerateLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		addr := ctx.PostForm("address")
		fmt.Println("postform:", addr)

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Linux",
			"error":   nil,
		})
	}
}

func GenerateWindowsGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Windows",
			"error":   nil,
		})
	}
}

func GenerateWindowsPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Windows",
			"error":   nil,
		})
	}
}

func GenerateS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-S3",
			"Regions": GetAWSRegions(),
			"Error":   nil,
		})
	}
}

func GenerateS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		endpoint := ctx.PostForm("endpoint")
		apiKey := ctx.PostForm("apikey")
		apiSecret := ctx.PostForm("apisecret")

		fmt.Println("Endpoint:", endpoint)
		fmt.Println("API Key:", apiKey)
		fmt.Println("API Secret:", apiSecret)

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-S3",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func GenerateGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateMySQLGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MySQL",
			"error":   nil,
		})
	}
}

func GenerateMySQLPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MySQL",
			"error":   nil,
		})
	}
}

func GenerateDynamoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-DynamoDB",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func GenerateDynamoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-DynamoDB",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func GenerateFirestoreGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Firestore",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateFirestorePostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Firestore",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateMongoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MongoDB",
			"error":   nil,
		})
	}
}

func GenerateMongoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MongoDB",
			"error":   nil,
		})
	}
}
