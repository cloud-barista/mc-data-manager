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

// TaskController is a struct that holds a reference to the TaskService
type TaskController struct {
	TaskService *task.FileScheduleManager
}

// GetAllTasksHandler godoc
//
//	@ID 			GetAllTasksHandler
//	@Summary		Get all Tasks
//	@Description	Retrieve a list of all Tasks in the system.
//	@Tags			[Task]
//	@Produce		json
//	@Success		200		{array}		models.Task	"Successfully retrieved all Tasks"
//	@Failure		500		{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/task [get]
func (tc *TaskController) GetAllTasksHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit("Get-task-list", "Get an existing task", start)
	tasks, err := tc.TaskService.GetScheduleList()
	if err != nil {
		errStr := err.Error()
		logger.Error().Err(err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	return ctx.JSON(http.StatusOK, tasks)
}

// CreateTaskHandler godoc
//
//	@ID 			CreateTaskHandler
//	@Summary		Create a new Task
//	@Description	Create a new Task and store it in the system.
//	@Tags			[Task]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.Schedule	true	"Parameters required for creating a Task"
//	@Success		200			{object}	models.BasicResponse	"Successfully created a Task"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/task [post]
func (tc *TaskController) CreateTaskHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit("Create-task", "Creating a new task", start)
	params := models.Schedule{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		errStr := "Invalid request data"
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}
	if err := tc.TaskService.CreateSchedule(params); err != nil {
		errStr := err.Error()
		logger.Error().Err(err)
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

// GetTaskHandler godoc
//
//	@ID 			GetTaskHandler
//	@Summary		Get a Task by ID
//	@Description	Get the details of a Task using its ID.
//	@Tags			[Task]
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Task ID"
//	@Success		200		{object}	models.Task	"Successfully retrieved a Task"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/task/{id} [get]
func (tc *TaskController) GetTaskHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit("Get-task", "Get an existing task", start)
	id := ctx.Param("id")
	task, err := tc.TaskService.GetSchedule(id)
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

// UpdateTaskHandler godoc
//
//	@ID 			UpdateTaskHandler
//	@Summary		Update an existing Task
//	@Description	Update the details of an existing Task using its ID.
//	@Tags			[Task]
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Task ID"
//	@Param			RequestBody	body	models.Schedule	true	"Parameters required for updating a Task"
//	@Success		200			{object}	models.BasicResponse	"Successfully updated the Task"
//	@Failure		404			{object}	models.BasicResponse	"Task not found"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/task/{id} [put]
func (tc *TaskController) UpdateTaskHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit("Update-task", "Updating an existing task", start)
	id := ctx.Param("id")
	params := models.Schedule{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		errStr := "Invalid request data"
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	if err := tc.TaskService.UpdateSchedule(id, params); err != nil {
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

// DeleteTaskHandler godoc
//
//	@ID 			DeleteTaskHandler
//	@Summary		Delete a Task
//	@Description	Delete an existing Task using its ID.
//	@Tags			[Task]
//	@Produce		json
//	@Param			id		path	string	true	"Task ID"
//	@Success		200		{object}	models.BasicResponse	"Successfully deleted the Task"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/task/{id} [delete]
func (tc *TaskController) DeleteTaskHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit("Delete-task", "Delete an existing task", start)
	id := ctx.Param("id")
	if err := tc.TaskService.DeleteSchedule(id); err != nil {
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
