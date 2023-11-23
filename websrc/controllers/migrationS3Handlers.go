package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
)

// Object Storage

// FROM AWS S3
func MigrationS3ToLinuxGetHandler(ctx echo.Context) error {

	logger := getLogger("migs3lin")
	logger.Info("migs3lin get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-S3-Linux",
		"Regions": GetAWSRegions(),
		"error":   nil,
		"os":      runtime.GOOS,
	})
}

func MigrationS3ToLinuxPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migs3lin", "Export s3 data to linux", start)

	if !osCheck(logger, start, "linux") {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	awsOSC := getS3OSC(logger, start, "mig", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if !oscExport(logger, start, "s3", awsOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from S3 to Linux", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationS3ToWindowsGetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")

	logger := getLogger("migs3win")
	logger.Info("migs3win get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-S3-Windows",
		"Regions": GetAWSRegions(),
		"tmpPath": tmpPath,
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func MigrationS3ToWindowsPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("genlinux", "Export s3 data to windows", start)

	if !osCheck(logger, start, "windows") {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	awsOSC := getS3OSC(logger, start, "mig", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if !oscExport(logger, start, "s3", awsOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from S3 to Windows", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationS3ToGCPGetHandler(ctx echo.Context) error {

	logger := getLogger("migs3gcp")
	logger.Info("migs3gcp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content":    "Migration-S3-GCP",
		"AWSRegions": GetAWSRegions(),
		"GCPRegions": GetGCPRegions(),
		"os":         runtime.GOOS,
		"error":      nil,
	})
}

func MigrationS3ToGCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migs3gcp", "Export s3 data to gcp", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	awsOSC := getS3OSC(logger, start, "mig", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if err := awsOSC.Copy(gcpOSC); err != nil {
		end := time.Now()
		logger.Errorf("OSController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from s3 to gcp", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationS3ToNCPGetHandler(ctx echo.Context) error {

	logger := getLogger("migs3ncp")
	logger.Info("migs3ncp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content":    "Migration-S3-NCP",
		"AWSRegions": GetAWSRegions(),
		"NCPRegions": GetNCPRegions(),
		"os":         runtime.GOOS,
		"error":      nil,
	})
}
func MigrationS3ToNCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migs3ncp", "Export s3 data to ncp objectstorage", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	awsOSC := getS3OSC(logger, start, "mig", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	ncpOSC := getS3COSC(logger, start, "mig", params)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if err := awsOSC.Copy(ncpOSC); err != nil {
		end := time.Now()
		logger.Errorf("OSController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from s3 to ncp", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}
