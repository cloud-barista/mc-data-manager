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

// GetSystemReadyHandler godoc
//
//	@ID 			GetSystemReadyHandler
//	@Summary		Get System Ready Handler
//	@Description	Get System Ready
//	@Tags			[Already Check System]
//	@Produce		json
//	@Success		200		{object}	models.BasicResponse	"System is Ready"
//	@Failure		404		{object}	models.BasicResponse	"Profile Load , Failed: err"
//	@Router			/readyZ [Get]
func GetSystemReadyHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "healthcheck-task", "Ready?", start)

	// TODO - db 헬스체크 추가 예정

	// credentailManger := config.NewProfileManager()
	// err := credentailManger.ValidateProfiles()
	// if err != nil {
	// 	errStr := "Profile Load , Failed : " + err.Error()
	// 	logger.Error().Msg(errStr)
	// 	return ctx.JSON(http.StatusNotFound, models.BasicResponse{
	// 		Result: logstrings.String(),
	// 		Error:  &errStr,
	// 	})
	// }
	jobEnd(logger, "System is Ready", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}
