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

// FROM Google Cloud Storage
func MigrationGCSToLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCS-Linux",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationGCSToLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if runtime.GOOS != "linux" {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-GCS-Linux",
				"Regions": GetGCPRegions(),
				"error":   errors.New("your current operating system is not linux"),
			})
			return
		}

		region := ctx.PostForm("region")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Migration-GCS-Linux",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-GCS-Linux",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")
		bucket := ctx.PostForm("bucket")
		path := ctx.PostForm("path")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-GCS-Linux",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("failed to save uploaded file : %v", err),
			})
			return
		}

		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-GCS-Linux",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		gcpOSC, err := osc.New(gcsfs.New(gc, projectid, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-GCS-Linux",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		if err := gcpOSC.MGet(path); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-GCS-Linux",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCS-Linux",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationGCSToWindowsGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCS-Windows",
			"Regions": GetGCPRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationGCSToWindowsPostHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		if runtime.GOOS != "windows" {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-GCS-Windows",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   errors.New("your current operating system is not windows"),
			})
			return
		}

		region := ctx.PostForm("region")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Migration-GCS-Windows",
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
				"Content": "Migration-GCS-Windows",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")
		bucket := ctx.PostForm("bucket")
		path := ctx.PostForm("path")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-GCS-Windows",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   fmt.Errorf("failed to save uploaded file : %v", err),
			})
			return
		}

		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-GCS-Windows",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   nil,
			})
			return
		}

		gcpOSC, err := osc.New(gcsfs.New(gc, projectid, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-GCS-Windows",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   nil,
			})
			return
		}

		if err := gcpOSC.MGet(path); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-GCS-Windows",
				"Regions": GetGCPRegions(),
				"tmpPath": tmpPath,
				"error":   nil,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-GCS-Windows",
			"Regions": GetGCPRegions(),
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationGCSToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCS-S3",
			"GCPRegions": GetGCPRegions(),
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationGCSToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gcpRegion := ctx.PostForm("gcpRegion")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content":    "Migration-GCS-S3",
				"GCPRegions": GetGCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-GCS-S3",
				"GCPRegions": GetGCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")
		gcsBucket := ctx.PostForm("gcsBucket")
		awsRegion := ctx.PostForm("awsRegion")
		s3AccessKey := ctx.PostForm("s3AccessKey")
		s3SecretKey := ctx.PostForm("s3SecretKey")
		s3Bucket := ctx.PostForm("s3Bucket")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-GCS-S3",
				"GCPRegions": GetGCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      fmt.Errorf("failed to save uploaded file : %v", err),
			})
			return
		}

		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-GCS-S3",
				"GCPRegions": GetGCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		gcpOSC, err := osc.New(gcsfs.New(gc, projectid, gcsBucket, gcpRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-GCS-S3",
				"GCPRegions": GetGCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		ac, err := config.NewS3Client(s3AccessKey, s3SecretKey, awsRegion)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-GCS-S3",
				"GCPRegions": GetGCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, ac, s3Bucket, awsRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-GCS-S3",
				"GCPRegions": GetGCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		if err := gcpOSC.Copy(awsOSC); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-GCS-S3",
				"GCPRegions": GetGCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCS-S3",
			"GCPRegions": GetGCPRegions(),
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationGCSToNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCS-NCS",
			"GCPRegions": GetGCPRegions(),
			"NCPRegions": GetNCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationGCSToNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gcpRegion := ctx.PostForm("gcpRegion")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content":    "Migration-GCS-NCS",
				"GCPRegions": GetGCPRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-GCS-NCS",
				"GCPRegions": GetGCPRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")
		gcsBucket := ctx.PostForm("gcsBucket")
		ncsRegion := ctx.PostForm("ncsRegion")
		ncsAccessKey := ctx.PostForm("ncsAccessKey")
		ncsSecretKey := ctx.PostForm("ncsSecretKey")
		ncsBucket := ctx.PostForm("ncsBucket")
		ncsEndpoint := ctx.PostForm("ncsEndpoint")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content":    "Migration-GCS-NCS",
				"GCPRegions": GetGCPRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      fmt.Errorf("failed to save uploaded file : %v", err),
			})
			return
		}

		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-GCS-NCS",
				"GCPRegions": GetGCPRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      nil,
			})
			return
		}

		gcpOSC, err := osc.New(gcsfs.New(gc, projectid, gcsBucket, gcpRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-GCS-NCS",
				"GCPRegions": GetGCPRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      nil,
			})
			return
		}

		ac, err := config.NewS3ClientWithEndpoint(ncsAccessKey, ncsSecretKey, ncsRegion, ncsEndpoint)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-GCS-NCS",
				"GCPRegions": GetGCPRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      nil,
			})
			return
		}

		ncsOSC, err := osc.New(s3fs.New(utils.AWS, ac, ncsBucket, ncsRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-GCS-NCS",
				"GCPRegions": GetGCPRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      nil,
			})
			return
		}

		if err := gcpOSC.Copy(ncsOSC); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-GCS-NCS",
				"GCPRegions": GetGCPRegions(),
				"NCPRegions": GetNCPRegions(),
				"error":      nil,
			})
			return
		}
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-GCS-NCS",
			"GCPRegions": GetGCPRegions(),
			"NCPRegions": GetNCPRegions(),
			"error":      nil,
		})
	}
}
