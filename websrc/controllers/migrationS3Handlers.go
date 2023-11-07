package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloud-barista/cm-data-mold/config"
	"github.com/cloud-barista/cm-data-mold/pkg/objectstorage/gcsfs"
	"github.com/cloud-barista/cm-data-mold/pkg/objectstorage/s3fs"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/cloud-barista/cm-data-mold/service/osc"
	"github.com/gin-gonic/gin"
)

// Object Storage

// FROM AWS S3
func MigrationS3ToLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-S3-Linux",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationS3ToLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if runtime.GOOS != "linux" {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-S3-Linux",
				"Regions": GetAWSRegions(),
				"error":   errors.New("your current operating system is not linux"),
			})
			return
		}

		region := ctx.PostForm("region")
		accessKey := ctx.PostForm("accessKey")
		secretKey := ctx.PostForm("secretKey")
		bucket := ctx.PostForm("bucket")
		path := ctx.PostForm("path")

		sc, err := config.NewS3Client(accessKey, secretKey, region)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-S3-Linux",
				"Regions": GetAWSRegions(),
				"error":   err,
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, sc, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-S3-Linux",
				"Regions": GetAWSRegions(),
				"error":   err,
			})
			return
		}

		if err := awsOSC.MGet(path); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-S3-Linux",
				"Regions": GetAWSRegions(),
				"error":   err,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-S3-Linux",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func MigrationS3ToWindowsGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-S3-Windows",
			"Regions": GetAWSRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationS3ToWindowsPostHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		if runtime.GOOS != "windows" {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-S3-Windows",
				"Regions": GetAWSRegions(),
				"tmpPath": tmpPath,
				"error":   errors.New("your current operating system is not windows"),
			})
			return
		}

		region := ctx.PostForm("region")
		accessKey := ctx.PostForm("accessKey")
		secretKey := ctx.PostForm("secretKey")
		bucket := ctx.PostForm("bucket")
		path := ctx.PostForm("path")

		sc, err := config.NewS3Client(accessKey, secretKey, region)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-S3-Windows",
				"Regions": GetAWSRegions(),
				"tmpPath": tmpPath,
				"error":   err,
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, sc, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-S3-Windows",
				"Regions": GetAWSRegions(),
				"tmpPath": tmpPath,
				"error":   err,
			})
			return
		}

		if err := awsOSC.MGet(path); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-S3-Windows",
				"Regions": GetAWSRegions(),
				"tmpPath": tmpPath,
				"error":   err,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-S3-Windows",
			"Regions": GetAWSRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationS3ToGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-S3-GCS",
			"AWSRegions": GetAWSRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationS3ToGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		awsRegion := ctx.PostForm("awsRegion")
		s3AccessKey := ctx.PostForm("s3AccessKey")
		s3SecretKey := ctx.PostForm("s3SecretKey")
		s3Bucket := ctx.PostForm("s3Bucket")
		gcpRegion := ctx.PostForm("gcpRegion")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content":    "Migration-S3-GCS",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-GCS",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")
		gcsBucket := ctx.PostForm("gcsBucket")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-GCS",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		sc, err := config.NewS3Client(s3AccessKey, s3SecretKey, awsRegion)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-GCS",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, sc, s3Bucket, awsRegion))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-GCS",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-GCS",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		gcsOSC, err := osc.New(gcsfs.New(gc, projectid, gcsBucket, gcpRegion))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-GCS",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		if err := awsOSC.Copy(gcsOSC); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-GCS",
				"AWSRegions": GetAWSRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-S3-GCS",
			"AWSRegions": GetAWSRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationS3ToNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-S3-NCS",
			"AWSRegions": GetAWSRegions(),
			"NCPRegions": GetNCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationS3ToNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		awsRegion := ctx.PostForm("awsRegion")
		s3AccessKey := ctx.PostForm("s3AccessKey")
		s3SecretKey := ctx.PostForm("s3SecretKey")
		s3Bucket := ctx.PostForm("s3Bucket")
		ncsRegion := ctx.PostForm("ncsRegion")
		ncsAccessKey := ctx.PostForm("ncsAccessKey")
		ncsSecretKey := ctx.PostForm("ncsSecretKey")
		ncsEndpoint := ctx.PostForm("ncsEndpoint")
		ncsBucket := ctx.PostForm("ncsBucket")

		sc, err := config.NewS3Client(s3AccessKey, s3SecretKey, awsRegion)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-NCS",
				"AWSRegions": GetAWSRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      err,
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, sc, s3Bucket, ncsRegion))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-NCS",
				"AWSRegions": GetAWSRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      err,
			})
			return
		}

		nc, err := config.NewS3ClientWithEndpoint(ncsAccessKey, ncsSecretKey, ncsRegion, ncsEndpoint)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-NCS",
				"AWSRegions": GetAWSRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      err,
			})
			return
		}

		ncpOSC, err := osc.New(s3fs.New(utils.NCP, nc, ncsBucket, ncsRegion))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-NCS",
				"AWSRegions": GetAWSRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      err,
			})
			return
		}

		if err := awsOSC.Copy(ncpOSC); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-S3-NCS",
				"AWSRegions": GetAWSRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      err,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-S3-NCS",
			"AWSRegions": GetAWSRegions(),
			"NCPRegions": GetNCPRegions(),
			"error":      nil,
		})
	}
}
