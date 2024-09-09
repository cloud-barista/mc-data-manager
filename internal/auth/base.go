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
	credentailMangeer := config.NewProfileManager()
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
	credentailManger := config.NewProfileManager()
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
	credentailManger := config.NewProfileManager()
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
