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
	"github.com/labstack/echo/v4"
)

// BackupOSPostHandler godoc
//
//	@ID 			BackupOSPostHandler
//	@Summary		Export data from objectstorage
//	@Description	Export data from a objectstorage  to files.
//	@Tags			[Backup]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.BackupTask	true	"Parameters required for backup"
//	@Success		200			{object}	models.BasicResponse	"Successfully backup data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/backup/objectstorage [post]
func BackupOSPostHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Backup", "Backup Objectstorage", start)

	var req models.BackupObjectStorageRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	missing := map[string]string{
		"sourcePoint.provider":     req.SourcePoint.Provider,
		"sourcePoint.region":       req.SourcePoint.Region,
		"sourcePoint.instanceId":   req.SourcePoint.InstanceId,
		"sourcePoint.instanceName": req.SourcePoint.InstanceName,
		"sourcePoint.bucket":       req.SourcePoint.Bucket,
		"targetPoint.backupName":   req.TargetPoint.BackupName,
	}
	for field, val := range missing {
		if val == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": field + " is required"})
		}
	}
	if req.SourcePoint.Provider == "ncp" && req.SourcePoint.Endpoint == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "sourcePoint.endpoint is required"})
	}

	record, err := backup.CreateBackup("objectstorage", req.SourcePoint.BackupSourceCommon, req.TargetPoint.BackupName, "")
	if err != nil {
		errStr := err.Error()
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: &errStr})
	}

	var params models.DataTask
	params.TaskMeta.TaskID = record.ID
	params.TaskMeta.TaskType = models.Backup
	params.TaskMeta.ServiceType = models.ObejectStorage
	params.SourcePoint.Provider = req.SourcePoint.Provider
	params.SourcePoint.Region = req.SourcePoint.Region
	params.SourcePoint.ObjectStorageParams = req.SourcePoint.ObjectStorageParams
	params.TargetPoint.Path = record.Path
	params.SourceFilter = req.SourceFilter

	manager := task.GetFileScheduleManager()
	success := manager.RunTaskOnce(params, traceIDFromCtx(ctx))
	status, err := backup.MarkStatus(record.ID, success)
	if err != nil {
		errStr := err.Error()
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: &errStr})
	}
	record.Status = status
	if !success {
		errStr := taskErrMsg(ctx, "backup failed")
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: &errStr})
	}

	jobEnd(logger, "Successfully Backup data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
}

// BackupRDBPostHandler godoc
//
//	@ID 			BackupRDBPostHandler
//	@Summary		Export data from MySQL
//	@Description	Export data from a MySQL database to SQL files.
//	@Tags			[Backup]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.BackupTask	true	"Parameters required for backup"
//	@Success		200			{object}	models.BasicResponse	"Successfully backup data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/backup/rdbms [post]
func BackupRDBPostHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Backup", "Backup RDBMS", start)

	var req models.BackupRDBRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	missing := map[string]string{
		"sourcePoint.provider":     req.SourcePoint.Provider,
		"sourcePoint.region":       req.SourcePoint.Region,
		"sourcePoint.instanceId":   req.SourcePoint.InstanceId,
		"sourcePoint.instanceName": req.SourcePoint.InstanceName,
		"sourcePoint.host":         req.SourcePoint.Host,
		"sourcePoint.port":         req.SourcePoint.Port,
		"sourcePoint.username":     req.SourcePoint.User,
		"sourcePoint.password":     req.SourcePoint.Password,
		"sourcePoint.databaseName": req.SourcePoint.DatabaseName,
		"targetPoint.backupName":   req.TargetPoint.BackupName,
	}
	for field, val := range missing {
		if val == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": field + " is required"})
		}
	}

	record, err := backup.CreateBackup("rdbms", req.SourcePoint.BackupSourceCommon, req.TargetPoint.BackupName, req.SourcePoint.DatabaseName)
	if err != nil {
		errStr := err.Error()
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: &errStr})
	}

	var params models.DataTask
	params.TaskMeta.TaskID = record.ID
	params.TaskMeta.TaskType = models.Backup
	params.TaskMeta.ServiceType = models.RDBMS
	params.SourcePoint.Provider = req.SourcePoint.Provider
	params.SourcePoint.Region = req.SourcePoint.Region
	params.SourcePoint.MySQLParams = req.SourcePoint.MySQLParams
	params.TargetPoint.Path = record.Path

	manager := task.GetFileScheduleManager()
	success := manager.RunTaskOnce(params, traceIDFromCtx(ctx))
	status, err := backup.MarkStatus(record.ID, success)
	if err != nil {
		errStr := err.Error()
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: &errStr})
	}
	record.Status = status
	if !success {
		errStr := taskErrMsg(ctx, "backup failed")
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: &errStr})
	}

	jobEnd(logger, "Successfully Backup data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
}

