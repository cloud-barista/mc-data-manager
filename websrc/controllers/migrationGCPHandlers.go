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

// MigrationGCPToLinuxPostHandler godoc
//
//	@Summary		Migrate data from GCP to Linux
//	@Description	Migrate data stored in GCP Cloud Storage to a Linux-based system.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/gcp/linux [post]
func MigrationGCPToLinuxPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("miggcplin", "Export gcp data to windows", start)

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

	gcpOSC := getGCPCOSC(logger, start, "mig", params.SourcePoint)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !oscExport(logger, start, "gcp", gcpOSC, params.TargetPoint.Path) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from gcp to linux", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationGCPToWindowsPostHandler godoc
//
//	@Summary		Migrate data from GCP to Windows
//	@Description	Migrate data stored in GCP Cloud Storage to a Windows-based system.
//	@Tags			[Data Migration]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/gcp/windows [post]
func MigrationGCPToWindowsPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("miggcpwin", "Export gcp data to windows", start)

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

	gcpOSC := getGCPCOSC(logger, start, "mig", params.SourcePoint)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !oscExport(logger, start, "gcp", gcpOSC, params.TargetPoint.Path) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from gcp to windows", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationGCPToS3PostHandler godoc
//
//	@Summary		Migrate data from GCP to AWS S3
//	@Description	Migrate data stored in GCP Cloud Storage to AWS S3.
//	@Tags			[Data Migration], [Object Storage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/gcp/aws [post]
func MigrationGCPToS3PostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("genlinux", "Export gcp data to s3", start)

	params := MigrateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	gcpOSC := getGCPCOSC(logger, start, "mig", params.SourcePoint)
	if gcpOSC == nil {
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

	logger.Infof("Start migration of GCP Cloud Storage to AWS S3")
	if err := gcpOSC.Copy(awsOSC); err != nil {
		end := time.Now()
		logger.Errorf("OSController migration failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from gcp to s3", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationGCPToNCPPostHandler godoc
//
//	@Summary		Migrate data from GCP to NCP Object Storage
//	@Description	Migrate data stored in GCP Cloud Storage to NCP Object Storage.
//	@Tags			[Data Migration], [Object Storage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/gcp/ncp [post]
func MigrationGCPToNCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("miggcpncp", "Export gcp data to ncp objectstorage", start)

	params := MigrateTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	gcpOSC := getGCPCOSC(logger, start, "mig", params.SourcePoint)
	if gcpOSC == nil {
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

	logger.Infof("Start migration of GCP Cloud Storage to NCP Object Storage")
	if err := gcpOSC.Copy(ncpOSC); err != nil {
		end := time.Now()
		logger.Errorf("OSController migration failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully migrated data from gcp to ncp", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}
