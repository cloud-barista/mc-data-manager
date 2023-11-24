package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
)

func MigrationLinuxToS3GetHandler(ctx echo.Context) error {
	logger := getLogger("miglins3")
	logger.Info("miglinux get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-Linux-S3",
		"Regions": GetAWSRegions(),
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func MigrationLinuxToS3PostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("miglins3", "Import linux data to s3", start)

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

	if !oscImport(logger, start, "s3", awsOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Linux to s3", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationLinuxToGCPGetHandler(ctx echo.Context) error {
	logger := getLogger("miglingcp")
	logger.Info("miglingcp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-Linux-GCP",
		"Regions": GetGCPRegions(),
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func MigrationLinuxToGCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("miglingcp", "Import linux data to gcp", start)

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

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})

	}
	defer os.RemoveAll(credTmpDir)

	gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})

	}

	if !oscImport(logger, start, "gcp", gcpOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})

	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Linux to gcp", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationLinuxToNCPGetHandler(ctx echo.Context) error {

	logger := getLogger("miglinncp")
	logger.Info("miglinncp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-Linux-NCP",
		"Regions": GetNCPRegions(),
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func MigrationLinuxToNCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("miglinncp", "Import linux data to ncp objectstorage", start)

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

	if !oscImport(logger, start, "ncp", ncpOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Linux to ncp objectstorage", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationWindowsToS3GetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	logger := getLogger("migwins3")
	logger.Info("migwins3 get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-Windows-S3",
		"Regions": GetAWSRegions(),
		"os":      runtime.GOOS,
		"tmpPath": tmpPath,
		"error":   nil,
	})

}

func MigrationWindowsToS3PostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migwins3", "Import windows data to s3", start)

	if !osCheck(logger, start, "windows") {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
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

	if !oscImport(logger, start, "s3", awsOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})

	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Windows to s3", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationWindowsToGCPGetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	logger := getLogger("migwingcp")
	logger.Info("migwingcp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-Windows-GCP",
		"Regions": GetGCPRegions(),
		"os":      runtime.GOOS,
		"tmpPath": tmpPath,
		"error":   nil,
	})
}

func MigrationWindowsToGCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("migwingcp", "Import windows data to gcp", start)

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

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if !oscImport(logger, start, "gcp", gcpOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Windows to gcp", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

func MigrationWindowsToNCPGetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")

	logger := getLogger("migwinncp")
	logger.Info("migwinncp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-Windows-NCP",
		"Regions": GetNCPRegions(),
		"os":      runtime.GOOS,
		"tmpPath": tmpPath,
		"error":   nil,
	})
}

func MigrationWindowsToNCPPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migwinncp", "Import linux data to ncp objectstorage", start)

	if !osCheck(logger, start, "windows") {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	params := MigrationForm{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
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

	if !oscImport(logger, start, "ncp", ncpOSC, params.Path) {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// migration success. Send result to client
	jobEnd(logger, "Successfully migrated data from Windows to ncp objectstorage", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}

// SQL Database
func MigrationMySQLGetHandler(ctx echo.Context) error {

	logger := getLogger("migmysql")
	logger.Info("migmysql get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Migration-MySQL",
		"error":   nil,
		"os":      runtime.GOOS,
	})
}

func MigrationMySQLPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("migmysql", "Import mysql to mysql", start)

	formdata := MigrationMySQLForm{}
	if !getDataWithBind(logger, start, ctx, &formdata) {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
	params := GetMigrationParamsFormFormData(formdata)

	srdbc := getMysqlRDBC(logger, start, "smig", params)
	if srdbc == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	trdbc := getMysqlRDBC(logger, start, "tmig", params)
	if trdbc == nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	if err := srdbc.Copy(trdbc); err != nil {
		end := time.Now()
		logger.Errorf("RDBController copy failed : %v", err)
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}

	// mysql migration success result send to client
	jobEnd(logger, "Successfully migrated data from mysql to mysql", start)
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"Result": logstrings.String(),
		"Error":  nil,
	})
}
