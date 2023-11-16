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

// FROM Naver Object Storage
func MigrationNCPToLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migncplin")
		logger.Info("migncplin get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-NCP-Linux",
			"Regions": GetNCPRegions(),
			"os":      runtime.GOOS,
			"error":   nil,
		})
	}
}

func MigrationNCPToLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migncplin", "Export ncp data to linux", start)

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

		ncpOSC := getS3COSC(logger, start, "mig", params)
		if ncpOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscExport(logger, start, "ncp", ncpOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jobEnd(logger, "Successfully migrated data from ncp objectstorage to linux", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationNCPToWindowsGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		logger := getLogger("migncpwin")
		logger.Info("migncpwin get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-NCP-Windows",
			"Regions": GetNCPRegions(),
			"os":      runtime.GOOS,
			"tmpPaht": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationNCPToWindowsPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migncpwin", "Export ncp data to windows", start)

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

		ncpOSC := getS3COSC(logger, start, "mig", params)
		if ncpOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscExport(logger, start, "ncp", ncpOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from ncp to windows", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationNCPToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migncps3")
		logger.Info("migncps3 get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-NCP-S3",
			"NCPRegions": GetNCPRegions(),
			"os":         runtime.GOOS,
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationNCPToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migncps3", "Export ncp data to s3", start)

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, &params) {
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

		awsOSC := getS3OSC(logger, start, "mig", params)
		if awsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := ncpOSC.Copy(awsOSC); err != nil {
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
		jobEnd(logger, "Successfully migrated data from ncp to s3", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationNCPToGCPGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migncpgcp")
		logger.Info("migncpgcp get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-NCP-GCP",
			"NCPRegions": GetNCPRegions(),
			"os":         runtime.GOOS,
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationNCPToGCPPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migncpgcp", "Export ncp data to gcp", start)

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

		ncpOSC := getS3COSC(logger, start, "mig", params)
		if ncpOSC == nil {
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

		if err := ncpOSC.Copy(gcpOSC); err != nil {
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

		jobEnd(logger, "Successfully migrated data from ncp to gcp", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}
