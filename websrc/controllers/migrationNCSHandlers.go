package controllers

import (
	"errors"
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

// FROM Naver Cloud Storage
func MigrationNCSToLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-NCS-Linux",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationNCSToLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if runtime.GOOS != "linux" {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-NCS-Linux",
				"Regions": GetNCPRegions(),
				"error":   errors.New("your current operating system is not linux"),
			})
			return
		}

		region := ctx.PostForm("region")
		accessKey := ctx.PostForm("accessKey")
		secretKey := ctx.PostForm("secretKey")
		endpoint := ctx.PostForm("endpoint")
		bucket := ctx.PostForm("bucket")
		path := ctx.PostForm("path")

		nc, err := config.NewS3ClientWithEndpoint(accessKey, secretKey, region, endpoint)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-NCS-Linux",
				"Regions": GetNCPRegions(),
				"error":   nil,
			})
			return
		}

		ncsOSC, err := osc.New(s3fs.New(utils.NCP, nc, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-NCS-Linux",
				"Regions": GetNCPRegions(),
				"error":   nil,
			})
			return
		}

		if err := ncsOSC.MGet(path); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-NCS-Linux",
				"Regions": GetNCPRegions(),
				"error":   nil,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-NCS-Linux",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func MigrationNCSToWindowsGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-NCS-Windows",
			"Regions": GetNCPRegions(),
			"tmpPaht": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationNCSToWindowsPostHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
		if runtime.GOOS != "windows" {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Migration-NCS-Windows",
				"Regions": GetNCPRegions(),
				"tmpPath": tmpPath,
				"error":   errors.New("your current operating system is not windows"),
			})
			return
		}

		region := ctx.PostForm("region")
		accessKey := ctx.PostForm("accessKey")
		secretKey := ctx.PostForm("secretKey")
		endpoint := ctx.PostForm("endpoint")
		bucket := ctx.PostForm("bucket")
		path := ctx.PostForm("path")

		nc, err := config.NewS3ClientWithEndpoint(accessKey, secretKey, region, endpoint)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-NCS-Windows",
				"Regions": GetNCPRegions(),
				"tmpPaht": tmpPath,
				"error":   nil,
			})
			return
		}

		ncsOSC, err := osc.New(s3fs.New(utils.NCP, nc, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-NCS-Windows",
				"Regions": GetNCPRegions(),
				"tmpPaht": tmpPath,
				"error":   nil,
			})
			return
		}

		if err := ncsOSC.MGet(path); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Migration-NCS-Windows",
				"Regions": GetNCPRegions(),
				"tmpPaht": tmpPath,
				"error":   nil,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Migration-NCS-Windows",
			"Regions": GetNCPRegions(),
			"tmpPaht": tmpPath,
			"error":   nil,
		})
	}
}

func MigrationNCSToS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-NCS-S3",
			"NCPRegions": GetNCPRegions(),
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationNCSToS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ncsRegion := ctx.PostForm("ncsRegion")
		ncsAccessKey := ctx.PostForm("ncsAccessKey")
		ncsSecretKey := ctx.PostForm("ncsSecretKey")
		ncsEndpoint := ctx.PostForm("ncsEndpoint")
		ncsBucket := ctx.PostForm("ncsBucket")
		awsRegion := ctx.PostForm("awsRegion")
		s3AccessKey := ctx.PostForm("s3AccessKey")
		s3SecretKey := ctx.PostForm("s3SecretKey")
		s3Bucket := ctx.PostForm("s3Bucket")

		nc, err := config.NewS3ClientWithEndpoint(ncsAccessKey, ncsSecretKey, ncsRegion, ncsEndpoint)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-S3",
				"NCPRegions": GetNCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		ncsOSC, err := osc.New(s3fs.New(utils.NCP, nc, ncsBucket, ncsRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-S3",
				"NCPRegions": GetNCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		ac, err := config.NewS3Client(s3AccessKey, s3SecretKey, awsRegion)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-S3",
				"NCPRegions": GetNCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, ac, s3Bucket, awsRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-S3",
				"NCPRegions": GetNCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		if err := ncsOSC.Copy(awsOSC); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-S3",
				"NCPRegions": GetNCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-NCS-S3",
			"NCPRegions": GetNCPRegions(),
			"AWSRegions": GetAWSRegions(),
			"error":      nil,
		})
	}
}

func MigrationNCSToGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-NCS-GCS",
			"NCPRegions": GetNCPRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}

func MigrationNCSToGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ncsRegion := ctx.PostForm("ncsRegion")
		ncsAccessKey := ctx.PostForm("ncsAccessKey")
		ncsSecretKey := ctx.PostForm("ncsSecretKey")
		ncsEndpoint := ctx.PostForm("ncsEndpoint")
		ncsBucket := ctx.PostForm("ncsBucket")
		gcpRegion := ctx.PostForm("gcpRegion")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-GCS",
				"NCPRegions": GetNCPRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-GCS",
				"NCPRegions": GetNCPRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")
		gcsBucket := ctx.PostForm("gcsBucket")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-GCS",
				"NCPRegions": GetNCPRegions(),
				"GCPRegions": GetGCPRegions(),
				"error":      nil,
			})
			return
		}

		nc, err := config.NewS3ClientWithEndpoint(ncsAccessKey, ncsSecretKey, ncsRegion, ncsEndpoint)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-S3",
				"NCPRegions": GetNCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		ncsOSC, err := osc.New(s3fs.New(utils.NCP, nc, ncsBucket, ncsRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-S3",
				"NCPRegions": GetNCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-S3",
				"NCPRegions": GetNCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		gcsOSC, err := osc.New(gcsfs.New(gc, projectid, gcsBucket, gcpRegion))
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-S3",
				"NCPRegions": GetNCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		if err := ncsOSC.Copy(gcsOSC); err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content":    "Migration-NCS-S3",
				"NCPRegions": GetNCPRegions(),
				"AWSRegions": GetAWSRegions(),
				"error":      nil,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content":    "Migration-NCS-GCS",
			"NCPRegions": GetNCPRegions(),
			"GCPRegions": GetGCPRegions(),
			"error":      nil,
		})
	}
}
