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
	"github.com/labstack/echo/v4"
)

// DeleteServiceAndTaskAllHandler godoc
//
//	@ID 			DeleteServiceAndTaskAllHandler
//	@Summary		Delete a Task
//	@Description	Delete an All Service and Task.
//	@Tags			[Service]
//	@Produce		json
//	@Success		200		{object}	models.BasicResponse	"Successfully deleted the All Service"
//	@Failure		404		{object}	models.BasicResponse	"Clear All Task , Failed"
//	@Router			/service/clearAll [delete]
func (tc *TaskController) DeleteServiceAndTaskAllHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Delete-task", "Delete an existing task", start)
	if err := tc.TaskService.ClearServiceAndTaskAll(); err != nil {
		errStr := "Clear All Task , Failed"
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
