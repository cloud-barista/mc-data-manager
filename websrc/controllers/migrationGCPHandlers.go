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
func MigrationGCPToLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("miggcplin")
		logger.Info("miggcplin get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCP-Linux",
			"os":      runtime.GOOS,
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationGCPToLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("miggcplin", "Export gcp data to windows", start)

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

		gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
		if gcpOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscExport(logger, start, "gcp", gcpOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from gcp to linux", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationGCPToWindowsGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		logger := getLogger("miggcpwin")
		logger.Info("miggcpwin get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCP-Windows",
			"os":      runtime.GOOS,
			"Regions": GetGCPRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationGCPToWindowsPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("miggcpwin", "Export gcp data to windows", start)

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

		gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
		if gcpOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscExport(logger, start, "gcp", gcpOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from gcp to windows", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationGCPToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCP-S3",
			"os":         runtime.GOOS,
			"GCPRegions": GetGCPRegions(),
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationGCPToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genlinux", "Export gcp data to s3", start)

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

		gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
		if gcpOSC == nil {
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

		if err := gcpOSC.Copy(awsOSC); err != nil {
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
		jobEnd(logger, "Successfully migrated data from gcp to s3", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationGCPToNCPGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("miggcpncp")
		logger.Info("miggcpncp get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCP-NCP",
			"os":         runtime.GOOS,
			"GCPRegions": GetGCPRegions(),
			"NCPRegions": GetNCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationGCPToNCPPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("miggcpncp", "Export gcp data to ncp objectstorage", start)

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

		gcpOSC := getGCPCOSC(logger, start, "mig", params, credFileName)
		if gcpOSC == nil {
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

		if err := gcpOSC.Copy(ncpOSC); err != nil {
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

		jobEnd(logger, "Successfully migrated data from gcp to ncp", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}