// BackupNRDBPostHandler godoc
//
//	@ID 			BackupNRDBPostHandler
//	@Summary		Export data from MySQL
//	@Description	Export data from a MySQL database to SQL files.
//	@Tags			[Backup]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.BackupTask	true	"Parameters required for backup"
//	@Success		200			{object}	models.BasicResponse	"Successfully backup data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/backup/nrdbms [post]
func BackupNRDBPostHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Backup", "Backup NRDBMS", start)

	var req models.BackupNRDBRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.SourcePoint.Provider == "" || req.SourcePoint.Region == "" || req.TargetPoint.BackupName == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "provider, region, backupName are required"})
	}
	if req.SourcePoint.Provider != "aws" && (req.SourcePoint.InstanceId == "" || req.SourcePoint.InstanceName == "") {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "instanceId, instanceName"})
	}

	switch req.SourcePoint.Provider {
	case "aws":
	case "gcp":
		if req.SourcePoint.DatabaseID == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "databaseId are required"})
		}
	case "ncp", "alibaba":
		if req.SourcePoint.Host == "" || req.SourcePoint.Port == "" || req.SourcePoint.User == "" || req.SourcePoint.Password == "" || req.SourcePoint.DatabaseName == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "host, port, username, password, databaseName are required"})
		}
	default:
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "unsupported provider: " + req.SourcePoint.Provider})
	}

	record, err := backup.CreateBackup("nrdbms", req.SourcePoint.BackupSourceCommon, req.TargetPoint.BackupName, req.SourcePoint.DatabaseName)
	if err != nil {
		errStr := err.Error()
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: &errStr})
	}

	var params models.DataTask
	params.TaskMeta.TaskID = record.ID
	params.TaskMeta.TaskType = models.Backup
	params.TaskMeta.ServiceType = models.NRDBMS
	params.SourcePoint.Provider = req.SourcePoint.Provider
	params.SourcePoint.Region = req.SourcePoint.Region
	params.SourcePoint.MySQLParams = req.SourcePoint.MySQLParams
	params.SourcePoint.DatabaseID = req.SourcePoint.DatabaseID
	params.TargetPoint.Path = record.Path

	manager := task.GetFileScheduleManager()
	success := manager.RunTaskOnce(params, traceIDFromCtx(ctx))
	status, err := backup.MarkStatus(record.ID, success)
	if err != nil {
		errStr := err.Error()
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: &errStr})
	}
	record.Status = status
	if !success {
		errStr := taskErrMsg(ctx, "backup failed")
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: &errStr})
	}

	jobEnd(logger, "Successfully Backup data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
}

// GetAllBackupHandler godoc
//
//	@ID 			GetAllBackupHandler
//	@Summary		Get all Tasks
//	@Description	Retrieve a list of all Tasks in the system.
//	@Tags			[Backup]
//	@Produce		json
//	@Success		200		{array}		models.Task	"Successfully retrieved all Tasks"
//	@Failure		500		{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/backup [get]
func GetAllBackupHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Get-task-list", "Get an existing task", start)
	manager := task.GetFileScheduleManager()
	tasks, err := manager.GetTasksByTypeList(models.Backup)
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

