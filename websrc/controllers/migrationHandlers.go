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

	logger, logstrings := pageLogInit("miglins3", "Import linux data to s3", start)

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
	logger.Debug().Msgf("%v", params)
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

	logger, logstrings := pageLogInit("miglingcp", "Import linux data to gcp", start)

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

	logger, logstrings := pageLogInit("miglinncp", "Import linux data to ncp objectstorage", start)

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

	logger, logstrings := pageLogInit("migwins3", "Import windows data to s3", start)

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

	logger, logstrings := pageLogInit("migwingcp", "Import windows data to gcp", start)

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

	logger, logstrings := pageLogInit("migwinncp", "Import linux data to ncp objectstorage", start)

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
//	@Tags			[Data Migration], [RDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	MigrateTask	true	"Parameters required for migration"
//	@Success		200			{object}	models.BasicResponse	"Successfully migrated data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/migration/mysql [post]
func MigrationMySQLPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migmysql", "Import mysql to mysql", start)

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
