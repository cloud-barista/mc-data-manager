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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cloud-barista/cm-data-mold/websrc/models"
	"github.com/labstack/echo/v4"
)

// GenerateLinuxPostHandler godoc
// @Summary Generate test data on on-premise Linux
// @Description Generate test data on on-premise Linux.
// @Tags [Test Data Generation] On-premise Linux
// @Accept  json
// @Produce  json
// @Param RequestBody body GenDataParams true "Parameters required to generate test data"
// @Success 200 {object} models.BasicResponse "Successfully generated test data"
// @Failure 400 {object} models.BasicResponse "Invalid Request"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /generate/linux [post]
func GenerateLinuxPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("genlinux", "Create dummy data in linux", start)

	if !osCheck(logger, start, "linux") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	params := getData("gen", ctx).(GenDataParams)

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully creating a dummy with Linux", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GenerateWindowsPostHandler godoc
// @Summary Generate test data on on-premise Windows
// @Description Generate test data on on-premise Windows.
// @Tags [Test Data Generation] On-premise Windows
// @Accept  json
// @Produce  json
// @Param RequestBody body GenDataParams true "Parameters required to generate test data"
// @Success 200 {object} models.BasicResponse "Successfully generated test data"
// @Failure 400 {object} models.BasicResponse "Invalid Request"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /generate/windows [post]
func GenerateWindowsPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("genwindows", "Create dummy data in windows", start)

	if !osCheck(logger, start, "windows") {
		fmt.Println("test")
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	params := getData("gen", ctx).(GenDataParams)

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	jobEnd(logger, "Successfully creating a dummy with Windows", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

type GenerateS3PostHandlerResponseBody struct {
	models.BasicResponse
}

// GenerateS3PostHandler godoc
// @Summary Generate test data on AWS S3
// @Description Generate test data on AWS S3.
// @Tags [Test Data Generation] AWS S3
// @Accept  json
// @Produce  json
// @Param RequestBody body GenDataParams true "Parameters required to generate test data"
// @Success 200 {object} models.BasicResponse "Successfully generated test data"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /generate/s3 [post]
func GenerateS3PostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genS3", "Create dummy data and import to s3", start)

	params := getData("gen", ctx).(GenDataParams)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	defer os.RemoveAll(tmpDir)
	params.DummyPath = tmpDir

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	awsOSC := getS3OSC(logger, start, "gen", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "s3", awsOSC, params.DummyPath) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	jobEnd(logger, "Dummy creation and import successful with s3", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GenerateGCPPostHandler godoc
// @Summary Generate test data on GCP Cloud Storage
// @Description Generate test data on GCP Cloud Storage.
// @Tags [Test Data Generation] GCP Cloud Storage
// @Accept  json
// @Produce  json
// @Param RequestBody body GenDataParams true "Parameters required to generate test data"
// @Param CredentialGCP formData file true "Parameters required to generate test data"
// @Success 200 {object} models.BasicResponse "Successfully generated test data"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /generate/gcp [post]
func GenerateGCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genGCP", "Create dummy data and import to gcp", start)

	params := getData("gen", ctx).(GenDataParams)

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

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	defer os.RemoveAll(tmpDir)
	params.DummyPath = tmpDir

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	gcpOSC := getGCPCOSC(logger, start, "gen", params, credFileName)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "gcp", gcpOSC, params.DummyPath) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Dummy creation and import successful with gcp", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GenerateNCPPostHandler godoc
// @Summary Generate test data on NCP Object Storage
// @Description Generate test data on NCP Object Storage.
// @Tags [Test Data Generation] NCP Object Storage
// @Accept  json
// @Produce  json
// @Param RequestBody body GenDataParams true "Parameters required to generate test data"
// @Success 200 {object} models.BasicResponse "Successfully generated test data"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /generate/ncp [post]
func GenerateNCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genNCP", "Create dummy data and import to ncp objectstorage", start)

	params := getData("gen", ctx).(GenDataParams)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	defer os.RemoveAll(tmpDir)

	params.DummyPath = tmpDir

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	ncpOSC := getS3COSC(logger, start, "gen", params)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "ncp", ncpOSC, params.DummyPath) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Create dummy data and import to ncp objectstorage", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GenerateMySQLPostHandler godoc
// @Summary Generate test data on MySQL
// @Description Generate test data on MySQL.
// @Tags [Test Data Generation] MySQL
// @Accept  json
// @Produce  json
// @Param RequestBody body GenDataParams true "Parameters required to generate test data"
// @Success 200 {object} models.BasicResponse "Successfully generated test data"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /generate/mysql [post]
func GenerateMySQLPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genmysql", "Create dummy data and import to mysql", start)

	params := getData("gen", ctx).(GenDataParams)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.DummyPath = tmpDir
	params.CheckServerSQL = "on"
	params.SizeServerSQL = "5"

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	rdbc := getMysqlRDBC(logger, start, "gen", params)
	if rdbc == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	sqlList := []string{}
	if !walk(logger, start, &sqlList, params.DummyPath, ".sql") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	logger.Info("Start Import with mysql")
	for _, sql := range sqlList {
		logger.Infof("Read sql file : %s", sql)
		data, err := os.ReadFile(sql)
		if err != nil {
			end := time.Now()
			logger.Errorf("os ReadFile failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  nil,
			})

		}

		logger.Infof("Put start : %s", filepath.Base(sql))
		if err := rdbc.Put(string(data)); err != nil {
			end := time.Now()
			logger.Errorf("RDBController import failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  nil,
			})
		}
		logger.Infof("sql put success : %s", filepath.Base(sql))
	}

	jobEnd(logger, "Dummy creation and import successful with mysql", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GenerateDynamoDBPostHandler godoc
// @Summary Generate test data on AWS DynamoDB
// @Description Generate test data on AWS DynamoDB.
// @Tags [Test Data Generation] AWS DynamoDB
// @Accept  json
// @Produce  json
// @Param RequestBody body GenDataParams true "Parameters required to generate test data"
// @Success 200 {object} models.BasicResponse "Successfully generated test data"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /generate/dynamodb [post]
func GenerateDynamoDBPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("gendynamodb", "Create dummy data and import to dynamoDB", start)

	params := getData("gen", ctx).(GenDataParams)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.DummyPath = tmpDir
	params.CheckServerJSON = "on"
	params.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	awsNRDB := getDynamoNRDBC(logger, start, "gen", params)
	if awsNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !nrdbPutWorker(logger, start, "DynamoDB", awsNRDB, jsonList) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Dummy creation and import successful with dynamoDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GenerateFirestorePostHandler godoc
// @Summary Generate test data on GCP Firestore
// @Description Generate test data on GCP Firestore.
// @Tags [Test Data Generation] GCP Firestore
// @Accept  json
// @Produce  json
// @Param RequestBody body GenDataParams true "Parameters required to generate test data"
// @Param CredentialGCP formData file true "Parameters required to generate test data"
// @Success 200 {object} models.BasicResponse "Successfully generated test data"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /generate/firestore [post]
func GenerateFirestorePostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genfirestore", "Create dummy data and import to firestoreDB", start)

	params := GenDataParams{}
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

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.DummyPath = tmpDir
	params.CheckServerJSON = "on"
	params.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	gcpNRDB := getFirestoreNRDBC(logger, start, "gen", params, credFileName)
	if gcpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !nrdbPutWorker(logger, start, "FirestoreDB", gcpNRDB, jsonList) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Dummy creation and import successful with firestoreDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GenerateMongoDBPostHandler godoc
// @Summary Generate test data on NCP MongoDB
// @Description Generate test data on NCP MongoDB.
// @Tags [Test Data Generation] NCP MongoDB
// @Accept  json
// @Produce  json
// @Param RequestBody body GenDataParams true "Parameters required to generate test data"
// @Success 200 {object} models.BasicResponse "Successfully generated test data"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /generate/mongodb [post]
func GenerateMongoDBPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genmongodb", "Create dummy data and import to mongoDB", start)

	params := getData("gen", ctx).(GenDataParams)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.DummyPath = tmpDir
	params.CheckServerJSON = "on"
	params.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	ncpNRDB := getMongoNRDBC(logger, start, "gen", params)
	if ncpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !nrdbPutWorker(logger, start, "MongoDB", ncpNRDB, jsonList) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Dummy creation and import successful with mongoDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}
