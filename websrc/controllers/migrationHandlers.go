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
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/service/task"
	"github.com/labstack/echo/v4"
)

// MigrationObjectstoragePostHandler godoc
//
//	@ID 			MigrationObjectstoragePostHandler
//	@Summary		Migrate data from ObjectStorage to ObjectStorage
//	@Description	Migrate data from ObjectStorage to ObjectStorage.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/objectstorage [post]
func MigrationObjectstoragePostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "migration", "Migration linux objectstorage to objectstorage", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	params.TaskMeta.TaskID = params.OperationId
	params.TaskMeta.TaskType = models.Migrate
	params.TaskMeta.ServiceType = models.ObejectStorage
	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationObjectstoragePostHandler godoc
//
//	@ID 			MigrationObjectstoragePostHandler
//	@Summary		Migrate data from ObjectStorage to ObjectStorage
//	@Description	Migrate data from ObjectStorage to ObjectStorage.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/nrdbms [post]
func MigrationNRDBMSPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "migration", "Migration linux NRDBMS to NRDBMS", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	params.TaskMeta.TaskID = params.OperationId
	params.TaskMeta.TaskType = models.Migrate
	params.TaskMeta.ServiceType = models.NRDBMS
	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationObjectstoragePostHandler godoc
//
//	@ID 			MigrationObjectstoragePostHandler
//	@Summary		Migrate data from ObjectStorage to ObjectStorage
//	@Description	Migrate data from ObjectStorage to ObjectStorage.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/rdbms [post]
func MigrationRDBMSPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "migration", "Migration linux NRDBMS to NRDBMS", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	params.TaskMeta.TaskID = params.OperationId
	params.TaskMeta.TaskType = models.Migrate
	params.TaskMeta.ServiceType = models.RDBMS
	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GetAllMigrateHandler godoc
//
//	@ID 			GetAllMigrateHandler
//	@Summary		Get all Tasks
//	@Description	Retrieve a list of all Tasks in the system.
//	@Tags			[Migrate]
//	@Produce		json
//	@Success		200		{array}		models.Task	"Successfully retrieved all Tasks"
//	@Failure		500		{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migrate [get]
func GetAllMigrateHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Get-task-list", "Get an existing task", start)
	manager := task.GetFileScheduleManager()
	tasks, err := manager.GetTasksByTypeList(models.Migrate)
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

// GetMigrateHandler godoc
//
//	@ID 			GetMigrateHandler
//	@Summary		Get a Task by ID
//	@Description	Get the details of a Task using its ID.
//	@Tags			[Migrate]
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Task ID"
//	@Success		200		{object}	models.Task	"Successfully retrieved a Task"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/migrate/{id} [get]
func GetMigrateHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Get-task", "Get an existing task", start)
	id := ctx.Param("id")
	manager := task.GetFileScheduleManager()
	task, err := manager.GetTasksByType(models.Migrate, id)
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

// UpdateMigrateHandler godoc
//
//	@ID 			UpdateMigrateHandler
//	@Summary		Update an existing Task
//	@Description	Update the details of an existing Task using its ID.
//	@Tags			[Migrate]
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Task ID"
//	@Param			RequestBody	body	models.Schedule	true	"Parameters required for updating a Task"
//	@Success		200			{object}	models.BasicResponse	"Successfully updated the Task"
//	@Failure		404			{object}	models.BasicResponse	"Task not found"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migrate/{id} [put]
func UpdateMigrateHandler(ctx echo.Context) error {
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
	if err := manager.UpdateTasksByType(models.Migrate, id, params.BasicDataTask); err != nil {
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

// DeleteMigratekHandler godoc
//
//	@ID 			DeleteMigratekHandler
//	@Summary		Delete a Task
//	@Description	Delete an existing Task using its ID.
//	@Tags			[Migrate]
//	@Produce		json
//	@Param			id		path	string	true	"Task ID"
//	@Success		200		{object}	models.BasicResponse	"Successfully deleted the Task"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/migrate/{id} [delete]
func DeleteMigratekHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Delete-task", "Delete an existing task", start)
	id := ctx.Param("id")
	manager := task.GetFileScheduleManager()
	if err := manager.DeleteTasksByType(models.Migrate, id); err != nil {
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

// MigrationLinuxToS3PostHandler godoc
//
//	@ID 			MigrationLinuxToS3PostHandler
//	@Summary		Migrate data from Linux to AWS S3
//	@Description	Migrate data stored in a Linux-based system to AWS S3.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/linux/aws [post]
func MigrationLinuxToS3PostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "miglins3", "Import linux data to s3", start)

	if !osCheck(logger, start, "linux") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	params := MigrateTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	awsOSC := getS3OSC(logger, start, "mig", params.TargetPoint)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !oscImport(logger, start, "s3", awsOSC, params.SourcePoint.Path) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Linux to s3", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationLinuxToGCPPostHandler godoc
//
//	@ID 			MigrationLinuxToGCPPostHandler
//	@Summary		Migrate data from Linux to GCP Cloud Storage
//	@Description	Migrate data stored in a Linux-based system to GCP Cloud Storage.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/linux/gcp [post]
func MigrationLinuxToGCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "miglingcp", "Import linux data to gcp", start)

	if !osCheck(logger, start, "linux") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	params := MigrateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	gcpOSC := getGCPCOSC(logger, start, "mig", params.TargetPoint)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "gcp", gcpOSC, params.SourcePoint.Path) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Linux to gcp", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationLinuxToNCPPostHandler godoc
