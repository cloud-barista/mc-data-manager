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
	"github.com/cloud-barista/mc-data-manager/service/task"
	"github.com/labstack/echo/v4"
)

type GenerateS3PostHandlerResponseBody struct {
	models.BasicResponse
}

// GenerateObjectStoragePostHandler godoc
//
//	@ID 			GenerateObjectStoragePostHandler
//	@Summary		Generate test data on Object Storage
//	@Description	Generate test data on Object Storage
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/objectstorage [post]
func GenerateObjectStoragePostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genS3", "Create dummy data and import to s3", start)

	params := models.DataTask{}
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
	params.Dummy.DummyPath = tmpDir
	if !dummyCreate(logger, start, params.Dummy) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	params.Dummy.DummyPath = tmpDir
	params.TaskMeta.TaskID = params.OperationId
	params.TaskMeta.TaskType = models.Generate
	params.TaskMeta.ServiceType = models.ObejectStorage

	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(params) {
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

// GenerateRDBMSPostHandler godoc
//
//	@ID 			GenerateRDBMSPostHandler
//	@Summary		Generate test data on RDBMS
//	@Description	Generate test data on RDBMS
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/rdbms [post]
func GenerateRDBMSPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "gendynamodb", "Create dummy data and import to SQLDB", start)

	params := models.DataTask{}
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

	params.Dummy.DummyPath = tmpDir
	params.Dummy.CheckServerSQL = true
	params.Dummy.SizeServerSQL = "1"

	if !dummyCreate(logger, start, params.Dummy) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	params.TaskMeta.TaskID = params.OperationId
	params.TaskMeta.TaskType = models.Generate
	params.TaskMeta.ServiceType = models.RDBMS
	dparam := models.CommandTask{}
	if !getDataWithReBind(logger, start, ctx, &dparam) {
	}

	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Dummy creation and import successful with SQLDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GenerateNRDBMSPostHandler godoc
//
//	@ID 			GenerateNRDBMSPostHandler
//	@Summary		Generate test data on Object Storage
//	@Description	Generate test data on Object Storage
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/nrdbms [post]
func GenerateNRDBMSPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "gendynamodb", "Create dummy data and import to NoSQLDB", start)

	params := models.DataTask{}
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

	params.Dummy.DummyPath = tmpDir
	params.Dummy.CheckServerJSON = true
	params.Dummy.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params.Dummy) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	params.TaskMeta.TaskID = params.OperationId
	params.TaskMeta.TaskType = models.Generate
	params.TaskMeta.ServiceType = models.NRDBMS
	dparam := models.CommandTask{}
	if !getDataWithReBind(logger, start, ctx, &dparam) {
	}

	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Dummy creation and import successful with NoSQLDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GetAllGenerateHandler godoc
//
//	@ID 			GetAllGenerateHandler
//	@Summary		Get all Tasks
//	@Description	Retrieve a list of all Tasks in the system.
//	@Tags			[Generate]
//	@Produce		json
//	@Success		200		{array}		models.Task	"Successfully retrieved all Tasks"
//	@Failure		500		{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate [get]
func GetAllGenerateHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Get-task-list", "Get an existing task", start)
	manager := task.GetFileScheduleManager()
	tasks, err := manager.GetTasksByTypeList(models.Generate)
	if err != nil {
		errStr := err.Error()
		logger.Error().Err(err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}
	logger.Info().Msgf("%v", tasks)
	jobEnd(logger, "Successfully Get Task List", start)
	return ctx.JSON(http.StatusOK, tasks)
}

