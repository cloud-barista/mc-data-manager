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
	"os"
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/labstack/echo/v4"
)

// MigrationS3ToLinuxPostHandler godoc
// @Summary Migrate data from AWS S3 to Linux
// @Description Migrate data stored in AWS S3 to a Linux-based system.
// @Tags [Data Migration]
// @Accept json
// @Produce json
// @Param RequestBody body MigrationForm true "Parameters required for migration"
// @Success 200 {object} models.BasicResponse "Successfully migrated data"
// @Failure 400 {object} models.BasicResponse "Invalid Request"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /migration/s3/linux [post]
func MigrationS3ToLinuxPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migs3lin", "Export s3 data to linux", start)

	if !osCheck(logger, start, "linux") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	awsOSC := getS3OSC(logger, start, "mig", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !oscExport(logger, start, "s3", awsOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from S3 to Linux", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationS3ToWindowsPostHandler godoc
// @Summary Migrate data from AWS S3 to Windows
// @Description Migrate data stored in AWS S3 to a Windows-based system.
// @Tags [Data Migration]
// @Accept json
// @Produce json
// @Param RequestBody body MigrationForm true "Parameters required for migration"
// @Success 200 {object} models.BasicResponse "Successfully migrated data"
// @Failure 400 {object} models.BasicResponse "Invalid Request"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /migration/s3/windows [post]
func MigrationS3ToWindowsPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("genlinux", "Export s3 data to windows", start)

	if !osCheck(logger, start, "windows") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	awsOSC := getS3OSC(logger, start, "mig", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !oscExport(logger, start, "s3", awsOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from S3 to Windows", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationS3ToGCPPostHandler godoc
// @Summary Migrate data from AWS S3 to GCP
// @Description Migrate data stored in AWS S3 to Google Cloud Storage.
// @Tags [Data Migration]
// @Accept multipart/form-data
// @Produce json
// @Param RequestBody 	formData MigrationForm	true  "Parameters required for migration"
// @Param gcpCredential	formData file 			false "Parameters required to generate test data"
// @Success 200 {object} models.BasicResponse "Successfully migrated data"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /migration/s3/gcp [post]
func MigrationS3ToGCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migs3gcp", "Export s3 data to gcp", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok && params.GCPCredentialJson == "" {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	awsOSC := getS3OSC(logger, start, "mig", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	logger.Infof("Start migration of AWS S3 to GCP Cloud Storage")
	if err := awsOSC.Copy(gcpOSC); err != nil {
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
	jobEnd(logger, "Successfully migrated data from s3 to gcp", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// MigrationS3ToNCPPostHandler godoc
// @Summary Migrate data from AWS S3 to NCP
// @Description Migrate data stored in AWS S3 to Naver Cloud Object Storage.
// @Tags [Data Migration]
// @Accept json
// @Produce json
// @Param RequestBody body MigrationForm true "Parameters required for migration"
// @Success 200 {object} models.BasicResponse "Successfully migrated data"
// @Failure 500 {object} models.BasicResponse "Internal Server Error"
// @Router /migration/s3/ncp [post]
func MigrationS3ToNCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migs3ncp", "Export s3 data to ncp objectstorage", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	awsOSC := getS3OSC(logger, start, "mig", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	ncpOSC := getS3COSC(logger, start, "mig", params)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	logger.Info("Start migration of AWS S3 to NCP Objest Storage")
	if err := awsOSC.Copy(ncpOSC); err != nil {
		end := time.Now()
		logger.Errorf("OSController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from s3 to ncp", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}
