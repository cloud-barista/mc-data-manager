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
package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

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
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

func GetConfig(credPath string, ConfigData *models.CommandTask) error {
	data, err := os.ReadFile(credPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, ConfigData)
	if err != nil {
		return err
	}
	return nil
}

func preRunProfileE(pName, cmdName string, params *models.ProviderConfig) error {
	logrus.Info("initiate a profile scan")
	credentailMangeer := config.NewFileCredentialsManager()
	if srcCreds, err := credentailMangeer.LoadCredentialsByProfile(params.ProfileName, params.Provider); err != nil {
		return fmt.Errorf("get config error : %s", err)

	} else {
		logrus.Infof("initiate a profile scan %v", srcCreds)
	}

	switch cmdName {
	case "objectstorage":
		return nil
	case "rdbms":
		return nil
	case "nrdbms":
		return nil
	default:
		return errors.New("not support service")
	}
}

func preRunE(pName string, cmdName string, params *models.CommandTask) error {
	logrus.Info("initiate a configuration scan")
	if err := GetConfig(params.TaskFilePath, params); err != nil {
		return fmt.Errorf("get config error: %s", err)
	}

	switch cmdName {
	case "objectstorage":
		return nil
	case "rdbms":
		return nil
	case "nrdbms":
		return nil
	default:
		return errors.New("not support service")
	}
}

func PreRun(task string, datamoldParams *models.CommandTask, use string) {
	logrus.SetFormatter(&log.CustomTextFormatter{CmdName: use, JobName: task})
	logrus.Infof("launch an %s to %s", use, task)
	err := preRunE(use, task, datamoldParams)
	if err != nil {
		logrus.Errorf("Pre-check for %s operation errors : %v", task, err)
		os.Exit(1)
	}
	logrus.Infof("successful pre-check %s into %s", use, task)
}