// GetBackupHandler godoc
//
//	@ID 			GetBackupHandler
//	@Summary		Get a Task by ID
//	@Description	Get the details of a Task using its ID.
//	@Tags			[Backup]
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Task ID"
//	@Success		200		{object}	models.Task	"Successfully retrieved a Task"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/backup/{id} [get]
func GetBackupHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Get-task", "Get an existing task", start)
	id := ctx.Param("id")
	manager := task.GetFileScheduleManager()
	task, err := manager.GetTasksByType(models.Backup, id)
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

// UpdateBackupHandler godoc
//
//	@ID 			UpdateBackupHandler
//	@Summary		Update an existing Task
//	@Description	Update the details of an existing Task using its ID.
//	@Tags			[Backup]
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Task ID"
//	@Param			RequestBody	body	models.Schedule	true	"Parameters required for updating a Task"
//	@Success		200			{object}	models.BasicResponse	"Successfully updated the Task"
//	@Failure		404			{object}	models.BasicResponse	"Task not found"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/backup/{id} [put]
func UpdateBackupHandler(ctx echo.Context) error {
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
	if err := manager.UpdateTasksByType(models.Backup, id, params.BasicDataTask); err != nil {
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

// DeleteBackupkHandler godoc
//
//	@ID 			DeleteBackupkHandler
//	@Summary		Delete a Task
//	@Description	Delete an existing Task using its ID.
//	@Tags			[Backup]
//	@Produce		json
//	@Param			id		path	string	true	"Task ID"
//	@Success		200		{object}	models.BasicResponse	"Successfully deleted the Task"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/backup/{id} [delete]
func DeleteBackupkHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Delete-task", "Delete an existing task", start)
	id := ctx.Param("id")
	manager := task.GetFileScheduleManager()
	if err := manager.DeleteTasksByType(models.Backup, id); err != nil {
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

// ListBackupObjectStorageHandler godoc
//
//	@ID 			ListBackupObjectStorageHandler
//	@Summary		List Object Storage backups
//	@Description	Returns the current namespace's Object Storage backup catalog entries.
//	@Tags			[Backup]
//	@Produce		json
//	@Success		200	{array}		models.BackupRecord	"Backup catalog entries"
//	@Failure		500	{object}	map[string]string		"Internal Server Error"
//	@Router			/backup/objectstorage [get]
func ListBackupObjectStorageHandler(ctx echo.Context) error {
	backups, err := backup.ListBackups(string(models.ObejectStorage))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, backups)
}

// ListBackupRDBHandler godoc
//
//	@ID 			ListBackupRDBHandler
//	@Summary		List RDB backups
//	@Description	Returns the current namespace's RDB backup catalog entries.
//	@Tags			[Backup]
//	@Produce		json
//	@Success		200	{array}		models.BackupRecord	"Backup catalog entries"
//	@Failure		500	{object}	map[string]string		"Internal Server Error"
//	@Router			/backup/rdbms [get]
func ListBackupRDBHandler(ctx echo.Context) error {
	backups, err := backup.ListBackups(string(models.RDBMS))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, backups)
}

// ListBackupNRDBHandler godoc
//
//	@ID 			ListBackupNRDBHandler
//	@Summary		List NRDB backups
//	@Description	Returns the current namespace's NRDB backup catalog entries.
//	@Tags			[Backup]
//	@Produce		json
//	@Success		200	{array}		models.BackupRecord	"Backup catalog entries"
//	@Failure		500	{object}	map[string]string		"Internal Server Error"
//	@Router			/backup/nrdbms [get]
func ListBackupNRDBHandler(ctx echo.Context) error {
	backups, err := backup.ListBackups(string(models.NRDBMS))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, backups)
}

// DeleteBackupHandler godoc
//
//	@ID 			DeleteBackupHandler
//	@Summary		Delete a backup
//	@Description	Deletes a backup catalog entry and its backed-up files on disk.
//	@Tags			[Backup]
//	@Produce		json
//	@Param			id	path	string	true	"Backup ID"
//	@Success		200	{object}	map[string]string	"Backup deleted"
//	@Failure		500	{object}	map[string]string	"Internal Server Error"
//	@Router			/backup/{id} [delete]
func DeleteBackupHandler(ctx echo.Context) error {
	id := ctx.Param("id")
	if err := backup.DeleteBackup(id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, map[string]string{"result": "backup deleted"})
}
