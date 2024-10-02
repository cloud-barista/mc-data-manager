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

// ScheduleController is a struct that holds a reference to the ScheduleService
type ScheduleController struct {
	ScheduleService *task.FileScheduleManager
}

// GetAllSchedulesHandler godoc
//
//	@ID 			GetAllSchedulesHandler
//	@Summary		Get all Schedules
//	@Description	Retrieve a list of all Schedules in the system.
//	@Tags			[Schedule]
//	@Produce		json
//	@Success		200		{array}		models.Schedule	"Successfully retrieved all Schedules"
//	@Failure		500		{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/schedule [get]
func (tc *ScheduleController) GetAllSchedulesHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Get-schedule-list", "Get an existing schedule", start)
	schedules, err := tc.ScheduleService.GetScheduleList()
	if err != nil {
		errStr := err.Error()
		logger.Error().Err(err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}
	jobEnd(logger, "Successfully Get Schedule List", start)
	return ctx.JSON(http.StatusOK, schedules)
}

// CreateScheduleHandler godoc
//
//	@ID 			CreateScheduleHandler
//	@Summary		Create a new Schedule
//	@Description	Create a new Schedule and store it in the system.
//	@Tags			[Schedule]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.Schedule	true	"Parameters required for creating a Schedule"
//	@Success		200			{object}	models.BasicResponse	"Successfully created a Schedule"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/schedule [post]
func (tc *ScheduleController) CreateScheduleHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Create-schedule", "Creating a new schedule", start)
	params := models.Schedule{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		errStr := "Invalid request data"
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: "",
			Error:  &errStr,
		})
	}
	logger.Info().Msg("=====Create Schedule======")
	if err := tc.ScheduleService.CreateSchedule(params); err != nil {
		errStr := err.Error()
		logger.Error().Err(err).Msg(errStr)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}
	jobEnd(logger, "Successfully Register Schedule", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// GetScheduleHandler godoc
//
//	@ID 			GetScheduleHandler
//	@Summary		Get a Schedule by ID
//	@Description	Get the details of a Schedule using its ID.
//	@Tags			[Schedule]
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Schedule ID"
//	@Success		200		{object}	models.Schedule	"Successfully retrieved a Schedule"
//	@Failure		404		{object}	models.BasicResponse	"Schedule not found"
//	@Router			/schedule/{id} [get]
func (tc *ScheduleController) GetScheduleHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Get-schedule", "Get an existing schedule", start)
	id := ctx.Param("id")
	schedule, err := tc.ScheduleService.GetSchedule(id)
	if err != nil {
		errStr := err.Error()
		logger.Error().Err(err)
		return ctx.JSON(http.StatusNotFound, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	return ctx.JSON(http.StatusOK, schedule)
}

// UpdateScheduleHandler godoc
//
//	@ID 			UpdateScheduleHandler
//	@Summary		Update an existing Schedule
//	@Description	Update the details of an existing Schedule using its ID.
//	@Tags			[Schedule]
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Schedule ID"
//	@Param			RequestBody	body	models.Schedule	true	"Parameters required for updating a Schedule"
//	@Success		200			{object}	models.BasicResponse	"Successfully updated the Schedule"
//	@Failure		404			{object}	models.BasicResponse	"Schedule not found"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/schedule/{id} [put]
func (tc *ScheduleController) UpdateScheduleHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Update-schedule", "Updating an existing schedule", start)
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

	if err := tc.ScheduleService.UpdateSchedule(id, params); err != nil {
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

// DeleteScheduleHandler godoc
//
//	@ID 			DeleteScheduleHandler
//	@Summary		Delete a Schedule
//	@Description	Delete an existing Schedule using its ID.
//	@Tags			[Schedule]
//	@Produce		json
//	@Param			id		path	string	true	"Schedule ID"
//	@Success		200		{object}	models.BasicResponse	"Successfully deleted the Schedule"
//	@Failure		404		{object}	models.BasicResponse	"Schedule not found"
//	@Router			/schedule/{id} [delete]
func (tc *ScheduleController) DeleteScheduleHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Delete-schedule", "Delete an existing schedule", start)
	id := ctx.Param("id")
	if err := tc.ScheduleService.DeleteSchedule(id); err != nil {
		errStr := "Schedule not found"
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
