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

// FROM AWS S3
func MigrationS3ToLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migs3lin")
		logger.Info("migs3lin get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-S3-Linux",
			"Regions": GetAWSRegions(),
			"error":   nil,
			"os":      runtime.GOOS,
		})
	}
}

func MigrationS3ToLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migs3lin", "Export s3 data to linux", start)

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

		awsOSC := getS3OSC(logger, start, "mig", params)
		if awsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscExport(logger, start, "s3", awsOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from S3 to Linux", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationS3ToWindowsGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		logger := getLogger("migs3win")
		logger.Info("migs3win get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-S3-Windows",
			"Regions": GetAWSRegions(),
			"tmpPath": tmpPath,
			"os":      runtime.GOOS,
			"error":   nil,
		})
	}
}

func MigrationS3ToWindowsPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genlinux", "Export s3 data to windows", start)

		if !osCheck(logger, start, "windows") {
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

		awsOSC := getS3OSC(logger, start, "mig", params)
		if awsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscExport(logger, start, "s3", awsOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from S3 to Windows", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationS3ToGCPGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migs3gcp")
		logger.Info("migs3gcp get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-S3-GCP",
			"AWSRegions": GetAWSRegions(),
			"GCPRegions": GetGCPRegions(),
			"os":         runtime.GOOS,
			"error":      nil,
		})
	}
}

func MigrationS3ToGCPPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migs3gcp", "Export s3 data to gcp", start)

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

		awsOSC := getS3OSC(logger, start, "mig", params)
		if awsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
		if gcpOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := awsOSC.Copy(gcpOSC); err != nil {
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
		jobEnd(logger, "Successfully migrated data from s3 to gcp", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationS3ToNCPGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migs3ncp")
		logger.Info("migs3ncp get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-S3-NCP",
			"AWSRegions": GetAWSRegions(),
			"NCPRegions": GetNCPRegions(),
			"os":         runtime.GOOS,
			"error":      nil,
		})
	}
}

func MigrationS3ToNCPPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migs3ncp", "Export s3 data to ncp objectstorage", start)

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, &params) {
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

		ncpOSC := getS3COSC(logger, start, "mig", params)
		if ncpOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := awsOSC.Copy(ncpOSC); err != nil {
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
		jobEnd(logger, "Successfully migrated data from s3 to ncp", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}
