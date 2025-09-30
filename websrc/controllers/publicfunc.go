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
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbms/awsdnmdb"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbms/gcpfsdb"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbms/ncpmgdb"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/filtering"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/gcpfs"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/s3fs"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbms/mysql"
	"github.com/cloud-barista/mc-data-manager/service/nrdbc"
	"github.com/cloud-barista/mc-data-manager/service/osc"
	"github.com/cloud-barista/mc-data-manager/service/rdbc"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/mongo"
)

func getLoggerFromContext(c echo.Context) *zerolog.Logger {
	// Retrieve the logger from the request context
	return zerolog.Ctx(c.Request().Context())
}

// func pageLogInit(c echo.Context, pageName, pageInfo string, startTime time.Time) (*zerolog.Logger, *strings.Builder) {
// 	logger := getLoggerFromContext(c)
// 	var logstrings strings.Builder

// 	// Log page access information
// 	logger.Info().Msgf("%s post page accessed", pageName)
// 	logger.Info().Msg(pageInfo)
// 	logger.Info().Str("start time", startTime.Format(time.RFC3339))

// 	return logger, &logstrings
// }

func pageLogInit(c echo.Context, pageName, pageInfo string, startTime time.Time) (*zerolog.Logger, *strings.Builder) {
	parentLogger := getLoggerFromContext(c)
	var logstrings strings.Builder

	// Assuming the original output is os.Stderr
	originalOutput := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	messageOnlyWriter := zerolog.ConsoleWriter{
		Out:        &logstrings,
		TimeFormat: time.DateTime,
		FormatMessage: func(i interface{}) string {
			if i == nil {
				return ""
			}
			return fmt.Sprintf("%s", i)
		},
		NoColor: true, // Remove Color code
		// The others  NullString.
		// FormatLevel: func(i interface{}) string { return "" },
		// FormatTimestamp:     func(i interface{}) string { return "" },
		FormatCaller:        func(i interface{}) string { return "" },
		FormatFieldName:     func(i interface{}) string { return "" },
		FormatFieldValue:    func(i interface{}) string { return "" },
		FormatErrFieldName:  func(i interface{}) string { return "" },
		FormatErrFieldValue: func(i interface{}) string { return "" },
	}

	// Create a MultiWriter that writes to both the original output and the strings.Builder
	multiWriter := io.MultiWriter(originalOutput, messageOnlyWriter)

	// Create a child logger with the new output
	logger := parentLogger.Output(multiWriter)

	// Log page access information
	logger.Info().Msgf("%s page accessed", pageName)
	logger.Info().Msg(pageInfo)
	logger.Info().Str("start time", startTime.Format(time.RFC3339))

	return &logger, &logstrings
}

func osCheck(logger *zerolog.Logger, startTime time.Time, osName string) bool {
	logger.Info().Msg("Check the operating system")
	if runtime.GOOS != osName {
		end := time.Now()
		logger.Error().Msgf("Not a %s operating system", osName)
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return false
	}
	return true
}

func dummyCreate(logger *zerolog.Logger, startTime time.Time, params models.GenFileParams) bool {
	logger.Info().Msg("Start dummy generation")
	err := genData(params, logger)
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("Failed to generate dummy data")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return false
	}
	return true
}

func jobEnd(logger *zerolog.Logger, endInfo string, startTime time.Time) {
	end := time.Now()
	logger.Info().Msg(endInfo)
	logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
	logger.Info().Str("Elapsed time", end.Sub(startTime).String())
}

func createDummyTemp(logger *zerolog.Logger, startTime time.Time) (string, bool) {
	logger.Info().Msg("Create a temporary directory where dummy data will be created")
	tmpDir, err := os.MkdirTemp("", "datamold-dummy")
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("Failed to generate dummy data: failed to create tmpdir")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return "", false
	} else {
		return tmpDir, true
	}
}

