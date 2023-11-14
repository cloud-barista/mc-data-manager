package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func GenerateLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genlinux")
		logger.Info("genlinux get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Linux",
			"os":      runtime.GOOS,
			"error":   nil,
		})
	}
}

func GenerateLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genlinux", "Create dummy data in linux", start)

		if !osCheck(logger, start, "linux") {
			ctx.JSONP(http.StatusBadRequest, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		params := getData("gen", ctx).(GenDataParams)

		if !dummyCreate(logger, start, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jobEnd(logger, "Successfully creating a dummy with Linux", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateWindowsGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		logger := getLogger("genwindows")
		logger.Info("genwindows get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Windows",
			"os":      runtime.GOOS,
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func GenerateWindowsPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genwindows", "Create dummy data in windows", start)

		if !osCheck(logger, start, "windows") {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		params := getData("gen", ctx).(GenDataParams)

		if !dummyCreate(logger, start, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jobEnd(logger, "Successfully creating a dummy with Windows", start)
		ctx.JSONP(http.StatusInternalServerError, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genS3")
		logger.Info("genS3 get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-S3",
			"os":      runtime.GOOS,
			"Regions": GetAWSRegions(),
			"Error":   nil,
		})
	}
}

func GenerateS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genS3", "Create dummy data and import to s3", start)

		params := getData("gen", ctx).(GenDataParams)

		tmpDir, ok := createDummyTemp(logger, start)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)
		params.DummyPath = tmpDir

		if !dummyCreate(logger, start, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		awsOSC := getS3OSC(logger, start, "gen", params)
		if awsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscImport(logger, start, "s3", awsOSC, params.DummyPath) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jobEnd(logger, "Dummy creation and import successful with s3", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genGCS")
		logger.Info("genGCS get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-GCS",
			"os":      runtime.GOOS,
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genGCS", "Create dummy data and import to gcs", start)

		params := GenDataParams{}
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

		tmpDir, ok := createDummyTemp(logger, start)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)
		params.DummyPath = tmpDir

		if !dummyCreate(logger, start, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		gcsOSC := getGCSCOSC(logger, start, "gen", params, credFileName)
		if gcsOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscImport(logger, start, "gcs", gcsOSC, params.DummyPath) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jobEnd(logger, "Dummy creation and import successful with gcs", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genNCS")
		logger.Info("genNCS get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-NCS",
			"os":      runtime.GOOS,
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genNCS", "Create dummy data and import to ncp objectstorage", start)

		params := getData("gen", ctx).(GenDataParams)

		tmpDir, ok := createDummyTemp(logger, start)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir

		if !dummyCreate(logger, start, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		ncpOSC := getS3COSC(logger, start, "gen", params)
		if ncpOSC == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !oscImport(logger, start, "ncp", ncpOSC, params.DummyPath) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jobEnd(logger, "Create dummy data and import to ncp objectstorage", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateMySQLGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genmysql")
		logger.Info("genmysql get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MySQL",
			"os":      runtime.GOOS,
			"error":   nil,
		})
	}
}

func GenerateMySQLPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genmysql", "Create dummy data and import to mysql", start)

		params := getData("gen", ctx).(GenDataParams)

		tmpDir, ok := createDummyTemp(logger, start)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir
		params.CheckServerSQL = "on"
		params.SizeServerSQL = "5"

		if !dummyCreate(logger, start, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		rdbc := getMysqlRDBC(logger, start, "gen", params)
		if rdbc == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		sqlList := []string{}
		if !walk(logger, start, &sqlList, params.DummyPath, ".sql") {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Start Import with mysql")
		for _, sql := range sqlList {
			logger.Infof("Read sql file : %s", sql)
			data, err := os.ReadFile(sql)
			if err != nil {
				end := time.Now()
				logger.Errorf("os ReadFile failed : %v", err)
				logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
				logger.Infof("Elapsed time : %s", end.Sub(start).String())
				ctx.JSONP(http.StatusInternalServerError, gin.H{
					"Result": logstrings.String(),
					"Error":  nil,
				})
				return
			}

			logger.Infof("Put start : %s", filepath.Base(sql))
			if err := rdbc.Put(string(data)); err != nil {
				end := time.Now()
				logger.Errorf("RDBController import failed : %v", err)
				logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
				logger.Infof("Elapsed time : %s", end.Sub(start).String())
				ctx.JSONP(http.StatusInternalServerError, gin.H{
					"Result": logstrings.String(),
					"Error":  nil,
				})
				return
			}
			logger.Infof("sql put success : %s", filepath.Base(sql))
		}

		jobEnd(logger, "Dummy creation and import successful with mysql", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateDynamoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("gendynamodb")
		logger.Info("gendynamodb get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-DynamoDB",
			"os":      runtime.GOOS,
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func GenerateDynamoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("gendynamodb", "Create dummy data and import to dynamoDB", start)

		params := getData("gen", ctx).(GenDataParams)

		tmpDir, ok := createDummyTemp(logger, start)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir
		params.CheckServerJSON = "on"
		params.SizeServerJSON = "1"

		if !dummyCreate(logger, start, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jsonList := []string{}
		if !walk(logger, start, &jsonList, params.DummyPath, ".json") {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		awsNRDB := getDynamoNRDBC(logger, start, "gen", params)
		if awsNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !nrdbPutWorker(logger, start, "DynamoDB", awsNRDB, jsonList) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
		}

		jobEnd(logger, "Dummy creation and import successful with dynamoDB", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateFirestoreGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genfirestore")
		logger.Info("genfirestore get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Firestore",
			"os":      runtime.GOOS,
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateFirestorePostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genfirestore", "Create dummy data and import to firestoreDB", start)

		params := GenDataParams{}
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

		tmpDir, ok := createDummyTemp(logger, start)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir
		params.CheckServerJSON = "on"
		params.SizeServerJSON = "1"

		if !dummyCreate(logger, start, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jsonList := []string{}
		if !walk(logger, start, &jsonList, params.DummyPath, ".json") {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		gcpNRDB := getFirestoreNRDBC(logger, start, "gen", params, credFileName)
		if gcpNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !nrdbPutWorker(logger, start, "FirestoreDB", gcpNRDB, jsonList) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
		}

		jobEnd(logger, "Dummy creation and import successful with firestoreDB", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateMongoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genfirestore")
		logger.Info("genmongodb get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MongoDB",
			"os":      runtime.GOOS,
			"error":   nil,
		})
	}
}

func GenerateMongoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		logger, logstrings := pageLogInit("genmongodb", "Create dummy data and import to mongoDB", start)

		params := getData("gen", ctx).(GenDataParams)

		tmpDir, ok := createDummyTemp(logger, start)
		if !ok {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir
		params.CheckServerJSON = "on"
		params.SizeServerJSON = "1"

		if !dummyCreate(logger, start, params) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jsonList := []string{}
		if !walk(logger, start, &jsonList, params.DummyPath, ".json") {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		ncpNRDB := getMongoNRDBC(logger, start, "gen", params)
		if ncpNRDB == nil {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		if !nrdbPutWorker(logger, start, "MongoDB", ncpNRDB, jsonList) {
			ctx.JSONP(http.StatusInternalServerError, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jobEnd(logger, "Dummy creation and import successful with mongoDB", start)
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}
