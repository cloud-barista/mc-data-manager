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
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Migration-Linux-S3",
				"Regions": GetAWSRegions(),
				"error":   errors.New("your current operating system is not linux"),
			})
			return
		}

		path := ctx.PostForm("path")
		region := ctx.PostForm("region")
		accessKey := ctx.PostForm("accessKey")
		secretKey := ctx.PostForm("secretKey")
		bucket := ctx.PostForm("bucket")

		s3c, err := config.NewS3Client(accessKey, secretKey, region)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Linux-S3",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("s3 client creation failed : %v", err),
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, s3c, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Linux-S3",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("OSController creation failed : %v", err),
			})
			return
		}

		if err := awsOSC.MPut(path); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Linux-S3",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("OSController import failed : %v", err),
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-S3",
			"Regions": GetAWSRegions(),
			"error":   nil,
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
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Migration-Linux-GCS",
				"Regions": GetGCPRegions(),
				"error":   errors.New("your current operating system is not linux"),
			})
			return
		}
		path := ctx.PostForm("path")
		region := ctx.PostForm("region")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Migration-Linux-GCS",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Linux-GCS",
				"Regions": GetGCPRegions(),
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
				"Content": "Migration-Linux-GCS",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("failed to save uploaded file : %v", err),
			})
			return
		}

		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Linux-GCS",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("gcs client generate error : %v", err),
			})
			return
		}

		gcsOSC, err := osc.New(gcsfs.New(gc, projectid, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Linux-GCS",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("OSController generate error : %v", err),
			})
			return
		}

		if err := gcsOSC.MPut(path); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Linux-GCS",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("GCS MPut error : %v", err),
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
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
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Migration-Linux-NCS",
				"Regions": GetNCPRegions(),
				"error":   errors.New("your current operating system is not linux"),
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
				"Content": "Migration-Linux-NCS",
				"Regions": GetNCPRegions(),
				"error":   fmt.Errorf("s3 compatible client creation failed : %v", err),
			})
			return
		}

		ncpOSC, err := osc.New(s3fs.New(utils.NCP, s3c, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Linux-NCS",
				"Regions": GetNCPRegions(),
				"error":   fmt.Errorf("OSController creation failed : %v", err),
			})
			return
		}

		if err := ncpOSC.MPut(path); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Linux-NCS",
				"Regions": GetNCPRegions(),
				"error":   fmt.Errorf("OSController import failed : %v", err),
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Linux-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
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
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		if runtime.GOOS != "windows" {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Migration-Windows-S3",
				"Regions": GetAWSRegions(),
				"tmpPath": tmpPath,
				"error":   errors.New("your current operating system is not windows"),
			})
			return
		}

		path := ctx.PostForm("path")
		region := ctx.PostForm("region")
		accessKey := ctx.PostForm("accessKey")
		secretKey := ctx.PostForm("secretKey")
		bucket := ctx.PostForm("bucket")

		s3c, err := config.NewS3Client(accessKey, secretKey, region)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-S3",
				"Regions": GetAWSRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("s3 client creation failed : %v", err),
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, s3c, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-S3",
				"Regions": GetAWSRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("OSController creation failed : %v", err),
			})
			return
		}

		if err := awsOSC.MPut(path); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-Windows-S3",
				"Regions": GetAWSRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("OSController import failed : %v", err),
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-Windows-S3",
			"Regions": GetAWSRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
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
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		srcProvider := ctx.PostForm("srcProvider")
		srcHost := ctx.PostForm("srcHost")
		srcPort := ctx.PostForm("srcPort")
		srcUsername := ctx.PostForm("srcUsername")
		srcPassword := ctx.PostForm("srcPassword")
		dstProvider := ctx.PostForm("targetProvider")
		dstHost := ctx.PostForm("targetHost")
		dstPort := ctx.PostForm("targetPort")
		dstUsername := ctx.PostForm("targetUsername")
		dstPassword := ctx.PostForm("targetPassword")

		srcSqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", srcUsername, srcPassword, srcHost, srcPort))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-MySQL",
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("source sqldb open failed : %v", err),
			})
			return
		}

		srcRDBMS, err := rdbc.New(mysql.New(utils.Provider(srcProvider), srcSqlDB))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-MySQL",
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("source rdbms error : %v", err),
			})
			return
		}

		dstSqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", dstUsername, dstPassword, dstHost, dstPort))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-MySQL",
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("target sqldb open failed : %v", err),
			})
			return
		}

		dstRDBMS, err := rdbc.New(mysql.New(utils.Provider(dstProvider), dstSqlDB))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-MySQL",
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("target rdbms error : %v", err),
			})
			return
		}

		if err := srcRDBMS.Copy(dstRDBMS); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-MySQL",
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("migration error : %v", err),
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-MySQL",
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}
