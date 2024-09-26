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
	"path/filepath"
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/labstack/echo/v4"
)

// GenerateLinuxPostHandler godoc
//
//	@ID 			GenerateLinuxPostHandler
//	@Summary		Generate test data on on-premise Linux
//	@Description	Generate test data on on-premise Linux.
//	@Tags			[Data Generation]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/linux [post]
func GenerateLinuxPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genlinux", "Create dummy data in linux", start)

	if !osCheck(logger, start, "linux") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	params := GenarateTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	if !dummyCreate(logger, start, params.TargetPoint.GenFileParams) {
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
//
//	@ID 			GenerateWindowsPostHandler
//	@Summary		Generate test data on on-premise Windows
//	@Description	Generate test data on on-premise Windows.
//	@Tags			[Data Generation]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/windows [post]
func GenerateWindowsPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genwindows", "Create dummy data in windows", start)

	if !osCheck(logger, start, "windows") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	params := GenarateTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !dummyCreate(logger, start, params.TargetPoint.GenFileParams) {
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
//
//	@ID 			GenerateS3PostHandler
//	@Summary		Generate test data on AWS S3
//	@Description	Generate test data on AWS S3.
//	@Tags			[Data Generation], [Service Object Storage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/aws [post]
func GenerateS3PostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genS3", "Create dummy data and import to s3", start)

	params := GenarateTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	defer os.RemoveAll(tmpDir)
	params.TargetPoint.DummyPath = tmpDir
	if !dummyCreate(logger, start, params.TargetPoint.GenFileParams) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	logger.Info().Msgf("ProviderConfig %+v", params.TargetPoint.ProviderConfig)
	awsOSC := getS3OSC(logger, start, "gen", params.TargetPoint.ProviderConfig)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "s3", awsOSC, params.TargetPoint.DummyPath) {
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
//
//	@ID 			GenerateGCPPostHandler
//	@Summary		Generate test data on GCP Cloud Storage
//	@Description	Generate test data on GCP Cloud Storage.
//	@Tags			[Data Generation], [Service Object Storage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		GenarateTask			true	"Parameters required to generate test data"
//	@Success		200				{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		500				{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/gcp [post]
func GenerateGCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genGCP", "Create dummy data and import to gcp", start)

	params := GenarateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	defer os.RemoveAll(tmpDir)
	params.TargetPoint.DummyPath = tmpDir
	if !dummyCreate(logger, start, params.TargetPoint.GenFileParams) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	gcpOSC := getGCPCOSC(logger, start, "gen", params.TargetPoint.ProviderConfig)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "gcp", gcpOSC, params.TargetPoint.DummyPath) {
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
//
//	@ID 			GenerateNCPPostHandler
//	@Summary		Generate test data on NCP Object Storage
//	@Description	Generate test data on NCP Object Storage.
//	@Tags			[Data Generation], [Service Object Storage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/ncp [post]
func GenerateNCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genNCP", "Create dummy data and import to ncp objectstorage", start)

	params := GenarateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	defer os.RemoveAll(tmpDir)

	params.TargetPoint.DummyPath = tmpDir
	if !dummyCreate(logger, start, params.TargetPoint.GenFileParams) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	ncpOSC := getS3COSC(logger, start, "gen", params.TargetPoint.ProviderConfig)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	if !oscImport(logger, start, "ncp", ncpOSC, params.TargetPoint.DummyPath) {
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
//
//	@ID 			GenerateMySQLPostHandler
//	@Summary		Generate test data on MySQL
//	@Description	Generate test data on MySQL.
//	@Tags			[Data Generation], [Service RDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/mysql [post]
func GenerateMySQLPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genmysql", "Create dummy data and import to mysql", start)

	params := GenarateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.TargetPoint.DummyPath = tmpDir
	params.TargetPoint.CheckServerSQL = true
	params.TargetPoint.SizeServerSQL = "1"

	if !dummyCreate(logger, start, params.TargetPoint.GenFileParams) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	rdbc := getMysqlRDBC(logger, start, "gen", params.TargetPoint.ProviderConfig)
	if rdbc == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	sqlList := []string{}
	if !walk(logger, start, &sqlList, params.TargetPoint.DummyPath, ".sql") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	logger.Info().Msg("Start Import with mysql")
	for _, sql := range sqlList {
		logger.Info().Str("file", sql).Msg("Read sql file")
		data, err := os.ReadFile(sql)
		if err != nil {
			end := time.Now()
			logger.Error().Err(err).Msg("os ReadFile failed")
			logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Info().Str("Elapsed time", end.Sub(start).String())
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  nil,
			})

		}

		logger.Info().Str("file", filepath.Base(sql)).Msg("Put start")
		if err := rdbc.Put(string(data)); err != nil {
			end := time.Now()
			logger.Error().Err(err).Msg("RDBController import failed")
			logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Info().Str("Elapsed time", end.Sub(start).String())
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  nil,
			})
		}
		logger.Info().Str("file", filepath.Base(sql)).Msg("sql put success")
	}

	jobEnd(logger, "Dummy creation and import successful with mysql", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GenerateDynamoDBPostHandler godoc
//
//	@ID 			GenerateDynamoDBPostHandler
//	@Summary		Generate test data on AWS DynamoDB
//	@Description	Generate test data on AWS DynamoDB.
//	@Tags			[Data Generation], [Service NRDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/dynamodb [post]
func GenerateDynamoDBPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "gendynamodb", "Create dummy data and import to dynamoDB", start)

	params := GenarateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.TargetPoint.DummyPath = tmpDir
	params.TargetPoint.CheckServerJSON = true
	params.TargetPoint.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params.TargetPoint.GenFileParams) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.TargetPoint.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	awsNRDB := getDynamoNRDBC(logger, start, "gen", params.TargetPoint.ProviderConfig)
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
//
//	@ID 			GenerateFirestorePostHandler
//	@Summary		Generate test data on GCP Firestore
//	@Description	Generate test data on GCP Firestore.
//	@Tags			[Data Generation], [Service NRDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		GenarateTask				true	"Parameters required to generate test data"
//	@Success		200				{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500				{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/firestore [post]
func GenerateFirestorePostHandler(ctx echo.Context) error {
	start := time.Now()
	pageName := "genfirestore"
	pageInfo := "Create dummy data and import to firestoreDB"
	logger, logstrings := pageLogInit(ctx, pageName, pageInfo, start)

	params := GenarateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.TargetPoint.DummyPath = tmpDir
	params.TargetPoint.CheckServerJSON = true
	params.TargetPoint.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params.TargetPoint.GenFileParams) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.TargetPoint.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	gcpNRDB := getFirestoreNRDBC(logger, start, "gen", params.TargetPoint.ProviderConfig)
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
//
//	@ID 			GenerateMongoDBPostHandler
//	@Summary		Generate test data on NCP MongoDB
//	@Description	Generate test data on NCP MongoDB.
//	@Tags			[Data Generation], [Service NRDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/mongodb [post]
func GenerateMongoDBPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genmongodb", "Create dummy data and import to mongoDB", start)

	params := GenarateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.TargetPoint.DummyPath = tmpDir
	params.TargetPoint.CheckServerJSON = true
	params.TargetPoint.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params.TargetPoint.GenFileParams) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.TargetPoint.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	ncpNRDB := getMongoNRDBC(logger, start, "gen", params.TargetPoint.ProviderConfig)
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
