package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cloud-barista/cm-data-mold/config"
	"github.com/cloud-barista/cm-data-mold/pkg/nrdbms/awsdnmdb"
	"github.com/cloud-barista/cm-data-mold/pkg/nrdbms/gcpfsdb"
	"github.com/cloud-barista/cm-data-mold/pkg/nrdbms/ncpmgdb"
	"github.com/cloud-barista/cm-data-mold/service/nrdbc"
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
		awsRegion := ctx.PostForm("awsRegion")
		awsAccessKey := ctx.PostForm("awsAccessKey")
		awsSecretKey := ctx.PostForm("awsSecretKey")
		gcpRegion := ctx.PostForm("gcpRegion")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-DynamoDB-Firestore",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-DynamoDB-Firestore",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")
		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-DynamoDB-Firestore",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		ac, err := config.NewDynamoDBClient(awsAccessKey, awsSecretKey, awsRegion)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-DynamoDB-Firestore",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		awsNRDBC, err := nrdbc.New(awsdnmdb.New(ac, awsRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-DynamoDB-Firestore",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		gc, err := config.NewFireStoreClient(credFileName, projectid)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-DynamoDB-Firestore",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		gcpNRDBC, err := nrdbc.New(gcpfsdb.New(gc, gcpRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-DynamoDB-Firestore",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		if err := awsNRDBC.Copy(gcpNRDBC); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-DynamoDB-Firestore",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

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
		awsRegion := ctx.PostForm("awsRegion")
		awsAccessKey := ctx.PostForm("awsAccessKey")
		awsSecretKey := ctx.PostForm("awsSecretKey")
		host := ctx.PostForm("host")
		port := ctx.PostForm("port")
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		databaseName := ctx.PostForm("databaseName")

		ac, err := config.NewDynamoDBClient(awsAccessKey, awsSecretKey, awsRegion)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-DynamoDB-MongoDB",
				"Regions": GetAWSRegions(),
				"error":   nil,
			})
			return
		}

		awsNRDBC, err := nrdbc.New(awsdnmdb.New(ac, awsRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-DynamoDB-MongoDB",
				"Regions": GetAWSRegions(),
				"error":   nil,
			})
			return
		}

		Port, _ := strconv.Atoi(port)
		nc, err := config.NewNCPMongoDBClient(username, password, host, Port)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-DynamoDB-MongoDB",
				"Regions": GetAWSRegions(),
				"error":   nil,
			})
			return
		}

		ncpNRDBC, err := nrdbc.New(ncpmgdb.New(nc, databaseName))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-DynamoDB-MongoDB",
				"Regions": GetAWSRegions(),
				"error":   nil,
			})
			return
		}

		if err := awsNRDBC.Copy(ncpNRDBC); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-DynamoDB-MongoDB",
				"Regions": GetAWSRegions(),
				"error":   nil,
			})
			return
		}

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
		gcpRegion := ctx.PostForm("gcpRegion")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-Firestore-DynamoDB",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-Firestore-DynamoDB",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")
		awsRegion := ctx.PostForm("awsRegion")
		awsAccessKey := ctx.PostForm("awsAccessKey")
		awsSecretKey := ctx.PostForm("awsSecretKey")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-Firestore-DynamoDB",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		ac, err := config.NewDynamoDBClient(awsAccessKey, awsSecretKey, awsRegion)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-Firestore-DynamoDB",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		awsNRDBC, err := nrdbc.New(awsdnmdb.New(ac, awsRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-Firestore-DynamoDB",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		gc, err := config.NewFireStoreClient(credFileName, projectid)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-Firestore-DynamoDB",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		gcpNRDBC, err := nrdbc.New(gcpfsdb.New(gc, gcpRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-Firestore-DynamoDB",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		if err := gcpNRDBC.Copy(awsNRDBC); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-Firestore-DynamoDB",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

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
		gcpRegion := ctx.PostForm("gcpRegion")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")
		host := ctx.PostForm("host")
		port := ctx.PostForm("port")
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		databaseName := ctx.PostForm("databaseName")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		gc, err := config.NewFireStoreClient(credFileName, projectid)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		gcpNRDBC, err := nrdbc.New(gcpfsdb.New(gc, gcpRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		Port, _ := strconv.Atoi(port)
		nc, err := config.NewNCPMongoDBClient(username, password, host, Port)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		ncpNRDBC, err := nrdbc.New(ncpmgdb.New(nc, databaseName))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		if err := gcpNRDBC.Copy(ncpNRDBC); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

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
		host := ctx.PostForm("host")
		port := ctx.PostForm("port")
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		databaseName := ctx.PostForm("databaseName")
		awsRegion := ctx.PostForm("awsRegion")
		awsAccessKey := ctx.PostForm("awsAccessKey")
		awsSecretKey := ctx.PostForm("awsSecretKey")

		Port, _ := strconv.Atoi(port)
		nc, err := config.NewNCPMongoDBClient(username, password, host, Port)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-MongoDB-DynamoDB",
				"Regions": GetAWSRegions(),
				"error":   nil,
			})
			return
		}

		ncpNRDBC, err := nrdbc.New(ncpmgdb.New(nc, databaseName))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-MongoDB-DynamoDB",
				"Regions": GetAWSRegions(),
				"error":   nil,
			})
			return
		}

		ac, err := config.NewDynamoDBClient(awsAccessKey, awsSecretKey, awsRegion)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-MongoDB-DynamoDB",
				"Regions": GetAWSRegions(),
				"error":   nil,
			})
			return
		}

		awsNRDBC, err := nrdbc.New(awsdnmdb.New(ac, awsRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-MongoDB-DynamoDB",
				"Regions": GetAWSRegions(),
				"error":   nil,
			})
			return
		}

		if err := ncpNRDBC.Copy(awsNRDBC); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-MongoDB-DynamoDB",
				"Regions": GetAWSRegions(),
				"error":   nil,
			})
			return
		}

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
		host := ctx.PostForm("host")
		port := ctx.PostForm("port")
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		databaseName := ctx.PostForm("databaseName")
		gcpRegion := ctx.PostForm("gcpRegion")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		Port, _ := strconv.Atoi(port)
		nc, err := config.NewNCPMongoDBClient(username, password, host, Port)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		ncpNRDBC, err := nrdbc.New(ncpmgdb.New(nc, databaseName))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		gc, err := config.NewFireStoreClient(credFileName, projectid)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		gcpNRDBC, err := nrdbc.New(gcpfsdb.New(gc, gcpRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		if err := ncpNRDBC.Copy(gcpNRDBC); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-Firestore-MongoDB",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MongoDB-Firestore",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}
