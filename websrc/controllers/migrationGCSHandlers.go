package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// Object Storage

// FROM Google Cloud Storage
func MigrationGCSToLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("miglingcs")
		logger.Info("miglingcs get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCS-Linux",
			"os":      runtime.GOOS,
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationGCSToLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("miglingcs", "Export gcs data to windows", start)

		if !osCheck(logger, start, "linux") {
			ctx.JSONP(http.StatusBadRequest, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, &params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		gcsOSC := getGCSCOSC(logger, start, "mig", params, credFileName)
		if gcsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscExport(logger, start, "gcs", gcsOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from gcs to linux", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationGCSToWindowsGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		logger := getLogger("miggcswin")
		logger.Info("miggcswin get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCS-Windows",
			"os":      runtime.GOOS,
			"Regions": GetGCPRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationGCSToWindowsPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("miggcswin", "Export gcs data to windows", start)

		if !osCheck(logger, start, "linux") {
			ctx.JSONP(http.StatusBadRequest, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, &params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		gcsOSC := getGCSCOSC(logger, start, "mig", params, credFileName)
		if gcsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscExport(logger, start, "gcs", gcsOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from gcs to windows", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationGCSToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCS-S3",
			"os":         runtime.GOOS,
			"GCPRegions": GetGCPRegions(),
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationGCSToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genlinux", "Export gcs data to s3", start)

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, &params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		gcsOSC := getGCSCOSC(logger, start, "mig", params, credFileName)
		if gcsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		awsOSC := getS3OSC(logger, start, "mig", params)
		if awsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := gcsOSC.Copy(awsOSC); err != nil {
			end := time.Now()
			logger.Errorf("OSController copy failed : %v", err)
			logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from gcs to s3", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationGCSToNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("miggcsncp")
		logger.Info("miggcsncp get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCS-NCS",
			"os":         runtime.GOOS,
			"GCPRegions": GetGCPRegions(),
			"NCPRegions": GetNCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationGCSToNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("miggcsncp", "Export gcs data to ncp objectstorage", start)

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, &params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		gcsOSC := getGCSCOSC(logger, start, "mig", params, credFileName)
		if gcsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		ncpOSC := getS3COSC(logger, start, "mig", params)
		if ncpOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := gcsOSC.Copy(ncpOSC); err != nil {
			end := time.Now()
			logger.Errorf("OSController copy failed : %v", err)
			logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jobEnd(logger, "Successfully migrated data from gcs to ncp", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}
