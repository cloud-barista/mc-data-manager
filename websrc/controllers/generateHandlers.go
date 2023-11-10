package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloud-barista/cm-data-mold/config"
	"github.com/cloud-barista/cm-data-mold/internal/logformatter"
	"github.com/cloud-barista/cm-data-mold/pkg/nrdbms/awsdnmdb"
	"github.com/cloud-barista/cm-data-mold/pkg/nrdbms/gcpfsdb"
	"github.com/cloud-barista/cm-data-mold/pkg/nrdbms/ncpmgdb"
	"github.com/cloud-barista/cm-data-mold/pkg/objectstorage/gcsfs"
	"github.com/cloud-barista/cm-data-mold/pkg/objectstorage/s3fs"
	"github.com/cloud-barista/cm-data-mold/pkg/rdbms/mysql"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/cloud-barista/cm-data-mold/service/nrdbc"
	"github.com/cloud-barista/cm-data-mold/service/osc"
	"github.com/cloud-barista/cm-data-mold/service/rdbc"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func getLogger(jobName string) *logrus.Logger {
	logger := logrus.StandardLogger()
	logger.SetFormatter(&logformatter.CustomTextFormatter{CmdName: "server", JobName: jobName})
	return logger
}

func GenerateLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genlinux")
		logger.Info("genlinux get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Linux",
			"error":   nil,
		})
	}
}

func GenerateLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genlinux")
		start := time.Now()
		logrus.Info("genlinux post page accessed")

		var logstrings = strings.Builder{}
		logger.SetOutput(io.MultiWriter(logger.Out, &logstrings))

		logger.Info("Create dummy data in linux")
		logger.Infof("start time : %s", start.Format("2006-01-02T15:04:05-07:00"))

		logger.Info("Check the operating system")
		if runtime.GOOS != "linux" {
			end := time.Now()
			logger.Error("Not a Linux operating system")
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  "your current operating system is not linux",
			})
			return
		}

		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		logger.Info("Start dummy generation")
		err := genData(params, logger)
		if err != nil {
			end := time.Now()
			logger.Errorf("Failed to generate dummy data : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		end := time.Now()
		logger.Info("Dummy data generation successfully")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
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
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func GenerateWindowsPostHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		logger := getLogger("genwindows")
		start := time.Now()
		logger.Info("genwindows post page accessed")

		var logstrings = strings.Builder{}
		logger.SetOutput(io.MultiWriter(logger.Out, &logstrings))

		logger.Info("Create dummy data in windows")
		logger.Infof("start time : %s", start.Format("2006-01-02T15:04:05-07:00"))

		logger.Info("Check the operating system")
		if runtime.GOOS != "windows" {
			end := time.Now()
			logger.Error("Not a Windows operating system")
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result":  logstrings.String(),
				"tmpPath": tmpPath,
				"Error":   nil,
			})
			return
		}

		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		logger.Info("Start dummy generation")
		err := genData(params, logger)
		if err != nil {
			end := time.Now()
			logger.Errorf("Failed to generate dummy data : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result":  logstrings.String(),
				"tmpPath": tmpPath,
				"Error":   err,
			})
			return
		}

		end := time.Now()
		logger.Info("Dummy data generation successfully")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("gens3")
		logger.Info("genS3 get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-S3",
			"Regions": GetAWSRegions(),
			"Error":   nil,
		})
	}
}

