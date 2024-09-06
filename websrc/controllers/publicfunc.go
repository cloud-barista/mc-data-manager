/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/internal/log"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbms/awsdnmdb"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbms/gcpfsdb"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbms/ncpmgdb"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/gcpfs"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/s3fs"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbms/mysql"
	"github.com/cloud-barista/mc-data-manager/service/nrdbc"
	"github.com/cloud-barista/mc-data-manager/service/osc"
	"github.com/cloud-barista/mc-data-manager/service/rdbc"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/mongo"
)

func getLogger(jobName string) *logrus.Logger {
	logger := logrus.StandardLogger()
	logger.SetFormatter(&log.CustomTextFormatter{CmdName: "server", JobName: jobName})
	return logger
}

func pageLogInit(pageName, pageInfo string, startTime time.Time) (*logrus.Logger, *strings.Builder) {
	logger := getLogger(pageName)
	var logstrings = strings.Builder{}

	logger.Infof("%s post page accessed", pageName)

	logger.SetOutput(io.MultiWriter(logger.Out, &logstrings))

	logger.Info(pageInfo)
	logger.Infof("start time : %s", startTime.Format("2006-01-02T15:04:05-07:00"))

	return logger, &logstrings
}

func osCheck(logger *logrus.Logger, startTime time.Time, osName string) bool {
	logger.Info("Check the operating system")
	if runtime.GOOS != osName {
		end := time.Now()
		logger.Errorf("Not a %s operating system", osName)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return false
	}
	return true
}