func GetOS(params *models.ProviderConfig) (*osc.OSController, error) {
	var OSC *osc.OSController
	logrus.Infof("ProfileName : %s", params.ProfileName)
	logrus.Infof("Provider : %s", params.Provider)
	logrus.Info("Get  Credentail")
	credentailManger := config.NewFileCredentialsManager()
	creds, err := credentailManger.LoadCredentialsByProfile(params.ProfileName, params.Provider)
	if err != nil {
		logrus.Errorf("credentail load failed : %v", err)

		return nil, err
	}

	if params.Provider == "aws" {
		awsc, ok := creds.(models.AWSCredentials)
		if !ok {
			return nil, errors.New("credential load failed")
		}

		logrus.Infof("AccessKey : %s", awsc.AccessKey)
		logrus.Infof("SecretKey : %s", awsc.SecretKey)
		logrus.Infof("Region : %s", params.Region)
		logrus.Infof("BucketName : %s", params.Bucket)
		s3c, err := config.NewS3Client(awsc.AccessKey, awsc.SecretKey, params.Region)
		if err != nil {
			return nil, fmt.Errorf("NewS3Client error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(models.AWS, s3c, params.Bucket, params.Region), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	} else if params.Provider == "gcp" {
		gcpc, ok := creds.(models.GCPCredentials)
		if !ok {
			return nil, errors.New("credential load failed")
		}

		logrus.Infof("ProjectID : %s", gcpc.ProjectID)

		credentialsJson, err := json.Marshal(gcpc)
		if err != nil {
			return nil, err
		}

		logrus.Infof("Region : %s", params.Region)
		logrus.Infof("BucketName : %s", params.Bucket)

		gc, err := config.NewGCPClient(string(credentialsJson))
		if err != nil {
			return nil, fmt.Errorf("NewGCPClient error : %v", err)
		}

		OSC, err = osc.New(gcpfs.New(gc, gcpc.ProjectID, params.Bucket, params.Region), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	} else if params.Provider == "ncp" {

		ncpc, ok := creds.(models.NCPCredentials)
		if !ok {
			return nil, errors.New("credential load failed")
		}
		logrus.Infof("AccessKey : %s", ncpc.AccessKey)
		logrus.Infof("SecretKey : %s", ncpc.SecretKey)
		logrus.Infof("Endpoint : %s", params.Endpoint)
		logrus.Infof("Region : %s", params.Region)
		logrus.Infof("BucketName : %s", params.Bucket)
		s3c, err := config.NewS3ClientWithEndpoint(ncpc.AccessKey, ncpc.SecretKey, params.Region, params.Endpoint)
		if err != nil {
			return nil, fmt.Errorf("NewS3ClientWithEndpint error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(models.NCP, s3c, params.Bucket, params.Region), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	}
	return OSC, nil
}

func GetRDMS(params *models.ProviderConfig) (*rdbc.RDBController, error) {
	logrus.Infof("Provider : %s", params.Provider)
	logrus.Infof("Username : %s", params.User)
	logrus.Infof("Password : %s", params.Password)
	logrus.Infof("Host : %s", params.Host)
	logrus.Infof("Port : %s", params.Port)
	dst, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", params.User, params.Password, params.Host, params.Port))
	if err != nil {
		return nil, err
	}
	return rdbc.New(mysql.New(models.Provider(params.Provider), dst), rdbc.WithLogger(logrus.StandardLogger()))
}

func GetNRDMS(params *models.ProviderConfig) (*nrdbc.NRDBController, error) {
	var NRDBC *nrdbc.NRDBController
	logrus.Infof("ProfileName : %s", params.ProfileName)
	logrus.Infof("Provider : %s", params.Provider)

	logrus.Info("Get  Credentail")
	credentailManger := config.NewFileCredentialsManager()
	creds, err := credentailManger.LoadCredentialsByProfile(params.ProfileName, params.Provider)
	if err != nil {
		logrus.Errorf("credentail load failed : %v", err)

		return nil, err
	}

	if params.Provider == "aws" {
		awsc, ok := creds.(models.AWSCredentials)
		if !ok {
			return nil, errors.New("credential load failed")
		}

		logrus.Infof("AccessKey : %s", awsc.AccessKey)
		logrus.Infof("SecretKey : %s", awsc.SecretKey)
		logrus.Infof("Region : %s", params.Region)
		awsnrdb, err := config.NewDynamoDBClient(awsc.AccessKey, awsc.SecretKey, params.Region)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(awsdnmdb.New(awsnrdb, params.Region), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	} else if params.Provider == "gcp" {
		gcpc, ok := creds.(models.GCPCredentials)
		if !ok {
			return nil, errors.New("credential load failed")
		}

		logrus.Infof("ProjectID : %s", gcpc.ProjectID)
		logrus.Infof("Region : %s", params.Region)

		credentialsJson, err := json.Marshal(gcpc)
		if err != nil {
			return nil, err
		}

		gcpnrdb, err := config.NewFireStoreClient(string(credentialsJson), gcpc.ProjectID, params.DatabaseID)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(gcpfsdb.New(gcpnrdb, params.Region), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	} else if params.Provider == "ncp" {
		logrus.Infof("Username : %s", params.User)
		logrus.Infof("Password : %s", params.Password)
		logrus.Infof("Host : %s", params.Host)
		logrus.Infof("Port : %s", params.Port)
		port, err := strconv.Atoi(params.Port)
		if err != nil {
			return nil, err
		}

		ncpnrdb, err := config.NewNCPMongoDBClient(params.User, params.Password, params.Host, port)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(ncpmgdb.New(ncpnrdb, params.DatabaseName), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	}
	return NRDBC, nil
}

// func GetSrcOS(datamoldParams *CommandTask) (*osc.OSController, error) {
// 	var OSC *osc.OSController
// 	logrus.Infof("Provider : %s", datamoldParams.SrcProvider.Provider)
// 	if datamoldParams.SrcProvider.Provider == "aws" {
// 		logrus.Infof("AccessKey : %s", datamoldParams.SrcProvider.AccessKey)
// 		logrus.Infof("SecretKey : %s", datamoldParams.SrcProvider.SecretKey)
// 		logrus.Infof("Region : %s", datamoldParams.SrcProvider.Region)
// 		logrus.Infof("BucketName : %s", datamoldParams.SrcProvider.BucketName)
// 		s3c, err := config.NewS3Client(datamoldParams.SrcProvider.AccessKey, datamoldParams.SrcProvider.SecretKey, datamoldParams.SrcProvider.Region)
// 		if err != nil {
// 			return nil, fmt.Errorf("NewS3Client error : %v", err)
// 		}

// 		OSC, err = osc.New(s3fs.New(models.AWS, s3c, datamoldParams.SrcProvider.BucketName, datamoldParams.SrcProvider.Region), osc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, fmt.Errorf("osc error : %v", err)
// 		}
// 	} else if datamoldParams.SrcProvider.Provider == "gcp" {
// 		logrus.Infof("CredentialsFilePath : %s", datamoldParams.SrcProvider.GcpCredPath)
// 		logrus.Infof("ProjectID : %s", datamoldParams.SrcProvider.ProjectID)
// 		logrus.Infof("Region : %s", datamoldParams.SrcProvider.Region)
// 		logrus.Infof("BucketName : %s", datamoldParams.SrcProvider.BucketName)
// 		gc, err := config.NewGCPClient(datamoldParams.SrcProvider.GcpCredPath)
// 		if err != nil {
// 			return nil, fmt.Errorf("NewGCPClient error : %v", err)
// 		}

// 		OSC, err = osc.New(gcpfs.New(gc, datamoldParams.SrcProvider.ProjectID, datamoldParams.SrcProvider.BucketName, datamoldParams.SrcProvider.Region), osc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, fmt.Errorf("osc error : %v", err)
// 		}
// 	} else if datamoldParams.SrcProvider.Provider == "ncp" {
// 		logrus.Infof("AccessKey : %s", datamoldParams.SrcProvider.AccessKey)
// 		logrus.Infof("SecretKey : %s", datamoldParams.SrcProvider.SecretKey)
// 		logrus.Infof("Endpoint : %s", datamoldParams.SrcProvider.Endpoint)
// 		logrus.Infof("Region : %s", datamoldParams.SrcProvider.Region)
// 		logrus.Infof("BucketName : %s", datamoldParams.SrcProvider.BucketName)
// 		s3c, err := config.NewS3ClientWithEndpoint(datamoldParams.SrcProvider.AccessKey, datamoldParams.SrcProvider.SecretKey, datamoldParams.SrcProvider.Region, datamoldParams.SrcProvider.Endpoint)
// 		if err != nil {
// 			return nil, fmt.Errorf("NewS3ClientWithEndpint error : %v", err)
// 		}

// 		OSC, err = osc.New(s3fs.New(models.AWS, s3c, datamoldParams.SrcProvider.BucketName, datamoldParams.SrcProvider.Region), osc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, fmt.Errorf("osc error : %v", err)
// 		}
// 	}
// 	return OSC, nil
// }

// func GetDstOS(datamoldParams *CommandTask) (*osc.OSController, error) {
// 	var OSC *osc.OSController
// 	logrus.Infof("Provider : %s", datamoldParams.DstProvider.Provider)
// 	if datamoldParams.DstProvider.Provider == "aws" {
// 		logrus.Infof("AccessKey : %s", datamoldParams.DstProvider.AccessKey)
// 		logrus.Infof("SecretKey : %s", datamoldParams.DstProvider.SecretKey)
// 		logrus.Infof("Region : %s", datamoldParams.DstProvider.Region)
// 		logrus.Infof("BucketName : %s", datamoldParams.DstProvider.BucketName)
// 		s3c, err := config.NewS3Client(datamoldParams.DstProvider.AccessKey, datamoldParams.DstProvider.SecretKey, datamoldParams.DstProvider.Region)
// 		if err != nil {
// 			return nil, fmt.Errorf("NewS3Client error : %v", err)
// 		}

// 		OSC, err = osc.New(s3fs.New(models.AWS, s3c, datamoldParams.DstProvider.BucketName, datamoldParams.DstProvider.Region), osc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, fmt.Errorf("osc error : %v", err)
// 		}
// 	} else if datamoldParams.DstProvider.Provider == "gcp" {
// 		logrus.Infof("CredentialsFilePath : %s", datamoldParams.DstProvider.GcpCredPath)
// 		logrus.Infof("ProjectID : %s", datamoldParams.DstProvider.ProjectID)
// 		logrus.Infof("Region : %s", datamoldParams.DstProvider.Region)
// 		logrus.Infof("BucketName : %s", datamoldParams.DstProvider.BucketName)
// 		gc, err := config.NewGCPClient(datamoldParams.DstProvider.GcpCredPath)
// 		if err != nil {
// 			return nil, fmt.Errorf("NewGCPClient error : %v", err)
// 		}

// 		OSC, err = osc.New(gcpfs.New(gc, datamoldParams.DstProvider.ProjectID, datamoldParams.DstProvider.BucketName, datamoldParams.DstProvider.Region), osc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, fmt.Errorf("osc error : %v", err)
// 		}
// 	} else if datamoldParams.DstProvider.Provider == "ncp" {
// 		logrus.Infof("AccessKey : %s", datamoldParams.DstProvider.AccessKey)
// 		logrus.Infof("SecretKey : %s", datamoldParams.DstProvider.SecretKey)
// 		logrus.Infof("Endpoint : %s", datamoldParams.DstProvider.Endpoint)
// 		logrus.Infof("Region : %s", datamoldParams.DstProvider.Region)
// 		logrus.Infof("BucketName : %s", datamoldParams.DstProvider.BucketName)
// 		s3c, err := config.NewS3ClientWithEndpoint(datamoldParams.DstProvider.AccessKey, datamoldParams.DstProvider.SecretKey, datamoldParams.DstProvider.Region, datamoldParams.DstProvider.Endpoint)
// 		if err != nil {
// 			return nil, fmt.Errorf("NewS3ClientWithEndpint error : %v", err)
// 		}

// 		OSC, err = osc.New(s3fs.New(models.AWS, s3c, datamoldParams.DstProvider.BucketName, datamoldParams.DstProvider.Region), osc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, fmt.Errorf("osc error : %v", err)
// 		}
// 	}
// 	return OSC, nil
// }

// func GetSrcRDMS(datamoldParams *CommandTask) (*rdbc.RDBController, error) {
// 	logrus.Infof("Provider : %s", datamoldParams.SrcProvider.Provider)
// 	logrus.Infof("Username : %s", datamoldParams.SrcProvider.Username)
// 	logrus.Infof("Password : %s", datamoldParams.SrcProvider.Password)
// 	logrus.Infof("Host : %s", datamoldParams.SrcProvider.Host)
// 	logrus.Infof("Port : %s", datamoldParams.SrcProvider.Port)
// 	src, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", datamoldParams.SrcProvider.Username, datamoldParams.SrcProvider.Password, datamoldParams.SrcProvider.Host, datamoldParams.SrcProvider.Port))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return rdbc.New(mysql.New(models.Provider(datamoldParams.SrcProvider.Provider), src), rdbc.WithLogger(logrus.StandardLogger()))
// }

// func GetDstRDMS(datamoldParams *CommandTask) (*rdbc.RDBController, error) {
// 	logrus.Infof("Provider : %s", datamoldParams.DstProvider.Provider)
// 	logrus.Infof("Username : %s", datamoldParams.DstProvider.Username)
// 	logrus.Infof("Password : %s", datamoldParams.DstProvider.Password)
// 	logrus.Infof("Host : %s", datamoldParams.DstProvider.Host)
// 	logrus.Infof("Port : %s", datamoldParams.DstProvider.Port)
// 	dst, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", datamoldParams.DstProvider.Username, datamoldParams.DstProvider.Password, datamoldParams.DstProvider.Host, datamoldParams.DstProvider.Port))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return rdbc.New(mysql.New(models.Provider(datamoldParams.DstProvider.Provider), dst), rdbc.WithLogger(logrus.StandardLogger()))
// }

// func GetSrcNRDMS(datamoldParams *DatamoldParams) (*nrdbc.NRDBController, error) {
// 	var NRDBC *nrdbc.NRDBController
// 	logrus.Infof("Provider : %s", datamoldParams.SrcProvider.Provider)
// 	if datamoldParams.SrcProvider.Provider == "aws" {
// 		logrus.Infof("AccessKey : %s", datamoldParams.SrcProvider.AccessKey)
// 		logrus.Infof("SecretKey : %s", datamoldParams.SrcProvider.SecretKey)
// 		logrus.Infof("Region : %s", datamoldParams.SrcProvider.Region)
// 		awsnrdb, err := config.NewDynamoDBClient(datamoldParams.SrcProvider.AccessKey, datamoldParams.SrcProvider.SecretKey, datamoldParams.SrcProvider.Region)
// 		if err != nil {
// 			return nil, err
// 		}

// 		NRDBC, err = nrdbc.New(awsdnmdb.New(awsnrdb, datamoldParams.SrcProvider.Region), nrdbc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, err
// 		}
// 	} else if datamoldParams.SrcProvider.Provider == "gcp" {
// 		logrus.Infof("CredentialsFilePath : %s", datamoldParams.SrcProvider.GcpCredPath)
// 		logrus.Infof("ProjectID : %s", datamoldParams.SrcProvider.ProjectID)
// 		logrus.Infof("Region : %s", datamoldParams.SrcProvider.Region)
// 		gcpnrdb, err := config.NewFireStoreClient(datamoldParams.SrcProvider.GcpCredJson, datamoldParams.SrcProvider.ProjectID, datamoldParams.SrcProvider.DatabaseID)
// 		if err != nil {
// 			return nil, err
// 		}

// 		NRDBC, err = nrdbc.New(gcpfsdb.New(gcpnrdb, datamoldParams.SrcProvider.Region), nrdbc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, err
// 		}
// 	} else if datamoldParams.SrcProvider.Provider == "ncp" {
// 		logrus.Infof("Username : %s", datamoldParams.SrcProvider.Username)
// 		logrus.Infof("Password : %s", datamoldParams.SrcProvider.Password)
// 		logrus.Infof("Host : %s", datamoldParams.SrcProvider.Host)
// 		logrus.Infof("Port : %s", datamoldParams.SrcProvider.Port)
// 		port, err := strconv.Atoi(datamoldParams.SrcProvider.Port)
// 		if err != nil {
// 			return nil, err
// 		}

// 		ncpnrdb, err := config.NewNCPMongoDBClient(datamoldParams.SrcProvider.Username, datamoldParams.SrcProvider.Password, datamoldParams.SrcProvider.Host, port)
// 		if err != nil {
// 			return nil, err
// 		}

// 		NRDBC, err = nrdbc.New(ncpmgdb.New(ncpnrdb, datamoldParams.SrcProvider.DBName), nrdbc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return NRDBC, nil
// }

// func GetDstNRDMS(datamoldParams *DatamoldParams) (*nrdbc.NRDBController, error) {
// 	var NRDBC *nrdbc.NRDBController
// 	logrus.Infof("Provider : %s", datamoldParams.DstProvider.Provider)
// 	if datamoldParams.DstProvider.Provider == "aws" {
// 		logrus.Infof("AccessKey : %s", datamoldParams.DstProvider.AccessKey)
// 		logrus.Infof("SecretKey : %s", datamoldParams.DstProvider.SecretKey)
// 		logrus.Infof("Region : %s", datamoldParams.DstProvider.Region)
// 		awsnrdb, err := config.NewDynamoDBClient(datamoldParams.DstProvider.AccessKey, datamoldParams.DstProvider.SecretKey, datamoldParams.DstProvider.Region)
// 		if err != nil {
// 			return nil, err
// 		}

// 		NRDBC, err = nrdbc.New(awsdnmdb.New(awsnrdb, datamoldParams.DstProvider.Region), nrdbc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, err
// 		}
// 	} else if datamoldParams.DstProvider.Provider == "gcp" {
// 		logrus.Infof("CredentialsFilePath : %s", datamoldParams.DstProvider.GcpCredPath)
// 		logrus.Infof("ProjectID : %s", datamoldParams.DstProvider.ProjectID)
// 		logrus.Infof("Region : %s", datamoldParams.DstProvider.Region)
// 		gcpnrdb, err := config.NewFireStoreClient(datamoldParams.DstProvider.GcpCredJson, datamoldParams.DstProvider.ProjectID, datamoldParams.DstProvider.DatabaseID)
// 		if err != nil {
// 			return nil, err
// 		}

// 		NRDBC, err = nrdbc.New(gcpfsdb.New(gcpnrdb, datamoldParams.DstProvider.Region), nrdbc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, err
// 		}
// 	} else if datamoldParams.DstProvider.Provider == "ncp" {
// 		logrus.Infof("Username : %s", datamoldParams.DstProvider.Username)
// 		logrus.Infof("Password : %s", datamoldParams.DstProvider.Password)
// 		logrus.Infof("Host : %s", datamoldParams.DstProvider.Host)
// 		logrus.Infof("Port : %s", datamoldParams.DstProvider.Port)
// 		port, err := strconv.Atoi(datamoldParams.DstProvider.Port)
// 		if err != nil {
// 			return nil, err
// 		}

// 		ncpnrdb, err := config.NewNCPMongoDBClient(datamoldParams.DstProvider.Username, datamoldParams.DstProvider.Password, datamoldParams.DstProvider.Host, port)
// 		if err != nil {
// 			return nil, err
// 		}

// 		NRDBC, err = nrdbc.New(ncpmgdb.New(ncpnrdb, datamoldParams.DstProvider.DBName), nrdbc.WithLogger(logrus.StandardLogger()))
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return NRDBC, nil
// }

// func handleObjectStorage(pName string, datamoldParams *CommandTask) error {
// 	if value, ok := datamoldParams.ConfigData["objectstorage"]; ok {
// 		if !datamoldParams.TaskTarget {
// 			if src, ok := value["src"]; ok {
// 				if err := applyOSValue(src, "src", datamoldParams); err != nil {
// 					return err
// 				}
// 			}
// 		} else {
// 			if dst, ok := value["dst"]; ok {
// 				if err := applyOSValue(dst, "dst", datamoldParams); err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	} else {
// 		return errors.New("does not exist objectstorage")
// 	}

// 	return validateDestinationPath(pName, datamoldParams)
// }

// func handleRDBMS(pName string, datamoldParams *CommandTask) error {
// 	if value, ok := datamoldParams.ConfigData["rdbms"]; ok {
// 		if !datamoldParams.TaskTarget {
// 			if src, ok := value["src"]; ok {
// 				if err := applyRDMValue(src, "src", datamoldParams); err != nil {
// 					return err
// 				}
// 			}
// 		} else {
// 			if dst, ok := value["dst"]; ok {
// 				if err := applyRDMValue(dst, "dst", datamoldParams); err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	} else {
// 		return errors.New("does not exist rdbms src")
// 	}

// 	return validateDestinationPath(pName, datamoldParams)
// }

// func handleNRDBMS(pName string, datamoldParams *CommandTask) error {
// 	if value, ok := datamoldParams.ConfigData["nrdbms"]; ok {
// 		if !datamoldParams.TaskTarget {
// 			if src, ok := value["src"]; ok {
// 				if err := applyNRDMValue(src, "src", datamoldParams); err != nil {
// 					return err
// 				}
// 			}
// 		} else {
// 			if dst, ok := value["dst"]; ok {
// 				if err := applyNRDMValue(dst, "dst", datamoldParams); err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	} else {
// 		return errors.New("does not exist nrdbms src")
// 	}

// 	return validateDestinationPath(pName, datamoldParams)
// }

// func validateDestinationPath(pName string, datamoldParams *CommandTask) error {
// 	if pName != "migration" && pName != "delete" {
// 		if err := utils.IsDir(datamoldParams.DstPath); err != nil {
// 			return errors.New("dstPath error")
// 		}
// 	}
// 	return nil
// }
// func applyNRDMValue(src map[string]string, p string, datamoldParams *CommandTask) error {
// 	provider, ok := src["provider"]
// 	if !ok {
// 		return errors.New("does not exist provider")
// 	}

// 	if provider != "aws" && provider != "gcp" && provider != "ncp" {
// 		return fmt.Errorf("provider error : %s", provider)
// 	}

// 	var access, secret, region, cred, projectID, username, password, host, port, DBName string

// 	switch provider {
// 	case "aws":
// 		access, ok = src["assessKey"]
// 		if !ok {
// 			return errors.New("does not exist assessKey")
// 		}

// 		secret, ok = src["secretKey"]
// 		if !ok {
// 			return errors.New("does not exist secretKey")
// 		}

// 		region, ok = src["region"]
// 		if !ok {
// 			return errors.New("does not exist region")
// 		}

// 	case "gcp":
// 		cred, ok = src["gcpCredPath"]
// 		if !ok {
// 			return errors.New("does not exist gcpCredPath")
// 		}

// 		projectID, ok = src["projectID"]
// 		if !ok {
// 			return errors.New("does not exist projectID")
// 		}

// 		region, ok = src["region"]
// 		if !ok {
// 			return errors.New("does not exist region")
// 		}

// 	case "ncp":
// 		username, ok = src["username"]
// 		if !ok {
// 			return errors.New("does not exist username")
// 		}

// 		password, ok = src["password"]
// 		if !ok {
// 			return errors.New("does not exist password")
// 		}

// 		host, ok = src["host"]
// 		if !ok {
// 			return errors.New("does not exist host")
// 		}

// 		port, ok = src["port"]
// 		if !ok {
// 			return errors.New("does not exist port")
// 		}

// 		DBName, ok = src["databaseName"]
// 		if !ok {
// 			return errors.New("does not exist databaseName")
// 		}
// 	}

// 	if p == "src" {
// 		datamoldParams.SrcProvider.Provider = provider
// 		if provider == "aws" {
// 			datamoldParams.SrcProvider.AccessKey = access
// 			datamoldParams.SrcProvider.SecretKey = secret
// 			datamoldParams.SrcProvider.Region = region
// 		} else if provider == "gcp" {
// 			datamoldParams.SrcProvider.GcpCredPath = cred
// 			datamoldParams.SrcProvider.ProjectID = projectID
// 			datamoldParams.SrcProvider.Region = region
// 		} else if provider == "ncp" {
// 			datamoldParams.SrcProvider.Username = username
// 			datamoldParams.SrcProvider.Password = password
// 			datamoldParams.SrcProvider.Host = host
// 			datamoldParams.SrcProvider.Port = port
// 			datamoldParams.SrcProvider.DBName = DBName
// 		}
// 	} else {
// 		datamoldParams.DstProvider.Provider = provider
// 		if provider == "aws" {
// 			datamoldParams.DstProvider.AccessKey = access
// 			datamoldParams.DstProvider.SecretKey = secret
// 			datamoldParams.DstProvider.Region = region
// 		} else if provider == "gcp" {
// 			datamoldParams.DstProvider.GcpCredPath = cred
// 			datamoldParams.DstProvider.ProjectID = projectID
// 			datamoldParams.DstProvider.Region = region
// 		} else if provider == "ncp" {
// 			datamoldParams.DstProvider.Username = username
// 			datamoldParams.DstProvider.Password = password
// 			datamoldParams.DstProvider.Host = host
// 			datamoldParams.DstProvider.Port = port
// 			datamoldParams.DstProvider.DBName = DBName
// 		}
// 	}

// 	return nil
// }

// func applyRDMValue(src map[string]string, p string, datamoldParams *DatamoldParams) error {
// 	provider, ok := src["provider"]
// 	if !ok {
// 		return errors.New("does not exist provider")
// 	}

// 	if provider != "aws" && provider != "gcp" && provider != "ncp" {
// 		return fmt.Errorf("provider error : %s", provider)
// 	}

// 	var username, password, host, port string

// 	username, ok = src["username"]
// 	if !ok {
// 		return errors.New("does not exist username")
// 	}

// 	password, ok = src["password"]
// 	if !ok {
// 		return errors.New("does not exist password")
// 	}

// 	host, ok = src["host"]
// 	if !ok {
// 		return errors.New("does not exist host")
// 	}

// 	port, ok = src["port"]
// 	if !ok {
// 		return errors.New("does not exist port")
// 	}

// 	if p == "src" {
// 		datamoldParams.SrcProvider.Provider = provider
// 		datamoldParams.SrcProvider.Username = username
// 		datamoldParams.SrcProvider.Password = password
// 		datamoldParams.SrcProvider.Host = host
// 		datamoldParams.SrcProvider.Port = port
// 	} else {
// 		datamoldParams.DstProvider.Provider = provider
// 		datamoldParams.DstProvider.Username = username
// 		datamoldParams.DstProvider.Password = password
// 		datamoldParams.DstProvider.Host = host
// 		datamoldParams.DstProvider.Port = port
// 	}

// 	return nil
// }

// func applyOSValue(src map[string]string, p string, datamoldParams *DatamoldParams) error {
// 	type Provider string

// 	const (
// 		AWS Provider = "aws"
// 		GCP Provider = "gcp"
// 		NCP Provider = "ncp"
// 	)

// 	providerStr, ok := src["provider"]
// 	if !ok {
// 		return errors.New("does not exist provider")
// 	}

// 	provider := Provider(providerStr)
// 	switch provider {
// 	case AWS, GCP, NCP:
// 	default:
// 		return fmt.Errorf("provider[aws,gcp,ncp] error : %s", provider)
// 	}

// 	var access, secret, region, bktName, cred, projectID, endpoint string

// 	switch provider {
// 	case AWS, NCP:
// 		access, ok = src["assessKey"]
// 		if !ok {
// 			return errors.New("does not exist assessKey")
// 		}

// 		secret, ok = src["secretKey"]
// 		if !ok {
// 			return errors.New("does not exist secretKey")
// 		}

// 		region, ok = src["region"]
// 		if !ok {
// 			return errors.New("does not exist region")
// 		}

// 		bktName, ok = src["bucketName"]
// 		if !ok {
// 			return errors.New("does not exist bucketName")
// 		}

// 		if provider == NCP {
// 			endpoint, ok = src["endpoint"]
// 			if !ok {
// 				return errors.New("does not exist endpoint")
// 			}
// 		}

// 	case GCP:
// 		cred, ok = src["gcpCredPath"]
// 		if !ok {
// 			return errors.New("does not exist gcpCredPath")
// 		}

// 		projectID, ok = src["projectID"]
// 		if !ok {
// 			return errors.New("does not exist projectID")
// 		}

// 		region, ok = src["region"]
// 		if !ok {
// 			return errors.New("does not exist region")
// 		}

// 		bktName, ok = src["bucketName"]
// 		if !ok {
// 			return errors.New("does not exist bucketName")
// 		}
// 	}

// 	if p == "src" {
// 		datamoldParams.SrcProvider.Provider = providerStr
// 		datamoldParams.SrcProvider.AccessKey = access
// 		datamoldParams.SrcProvider.SecretKey = secret
// 		datamoldParams.SrcProvider.Region = region
// 		datamoldParams.SrcProvider.BucketName = bktName
// 		if provider == NCP {
// 			datamoldParams.SrcProvider.Endpoint = endpoint
// 		}
// 		if provider == GCP {
// 			datamoldParams.SrcProvider.GcpCredPath = cred
// 			datamoldParams.SrcProvider.ProjectID = projectID
// 		}
// 	} else {
// 		datamoldParams.DstProvider.Provider = providerStr
// 		datamoldParams.DstProvider.AccessKey = access
// 		datamoldParams.DstProvider.SecretKey = secret
// 		datamoldParams.DstProvider.Region = region
// 		datamoldParams.DstProvider.BucketName = bktName
// 		if provider == NCP {
// 			datamoldParams.DstProvider.Endpoint = endpoint
// 		}
// 		if provider == GCP {
// 			datamoldParams.DstProvider.GcpCredPath = cred
// 			datamoldParams.DstProvider.ProjectID = projectID
// 		}
// 	}

// 	return nil
// }
