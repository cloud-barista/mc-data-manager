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

// FROM Naver Object Storage
func MigrationNCPToLinuxGetHandler(ctx echo.Context) error {

	logger := getLogger("migncplin")
	logger.Info("migncplin get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-NCP-Linux",
		"Regions": GetNCPRegions(),
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func MigrationNCPToLinuxPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migncplin", "Export ncp data to linux", start)

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

	ncpOSC := getS3COSC(logger, start, "mig", params)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if !oscExport(logger, start, "ncp", ncpOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	jobEnd(logger, "Successfully migrated data from ncp objectstorage to linux", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationNCPToWindowsGetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")

	logger := getLogger("migncpwin")
	logger.Info("migncpwin get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-NCP-Windows",
		"Regions": GetNCPRegions(),
		"os":      runtime.GOOS,
		"tmpPath": tmpPath,
		"error":   nil,
	})
}

func MigrationNCPToWindowsPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migncpwin", "Export ncp data to windows", start)

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

	ncpOSC := getS3COSC(logger, start, "mig", params)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if !oscExport(logger, start, "ncp", ncpOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from ncp to windows", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationNCPToS3GetHandler(ctx echo.Context) error {

	logger := getLogger("migncps3")
	logger.Info("migncps3 get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content":    "Migration-NCP-S3",
		"NCPRegions": GetNCPRegions(),
		"os":         runtime.GOOS,
		"AWSRegions": GetAWSRegions(),
		"error":      nil,
	})
}

func MigrationNCPToS3PostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migncps3", "Export ncp data to s3", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
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

	awsOSC := getS3OSC(logger, start, "mig", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	logger.Infof("Start migration of NCP Object Storage to AWS S3")
	if err := ncpOSC.Copy(awsOSC); err != nil {
		end := time.Now()
		logger.Errorf("OSController migration failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from ncp to s3", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationNCPToGCPGetHandler(ctx echo.Context) error {

	logger := getLogger("migncpgcp")
	logger.Info("migncpgcp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content":    "Migration-NCP-GCP",
		"NCPRegions": GetNCPRegions(),
		"os":         runtime.GOOS,
		"GCPRegions": GetGCPRegions(),
		"error":      nil,
	})
}

func MigrationNCPToGCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migncpgcp", "Export ncp data to gcp", start)

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

	ncpOSC := getS3COSC(logger, start, "mig", params)
	if ncpOSC == nil {
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

	logger.Infof("Start migration of NCP Object Storage to GCP Cloud Storage")
	if err := ncpOSC.Copy(gcpOSC); err != nil {
		end := time.Now()
		logger.Errorf("OSController migration failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	jobEnd(logger, "Successfully migrated data from ncp to gcp", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}
