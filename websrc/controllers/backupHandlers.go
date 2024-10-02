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

	logger, logstrings := pageLogInit(ctx, "Bakcup", "Bakcup linux objectstorage to objectstorage", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	params.TaskMeta.TaskID = params.OperationId
	params.TaskMeta.TaskType = models.Backup
	params.TaskMeta.ServiceType = models.ObejectStorage
	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	// backup success. Send result to client
	jobEnd(logger, "Successfully Bakcup data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
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

	logger, logstrings := pageLogInit(ctx, "Bakcup", "Bakcup linux RDBMS to RDBMS", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	params.TaskMeta.TaskID = params.OperationId
	params.TaskMeta.TaskType = models.Backup
	params.TaskMeta.ServiceType = models.RDBMS
	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	// backup success. Send result to client
	jobEnd(logger, "Successfully Bakcup data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
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

	logger, logstrings := pageLogInit(ctx, "Bakcup", "Bakcup linux NRDBMS to NRDBMS", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	params.TaskMeta.TaskID = params.OperationId
	params.TaskMeta.TaskType = models.Backup
	params.TaskMeta.ServiceType = models.NRDBMS
	manager := task.GetFileScheduleManager()

	if !manager.RunTaskOnce(params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	// backup success. Send result to client
	jobEnd(logger, "Successfully Bakcup data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
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
