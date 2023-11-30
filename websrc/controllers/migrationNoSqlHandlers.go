package controllers

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
)

// AWS DynamoDB to others

func MigrationDynamoDBToFirestoreGetHandler(ctx echo.Context) error {

	logger := getLogger("migDNFS")
	logger.Info("migDNFS get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content":    "Migration-DynamoDB-Firestore",
		"AWSRegions": GetAWSRegions(),
		"os":         runtime.GOOS,
		"GCPRegions": GetGCPRegions(),
		"error":      nil,
	})
}

func MigrationDynamoDBToFirestorePostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migDNFS", "Export dynamoDB data to firestoreDB", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
	if awsNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
	if gcpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if err := awsNRDB.Copy(gcpNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from dynamoDB to firestoreDB", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationDynamoDBToMongoDBGetHandler(ctx echo.Context) error {

	logger := getLogger("migDNMG")
	logger.Info("migDNMG get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-DynamoDB-MongoDB",
		"Regions": GetAWSRegions(),
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func MigrationDynamoDBToMongoDBPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migDNMG", "Export dynamoDB data to mongoDB", start)
	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
	if awsNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
	if ncpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if err := awsNRDB.Copy(ncpNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from dynamoDB to ncp mongoDB", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

// Google Cloud Firestore to others

func MigrationFirestoreToDynamoDBGetHandler(ctx echo.Context) error {

	logger := getLogger("migFSDN")
	logger.Info("migFSDN get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content":    "Migration-Firestore-DynamoDB",
		"AWSRegions": GetAWSRegions(),
		"os":         runtime.GOOS,
		"GCPRegions": GetGCPRegions(),
		"error":      nil,
	})
}

func MigrationFirestoreToDynamoDBPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migFSDN", "Export firestoreDB data to dynamoDB", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
	if gcpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
	if awsNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if err := gcpNRDB.Copy(awsNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from firestoreDB to dynamoDB", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationFirestoreToMongoDBGetHandler(ctx echo.Context) error {

	logger := getLogger("migFSMG")
	logger.Info("migFSMG get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-Firestore-MongoDB",
		"Regions": GetGCPRegions(),
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func MigrationFirestoreToMongoDBPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migFSMG", "Export firestoreDB data to mongoDB", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
	if gcpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
	if ncpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if err := gcpNRDB.Copy(ncpNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from firestoreDB to mongoDB", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

// Naver Cloud MongoDB to others

func MigrationMongoDBToDynamoDBGetHandler(ctx echo.Context) error {

	logger := getLogger("migMGDN")
	logger.Info("migMGDN get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-MongoDB-DynamoDB",
		"Regions": GetAWSRegions(),
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func MigrationMongoDBToDynamoDBPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migMGDN", "Export mongoDB data to dynamoDB", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
	if ncpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
	if awsNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if err := ncpNRDB.Copy(awsNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from mongoDB to dynamoDB", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationMongoDBToFirestoreGetHandler(ctx echo.Context) error {

	logger := getLogger("migMGFS")
	logger.Info("migMGFS get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-MongoDB-Firestore",
		"Regions": GetGCPRegions(),
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func MigrationMongoDBToFirestorePostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migMGFS", "Export mongoDB data to firestoreDB", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
	if gcpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
	if ncpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if err := ncpNRDB.Copy(gcpNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from mongoDB to firestoreDB", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}
