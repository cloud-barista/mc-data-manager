package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AWS DynamoDB to others

func MigrationDynamoDBToFirestoreGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-DynamoDB-Firestore",
			"AWSRegions": GetAWSRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationDynamoDBToFirestorePostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-DynamoDB-Firestore",
			"AWSRegions": GetAWSRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationDynamoDBToMongoDBeGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-DynamoDB-MongoDB",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationDynamoDBToMongoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-DynamoDB-MongoDB",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

// Google Cloud Firestore to others

func MigrationFirestoreToDynamoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-Firestore-DynamoDB",
			"AWSRegions": GetAWSRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationFirestoreToDynamoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-Firestore-DynamoDB",
			"AWSRegions": GetAWSRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationFirestoreToMongoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Firestore-MongoDB",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationFirestoreToMongoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Firestore-MongoDB",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

// Naver Cloud MongoDB to others

func MigrationMongoDBToDynamoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MongoDB-DynamoDB",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationMongoDBToDynamoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MongoDB-DynamoDB",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationMongoDBToFirestoreGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MongoDB-Firestore",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationMongoDBToFirestorePostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: Get POST params
		// Migration data function

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MongoDB-Firestore",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}