func GenerateS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("gens3")
		start := time.Now()
		logger.Info("genS3 post page accessed")

		var logstrings = strings.Builder{}
		logger.SetOutput(io.MultiWriter(logger.Out, &logstrings))

		logger.Info("Create dummy data in s3")
		logger.Infof("start time : %s", start.Format("2006-01-02T15:04:05-07:00"))

		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		logger.Info("Create a temporary directory where dummy data will be created")
		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			end := time.Now()
			logger.Error("Failed to generate dummy data : failed to create tmpdir")
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir

		logger.Info("Start dummy generation")
		err = genData(params, logger)
		if err != nil {
			end := time.Now()
			logger.Errorf("Failed to generate dummy data : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Get S3 Client")
		s3c, err := config.NewS3Client(params.AccessKey, params.SecretKey, params.Region)
		if err != nil {
			end := time.Now()
			logger.Errorf("s3 client creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Set up the client as an OSController")
		awsOSC, err := osc.New(s3fs.New(utils.AWS, s3c, params.Bucket, params.Region), osc.WithLogger(logger))
		if err != nil {
			end := time.Now()
			logger.Errorf("OSController creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Start Import with s3")
		if err := awsOSC.MPut(tmpDir); err != nil {
			end := time.Now()
			logger.Errorf("OSController import failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		end := time.Now()
		logger.Info("Successfully generated dummy data with s3")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("gengcs")
		logger.Info("genGCS get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("gengcs")
		start := time.Now()
		logger.Info("genGCS post page accessed")

		var logstrings = strings.Builder{}
		logger.SetOutput(io.MultiWriter(logger.Out, &logstrings))

		logger.Info("Create dummy data in gcs")
		logger.Infof("start time : %s", start.Format("2006-01-02T15:04:05-07:00"))

		params := GenDataParams{}
		ctx.ShouldBind(&params)

		logger.Info("Create a temporary directory where credential files will be stored")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			end := time.Now()
			logger.Errorf("Get CredentialFile error : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			end := time.Now()
			logger.Errorf("Get CredentialFile error : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			end := time.Now()
			logger.Errorf("Get CredentialFile error : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Create a temporary directory where dummy data will be created")
		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			end := time.Now()
			logger.Error("Failed to generate dummy data : failed to create tmpdir")
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir

		err = genData(params, logger)
		if err != nil {
			end := time.Now()
			logger.Errorf("Failed to generate dummy data : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
		}

		logger.Info("Get GCS Client")
		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			end := time.Now()
			logger.Errorf("gcs client creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Set up the client as an OSController")
		gcsOSC, err := osc.New(gcsfs.New(gc, params.ProjectID, params.Bucket, params.Region), osc.WithLogger(logger))
		if err != nil {
			end := time.Now()
			logger.Errorf("OSController creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Start Import with gcs")
		if err := gcsOSC.MPut(tmpDir); err != nil {
			end := time.Now()
			logger.Errorf("OSController import failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		end := time.Now()
		logger.Info("Successfully generated dummy data with gcs")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("gens3")
		logger.Info("genS3 get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genncs")
		start := time.Now()
		logger.Info("genNCP post page accessed")

		var logstrings = strings.Builder{}
		logger.SetOutput(io.MultiWriter(logger.Out, &logstrings))

		logger.Info("Create dummy data in ncp objectstorage")
		logger.Infof("start time : %s", start.Format("2006-01-02T15:04:05-07:00"))

		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		logger.Info("Create a temporary directory where dummy data will be created")
		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			end := time.Now()
			logger.Error("Failed to generate dummy data : failed to create tmpdir")
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir

		logger.Info("Start dummy generation")
		err = genData(params, logger)
		if err != nil {
			end := time.Now()
			logger.Errorf("Failed to generate dummy data : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Get S3 Compatible Client")
		s3c, err := config.NewS3ClientWithEndpoint(params.AccessKey, params.SecretKey, params.Region, params.Endpoint)
		if err != nil {
			end := time.Now()
			logger.Errorf("s3 compatible client creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Set up the client as an OSController")
		ncpOSC, err := osc.New(s3fs.New(utils.NCP, s3c, params.Bucket, params.Region), osc.WithLogger(logger))
		if err != nil {
			end := time.Now()
			logger.Errorf("OSController creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Start Import with ncp objectstorage")
		if err := ncpOSC.MPut(tmpDir); err != nil {
			end := time.Now()
			logger.Errorf("OSController import failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		end := time.Now()
		logger.Info("Successfully generated dummy data with ncp objectstorage")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
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
			"error":   nil,
		})
	}
}

func GenerateMySQLPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genmysql")
		start := time.Now()
		logger.Info("genmysql post page accessed")

		var logstrings = strings.Builder{}
		logger.SetOutput(io.MultiWriter(logger.Out, &logstrings))

		logger.Info("Create dummy data in mysql")
		logger.Infof("start time : %s", start.Format("2006-01-02T15:04:05-07:00"))

		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		logger.Info("Create a temporary directory where dummy data will be created")
		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			end := time.Now()
			logger.Error("Failed to generate dummy data : failed to create tmpdir")
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir
		params.CheckServerSQL = "on"
		params.SizeServerSQL = "5"

		logger.Info("Start dummy generation")
		err = genData(params, logger)
		if err != nil {
			end := time.Now()
			logger.Errorf("Failed to generate dummy data : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Get sqlDB Client")
		sqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", params.DBUser, params.DBPassword, params.DBHost, params.DBPort))
		if err != nil {
			end := time.Now()
			logger.Errorf("sqlDB client creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Set up the client as an RDBController")
		rdbController, err := rdbc.New(mysql.New(utils.Provider(params.DBProvider), sqlDB), rdbc.WithLogger(logger))
		if err != nil {
			end := time.Now()
			logger.Errorf("RDBController creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		sqlList := []string{}
		err = filepath.Walk(tmpDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(path) == ".sql" {
				sqlList = append(sqlList, path)
			}

			return nil
		})
		if err != nil {
			end := time.Now()
			logger.Errorf("filepath walk failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
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
				ctx.JSONP(http.StatusOK, gin.H{
					"Result": logstrings.String(),
					"Error":  nil,
				})
				return
			}

			logger.Infof("Put start : %s", filepath.Base(sql))
			if err := rdbController.Put(string(data)); err != nil {
				end := time.Now()
				logger.Errorf("OSController import failed : %v", err)
				logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
				logger.Infof("Elapsed time : %s", end.Sub(start).String())
				ctx.JSONP(http.StatusOK, gin.H{
					"Result": logstrings.String(),
					"Error":  nil,
				})
				return
			}
			logger.Infof("sql put success : %s", filepath.Base(sql))
		}

		end := time.Now()
		logger.Info("Successfully generated dummy data with mysql")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
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
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func GenerateDynamoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("gendynamodb")
		start := time.Now()
		logrus.Info("gendynamodb post page accessed")

		var logstrings = strings.Builder{}
		logger.SetOutput(io.MultiWriter(logger.Out, &logstrings))

		logger.Info("Create dummy data in dynamodb")
		logger.Infof("start time : %s", start.Format("2006-01-02T15:04:05-07:00"))

		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		logger.Info("Create a temporary directory where dummy data will be created")
		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			end := time.Now()
			logger.Error("Failed to generate dummy data : failed to create tmpdir")
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir
		params.CheckServerJSON = "on"
		params.SizeServerJSON = "1"

		logger.Info("Start dummy generation")
		err = genData(params, logger)
		if err != nil {
			end := time.Now()
			logger.Errorf("Failed to generate dummy data : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Get DynamoDB Client")
		dc, err := config.NewDynamoDBClient(params.AccessKey, params.SecretKey, params.Region)
		if err != nil {
			end := time.Now()
			logger.Errorf("dynamoDB client creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Set up the client as an NRDBController")
		awsNRDB, err := nrdbc.New(awsdnmdb.New(dc, params.Region), nrdbc.WithLogger(logger))
		if err != nil {
			end := time.Now()
			logger.Errorf("NRDBController creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jsonList := []string{}
		err = filepath.Walk(tmpDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(path) == ".json" {
				jsonList = append(jsonList, path)
			}

			return nil
		})
		if err != nil {
			end := time.Now()
			logger.Errorf("filepath walk failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		ret := make(chan error)

		logger.Info("Start Import with dynamoDB")
		for _, j := range jsonList {
			wg.Add(1)
			go func(jPath string, jret chan<- error) {
				defer wg.Done()

				mu.Lock()
				logger.Infof("Read json file : %s", jPath)
				mu.Unlock()

				data, err := os.ReadFile(jPath)
				if err != nil {
					jret <- err
					return
				}

				mu.Lock()
				logger.Infof("data unmarshal : %s", filepath.Base(jPath))
				mu.Unlock()

				var jsonData []map[string]interface{}
				err = json.Unmarshal(data, &jsonData)
				if err != nil {
					jret <- err
					return
				}

				tableName := strings.TrimSuffix(filepath.Base(jPath), ".json")

				mu.Lock()
				logger.Infof("Put start : %s", filepath.Base(jPath))
				mu.Unlock()

				if err := awsNRDB.Put(tableName, &jsonData); err != nil {
					jret <- err
					return
				}

				jret <- nil
			}(j, ret)
		}

		go func() {
			wg.Wait()
			close(ret)
		}()

		for result := range ret {
			if result != nil {
				end := time.Now()
				logger.Errorf("NRDBController Import failed : %v", err)
				logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
				logger.Infof("Elapsed time : %s", end.Sub(start).String())
				ctx.JSONP(http.StatusOK, gin.H{
					"Result": logstrings.String(),
					"Error":  nil,
				})
				return
			}
		}

		end := time.Now()
		logger.Info("Successfully generated dummy data with dynamoDB")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
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
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateFirestorePostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genfirestore")
		start := time.Now()
		logger.Info("genfirestore post page accessed")

		var logstrings = strings.Builder{}
		logger.SetOutput(io.MultiWriter(logger.Out, &logstrings))

		logger.Info("Create dummy data in firestore")
		logger.Infof("start time : %s", start.Format("2006-01-02T15:04:05-07:00"))

		params := GenDataParams{}
		ctx.ShouldBind(&params)

		logger.Info("Create a temporary directory where credential files will be stored")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			end := time.Now()
			logger.Errorf("Get CredentialFile error : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			end := time.Now()
			logger.Errorf("Get CredentialFile error : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			end := time.Now()
			logger.Errorf("Get CredentialFile error : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Create a temporary directory where dummy data will be created")
		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			end := time.Now()
			logger.Error("Failed to generate dummy data : failed to create tmpdir")
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir
		params.CheckServerJSON = "on"
		params.SizeServerJSON = "1"

		err = genData(params, logger)
		if err != nil {
			end := time.Now()
			logger.Errorf("Failed to generate dummy data : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
		}

		jsonList := []string{}
		err = filepath.Walk(tmpDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(path) == ".json" {
				jsonList = append(jsonList, path)
			}

			return nil
		})
		if err != nil {
			end := time.Now()
			logger.Errorf("filepath walk failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		logger.Info("Get FirestoreDB Client")
		fc, err := config.NewFireStoreClient(credFileName, params.ProjectID)
		if err != nil {
			end := time.Now()
			logger.Errorf("firestoreDB client creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Set up the client as an NRDBController")
		gcpNRDB, err := nrdbc.New(gcpfsdb.New(fc, params.Region), nrdbc.WithLogger(logger))
		if err != nil {
			end := time.Now()
			logger.Errorf("NRDBController creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		ret := make(chan error)

		logger.Info("Start Import with firestoreDB")
		for _, j := range jsonList {
			wg.Add(1)
			go func(jPath string, jret chan<- error) {
				defer wg.Done()

				mu.Lock()
				logger.Infof("Read json file : %s", jPath)
				mu.Unlock()

				data, err := os.ReadFile(jPath)
				if err != nil {
					jret <- err
					return
				}

				logger.Infof("data unmarshal : %s", filepath.Base(jPath))
				var jsonData []map[string]interface{}
				err = json.Unmarshal(data, &jsonData)
				if err != nil {
					jret <- err
					return
				}

				tableName := strings.TrimSuffix(filepath.Base(jPath), ".json")

				mu.Lock()
				logger.Infof("Put start : %s", filepath.Base(jPath))
				mu.Unlock()

				if err := gcpNRDB.Put(tableName, &jsonData); err != nil {
					jret <- err
					return
				}

				jret <- nil
			}(j, ret)
		}

		go func() {
			wg.Wait()
			close(ret)
		}()

		for result := range ret {
			if result != nil {
				end := time.Now()
				logger.Errorf("NRDBController Import failed : %v", err)
				logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
				logger.Infof("Elapsed time : %s", end.Sub(start).String())
				ctx.JSONP(http.StatusOK, gin.H{
					"Result": logstrings.String(),
					"Error":  nil,
				})
				return
			}
		}

		end := time.Now()
		logger.Info("Successfully generated dummy data with firestoreDB")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}

func GenerateMongoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genmongodb")
		logger.Info("genmongodb get page accessed")
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MongoDB",
			"error":   nil,
		})
	}
}

func GenerateMongoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := getLogger("genmongodb")
		start := time.Now()
		logger.Info("genmongodb post page accessed")

		var logstrings = strings.Builder{}
		logger.SetOutput(io.MultiWriter(logger.Out, &logstrings))

		logger.Info("Create dummy data in mongodb")
		logger.Infof("start time : %s", start.Format("2006-01-02T15:04:05-07:00"))

		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		logger.Info("Create a temporary directory where dummy data will be created")
		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			end := time.Now()
			logger.Error("Failed to generate dummy data : failed to create tmpdir")
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir
		params.CheckServerJSON = "on"
		params.SizeServerJSON = "1"

		logger.Info("Start dummy generation")
		err = genData(params, logger)
		if err != nil {
			end := time.Now()
			logger.Errorf("Failed to generate dummy data : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		jsonList := []string{}
		err = filepath.Walk(tmpDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(path) == ".json" {
				jsonList = append(jsonList, path)
			}

			return nil
		})
		if err != nil {
			end := time.Now()
			logger.Errorf("filepath walk failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}
		Port, err := strconv.Atoi(params.DBPort)
		if err != nil {
			end := time.Now()
			logger.Errorf("port atoi failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Get MongoDB Client")
		mc, err := config.NewNCPMongoDBClient(params.DBUser, params.DBPassword, params.DBHost, Port)
		if err != nil {
			end := time.Now()
			logger.Errorf("mongoDB client creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		var wg sync.WaitGroup
		var mu sync.Mutex
		ret := make(chan error)

		logger.Info("Set up the client as an NRDBController")
		ncpNRDB, err := nrdbc.New(ncpmgdb.New(mc, params.DatabaseName), nrdbc.WithLogger(logger))
		if err != nil {
			end := time.Now()
			logger.Errorf("NRDBController creation failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": logstrings.String(),
				"Error":  nil,
			})
			return
		}

		logger.Info("Start Import with mongoDB")
		for _, j := range jsonList {
			wg.Add(1)
			go func(jPath string, jret chan<- error) {
				defer wg.Done()

				mu.Lock()
				logger.Infof("Read json file : %s", jPath)
				mu.Unlock()

				data, err := os.ReadFile(jPath)
				if err != nil {
					jret <- err
					return
				}

				logger.Infof("data unmarshal : %s", filepath.Base(jPath))
				var jsonData []map[string]interface{}
				err = json.Unmarshal(data, &jsonData)
				if err != nil {
					jret <- err
					return
				}

				tableName := strings.TrimSuffix(filepath.Base(jPath), ".json")

				mu.Lock()
				logger.Infof("Put start : %s", filepath.Base(jPath))
				mu.Unlock()

				if err := ncpNRDB.Put(tableName, &jsonData); err != nil {
					jret <- err
					return
				}

				jret <- nil
			}(j, ret)
		}

		go func() {
			wg.Wait()
			close(ret)
		}()

		for result := range ret {
			if result != nil {
				end := time.Now()
				logger.Errorf("NRDBController Import failed : %v", err)
				logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
				logger.Infof("Elapsed time : %s", end.Sub(start).String())
				ctx.JSONP(http.StatusOK, gin.H{
					"Result": logstrings.String(),
					"Error":  nil,
				})
				return
			}
		}

		end := time.Now()
		logger.Info("Successfully generated dummy data with mongoDB")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(start).String())
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": logstrings.String(),
			"Error":  nil,
		})
	}
}