//
//	@ID 			MigrationLinuxToNCPPostHandler
//	@Summary		Migrate data from Linux to NCP Object Storage
//	@Description	Migrate data stored in a Linux-based system to NCP Object Storage.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/linux/ncp [post]
func MigrationLinuxToNCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "miglinncp", "Import linux data to ncp objectstorage", start)

	if !osCheck(logger, start, "linux") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	params := MigrateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	ncpOSC := getS3COSC(logger, start, "mig", params.TargetPoint)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !oscImport(logger, start, "ncp", ncpOSC, params.SourcePoint.Path) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Linux to ncp objectstorage", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationWindowsToS3PostHandler godoc
//
//	@ID 			MigrationWindowsToS3PostHandler
//	@Summary		Migrate data from Windows to AWS S3
//	@Description	Migrate data stored in a Windows-based system to AWS S3.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/windows/aws [post]
func MigrationWindowsToS3PostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "migwins3", "Import windows data to s3", start)

	if !osCheck(logger, start, "windows") {
		return ctx.JSON(http.StatusOK, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	params := MigrateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	awsOSC := getS3OSC(logger, start, "mig", params.TargetPoint)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "s3", awsOSC, params.SourcePoint.Path) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Windows to s3", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationWindowsToGCPPostHandler godoc
//
//	@ID 			MigrationWindowsToGCPPostHandler
//	@Summary		Migrate data from Windows to GCP Cloud Storage
//	@Description	Migrate data stored in a Windows-based system to GCP Cloud Storage.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/windows/gcp [post]
func MigrationWindowsToGCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "migwingcp", "Import windows data to gcp", start)

	if !osCheck(logger, start, "windows") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	params := MigrateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	gcpOSC := getGCPCOSC(logger, start, "mig", params.TargetPoint)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !oscImport(logger, start, "gcp", gcpOSC, params.SourcePoint.Path) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Windows to gcp", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationWindowsToNCPPostHandler godoc
//
//	@ID 			MigrationWindowsToNCPPostHandler
//	@Summary		Migrate data from Windows to NCP Object Storage
//	@Description	Migrate data stored in a Windows-based system to NCP Object Storage.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/windows/ncp [post]
func MigrationWindowsToNCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "migwinncp", "Import linux data to ncp objectstorage", start)

	if !osCheck(logger, start, "windows") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	params := MigrateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusOK, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	ncpOSC := getS3COSC(logger, start, "mig", params.TargetPoint)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !oscImport(logger, start, "ncp", ncpOSC, params.SourcePoint.Path) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Windows to ncp objectstorage", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationMySQLPostHandler godoc
//
//	@ID 			MigrationMySQLPostHandler
//	@Summary		Migrate data from MySQL to MySQL
//	@Description	Migrate data from one MySQL database to another MySQL database.
//	@Tags			[Data Migration], [Service RDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/mysql [post]
func MigrationMySQLPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "migmysql", "Import mysql to mysql", start)

	params := MigrateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusOK, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	srdbc := getMysqlRDBC(logger, start, "smig", params.SourcePoint)
	if srdbc == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	trdbc := getMysqlRDBC(logger, start, "tmig", params.TargetPoint)
	if trdbc == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if err := srdbc.Copy(trdbc); err != nil {
		return ctx.JSON(http.StatusOK, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// mysql migration success result send to client
	jobEnd(logger, "Successfully migrated data from mysql to mysql", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}
