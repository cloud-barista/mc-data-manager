package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloud-barista/cm-data-mold/config"
	"github.com/cloud-barista/cm-data-mold/pkg/objectstorage/gcsfs"
	"github.com/cloud-barista/cm-data-mold/pkg/objectstorage/s3fs"
	"github.com/cloud-barista/cm-data-mold/pkg/rdbms/mysql"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/cloud-barista/cm-data-mold/service/osc"
	"github.com/cloud-barista/cm-data-mold/service/rdbc"
	"github.com/gin-gonic/gin"
)

func MigrationLinuxToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-S3",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationLinuxToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if runtime.GOOS != "linux" {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "your current operating system is not linux",
				"Error":  errors.New("your current operating system is not linux"),
			})
			return
		}

		params := MigrationForm{}
		if err := ctx.ShouldBind(&params); err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "failed to bind form data",
				"Error":  err,
			})
			return
		}

		s3c, err := config.NewS3Client(params.AWSAccessKey, params.AWSSecretKey, params.AWSRegion)
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "s3 client creation failed",
				"Error":  err,
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, s3c, params.AWSBucket, params.AWSRegion))
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "OSController creation failed",
				"Error":  err,
			})
			return
		}

		if err := awsOSC.MPut(params.Path); err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "OSController creation failed",
				"Error":  err,
			})
			return
		}

		// migration success. Send result to client
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": "",
			"Error":  nil,
		})
	}
}

func MigrationLinuxToGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationLinuxToGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if runtime.GOOS != "linux" {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "your current operating system is not linux",
				"Error":  errors.New("your current operating system is not linux"),
			})
			return
		}
		params := MigrationForm{}
		if err := ctx.ShouldBind(&params); err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "failed to bind form data",
				"Error":  err,
			})
			return
		}

		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcpCredential")
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "failed to upload crendential file",
				"Error":  err,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "failed",
				"Error":  err,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "failed to save uploaded file",
				"Error":  err,
			})
			return
		}

		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "gcs client generate error",
				"Error":  err,
			})
			return
		}

		gcsOSC, err := osc.New(gcsfs.New(gc, params.ProjectID, params.GCPBucket, params.GCPRegion))
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "OSController generate error",
				"Error":  err,
			})
			return
		}

		if err := gcsOSC.MPut(params.Path); err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "GCS MPut error",
				"Error":  err,
			})
			return
		}

		// migration success. Send result to client
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": "",
			"Error":  nil,
		})
	}
}

func MigrationLinuxToNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationLinuxToNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if runtime.GOOS != "linux" {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "your current operating system is not linux",
				"Error":  errors.New("your current operating system is not linux"),
			})
			return
		}

		params := MigrationForm{}
		if err := ctx.ShouldBind(&params); err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "failed to bind form data",
				"Error":  err,
			})
			return
		}

		s3c, err := config.NewS3ClientWithEndpoint(params.NCPAccessKey, params.NCPSecretKey, params.NCPRegion, params.NCPEndPoint)
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "s3 compatible client creation failed",
				"Error":  err,
			})
			return
		}

		ncpOSC, err := osc.New(s3fs.New(utils.NCP, s3c, params.NCPBucket, params.NCPRegion))
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "OSController creation failed",
				"Error":  err,
			})
			return
		}

		if err := ncpOSC.MPut(params.Path); err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "OSController import failed",
				"Error":  err,
			})
			return
		}

		// migration success. Send result to client
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": "",
			"Error":  nil,
		})
	}
}

func MigrationWindowsToS3GetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-S3",
			"Regions": GetAWSRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationWindowsToS3PostHandler() gin.HandlerFunc {
	filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		if runtime.GOOS != "windows" {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "your current operating system is not windows",
				"Error":  errors.New("your current operating system is not windows"),
			})
			return
		}

		params := MigrationForm{}
		if err := ctx.ShouldBind(&params); err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "failed to bind form data",
				"Error":  err,
			})
			return
		}

		s3c, err := config.NewS3Client(params.AWSAccessKey, params.AWSSecretKey, params.AWSRegion)
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "s3 client creation failed",
				"Error":  err,
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, s3c, params.AWSBucket, params.AWSRegion))
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "OSController creation failed",
				"Error":  err,
			})
			return
		}

		if err := awsOSC.MPut(params.Path); err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "OSController import failed",
				"Error":  err,
			})
			return
		}

		// migration success. Send result to client
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": "",
			"Error":  nil,
		})
	}
}

func MigrationWindowsToGCSGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-GCS",
			"Regions": GetGCPRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationWindowsToGCSPostHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		if runtime.GOOS != "windows" {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Migration-Windows-GCS",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   errors.New("your current operating system is not windows"),
			})
			return
		}
		path := ctx.PostForm("path")
		region := ctx.PostForm("region")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Migration-Windows-GCS",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-GCS",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")
		bucket := ctx.PostForm("bucket")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-GCS",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("failed to save uploaded file : %v", err),
			})
			return
		}

		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-GCS",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("gcs client generate error : %v", err),
			})
			return
		}

		gcsOSC, err := osc.New(gcsfs.New(gc, projectid, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-GCS",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("OSController generate error : %v", err),
			})
			return
		}

		if err := gcsOSC.MPut(path); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-GCS",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("GCS MPut error : %v", err),
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-GCS",
			"Regions": GetGCPRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationWindowsToNCSGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-NCS",
			"Regions": GetNCPRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationWindowsToNCSPostHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		if runtime.GOOS != "windows" {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Migration-Windows-NCS",
				"Regions": GetNCPRegions(),
				"tmpPath": tmpPath,
				"error":   errors.New("your current operating system is not windows"),
			})
			return
		}

		path := ctx.PostForm("path")
		region := ctx.PostForm("region")
		accessKey := ctx.PostForm("accessKey")
		secretKey := ctx.PostForm("secretKey")
		endpoint := ctx.PostForm("endpoint")
		bucket := ctx.PostForm("bucket")

		s3c, err := config.NewS3ClientWithEndpoint(accessKey, secretKey, region, endpoint)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-NCS",
				"Regions": GetNCPRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("s3 compatible client creation failed : %v", err),
			})
			return
		}

		ncpOSC, err := osc.New(s3fs.New(utils.NCP, s3c, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-NCS",
				"Regions": GetNCPRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("OSController creation failed : %v", err),
			})
			return
		}

		if err := ncpOSC.MPut(path); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-NCS",
				"Regions": GetNCPRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("OSController import failed : %v", err),
			})
			return
		}
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-NCS",
			"Regions": GetNCPRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

// SQL Database

func MigrationMySQLGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MySQL",
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationMySQLPostHandler() gin.HandlerFunc {
	filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		formdata := MigrationMySQLForm{}
		p := GetMigrationParamsFormFormData(formdata)

		if err := ctx.ShouldBind(&p); err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "failed to send Form data",
				"Error":  err,
			})
			return
		}

		srcSqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", p.Source.Username, p.Source.Password, p.Source.Host, p.Source.Port))
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "source sqldb open failed",
				"Error":  err,
			})
			return
		}

		srcRDBMS, err := rdbc.New(mysql.New(utils.Provider(p.Source.Provider), srcSqlDB))
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "source rdbms error",
				"Error":  err,
			})
			return
		}

		dstSqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", p.Dest.Username, p.Dest.Password, p.Dest.Host, p.Dest.Port))
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "dest sqldb open failed",
				"Error":  err,
			})
			return
		}

		dstRDBMS, err := rdbc.New(mysql.New(utils.Provider(p.Dest.Provider), dstSqlDB))
		if err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "dest rdbms error",
				"Error":  err,
			})
			return
		}

		if err := srcRDBMS.Copy(dstRDBMS); err != nil {
			ctx.JSONP(http.StatusOK, gin.H{
				"Result": "mysql migration failed. error occured.",
				"Error":  err,
			})
			return
		}

		// mysql migration success result send to client
		ctx.JSONP(http.StatusOK, gin.H{
			"Result": "",
			"Error":  err,
		})
	}
}
