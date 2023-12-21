/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/cloud-barista/cm-data-mold/websrc/models"
	"github.com/labstack/echo/v4"
)

// AWS DynamoDB to others

func MigrationDynamoDBToFirestorePostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migDNFS", "Export dynamoDB data to firestoreDB", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
	if awsNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
	if gcpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if err := awsNRDB.Copy(gcpNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from dynamoDB to firestoreDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func MigrationDynamoDBToMongoDBPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migDNMG", "Export dynamoDB data to mongoDB", start)
	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
	if awsNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
	if ncpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if err := awsNRDB.Copy(ncpNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from dynamoDB to ncp mongoDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// Google Cloud Firestore to others

func MigrationFirestoreToDynamoDBPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migFSDN", "Export firestoreDB data to dynamoDB", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
	if gcpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
	if awsNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if err := gcpNRDB.Copy(awsNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from firestoreDB to dynamoDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func MigrationFirestoreToMongoDBPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migFSMG", "Export firestoreDB data to mongoDB", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
	if gcpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
	if ncpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if err := gcpNRDB.Copy(ncpNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from firestoreDB to mongoDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// Naver Cloud MongoDB to others

func MigrationMongoDBToDynamoDBPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migMGDN", "Export mongoDB data to dynamoDB", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
	if ncpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	awsNRDB := getDynamoNRDBC(logger, start, "mig", params)
	if awsNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if err := ncpNRDB.Copy(awsNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from mongoDB to dynamoDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func MigrationMongoDBToFirestorePostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migMGFS", "Export mongoDB data to firestoreDB", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpNRDB := getFirestoreNRDBC(logger, start, "mig", params, credFileName)
	if gcpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	ncpNRDB := getMongoNRDBC(logger, start, "mig", params)
	if ncpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if err := ncpNRDB.Copy(gcpNRDB); err != nil {
		end := time.Now()
		logger.Errorf("NRDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from mongoDB to firestoreDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}