func dummyCreate(logger *logrus.Logger, startTime time.Time, params GenFileParams) bool {
	logger.Info("Start dummy generation")
	err := genData(params, logger)
	if err != nil {
		end := time.Now()
		logger.Errorf("Failed to generate dummy data : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return false
	}
	return true
}

func jobEnd(logger *logrus.Logger, endInfo string, startTime time.Time) {
	end := time.Now()
	logger.Info(endInfo)
	logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
	logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
}

func createDummyTemp(logger *logrus.Logger, startTime time.Time) (string, bool) {
	logger.Info("Create a temporary directory where dummy data will be created")
	tmpDir, err := os.MkdirTemp("", "datamold-dummy")
	if err != nil {
		end := time.Now()
		logger.Error("Failed to generate dummy data : failed to create tmpdir")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return "", false
	} else {
		return tmpDir, true
	}
}

func getS3OSC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}) *osc.OSController {
	gparam, _ := params.(ProviderConfig)
	var err error
	var s3c *s3.Client
	var awsOSC *osc.OSController
	logger.Infof("gmaraps : %v", gparam)
	logger.Info("Get S3 Client")
	credentailManger := config.NewFileCredentialsManager()
	creds, err := credentailManger.LoadCredentialsByProfile(gparam.ProfileName, gparam.Provider)
	if err != nil {
		end := time.Now()
		logger.Errorf("credentail load failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	awsc, ok := creds.(models.AWSCredentials)
	if !ok {
		end := time.Now()
		logger.Errorf("AWS client creation failed")
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	s3c, err = config.NewS3Client(awsc.AccessKey, awsc.SecretKey, gparam.Region)
	if err != nil {
		end := time.Now()
		logger.Errorf("s3 client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an OSController")
	if jobType == "gen" {
		awsOSC, err = osc.New(s3fs.New(models.AWS, s3c, gparam.Bucket, gparam.Region), osc.WithLogger(logger))
	} else {
		awsOSC, err = osc.New(s3fs.New(models.AWS, s3c, gparam.Bucket, gparam.Region), osc.WithLogger(logger))
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("OSController creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	return awsOSC
}

func getS3COSC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}) *osc.OSController {
	gparam, _ := params.(ProviderConfig)

	var err error
	var s3c *s3.Client
	var OSC *osc.OSController

	logger.Info("Get S3 Compataible Client")
	credentailManger := config.NewFileCredentialsManager()
	creds, err := credentailManger.LoadCredentialsByProfile(gparam.ProfileName, gparam.Provider)
	if err != nil {
		end := time.Now()
		logger.Errorf("S3 credentail load failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}
	ncpc, ok := creds.(models.NCPCredentials)
	if !ok {
		logger.Errorf(" credential load failed")
	}
	s3c, err = config.NewS3ClientWithEndpoint(ncpc.AccessKey, ncpc.SecretKey, gparam.Region, gparam.Endpoint)

	if err != nil {
		end := time.Now()
		logger.Errorf("S3 s3 compatible client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an OSController")
	OSC, err = osc.New(s3fs.New(models.NCP, s3c, gparam.Bucket, gparam.Region), osc.WithLogger(logger))
	if err != nil {
		end := time.Now()
		logger.Errorf("OSController creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	return OSC
}

func getGCPCOSC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}) *osc.OSController {
	gparam, _ := params.(ProviderConfig)

	var err error
	var gcpOSC *osc.OSController

	logger.Info("Get GCP Client")
	credentailManger := config.NewFileCredentialsManager()
	creds, err := credentailManger.LoadCredentialsByProfile(gparam.ProfileName, gparam.Provider)
	if err != nil {
		end := time.Now()
		logger.Errorf("gcp credentail load failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}
	gcpc, ok := creds.(models.GCPCredentials)
	if !ok {
		logger.Errorf(" credential load failed")
		return nil
	}
	credentialsJson, err := json.Marshal(gcpc)
	if err != nil {
		end := time.Now()
		logger.Errorf("gcp credentail json Marshal failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	gc, err := config.NewGCPClient(string(credentialsJson))
	if err != nil {
		end := time.Now()
		logger.Errorf("gcp client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an OSController")
	if jobType == "gen" {
		gcpOSC, err = osc.New(gcpfs.New(gc, gcpc.ProjectID, gparam.Bucket, gparam.Region), osc.WithLogger(logger))
	} else {
		gcpOSC, err = osc.New(gcpfs.New(gc, gcpc.ProjectID, gparam.Bucket, gparam.Region), osc.WithLogger(logger))
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("OSController creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	return gcpOSC
}

func getMysqlRDBC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}) *rdbc.RDBController {
	gparam, _ := params.(ProviderConfig)

	var err error
	var sqlDB *sql.DB
	var RDBC *rdbc.RDBController

	logger.Infof("Get SQL Client %v", jobType)
	sqlDB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", gparam.User, gparam.Password, gparam.Host, gparam.Port))
	if err != nil {
		end := time.Now()
		logger.Errorf("sqlDB client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an RDBController")
	RDBC, err = rdbc.New(mysql.New(models.Provider(gparam.Provider), sqlDB), rdbc.WithLogger(logger))
	if err != nil {
		end := time.Now()
		logger.Errorf("RDBController creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	return RDBC
}

func getDynamoNRDBC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}) *nrdbc.NRDBController {
	gparam, _ := params.(ProviderConfig)

	var err error
	var dc *dynamodb.Client
	var NRDBC *nrdbc.NRDBController

	logger.Info("Get DynamoDB Client")
	credentailManger := config.NewFileCredentialsManager()
	creds, err := credentailManger.LoadCredentialsByProfile(gparam.ProfileName, gparam.Provider)
	if err != nil {
		end := time.Now()
		logger.Errorf("aws credentail load failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}
	awsc, ok := creds.(models.AWSCredentials)
	if !ok {
		logger.Errorf(" credential load failed")
	}
	if jobType == "gen" {
		dc, err = config.NewDynamoDBClient(awsc.AccessKey, awsc.SecretKey, gparam.Region)
	} else {
		dc, err = config.NewDynamoDBClient(awsc.AccessKey, awsc.SecretKey, gparam.Region)
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("dynamoDB client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an NRDBController")
	if jobType == "gen" {
		NRDBC, err = nrdbc.New(awsdnmdb.New(dc, gparam.Region), nrdbc.WithLogger(logger))
	} else {
		NRDBC, err = nrdbc.New(awsdnmdb.New(dc, gparam.Region), nrdbc.WithLogger(logger))
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("NRDBController creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	return NRDBC
}

func getFirestoreNRDBC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}) *nrdbc.NRDBController {
	gparam, _ := params.(ProviderConfig)

	var err error
	var fc *firestore.Client
	var NRDBC *nrdbc.NRDBController

	logger.Info("Get FirestoreDB Client")

	credentailManger := config.NewFileCredentialsManager()
	creds, err := credentailManger.LoadCredentialsByProfile(gparam.ProfileName, gparam.Provider)
	if err != nil {
		end := time.Now()
		logger.Errorf("gcp credentail load failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}
	gcpc, ok := creds.(models.GCPCredentials)
	if !ok {
		logger.Errorf(" credential load failed")
		return nil
	}
	credentialsJson, err := json.Marshal(gcpc)
	if err != nil {
		end := time.Now()
		logger.Errorf("gcp credentail json Marshal failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	fc, err = config.NewFireStoreClient(string(credentialsJson), gcpc.ProjectID, gparam.DatabaseID)

	if err != nil {
		end := time.Now()
		logger.Errorf("firestoreDB client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an NRDBController")
	NRDBC, err = nrdbc.New(gcpfsdb.New(fc, gparam.Region), nrdbc.WithLogger(logger))

	if err != nil {
		end := time.Now()
		logger.Errorf("NRDBController creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	return NRDBC
}

func getMongoNRDBC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}) *nrdbc.NRDBController {
	gparam, _ := params.(ProviderConfig)

	var err error
	var mc *mongo.Client
	var NRDBC *nrdbc.NRDBController

	logger.Info("Get MongoDB Client")
	if jobType == "gen" {
		mc, err = config.NewNCPMongoDBClient(gparam.User, gparam.Password, gparam.Host, cast.ToInt(gparam.Port))
	} else {
		mc, err = config.NewNCPMongoDBClient(gparam.User, gparam.Password, gparam.Host, cast.ToInt(gparam.Port))
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("mongoDB client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an NRDBController")
	if jobType == "gen" {
		NRDBC, err = nrdbc.New(ncpmgdb.New(mc, gparam.DatabaseName), nrdbc.WithLogger(logger))
	} else {
		NRDBC, err = nrdbc.New(ncpmgdb.New(mc, gparam.DatabaseName), nrdbc.WithLogger(logger))
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("NRDBController creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	return NRDBC
}

func nrdbPutWorker(logger *logrus.Logger, startTime time.Time, dbType string, nrdbc *nrdbc.NRDBController, jsonList []string) bool {
	var wg sync.WaitGroup
	var mu sync.Mutex
	ret := make(chan error)

	logger.Infof("Start Import with %s", dbType)
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

			if err := nrdbc.Put(tableName, &jsonData); err != nil {
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
			logger.Errorf("NRDBController Import failed : %v", result)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
			return false
		}
	}

	return true
}

func walk(logger *logrus.Logger, startTime time.Time, list *[]string, dirPath string, ext string) bool {
	err := filepath.Walk(dirPath, func(path string, _ fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ext {
			*(list) = append(*(list), path)
		}

		return nil
	})
	if err != nil {
		end := time.Now()
		logger.Errorf("filepath walk failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return false
	}
	return true
}

func oscImport(logger *logrus.Logger, startTime time.Time, osType string, osc *osc.OSController, dstDir string) bool {
	logger.Infof("Start Import with %s", osType)
	if err := osc.MPut(dstDir); err != nil {
		end := time.Now()
		logger.Errorf("OSController import failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return false
	}
	return true
}

func oscExport(logger *logrus.Logger, startTime time.Time, osType string, osc *osc.OSController, dstDir string) bool {
	logger.Infof("Start Export with %s", osType)
	if err := osc.MGet(dstDir); err != nil {
		end := time.Now()
		logger.Errorf("OSController export failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return false
	}
	return true
}

func getData(jobtype string, ctx echo.Context) interface{} {
	if jobtype == "gen" {

		// Read the request body
		data, err := io.ReadAll(ctx.Request().Body)
		if err != nil {
			return err
		}
		// Reset the request body using io.NopCloser
		ctx.Request().Body = io.NopCloser(bytes.NewBuffer(data))

		params := GenDataParams{}
		json.Unmarshal(data, &params)
		return params
	} else {

		return nil
	}
}

func getFileData(jobtype string, ctx echo.Context) interface{} {
	if jobtype == "gen" {

		// Read the request body
		data, err := io.ReadAll(ctx.Request().Body)
		if err != nil {
			return err
		}
		// Reset the request body using io.NopCloser
		ctx.Request().Body = io.NopCloser(bytes.NewBuffer(data))

		params := GenFileParams{}
		json.Unmarshal(data, &params)
		return params
	} else {

		return nil
	}
}

func getDataWithBind(logger *logrus.Logger, startTime time.Time, ctx echo.Context, params interface{}) bool {
	if err := ctx.Bind(params); err != nil {
		end := time.Now()
		logger.Error("Failed to bind form data")
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return false
	}
	return true
}

func gcpCreateCredFile(logger *logrus.Logger, startTime time.Time, ctx echo.Context) (string, string, bool) {
	logger.Info("Create a temporary directory where credential files will be stored")
	// func (*http.Request).FormFile(key string) (multipart.File, *multipart.FileHeader, error)
	// gcpCredentialFile, gcpCredentialHeader, err := ctx.Request.FormFile("gcpCredential")
	gcpCredentialHeader, err := ctx.FormFile("gcpCredential")
	if err != nil {
		end := time.Now()
		logger.Errorf("Get CredentialFile error : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return "", "", false
	}

	credTmpDir, err := os.MkdirTemp("", "datamold-gcp-cred-")
	if err != nil {
		end := time.Now()
		logger.Errorf("Get CredentialFile error : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return "", "", false
	}

	credFileName := filepath.Join(credTmpDir, gcpCredentialHeader.Filename)
	gcpCredentialFile, err := gcpCredentialHeader.Open()
	// err = ctx.SaveUploadedFile(gcpCredentialHeader, credFileName)
	if err != nil {
		end := time.Now()
		logger.Errorf("Get CredentialFile error : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return "", "", false
	}
	defer gcpCredentialFile.Close()

	dst, err := os.Create(credFileName)
	if err != nil {
		end := time.Now()
		logger.Errorf("File create error : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return "", "", false
	}
	defer dst.Close()

	if _, err = io.Copy(dst, gcpCredentialFile); err != nil {
		end := time.Now()
		logger.Errorf("File copy error : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return "", "", false
	}

	return credTmpDir, credFileName, true
}