func getS3OSC(logger *zerolog.Logger, startTime time.Time, jobType string, params interface{}) *osc.OSController {
	gparam, _ := params.(models.ProviderConfig)
	var err error
	var s3c *s3.Client
	var awsOSC *osc.OSController
	logger.Info().Msg("Get S3 Client")
	credentailManger := config.AuthManager
	creds, err := credentailManger.LoadCredentialsById(uint64(gparam.CredentialId))
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("credentail load failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	awsc, ok := creds.(models.AWSCredentials)
	if !ok {
		end := time.Now()
		logger.Error().Msg("AWS client creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	s3c, err = config.NewS3Client(awsc.AccessKey, awsc.SecretKey, gparam.Region)
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("s3 client creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	logger.Info().Msg("Set up the client as an OSController")
	if jobType == "gen" {
		awsOSC, err = osc.New(s3fs.New(models.AWS, s3c, gparam.Bucket, gparam.Region))
	} else {
		awsOSC, err = osc.New(s3fs.New(models.AWS, s3c, gparam.Bucket, gparam.Region))
	}
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("OSController creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	return awsOSC
}

func getS3COSC(logger *zerolog.Logger, startTime time.Time, jobType string, params interface{}) *osc.OSController {
	gparam, _ := params.(models.ProviderConfig)

	var err error
	var s3c *s3.Client
	var OSC *osc.OSController

	logger.Info().Msg("Get S3 Compataible Client")
	credentailManger := config.AuthManager
	creds, err := credentailManger.LoadCredentialsById(uint64(gparam.CredentialId))
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("S3 credentail load failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}
	ncpc, ok := creds.(models.NCPCredentials)
	if !ok {
		logger.Error().Msg("credential load failed")
	}
	s3c, err = config.NewS3ClientWithEndpoint(ncpc.AccessKey, ncpc.SecretKey, gparam.Region, gparam.Endpoint)

	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("S3 s3 compatible client creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	logger.Info().Msg("Set up the client as an OSController")
	OSC, err = osc.New(s3fs.New(models.NCP, s3c, gparam.Bucket, gparam.Region))
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("OSController creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	return OSC
}

func getGCPCOSC(logger *zerolog.Logger, startTime time.Time, jobType string, params interface{}) *osc.OSController {
	gparam, _ := params.(models.ProviderConfig)

	var err error
	var gcpOSC *osc.OSController

	logger.Info().Msg("Get GCP Client")
	credentailManger := config.AuthManager
	creds, err := credentailManger.LoadCredentialsById(uint64(gparam.CredentialId))
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("gcp credentail load failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}
	gcpc, ok := creds.(models.GCPCredentials)
	if !ok {
		logger.Error().Msg("credential load failed")
		return nil
	}
	credentialsJson, err := json.Marshal(gcpc)
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("gcp credentail json Marshal failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	gc, err := config.NewGCPClient(string(credentialsJson))
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("gcp client creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	logger.Info().Msg("Set up the client as an OSController")
	if jobType == "gen" {
		gcpOSC, err = osc.New(gcpfs.New(gc, gcpc.ProjectID, gparam.Bucket, gparam.Region))
	} else {
		gcpOSC, err = osc.New(gcpfs.New(gc, gcpc.ProjectID, gparam.Bucket, gparam.Region))
	}
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("OSController creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	return gcpOSC
}

func getMysqlRDBC(logger *zerolog.Logger, startTime time.Time, jobType string, params interface{}) *rdbc.RDBController {
	gparam, _ := params.(models.ProviderConfig)

	var err error
	var sqlDB *sql.DB
	var RDBC *rdbc.RDBController

	logger.Info().Msgf("Get SQL Client %v", jobType)
	sqlDB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", gparam.User, gparam.Password, gparam.Host, gparam.Port))
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("sqlDB client creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	logger.Info().Msg("Set up the client as an RDBController")
	RDBC, err = rdbc.New(mysql.New(models.Provider(gparam.Provider), sqlDB))
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("RDBController creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	return RDBC
}

func getDynamoNRDBC(logger *zerolog.Logger, startTime time.Time, jobType string, params interface{}) *nrdbc.NRDBController {
	gparam, _ := params.(models.ProviderConfig)

	var err error
	var dc *dynamodb.Client
	var NRDBC *nrdbc.NRDBController

	logger.Info().Msg("Get DynamoDB Client")
	credentailManger := config.AuthManager
	creds, err := credentailManger.LoadCredentialsById(uint64(gparam.CredentialId))
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("aws credentail load failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}
	awsc, ok := creds.(models.AWSCredentials)
	if !ok {
		logger.Error().Msg("credential load failed")
	}
	if jobType == "gen" {
		dc, err = config.NewDynamoDBClient(awsc.AccessKey, awsc.SecretKey, gparam.Region)
	} else {
		dc, err = config.NewDynamoDBClient(awsc.AccessKey, awsc.SecretKey, gparam.Region)
	}
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("dynamoDB client creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	logger.Info().Msg("Set up the client as an NRDBController")
	if jobType == "gen" {
		NRDBC, err = nrdbc.New(awsdnmdb.New(dc, gparam.Region))
	} else {
		NRDBC, err = nrdbc.New(awsdnmdb.New(dc, gparam.Region))
	}
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("NRDBController creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	return NRDBC
}

func getFirestoreNRDBC(logger *zerolog.Logger, startTime time.Time, jobType string, params interface{}) *nrdbc.NRDBController {
	gparam, _ := params.(models.ProviderConfig)

	var err error
	var fc *firestore.Client
	var NRDBC *nrdbc.NRDBController

	logger.Info().Msg("Get FirestoreDB Client")

	credentailManger := config.AuthManager
	creds, err := credentailManger.LoadCredentialsById(uint64(gparam.CredentialId))
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("gcp credentail load failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}
	gcpc, ok := creds.(models.GCPCredentials)
	if !ok {
		logger.Error().Msg("credential load failed")
		return nil
	}
	credentialsJson, err := json.Marshal(gcpc)
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("gcp credentail json Marshal failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	fc, err = config.NewFireStoreClient(string(credentialsJson), gcpc.ProjectID, gparam.DatabaseID)

	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("firestoreDB client creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	logger.Info().Msg("Set up the client as an NRDBController")
	NRDBC, err = nrdbc.New(gcpfsdb.New(fc, gparam.Region))

	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("NRDBController creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	return NRDBC
}

func getMongoNRDBC(logger *zerolog.Logger, startTime time.Time, jobType string, params interface{}) *nrdbc.NRDBController {
	gparam, _ := params.(models.ProviderConfig)

	var err error
	var mc *mongo.Client
	var NRDBC *nrdbc.NRDBController

	logger.Info().Msg("Get MongoDB Client")
	if jobType == "gen" {
		mc, err = config.NewNCPMongoDBClient(gparam.User, gparam.Password, gparam.Host, cast.ToInt(gparam.Port))
	} else {
		mc, err = config.NewNCPMongoDBClient(gparam.User, gparam.Password, gparam.Host, cast.ToInt(gparam.Port))
	}
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("mongoDB client creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	logger.Info().Msg("Set up the client as an NRDBController")
	if jobType == "gen" {
		NRDBC, err = nrdbc.New(ncpmgdb.New(mc, gparam.DatabaseName))
	} else {
		NRDBC, err = nrdbc.New(ncpmgdb.New(mc, gparam.DatabaseName))
	}
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("NRDBController creation failed")
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return nil
	}

	return NRDBC
}

func nrdbPutWorker(logger *zerolog.Logger, startTime time.Time, dbType string, nrdbc *nrdbc.NRDBController, jsonList []string) bool {
	var wg sync.WaitGroup
	var mu sync.Mutex
	ret := make(chan error)

	logger.Info().Msgf("Start Import with %s", dbType)
	for _, j := range jsonList {
		wg.Add(1)
		go func(jPath string, jret chan<- error) {
			defer wg.Done()

			mu.Lock()
			logger.Info().Msgf("Read json file : %s", jPath)
			mu.Unlock()

			data, err := os.ReadFile(jPath)
			if err != nil {
				jret <- err
				return
			}

			logger.Info().Msgf("data unmarshal : %s", filepath.Base(jPath))
			var jsonData []map[string]interface{}
			err = json.Unmarshal(data, &jsonData)
			if err != nil {
				jret <- err
				return
			}

			tableName := strings.TrimSuffix(filepath.Base(jPath), ".json")

			mu.Lock()
			logger.Info().Msgf("Put start : %s", filepath.Base(jPath))
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
			logger.Error().Err(result).Msgf("NRDBController Import failedd: %v", result)
			logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Info().Str("Elapsed time", end.Sub(startTime).String())
			return false
		}
	}

	return true
}

func walk(logger *zerolog.Logger, startTime time.Time, list *[]string, dirPath string, ext string) bool {
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
		logger.Error().Err(err).Msgf("filepath walk failed: %v", err)
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return false
	}
	return true
}

func oscImport(logger *zerolog.Logger, startTime time.Time, osType string, osc *osc.OSController, dstDir string) bool {
	logger.Info().Msgf("Start Import with %s", osType)
	if err := osc.MPut(dstDir); err != nil {
		end := time.Now()
		logger.Error().Err(err).Msgf("OSController import failed : %v", err)
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
		return false
	}
	return true
}

func oscExport(logger *zerolog.Logger, startTime time.Time, osType string, osc *osc.OSController, dstDir string, flt *filtering.ObjectFilter) bool {
	logger.Info().Msgf("Start Export with %s", osType)
	if err := osc.MGet(dstDir, flt); err != nil {
		end := time.Now()
		logger.Error().Err(err).Msgf("OSController export failed: %v", err)
		logger.Info().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Info().Str("Elapsed time", end.Sub(startTime).String())
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

		params := models.GenDataParams{}
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

		params := models.GenFileParams{}
		json.Unmarshal(data, &params)
		return params
	} else {

		return nil
	}
}

// Bind onetime
func getDataWithBind(logger *zerolog.Logger, startTime time.Time, ctx echo.Context, params interface{}) bool {

	if err := ctx.Bind(params); err != nil {
		end := time.Now()
		logger.Error().Msg("Failed to bind form data")
		logger.Error().Msgf("params : %+v", ctx.Request().Body)
		logger.Error().Str("End time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Error().Str("Elapsed time", end.Sub(startTime).String())
		return false
	}
	return true
}

// For Rebind
func getDataWithReBind(logger *zerolog.Logger, startTime time.Time, ctx echo.Context, params interface{}) bool {

	bodyBytes, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		logger.Error().Msg("Failed to read request body")
		return false
	}

	ctx.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := ctx.Bind(params); err != nil {
		end := time.Now()
		// fmt.Println("error: ", err.Error())
		logger.Error().Err(err)
		logger.Error().Msg("Failed to bind form data")
		logger.Error().Interface("Params", string(bodyBytes))
		logger.Error().Str("End time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Error().Str("Elapsed time", end.Sub(startTime).String())
		return false
	}

	ctx.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return true
}

func gcpCreateCredFile(logger *zerolog.Logger, startTime time.Time, ctx echo.Context) (string, string, bool) {
	logger.Info().Msg("Create a temporary directory where credential files will be stored")
	gcpCredentialHeader, err := ctx.FormFile("gcpCredential")
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("Get CredentialFile error")
		logger.Error().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Error().Str("Elapsed time", end.Sub(startTime).String())
		return "", "", false
	}

	credTmpDir, err := os.MkdirTemp("", "datamold-gcp-cred-")
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("Get CredentialFile error")
		logger.Error().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Error().Str("Elapsed time", end.Sub(startTime).String())
		return "", "", false
	}

	credFileName := filepath.Join(credTmpDir, gcpCredentialHeader.Filename)
	gcpCredentialFile, err := gcpCredentialHeader.Open()
	// err = ctx.SaveUploadedFile(gcpCredentialHeader, credFileName)
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("Get CredentialFile error")
		logger.Error().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Error().Str("Elapsed time", end.Sub(startTime).String())
		return "", "", false
	}
	defer gcpCredentialFile.Close()

	dst, err := os.Create(credFileName)
	if err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("File create error")
		logger.Error().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Error().Str("Elapsed time", end.Sub(startTime).String())
		return "", "", false
	}
	defer dst.Close()

	if _, err = io.Copy(dst, gcpCredentialFile); err != nil {
		end := time.Now()
		logger.Error().Err(err).Msg("File copy error")
		logger.Error().Str("end time", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Error().Str("Elapsed time", end.Sub(startTime).String())
		return "", "", false
	}

	return credTmpDir, credFileName, true
}
