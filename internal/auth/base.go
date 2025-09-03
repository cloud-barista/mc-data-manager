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
	"github.com/rs/zerolog/log"
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
	log.Info().Msg("initiate a profile scan")
	credentailMangeer := config.NewProfileManager()
	if srcCreds, err := credentailMangeer.LoadCredentialsByProfile(params.ProfileName, params.Provider); err != nil {
		return fmt.Errorf("get config error : %s", err)

	} else {
		log.Info().Interface("credentials", srcCreds).Msg("initiate a profile scan")
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
	log.Info().Msg("initiate a configuration scan")
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
	log.Info().Msgf("launch an %s to %s", use, task)
	err := preRunE(use, task, datamoldParams)
	if err != nil {
		log.Error().Err(err).Msgf("Pre-check for %s operation errors", task)
		os.Exit(1)
	}
	log.Info().Msgf("successful pre-check %s into %s", use, task)
}

func GetOS(params *models.ProviderConfig) (*osc.OSController, error) {
	var OSC *osc.OSController
	log.Info().Str("ProfileName", params.ProfileName).Msg("GetOS")
	log.Info().Str("Provider", params.Provider).Msg("GetOS")
	log.Info().Msg("Get  Credential")
	credentailManger := config.NewProfileManager()
	// creds, err := credentailManger.LoadCredentialsByProfile(params.ProfileName, params.Provider)
	creds, err := credentailManger.LoadCredentialsById(uint64(params.CredentialId), params.Provider)
	if err != nil {
		log.Error().Err(err).Msg("credential load failed")
		return nil, err
	}

	switch params.Provider {
	case "aws":
		awsc, ok := creds.(models.AWSCredentials)
		if !ok {
			return nil, errors.New("credential load failed")
		}

		log.Info().Str("AccessKey", awsc.AccessKey).Msg("AWS Credentials")
		log.Info().Str("SecretKey", awsc.SecretKey).Msg("AWS Credentials")
		log.Info().Str("Region", params.Region).Msg("AWS Region")
		log.Info().Str("BucketName", params.Bucket).Msg("AWS BucketName")
		s3c, err := config.NewS3Client(awsc.AccessKey, awsc.SecretKey, params.Region)
		if err != nil {
			return nil, fmt.Errorf("NewS3Client error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(models.AWS, s3c, params.Bucket, params.Region))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	case "gcp":
		gcpc, ok := creds.(models.GCPCredentials)
		if !ok {
			return nil, errors.New("credential load failed")
		}

		log.Info().Str("ProjectID", gcpc.ProjectID).Msg("GCP Project")
		credentialsJson, err := json.Marshal(gcpc)
		if err != nil {
			return nil, err
		}

		log.Info().Str("Region", params.Region).Msg("GCP Region")
		log.Info().Str("BucketName", params.Bucket).Msg("GCP BucketName")

		gc, err := config.NewGCPClient(string(credentialsJson))
		if err != nil {
			return nil, fmt.Errorf("NewGCPClient error : %v", err)
		}

		OSC, err = osc.New(gcpfs.New(gc, gcpc.ProjectID, params.Bucket, params.Region))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	case "ncp":
		ncpc, ok := creds.(models.NCPCredentials)
		if !ok {
			return nil, errors.New("credential load failed")
		}
		log.Info().Str("AccessKey", ncpc.AccessKey).Msg("NCP Credentials")
		log.Info().Str("SecretKey", ncpc.SecretKey).Msg("NCP Credentials")
		log.Info().Str("Endpoint", params.Endpoint).Msg("NCP Endpoint")
		log.Info().Str("Region", params.Region).Msg("NCP Region")
		log.Info().Str("BucketName", params.Bucket).Msg("NCP BucketName")
		s3c, err := config.NewS3ClientWithEndpoint(ncpc.AccessKey, ncpc.SecretKey, params.Region, params.Endpoint)
		if err != nil {
			return nil, fmt.Errorf("NewS3ClientWithEndpint error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(models.NCP, s3c, params.Bucket, params.Region))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	}
	return OSC, nil
}
func GetRDMS(params *models.ProviderConfig) (*rdbc.RDBController, error) {
	log.Info().Str("Provider", params.Provider).Msg("GetRDMS")
	log.Info().Str("Username", params.User).Msg("GetRDMS")
	log.Info().Str("Password", params.Password).Msg("GetRDMS")
	log.Info().Str("Host", params.Host).Msg("GetRDMS")
	log.Info().Str("Port", params.Port).Msg("GetRDMS")
	dst, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", params.User, params.Password, params.Host, params.Port))
	if err != nil {
		return nil, err
	}
	return rdbc.New(mysql.New(models.Provider(params.Provider), dst))
}

func GetNRDMS(params *models.ProviderConfig) (*nrdbc.NRDBController, error) {
	var NRDBC *nrdbc.NRDBController
	log.Info().Str("ProfileName", params.ProfileName).Msg("GetNRDMS")
	log.Info().Str("Provider", params.Provider).Msg("GetNRDMS")
	log.Info().Msg("Get  Credential")
	credentailManger := config.NewProfileManager()
	// creds, err := credentailManger.LoadCredentialsByProfile(params.ProfileName, params.Provider)
	creds, err := credentailManger.LoadCredentialsById(uint64(params.CredentialId), params.Provider)
	if err != nil {
		log.Error().Err(err).Msg("credential load failed")
		return nil, err
	}

	switch params.Provider {
	case "aws":
		awsc, ok := creds.(models.AWSCredentials)
		if !ok {
			return nil, errors.New("credential load failed")
		}

		log.Info().Str("AccessKey", awsc.AccessKey).Msg("AWS Credentials")
		log.Info().Str("SecretKey", awsc.SecretKey).Msg("AWS Credentials")
		log.Info().Str("Region", params.Region).Msg("AWS Region")
		awsnrdb, err := config.NewDynamoDBClient(awsc.AccessKey, awsc.SecretKey, params.Region)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(awsdnmdb.New(awsnrdb, params.Region))
		if err != nil {
			return nil, err
		}
	case "gcp":
		gcpc, ok := creds.(models.GCPCredentials)
		if !ok {
			return nil, errors.New("credential load failed")
		}

		log.Info().Str("ProjectID", gcpc.ProjectID).Msg("GCP Project")
		log.Info().Str("Region", params.Region).Msg("GCP Region")

		credentialsJson, err := json.Marshal(gcpc)
		if err != nil {
			return nil, err
		}

		gcpnrdb, err := config.NewFireStoreClient(string(credentialsJson), gcpc.ProjectID, params.DatabaseID)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(gcpfsdb.New(gcpnrdb, params.Region))
		if err != nil {
			return nil, err
		}
	case "ncp":
		log.Info().Str("Username", params.User).Msg("NCP Credentials")
		log.Info().Str("Password", params.Password).Msg("NCP Credentials")
		log.Info().Str("Host", params.Host).Msg("NCP Host")
		log.Info().Str("Port", params.Port).Msg("NCP Port")
		port, err := strconv.Atoi(params.Port)
		if err != nil {
			return nil, err
		}

		ncpnrdb, err := config.NewNCPMongoDBClient(params.User, params.Password, params.Host, port)
		if err != nil {
			return nil, err
		}
		NRDBC, err = nrdbc.New(ncpmgdb.New(ncpnrdb, params.DatabaseName))
		if err != nil {
			return nil, err
		}
	}
	return NRDBC, nil
}
