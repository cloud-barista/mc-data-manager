package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func MigrationLinuxToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("miglins3")
		logger.Info("miglinux get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-S3",
			"Regions": GetAWSRegions(),
			"os":      runtime.GOOS,
			"error":   nil,
		})
	}
}

func MigrationLinuxToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("miglins3", "Import linux data to s3", start)

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

		if !oscImport(logger, start, "s3", awsOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from Linux to s3", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationLinuxToGCPGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("miglingcp")
		logger.Info("miglingcp get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-GCP",
			"Regions": GetGCPRegions(),
			"os":      runtime.GOOS,
			"error":   nil,
		})
	}
}

func MigrationLinuxToGCPPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("miglingcp", "Import linux data to gcp", start)

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

		if !oscImport(logger, start, "gcp", gcpOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from Linux to gcp", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationLinuxToNCPGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("miglinncp")
		logger.Info("miglinncp get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-NCP",
			"Regions": GetNCPRegions(),
			"os":      runtime.GOOS,
			"error":   nil,
		})
	}
}

func MigrationLinuxToNCPPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("miglinncp", "Import linux data to ncp objectstorage", start)

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

		if !oscImport(logger, start, "ncp", ncpOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from Linux to ncp objectstorage", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationWindowsToS3GetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		logger := getLogger("migwins3")
		logger.Info("migwins3 get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-S3",
			"Regions": GetAWSRegions(),
			"os":      runtime.GOOS,
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationWindowsToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migwins3", "Import windows data to s3", start)

		if !osCheck(logger, start, "windows") {
			ctx.JSONP(http.StatusOK, gin.H{
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

		if !oscImport(logger, start, "s3", awsOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from Windows to s3", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationWindowsToGCPGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		logger := getLogger("migwingcp")
		logger.Info("migwingcp get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-GCP",
			"Regions": GetGCPRegions(),
			"os":      runtime.GOOS,
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationWindowsToGCPPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migwingcp", "Import windows data to gcp", start)

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

		if !oscImport(logger, start, "gcp", gcpOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from Windows to gcp", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func MigrationWindowsToNCPGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		logger := getLogger("migwinncp")
		logger.Info("migwinncp get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-NCP",
			"Regions": GetNCPRegions(),
			"os":      runtime.GOOS,
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationWindowsToNCPPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migwinncp", "Import linux data to ncp objectstorage", start)

		if !osCheck(logger, start, "windows") {
			ctx.JSONP(http.StatusBadRequest, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		params := MigrationForm{}
		if !getDataWithBind(logger, start, ctx, &params) {
			ctx.JSONP(http.StatusOK, gin.H{
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

		if !oscImport(logger, start, "ncp", ncpOSC, params.Path) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// migration success. Send result to client
		jobEnd(logger, "Successfully migrated data from Windows to ncp objectstorage", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

// SQL Database
func MigrationMySQLGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("migmysql")
		logger.Info("migmysql get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MySQL",
			"error":   nil,
			"os":      runtime.GOOS,
		})
	}
}

func MigrationMySQLPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("migmysql", "Import mysql to mysql", start)

		formdata := MigrationMySQLForm{}
		if !getDataWithBind(logger, start, ctx, &formdata) {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		params := GetMigrationParamsFormFormData(formdata)

		srdbc := getMysqlRDBC(logger, start, "smig", params)
		if srdbc == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		trdbc := getMysqlRDBC(logger, start, "tmig", params)
		if trdbc == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if err := srdbc.Copy(trdbc); err != nil {
			end := time.Now()
			logger.Errorf("RDBController copy failed : %v", err)
			logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		// mysql migration success result send to client
		jobEnd(logger, "Successfully migrated data from mysql to mysql", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}
