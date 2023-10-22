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

func GenerateLinuxSubmitPostHandler() gin.HandlerFunc {
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

func GenerateWindowsSubmitPostHandler() gin.HandlerFunc {
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

func GenerateS3SubmitPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		endpoint := ctx.PostForm("endpoint")
		apiKey := ctx.PostForm("apikey")
		apiSecret := ctx.PostForm("apisecret")

		fmt.Println("Endpoint:", endpoint)
		fmt.Println("API Key:", apiKey)
		fmt.Println("API Secret:", apiSecret)

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-S3",
			"error":   nil,
		})
	}
}

func GenerateGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-GCS",
			"error":   nil,
		})
	}
}

func GenerateGCSSubmitPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-GCS",
			"error":   nil,
		})
	}
}

func GenerateNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-NCS",
			"error":   nil,
		})
	}
}

func GenerateNCSSubmitPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-NCS",
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

func GenerateMySQLSubmitPostHandler() gin.HandlerFunc {
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
			"error":   nil,
		})
	}
}

func GenerateDynamoDBSubmitPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-DynamoDB",
			"error":   nil,
		})
	}
}

func GenerateFirestoreGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Firestore",
			"error":   nil,
		})
	}
}

func GenerateFirestoreSubmitPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Firestore",
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

func GenerateMongoDBSubmitPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Generate data function
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MongoDB",
			"error":   nil,
		})
	}
}