// GetGenerateHandler godoc
//
//	@ID 			GetGenerateHandler
//	@Summary		Get a Task by ID
//	@Description	Get the details of a Task using its ID.
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Task ID"
//	@Success		200		{object}	models.Task	"Successfully retrieved a Task"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/generate/{id} [get]
func GetGenerateHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Get-task", "Get an existing task", start)
	id := ctx.Param("id")
	manager := task.GetFileScheduleManager()
	task, err := manager.GetTasksByType(models.Generate, id)
	if err != nil {
		errStr := err.Error()
		logger.Error().Err(err)
		return ctx.JSON(http.StatusNotFound, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	return ctx.JSON(http.StatusOK, task)
}

// UpdateGenerateHandler godoc
//
//	@ID 			UpdateGenerateHandler
//	@Summary		Update an existing Task
//	@Description	Update the details of an existing Task using its ID.
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Task ID"
//	@Param			RequestBody	body	models.Schedule	true	"Parameters required for updating a Task"
//	@Success		200			{object}	models.BasicResponse	"Successfully updated the Task"
//	@Failure		404			{object}	models.BasicResponse	"Task not found"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/{id} [put]
func UpdateGenerateHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Update-task", "Updating an existing task", start)
	id := ctx.Param("id")
	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		errStr := "Invalid request data"
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}
	manager := task.GetFileScheduleManager()
	if err := manager.UpdateTasksByType(models.Generate, id, params.BasicDataTask); err != nil {
		errStr := err.Error()
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// DeleteGeneratekHandler godoc
//
//	@ID 			DeleteGeneratekHandler
//	@Summary		Delete a Task
//	@Description	Delete an existing Task using its ID.
//	@Tags			[Generate]
//	@Produce		json
//	@Param			id		path	string	true	"Task ID"
//	@Success		200		{object}	models.BasicResponse	"Successfully deleted the Task"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/generate/{id} [delete]
func DeleteGeneratekHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Delete-task", "Delete an existing task", start)
	id := ctx.Param("id")

	manager := task.GetFileScheduleManager()
	if err := manager.DeleteTasksByType(models.Generate, id); err != nil {
		errStr := "Task not found"
		logger.Error().Msg(errStr)

		return ctx.JSON(http.StatusNotFound, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GenerateLinuxPostHandler godoc
//
//	@ID 			GenerateLinuxPostHandler
//	@Summary		Generate test data on on-premise Linux
//	@Description	Generate test data on on-premise Linux.
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
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

	params := models.GenarateTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	logger.Info().Msgf("%v", params.TargetPoint)
	if !dummyCreate(logger, start, params.Dummy) {
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
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
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

	params := models.GenarateTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !dummyCreate(logger, start, params.Dummy) {
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

// GenerateS3PostHandler godoc
//
//	@ID 			GenerateS3PostHandler
//	@Summary		Generate test data on AWS S3
//	@Description	Generate test data on AWS S3.
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/aws [post]
func GenerateS3PostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genS3", "Create dummy data and import to s3", start)

	params := models.GenarateTask{}
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
	params.Dummy.DummyPath = tmpDir
	if !dummyCreate(logger, start, params.Dummy) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	dparams := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &dparams) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	dparams.Dummy.DummyPath = tmpDir
	dparams.TargetPoint = params.TargetPoint

	dparams.TaskMeta.TaskID = dparams.OperationId
	dparams.TaskMeta.TaskType = models.Generate
	dparams.TaskMeta.ServiceType = models.ObejectStorage
	// dparams.Status = models.StatusInactive
	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(dparams) {
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
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
//	@Success		200				{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		500				{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/gcp [post]
func GenerateGCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genGCP", "Create dummy data and import to gcp", start)

	params := models.GenarateTask{}
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
	params.Dummy.DummyPath = tmpDir
	if !dummyCreate(logger, start, params.Dummy) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	gcpOSC := getGCPCOSC(logger, start, "gen", params.TargetPoint)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "gcp", gcpOSC, params.Dummy.DummyPath) {
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
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/ncp [post]
func GenerateNCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genNCP", "Create dummy data and import to ncp objectstorage", start)

	params := models.GenarateTask{}
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

	params.Dummy.DummyPath = tmpDir
	if !dummyCreate(logger, start, params.Dummy) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	ncpOSC := getS3COSC(logger, start, "gen", params.TargetPoint)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	if !oscImport(logger, start, "ncp", ncpOSC, params.Dummy.DummyPath) {
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
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/mysql [post]
func GenerateMySQLPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genmysql", "Create dummy data and import to mysql", start)

	params := models.GenarateTask{}
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

	params.Dummy.DummyPath = tmpDir
	params.Dummy.CheckServerSQL = true
	params.Dummy.SizeServerSQL = "1"

	if !dummyCreate(logger, start, params.Dummy) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	rdbc := getMysqlRDBC(logger, start, "gen", params.TargetPoint)
	if rdbc == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	sqlList := []string{}
	if !walk(logger, start, &sqlList, params.Dummy.DummyPath, ".sql") {
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
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/dynamodb [post]
func GenerateDynamoDBPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "gendynamodb", "Create dummy data and import to dynamoDB", start)

	params := models.GenarateTask{}
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

	params.Dummy.DummyPath = tmpDir
	params.Dummy.CheckServerJSON = true
	params.Dummy.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params.Dummy) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.Dummy.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	awsNRDB := getDynamoNRDBC(logger, start, "gen", params.TargetPoint)
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
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask				true	"Parameters required to generate test data"
//	@Success		200				{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500				{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/firestore [post]
func GenerateFirestorePostHandler(ctx echo.Context) error {
	start := time.Now()
	pageName := "genfirestore"
	pageInfo := "Create dummy data and import to firestoreDB"
	logger, logstrings := pageLogInit(ctx, pageName, pageInfo, start)

	params := models.GenarateTask{}
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

	params.Dummy.DummyPath = tmpDir
	params.Dummy.CheckServerJSON = true
	params.Dummy.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params.Dummy) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.Dummy.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	gcpNRDB := getFirestoreNRDBC(logger, start, "gen", params.TargetPoint)
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
//	@Tags			[Generate]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		 models.GenarateTask			true	"Parameters required to generate test data"
//	@Success		200			{object}	models.BasicResponse	"Successfully generated test data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/generate/mongodb [post]
func GenerateMongoDBPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "genmongodb", "Create dummy data and import to mongoDB", start)

	params := models.GenarateTask{}
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

	params.Dummy.DummyPath = tmpDir
	params.Dummy.CheckServerJSON = true
	params.Dummy.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params.Dummy) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.Dummy.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	ncpNRDB := getMongoNRDBC(logger, start, "gen", params.TargetPoint)
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
