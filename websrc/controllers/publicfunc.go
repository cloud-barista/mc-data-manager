package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	"go.mongodb.org/mongo-driver/mongo"
)

func getLogger(jobName string) *logrus.Logger {
	logger := logrus.StandardLogger()
	logger.SetFormatter(&logformatter.CustomTextFormatter{CmdName: "server", JobName: jobName})
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

func dummyCreate(logger *logrus.Logger, startTime time.Time, params GenDataParams) bool {
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
	gparam, _ := params.(GenDataParams)
	mparam, _ := params.(MigrationForm)

	var err error
	var s3c *s3.Client
	var awsOSC *osc.OSController

	logger.Info("Get S3 Client")
	if jobType == "gen" {
		s3c, err = config.NewS3Client(gparam.AccessKey, gparam.SecretKey, gparam.Region)
	} else {
		s3c, err = config.NewS3Client(mparam.AWSAccessKey, mparam.AWSSecretKey, mparam.AWSRegion)
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("s3 client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an OSController")
	if jobType == "gen" {
		awsOSC, err = osc.New(s3fs.New(utils.AWS, s3c, gparam.Bucket, gparam.Region), osc.WithLogger(logger))
	} else {
		awsOSC, err = osc.New(s3fs.New(utils.AWS, s3c, mparam.AWSBucket, mparam.AWSRegion), osc.WithLogger(logger))
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
	gparam, _ := params.(GenDataParams)
	mparam, _ := params.(MigrationForm)

	var err error
	var s3c *s3.Client
	var OSC *osc.OSController

	logger.Info("Get S3 Compataible Client")
	if jobType == "gen" {
		s3c, err = config.NewS3ClientWithEndpoint(gparam.AccessKey, gparam.SecretKey, gparam.Region, gparam.Endpoint)
	} else {
		s3c, err = config.NewS3ClientWithEndpoint(mparam.NCPAccessKey, mparam.NCPSecretKey, mparam.NCPRegion, mparam.NCPEndPoint)
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("s3 compatible client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an OSController")
	if jobType == "gen" {
		OSC, err = osc.New(s3fs.New(utils.AWS, s3c, gparam.Bucket, gparam.Region), osc.WithLogger(logger))
	} else {
		OSC, err = osc.New(s3fs.New(utils.AWS, s3c, mparam.AWSBucket, mparam.AWSRegion), osc.WithLogger(logger))
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("OSController creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	return OSC
}

func getGCSCOSC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}, credFileName string) *osc.OSController {
	gparam, _ := params.(GenDataParams)
	mparam, _ := params.(MigrationForm)

	var err error
	var gcsOSC *osc.OSController

	logger.Info("Get GCS Client")
	gc, err := config.NewGCSClient(credFileName)
	if err != nil {
		end := time.Now()
		logger.Errorf("gcs client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an OSController")
	if jobType == "gen" {
		gcsOSC, err = osc.New(gcsfs.New(gc, gparam.ProjectID, gparam.Bucket, gparam.Region), osc.WithLogger(logger))
	} else {
		gcsOSC, err = osc.New(gcsfs.New(gc, mparam.ProjectID, mparam.GCPBucket, mparam.GCPRegion), osc.WithLogger(logger))
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("OSController creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	return gcsOSC
}

func getMysqlRDBC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}) *rdbc.RDBController {
	gparam, _ := params.(GenDataParams)
	mparam, _ := params.(MigrationMySQLParams)

	var err error
	var sqlDB *sql.DB
	var RDBC *rdbc.RDBController

	if jobType == "gen" {
		logger.Info("Get SQL Client")
		sqlDB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", gparam.DBUser, gparam.DBPassword, gparam.DBHost, gparam.DBPort))
	} else if jobType == "smig" {
		logger.Info("Get Source SQL Client")
		sqlDB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", mparam.Source.Username, mparam.Source.Password, mparam.Source.Host, mparam.Source.Port))
	} else if jobType == "tmig" {
		logger.Info("Get Target SQL Client")
		sqlDB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", mparam.Dest.Username, mparam.Dest.Password, mparam.Dest.Host, mparam.Dest.Port))
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("sqlDB client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an RDBController")
	if jobType == "gen" {
		logger.Info("Set up the client as an RDBController")
		RDBC, err = rdbc.New(mysql.New(utils.Provider(gparam.DBProvider), sqlDB), rdbc.WithLogger(logger))
	} else if jobType == "smig" {
		logger.Info("Set up the client as an Source RDBController")
		RDBC, err = rdbc.New(mysql.New(utils.Provider(mparam.Source.Provider), sqlDB), rdbc.WithLogger(logger))
	} else if jobType == "tmig" {
		logger.Info("Set up the client as an Target RDBController")
		RDBC, err = rdbc.New(mysql.New(utils.Provider(mparam.Dest.Provider), sqlDB), rdbc.WithLogger(logger))
	}

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
	gparam, _ := params.(GenDataParams)
	mparam, _ := params.(MigrationForm)

	var err error
	var dc *dynamodb.Client
	var NRDBC *nrdbc.NRDBController

	logger.Info("Get DynamoDB Client")
	if jobType == "gen" {
		dc, err = config.NewDynamoDBClient(gparam.AccessKey, gparam.SecretKey, gparam.Region)
	} else {
		dc, err = config.NewDynamoDBClient(mparam.AWSAccessKey, mparam.AWSSecretKey, mparam.AWSRegion)
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
		NRDBC, err = nrdbc.New(awsdnmdb.New(dc, mparam.AWSRegion), nrdbc.WithLogger(logger))
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

func getFirestoreNRDBC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}, credFileName string) *nrdbc.NRDBController {
	gparam, _ := params.(GenDataParams)
	mparam, _ := params.(MigrationForm)

	var err error
	var fc *firestore.Client
	var NRDBC *nrdbc.NRDBController

	logger.Info("Get FirestoreDB Client")
	if jobType == "gen" {
		fc, err = config.NewFireStoreClient(credFileName, gparam.ProjectID)
	} else {
		fc, err = config.NewFireStoreClient(credFileName, mparam.ProjectID)
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("firestoreDB client creation failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Set up the client as an NRDBController")
	if jobType == "gen" {
		NRDBC, err = nrdbc.New(gcpfsdb.New(fc, gparam.Region), nrdbc.WithLogger(logger))
	} else {
		NRDBC, err = nrdbc.New(gcpfsdb.New(fc, mparam.GCPRegion), nrdbc.WithLogger(logger))
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

func getMongoNRDBC(logger *logrus.Logger, startTime time.Time, jobType string, params interface{}) *nrdbc.NRDBController {
	gparam, _ := params.(GenDataParams)
	mparam, _ := params.(MigrationForm)

	var Port int
	var err error
	var mc *mongo.Client
	var NRDBC *nrdbc.NRDBController

	if jobType == "gen" {
		Port, err = strconv.Atoi(gparam.DBPort)
	} else {
		Port, err = strconv.Atoi(mparam.MongoPort)
	}
	if err != nil {
		end := time.Now()
		logger.Errorf("port atoi failed : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return nil
	}

	logger.Info("Get MongoDB Client")
	if jobType == "gen" {
		mc, err = config.NewNCPMongoDBClient(gparam.DBUser, gparam.DBPassword, gparam.DBHost, Port)
	} else {
		mc, err = config.NewNCPMongoDBClient(mparam.MongoUsername, mparam.MongoPassword, mparam.MongoHost, Port)
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
		NRDBC, err = nrdbc.New(ncpmgdb.New(mc, mparam.MongoDBName), nrdbc.WithLogger(logger))
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

func getData(jobtype string, ctx *gin.Context) interface{} {
	if jobtype == "gen" {
		data, _ := ctx.GetRawData()
		params := GenDataParams{}
		json.Unmarshal(data, &params)
		return params
	} else {

		return nil
	}
}

func getDataWithBind(logger *logrus.Logger, startTime time.Time, ctx *gin.Context, params interface{}) bool {
	if err := ctx.ShouldBind(params); err != nil {
		end := time.Now()
		logger.Error("Failed to bind form data")
		logger.Infof("End time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return false
	}
	return true
}

func gcpCreateCredFile(logger *logrus.Logger, startTime time.Time, ctx *gin.Context) (string, string, bool) {
	logger.Info("Create a temporary directory where credential files will be stored")
	gcsCredentialFile, gcsCredentialHeader, err := ctx.Request.FormFile("gcsCredential")
	if err != nil {
		end := time.Now()
		logger.Errorf("Get CredentialFile error : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return "", "", false
	}
	defer gcsCredentialFile.Close()

	credTmpDir, err := os.MkdirTemp("", "datamold-gcs-cred-")
	if err != nil {
		end := time.Now()
		logger.Errorf("Get CredentialFile error : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return "", "", false
	}

	credFileName := filepath.Join(credTmpDir, gcsCredentialHeader.Filename)
	err = ctx.SaveUploadedFile(gcsCredentialHeader, credFileName)
	if err != nil {
		end := time.Now()
		logger.Errorf("Get CredentialFile error : %v", err)
		logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
		logger.Infof("Elapsed time : %s", end.Sub(startTime).String())
		return "", "", false
	}
	return credTmpDir, credFileName, true
}
