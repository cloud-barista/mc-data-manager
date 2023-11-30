package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

// Object Storage

// FROM Google Cloud Storage
func MigrationGCPToLinuxGetHandler(ctx echo.Context) error {

	logger := getLogger("miggcplin")
	logger.Info("miggcplin get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-GCP-Linux",
		"os":      runtime.GOOS,
		"Regions": GetGCPRegions(),
		"error":   nil,
	})
}

func MigrationGCPToLinuxPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("miggcplin", "Export gcp data to windows", start)

	if !osCheck(logger, start, "linux") {
		return ctx.JSON(http.StatusBadRequest, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if !oscExport(logger, start, "gcp", gcpOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from gcp to linux", start)
	return ctx.JSON(http.StatusOK, gin.H{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationGCPToWindowsGetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")

	logger := getLogger("miggcpwin")
	logger.Info("miggcpwin get page accessed")
	return ctx.Render(http.StatusOK, "index.html", gin.H{
		"Content": "Migration-GCP-Windows",
		"os":      runtime.GOOS,
		"Regions": GetGCPRegions(),
		"tmpPath": tmpPath,
		"error":   nil,
	})
}

func MigrationGCPToWindowsPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("miggcpwin", "Export gcp data to windows", start)

	if !osCheck(logger, start, "windows") {
		return ctx.JSON(http.StatusBadRequest, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if !oscExport(logger, start, "gcp", gcpOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from gcp to windows", start)
	return ctx.JSON(http.StatusOK, gin.H{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationGCPToS3GetHandler(ctx echo.Context) error {

	return ctx.Render(http.StatusOK, "index.html", gin.H{
		"Content":    "Migration-GCP-S3",
		"os":         runtime.GOOS,
		"GCPRegions": GetGCPRegions(),
		"AWSRegions": GetAWSRegions(),
		"error":      nil,
	})
}

func MigrationGCPToS3PostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("genlinux", "Export gcp data to s3", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	awsOSC := getS3OSC(logger, start, "mig", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	logger.Infof("Start migration of GCP Cloud Storage to AWS S3")
	if err := gcpOSC.Copy(awsOSC); err != nil {
		end := time.Now()
		logger.Errorf("OSController migration failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from gcp to s3", start)
	return ctx.JSON(http.StatusOK, gin.H{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationGCPToNCPGetHandler(ctx echo.Context) error {

	logger := getLogger("miggcpncp")
	logger.Info("miggcpncp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", gin.H{
		"Content":    "Migration-GCP-NCP",
		"os":         runtime.GOOS,
		"GCPRegions": GetGCPRegions(),
		"NCPRegions": GetNCPRegions(),
		"error":      nil,
	})
}

func MigrationGCPToNCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("miggcpncp", "Export gcp data to ncp objectstorage", start)

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	ncpOSC := getS3COSC(logger, start, "mig", params)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	logger.Infof("Start migration of GCP Cloud Storage to NCP Object Storage")
	if err := gcpOSC.Copy(ncpOSC); err != nil {
		end := time.Now()
		logger.Errorf("OSController migration failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	jobEnd(logger, "Successfully migrated data from gcp to ncp", start)
	return ctx.JSON(http.StatusOK, gin.H{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}
