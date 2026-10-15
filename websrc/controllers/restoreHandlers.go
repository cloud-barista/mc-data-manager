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
	"github.com/cloud-barista/mc-data-manager/service/backup"
	"github.com/cloud-barista/mc-data-manager/service/task"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RestoreOSPostHandler godoc
//
//	@ID 			RestoreOSPostHandler
//	@Summary		Restore data from objectstorage
//	@Description	Restore objectstorage from files to a objectstorage
//	@Tags			[Restore]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.RestoreObjectStorageRequest	true	"Parameters required for Restore"
//	@Success		200			{object}	models.BasicResponse	"Successfully Restore data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/restore/objectstorage [post]
func RestoreOSPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "Restore", "Restore objectstorage", start)

	var req models.RestoreObjectStorageRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	missing := map[string]string{
		"sourcePoint.backupId": req.SourcePoint.BackupId,
		"targetPoint.provider": req.TargetPoint.Provider,
		"targetPoint.region":   req.TargetPoint.Region,
		"targetPoint.bucket":   req.TargetPoint.Bucket,
	}
	for field, val := range missing {
		if val == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": field + " is required"})
		}
	}
	if req.TargetPoint.Provider == "ncp" && req.TargetPoint.Endpoint == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "targetPoint.endpoint is required"})
	}

	record, err := backup.GetRestorableBackup(req.SourcePoint.BackupId, "objectstorage")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var params models.DataTask
	params.TaskMeta.TaskID = uuid.New().String()
	params.TaskMeta.TaskType = models.Restore
	params.TaskMeta.ServiceType = models.ObejectStorage
	params.TargetPoint.Provider = req.TargetPoint.Provider
	params.TargetPoint.Region = req.TargetPoint.Region
	params.TargetPoint.ObjectStorageParams = req.TargetPoint.ObjectStorageParams
	params.SourcePoint.Path = record.Path

	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(params, traceIDFromCtx(ctx)) {
		errStr := taskErrMsg(ctx, "task failed")
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}
	// backup success. Send result to client
	jobEnd(logger, "Successfully Restore data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// RestoreRDBPostHandler godoc
//
//	@ID 			RestoreRDBPostHandler
//	@Summary		Restore data from MySQL
//	@Description	Restore MySQL from MySQL files to a MySQL database
//	@Tags			[Restore]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.RestoreTask	true	"Parameters required for Restore"
//	@Success		200			{object}	models.BasicResponse	"Successfully Restore data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/restore/rdbms [post]
func RestoreRDBPostHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Restore", "Restore RDBMS", start)

	var req models.RestoreRDBRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	missing := map[string]string{
		"sourcePoint.backupId": req.SourcePoint.BackupId,
		"targetPoint.provider": req.TargetPoint.Provider,
		"targetPoint.region":   req.TargetPoint.Region,
		"targetPoint.host":     req.TargetPoint.Host,
		"targetPoint.port":     req.TargetPoint.Port,
		"targetPoint.username": req.TargetPoint.User,
		"targetPoint.password": req.TargetPoint.Password,
	}
	for field, val := range missing {
		if val == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": field + " is required"})
		}
	}

	record, err := backup.GetRestorableBackup(req.SourcePoint.BackupId, "rdbms")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var params models.DataTask
	params.TaskMeta.TaskID = uuid.New().String()
	params.TaskMeta.TaskType = models.Restore
	params.TaskMeta.ServiceType = models.RDBMS
	params.TargetPoint.Provider = req.TargetPoint.Provider
	params.TargetPoint.Region = req.TargetPoint.Region
	params.TargetPoint.MySQLParams = req.TargetPoint.MySQLParams
	// restored DB name for the execution log (informational; restore replays the dump).
	if record.DatabaseName != nil {
		params.TargetPoint.DatabaseName = *record.DatabaseName
	}
	params.SourcePoint.Path = record.Path

	manager := task.GetFileScheduleManager()
	if !manager.RunTaskOnce(params, traceIDFromCtx(ctx)) {
		errStr := taskErrMsg(ctx, "task failed")
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	jobEnd(logger, "Successfully Restore data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// RestoreNRDBPostHandler godoc
//
//	@ID 			RestoreNRDBPostHandler
//	@Summary		Restore NoSQL from data to NoSQL
//	@Description	Restore NoSQL from SQL files to a NoSQL database
//	@Tags			[Restore]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.RestoreTask	true	"Parameters required for Restore"
//	@Success		200			{object}	models.BasicResponse	"Successfully Restore data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/restore/nrdbms [post]
func RestoreNRDBPostHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Restore", "Restore NRDBMS", start)

	var req models.RestoreNRDBRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.SourcePoint.BackupId == "" || req.TargetPoint.Provider == "" || req.TargetPoint.Region == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "sourcePoint.backupId, targetPoint.provider, targetPoint.region are required"})
	}

	switch req.TargetPoint.Provider {
	case "aws":
	case "gcp":
		if req.TargetPoint.DatabaseID == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "targetPoint.databaseId is required"})
		}
	case "ncp", "alibaba":
		if req.TargetPoint.Host == "" || req.TargetPoint.Port == "" || req.TargetPoint.User == "" || req.TargetPoint.Password == "" || req.TargetPoint.DatabaseName == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "targetPoint.host, targetPoint.port, targetPoint.username, targetPoint.password, targetPoint.databaseName are required"})
		}
	default:
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "unsupported provider: " + req.TargetPoint.Provider})
	}

	record, err := backup.GetRestorableBackup(req.SourcePoint.BackupId, "nrdbms")
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var params models.DataTask
	params.TaskMeta.TaskID = uuid.New().String()
	params.TaskMeta.TaskType = models.Restore
	params.TaskMeta.ServiceType = models.NRDBMS
	params.TargetPoint.Provider = req.TargetPoint.Provider
	params.TargetPoint.Region = req.TargetPoint.Region
	params.TargetPoint.MySQLParams = req.TargetPoint.MySQLParams
	params.TargetPoint.DatabaseID = req.TargetPoint.DatabaseID
	params.SourcePoint.Path = record.Path

	manager := task.GetFileScheduleManager()
	if !manager.RunTaskOnce(params, traceIDFromCtx(ctx)) {
		errStr := taskErrMsg(ctx, "task failed")
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	jobEnd(logger, "Successfully Restore data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GetAllRestoreHandler godoc
//
//	@ID 			GetAllRestoreHandler
//	@Summary		Get all Tasks
//	@Description	Retrieve a list of all Tasks in the system.
//	@Tags			[Restore]
//	@Produce		json
//	@Success		200		{array}		models.Task	"Successfully retrieved all Tasks"
//	@Failure		500		{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/restore [get]
func GetAllRestoreHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Get-task-list", "Get an existing task", start)
	manager := task.GetFileScheduleManager()
	tasks, err := manager.GetTasksByTypeList(models.Restore)
	if err != nil {
		errStr := err.Error()
		logger.Error().Err(err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}
	jobEnd(logger, "Successfully Get Task List", start)
	return ctx.JSON(http.StatusOK, tasks)
}

// GetRestoreHandler godoc
//
//	@ID 			GetRestoreHandler
//	@Summary		Get a Task by ID
//	@Description	Get the details of a Task using its ID.
//	@Tags			[Restore]
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Task ID"
//	@Success		200		{object}	models.Task	"Successfully retrieved a Task"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/restore/{id} [get]
func GetRestoreHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Get-task", "Get an existing task", start)
	id := ctx.Param("id")
	manager := task.GetFileScheduleManager()
	task, err := manager.GetTasksByType(models.Restore, id)
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

// UpdateRestoreHandler godoc
//
//	@ID 			UpdateRestoreHandler
//	@Summary		Update an existing Task
//	@Description	Update the details of an existing Task using its ID.
//	@Tags			[Restore]
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Task ID"
//	@Param			RequestBody	body	models.Schedule	true	"Parameters required for updating a Task"
//	@Success		200			{object}	models.BasicResponse	"Successfully updated the Task"
//	@Failure		404			{object}	models.BasicResponse	"Task not found"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/restore/{id} [put]
func UpdateRestoreHandler(ctx echo.Context) error {
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
	if err := manager.UpdateTasksByType(models.Restore, id, params.BasicDataTask); err != nil {
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

// DeleteRestorekHandler godoc
//
//	@ID 			DeleteRestorekHandler
//	@Summary		Delete a Task
//	@Description	Delete an existing Task using its ID.
//	@Tags			[Restore]
//	@Produce		json
//	@Param			id		path	string	true	"Task ID"
//	@Success		200		{object}	models.BasicResponse	"Successfully deleted the Task"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/restore/{id} [delete]
func DeleteRestorekHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Delete-task", "Delete an existing task", start)
	id := ctx.Param("id")
	manager := task.GetFileScheduleManager()
	if err := manager.DeleteTasksByType(models.Restore, id); err != nil {
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
