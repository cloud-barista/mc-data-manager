package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/cloud-barista/cm-data-mold/config"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/semistructed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/structed"
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
)

func GenerateLinuxGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Linux",
			"error":   nil,
		})
	}
}

func GenerateLinuxPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if runtime.GOOS != "linux" {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Generate-Linux",
				"error":   errors.New("your current operating system is not linux"),
			})
			return
		}

		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		err := genData(params)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Generate-Linux",
				"error":   err,
			})
			return
		}

		ctx.JSONP(http.StatusOK, gin.H{
			"Data":  "dddd",
			"Error": nil,
		})
	}
}

func GenerateWindowsGetHandler() gin.HandlerFunc {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	return func(ctx *gin.Context) {
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
		if runtime.GOOS != "windows" {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Generate-Windows",
				"tmpPath": tmpPath,
				"error":   errors.New("your current operating system is not windows"),
			})
		}

		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		err := genData(params)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Generate-Windows",
				"tmpPath": tmpPath,
				"error":   err,
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Windows",
			"tmpPath": tmpPath,
			"error":   nil,
		})
	}
}

func GenerateS3GetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-S3",
			"Regions": GetAWSRegions(),
			"Error":   nil,
		})
	}
}

func GenerateS3PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-S3",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir

		err = genData(params)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Generate-S3",
				"Regions": GetAWSRegions(),
				"error":   err,
			})
			return
		}

		s3c, err := config.NewS3Client(params.AccessKey, params.SecretKey, params.Region)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-S3",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("s3 client creation failed : %v", err),
			})
			return
		}

		awsOSC, err := osc.New(s3fs.New(utils.AWS, s3c, params.Bucket, params.Region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-S3",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("OSController creation failed : %v", err),
			})
			return
		}

		if err := awsOSC.MPut(tmpDir); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-S3",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("OSController import failed : %v", err),
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-S3",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func GenerateGCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateGCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		region := ctx.PostForm("region")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Generate-GCS",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-GCS",
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
				"Content": "Generate-GCS",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("failed to save uploaded file : %v", err),
			})
			return
		}

		params := GenDataParams{
			CheckSQL:  ctx.PostForm("checkSQL"),
			SizeSQL:   ctx.PostForm("sizeSQL"),
			CheckCSV:  ctx.PostForm("checkCSV"),
			SizeCSV:   ctx.PostForm("sizeCSV"),
			CheckTXT:  ctx.PostForm("checkTXT"),
			SizeTXT:   ctx.PostForm("sizeTXT"),
			CheckPNG:  ctx.PostForm("checkPNG"),
			SizePNG:   ctx.PostForm("sizePNG"),
			CheckGIF:  ctx.PostForm("checkGIF"),
			SizeGIF:   ctx.PostForm("sizeGIF"),
			CheckZIP:  ctx.PostForm("checkZIP"),
			SizeZIP:   ctx.PostForm("sizeZIP"),
			CheckJSON: ctx.PostForm("checkJSON"),
			SizeJSON:  ctx.PostForm("sizeJSON"),
			CheckXML:  ctx.PostForm("checkXML"),
			SizeXML:   ctx.PostForm("sizeXML"),
		}

		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-GCS",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir

		err = genData(params)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Generate-GCS",
				"Regions": GetGCPRegions(),
				"error":   err,
			})
			return
		}

		gc, err := config.NewGCSClient(credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-GCS",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("gcs client generate error : %v", err),
			})
			return
		}

		gcsOSC, err := osc.New(gcsfs.New(gc, projectid, bucket, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-GCS",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("OSController generate error : %v", err),
			})
			return
		}

		if err := gcsOSC.MPut(tmpDir); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-GCS",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("GCS MPut error : %v", err),
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-GCS",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateNCSGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateNCSPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-NCS",
				"Regions": GetNCPRegions(),
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir

		err = genData(params)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"Content": "Generate-NCS",
				"Regions": GetNCPRegions(),
				"error":   err,
			})
			return
		}

		s3c, err := config.NewS3ClientWithEndpoint(params.AccessKey, params.SecretKey, params.Region, params.Endpoint)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-NCS",
				"Regions": GetNCPRegions(),
				"error":   fmt.Errorf("s3 compatible client creation failed : %v", err),
			})
			return
		}

		ncpOSC, err := osc.New(s3fs.New(utils.NCP, s3c, params.Bucket, params.Region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-NCS",
				"Regions": GetNCPRegions(),
				"error":   fmt.Errorf("OSController creation failed : %v", err),
			})
			return
		}

		if err := ncpOSC.MPut(tmpDir); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-NCS",
				"Regions": GetNCPRegions(),
				"error":   fmt.Errorf("OSController import failed : %v", err),
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-NCS",
			"Regions": GetNCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateMySQLGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MySQL",
			"error":   nil,
		})
	}
}

func GenerateMySQLPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)

		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MySQL",
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		params.DummyPath = tmpDir

		if err := structed.GenerateRandomSQL(tmpDir, 1); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MySQL",
				"error":   fmt.Errorf("sql generate error : %v", err),
			})
			return
		}

		sqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", params.DBUser, params.DBPassword, params.DBHost, params.DBPort))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MySQL",
				"error":   fmt.Errorf("sqlDB generate error : %v", err),
			})
			return
		}

		rdbController, err := rdbc.New(mysql.New(utils.Provider(params.DBProvider), sqlDB))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MySQL",
				"error":   fmt.Errorf("RDBController generate error : %v", err),
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
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MySQL",
				"error":   fmt.Errorf("walk error : %v", err),
			})
			return
		}

		for _, sql := range sqlList {
			data, err := os.ReadFile(sql)
			if err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-MySQL",
					"error":   fmt.Errorf("readfile error : %v", err),
				})
				return
			}

			if err := rdbController.Put(string(data)); err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-MySQL",
					"error":   fmt.Errorf("put error : %v", err),
				})
				return
			}
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MySQL",
			"error":   nil,
		})
	}
}

func GenerateDynamoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-DynamoDB",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func GenerateDynamoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		region := ctx.PostForm("region")
		accessKey := ctx.PostForm("accessKey")
		secretKey := ctx.PostForm("secretKey")

		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-DynamoDB",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		if err := semistructed.GenerateRandomJSON(tmpDir, 1); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-DynamoDB",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("json generate error : %v", err),
			})
			return
		}

		dc, err := config.NewDynamoDBClient(accessKey, secretKey, region)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-DynamoDB",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("dynamoDB client generate error : %v", err),
			})
			return
		}

		awsNRDB, err := nrdbc.New(awsdnmdb.New(dc, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-DynamoDB",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("NRDBController generate error : %v", err),
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
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-DynamoDB",
				"Regions": GetAWSRegions(),
				"error":   fmt.Errorf("walk error : %v", err),
			})
			return
		}

		for _, j := range jsonList {
			data, err := os.ReadFile(j)
			if err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-DynamoDB",
					"Regions": GetAWSRegions(),
					"error":   fmt.Errorf("read json file error : %v", err),
				})
				return
			}

			var jsonData []map[string]interface{}
			err = json.Unmarshal(data, &jsonData)
			if err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-DynamoDB",
					"Regions": GetAWSRegions(),
					"error":   fmt.Errorf("json unmarshal error : %v", err),
				})
				return
			}

			tableName := strings.TrimSuffix(filepath.Base(j), ".json")

			if err := awsNRDB.Put(tableName, &jsonData); err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-DynamoDB",
					"Regions": GetAWSRegions(),
					"error":   fmt.Errorf("put error : %v", err),
				})
				return
			}
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-DynamoDB",
			"Regions": GetAWSRegions(),
			"error":   nil,
		})
	}
}

func GenerateFirestoreGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Firestore",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateFirestorePostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		region := ctx.PostForm("region")
		gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "index.html", gin.H{
				"Content": "Generate-Firestore",
				"Regions": GetGCPRegions(),
				"error":   nil,
			})
			return
		}
		defer gcsCredentialFile.Close()

		credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-Firestore",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(credTmpDir)

		projectid := ctx.PostForm("projectid")

		credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
		err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-Firestore",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("failed to save uploaded file : %v", err),
			})
			return
		}

		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-Firestore",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		if err := semistructed.GenerateRandomJSON(tmpDir, 1); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-Firestore",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("json generate error : %v", err),
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
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-Firestore",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("walk error : %v", err),
			})
			return
		}

		fc, err := config.NewFireStoreClient(credFileName, projectid)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-Firestore",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("walk error : %v", err),
			})
			return
		}

		gcpNRDB, err := nrdbc.New(gcpfsdb.New(fc, region))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-Firestore",
				"Regions": GetGCPRegions(),
				"error":   fmt.Errorf("firestore error : %v", err),
			})
			return
		}

		for _, j := range jsonList {
			data, err := os.ReadFile(j)
			if err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-Firestore",
					"Regions": GetGCPRegions(),
					"error":   fmt.Errorf("read json file error : %v", err),
				})
				return
			}

			var jsonData []map[string]interface{}
			err = json.Unmarshal(data, &jsonData)
			if err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-Firestore",
					"Regions": GetGCPRegions(),
					"error":   fmt.Errorf("json unmarshal error : %v", err),
				})
				return
			}

			tableName := strings.TrimSuffix(filepath.Base(j), ".json")

			if err := gcpNRDB.Put(tableName, &jsonData); err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-Firestore",
					"Regions": GetGCPRegions(),
					"error":   fmt.Errorf("put error : %v", err),
				})
				return
			}
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-Firestore",
			"Regions": GetGCPRegions(),
			"error":   nil,
		})
	}
}

func GenerateMongoDBGetHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MongoDB",
			"error":   nil,
		})
	}
}

func GenerateMongoDBPostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		host := ctx.PostForm("host")
		port := ctx.PostForm("port")
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		databaseName := ctx.PostForm("databaseName")

		tmpDir, err := os.MkdirTemp("", "datamold-dummy")
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MongoDB",
				"error":   fmt.Errorf("failed to create tmpdir : %v", err),
			})
			return
		}
		defer os.RemoveAll(tmpDir)

		if err := semistructed.GenerateRandomJSON(tmpDir, 1); err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MongoDB",
				"error":   fmt.Errorf("json generate error : %v", err),
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
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MongoDB",
				"error":   fmt.Errorf("walk error : %v", err),
			})
			return
		}
		Port, err := strconv.Atoi(port)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MongoDB",
				"error":   fmt.Errorf("atoi error : %v", err),
			})
			return
		}

		mc, err := config.NewNCPMongoDBClient(username, password, host, Port)
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MongoDB",
				"error":   fmt.Errorf("mongodb client generate error : %v", err),
			})
			return
		}

		ncpNRDB, err := nrdbc.New(ncpmgdb.New(mc, databaseName))
		if err != nil {
			ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"Content": "Generate-MongoDB",
				"error":   fmt.Errorf("NRDBController generate error : %v", err),
			})
			return
		}

		for _, j := range jsonList {
			data, err := os.ReadFile(j)
			if err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-MongoDB",
					"error":   fmt.Errorf("read json file error : %v", err),
				})
				return
			}

			var jsonData []map[string]interface{}
			err = json.Unmarshal(data, &jsonData)
			if err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-MongoDB",
					"error":   fmt.Errorf("json unmarshal error : %v", err),
				})
				return
			}

			tableName := strings.TrimSuffix(filepath.Base(j), ".json")

			if err := ncpNRDB.Put(tableName, &jsonData); err != nil {
				ctx.HTML(http.StatusInternalServerError, "index.html", gin.H{
					"Content": "Generate-MongoDB",
					"error":   fmt.Errorf("put error : %v", err),
				})
				return
			}
		}

		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"Content": "Generate-MongoDB",
			"error":   nil,
		})
	}
}
