package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// AWS DynamoDB to others

func MigrationDynamoDBToFirestoreGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migDNFS")
		logger.Info("migDNFS get page accessed")
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
		start := time.Now()

		logger, logstrings := pageLogInit("migDNFS", "Export dynamoDB data to firestoreDB", start)

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
		if awsNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
		if gcpNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := awsNRDB.Copy(gcpNRDB); err != nil {
			end := time.Now()
			logger.Errorf("NRDBController copy failed : %v", err)
			logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from dynamoDB to firestoreDB", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationDynamoDBToMongoDBeGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migDNMG")
		logger.Info("migDNMG get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-DynamoDB-MongoDB",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationDynamoDBToMongoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migDNMG", "Export dynamoDB data to mongoDB", start)
		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
		if awsNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
		if ncpNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := awsNRDB.Copy(ncpNRDB); err != nil {
			end := time.Now()
			logger.Errorf("NRDBController copy failed : %v", err)
			logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from dynamoDB to ncp mongoDB", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

// Google Cloud Firestore to others

func MigrationFirestoreToDynamoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migFSDN")
		logger.Info("migFSDN get page accessed")
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
		start := time.Now()

		logger, logstrings := pageLogInit("migFSDN", "Export firestoreDB data to dynamoDB", start)

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
		if gcpNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
		if awsNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := gcpNRDB.Copy(awsNRDB); err != nil {
			end := time.Now()
			logger.Errorf("NRDBController copy failed : %v", err)
			logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from firestoreDB to dynamoDB", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationFirestoreToMongoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migFSMG")
		logger.Info("migFSMG get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Firestore-MongoDB",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationFirestoreToMongoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migFSMG", "Export firestoreDB data to mongoDB", start)

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
		if gcpNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
		if ncpNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := gcpNRDB.Copy(ncpNRDB); err != nil {
			end := time.Now()
			logger.Errorf("NRDBController copy failed : %v", err)
			logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from firestoreDB to mongoDB", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

// Naver Cloud MongoDB to others

func MigrationMongoDBToDynamoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migMGDN")
		logger.Info("migMGDN get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MongoDB-DynamoDB",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationMongoDBToDynamoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migMGDN", "Export mongoDB data to dynamoDB", start)

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
		if ncpNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
		if awsNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := ncpNRDB.Copy(awsNRDB); err != nil {
			end := time.Now()
			logger.Errorf("NRDBController copy failed : %v", err)
			logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from mongoDB to dynamoDB", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationMongoDBToFirestoreGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migMGFS")
		logger.Info("migMGFS get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MongoDB-Firestore",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationMongoDBToFirestorePostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migMGFS", "Export mongoDB data to firestoreDB", start)

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
		if gcpNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
		if ncpNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := ncpNRDB.Copy(gcpNRDB); err != nil {
			end := time.Now()
			logger.Errorf("NRDBController copy failed : %v", err)
			logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from mongoDB to firestoreDB", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}
