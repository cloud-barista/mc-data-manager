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
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbms/awsdnmdb"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbms/gcpfsdb"
	"github.com/cloud-barista/mc-data-manager/pkg/nrdbms/ncpmgdb"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/gcpfs"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/s3fs"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbms/mysql"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/cloud-barista/mc-data-manager/service/nrdbc"
	"github.com/cloud-barista/mc-data-manager/service/osc"
	"github.com/cloud-barista/mc-data-manager/service/rdbc"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

func PreRun(task string, datamoldParams *DatamoldParams, use string) {
	logrus.SetFormatter(&log.CustomTextFormatter{CmdName: use, JobName: task})
	logrus.Infof("launch an %s to %s", use, task)
	err := preRunE(use, task, datamoldParams)
	if err != nil {
		logrus.Errorf("Pre-check for %s operation errors : %v", task, err)
		os.Exit(1)
	}
	logrus.Infof("successful pre-check %s into %s", use, task)
}

func GetSrcOS(datamoldParams *DatamoldParams) (*osc.OSController, error) {
	var OSC *osc.OSController
	logrus.Infof("Provider : %s", datamoldParams.SrcProvider)
	if datamoldParams.SrcProvider == "aws" {
		logrus.Infof("AccessKey : %s", datamoldParams.SrcAccessKey)
		logrus.Infof("SecretKey : %s", datamoldParams.SrcSecretKey)
		logrus.Infof("Region : %s", datamoldParams.SrcRegion)
		logrus.Infof("BucketName : %s", datamoldParams.SrcBucketName)
		s3c, err := config.NewS3Client(datamoldParams.SrcAccessKey, datamoldParams.SrcSecretKey, datamoldParams.SrcRegion)
		if err != nil {
			return nil, fmt.Errorf("NewS3Client error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(utils.AWS, s3c, datamoldParams.SrcBucketName, datamoldParams.SrcRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	} else if datamoldParams.SrcProvider == "gcp" {
		logrus.Infof("CredentialsFilePath : %s", datamoldParams.SrcGcpCredPath)
		logrus.Infof("ProjectID : %s", datamoldParams.SrcProjectID)
		logrus.Infof("Region : %s", datamoldParams.SrcRegion)
		logrus.Infof("BucketName : %s", datamoldParams.SrcBucketName)
		gc, err := config.NewGCPClient(datamoldParams.SrcGcpCredPath)
		if err != nil {
			return nil, fmt.Errorf("NewGCPClient error : %v", err)
		}

		OSC, err = osc.New(gcpfs.New(gc, datamoldParams.SrcProjectID, datamoldParams.SrcBucketName, datamoldParams.SrcRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	} else if datamoldParams.SrcProvider == "ncp" {
		logrus.Infof("AccessKey : %s", datamoldParams.SrcAccessKey)
		logrus.Infof("SecretKey : %s", datamoldParams.SrcSecretKey)
		logrus.Infof("Endpoint : %s", datamoldParams.SrcEndpoint)
		logrus.Infof("Region : %s", datamoldParams.SrcRegion)
		logrus.Infof("BucketName : %s", datamoldParams.SrcBucketName)
		s3c, err := config.NewS3ClientWithEndpoint(datamoldParams.SrcAccessKey, datamoldParams.SrcSecretKey, datamoldParams.SrcRegion, datamoldParams.SrcEndpoint)
		if err != nil {
			return nil, fmt.Errorf("NewS3ClientWithEndpint error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(utils.AWS, s3c, datamoldParams.SrcBucketName, datamoldParams.SrcRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	}
	return OSC, nil
}

func GetDstOS(datamoldParams *DatamoldParams) (*osc.OSController, error) {
	var OSC *osc.OSController
	logrus.Infof("Provider : %s", datamoldParams.DstProvider)
	if datamoldParams.DstProvider == "aws" {
		logrus.Infof("AccessKey : %s", datamoldParams.DstAccessKey)
		logrus.Infof("SecretKey : %s", datamoldParams.DstSecretKey)
		logrus.Infof("Region : %s", datamoldParams.DstRegion)
		logrus.Infof("BucketName : %s", datamoldParams.DstBucketName)
		s3c, err := config.NewS3Client(datamoldParams.DstAccessKey, datamoldParams.DstSecretKey, datamoldParams.DstRegion)
		if err != nil {
			return nil, fmt.Errorf("NewS3Client error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(utils.AWS, s3c, datamoldParams.DstBucketName, datamoldParams.DstRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	} else if datamoldParams.DstProvider == "gcp" {
		logrus.Infof("CredentialsFilePath : %s", datamoldParams.DstGcpCredPath)
		logrus.Infof("ProjectID : %s", datamoldParams.DstProjectID)
		logrus.Infof("Region : %s", datamoldParams.DstRegion)
		logrus.Infof("BucketName : %s", datamoldParams.DstBucketName)
		gc, err := config.NewGCPClient(datamoldParams.DstGcpCredPath)
		if err != nil {
			return nil, fmt.Errorf("NewGCPClient error : %v", err)
		}

		OSC, err = osc.New(gcpfs.New(gc, datamoldParams.DstProjectID, datamoldParams.DstBucketName, datamoldParams.DstRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	} else if datamoldParams.DstProvider == "ncp" {
		logrus.Infof("AccessKey : %s", datamoldParams.DstAccessKey)
		logrus.Infof("SecretKey : %s", datamoldParams.DstSecretKey)
		logrus.Infof("Endpoint : %s", datamoldParams.DstEndpoint)
		logrus.Infof("Region : %s", datamoldParams.DstRegion)
		logrus.Infof("BucketName : %s", datamoldParams.DstBucketName)
		s3c, err := config.NewS3ClientWithEndpoint(datamoldParams.DstAccessKey, datamoldParams.DstSecretKey, datamoldParams.DstRegion, datamoldParams.DstEndpoint)
		if err != nil {
			return nil, fmt.Errorf("NewS3ClientWithEndpint error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(utils.AWS, s3c, datamoldParams.DstBucketName, datamoldParams.DstRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	}
	return OSC, nil
}

func GetSrcRDMS(datamoldParams *DatamoldParams) (*rdbc.RDBController, error) {
	logrus.Infof("Provider : %s", datamoldParams.SrcProvider)
	logrus.Infof("Username : %s", datamoldParams.SrcUsername)
	logrus.Infof("Password : %s", datamoldParams.SrcPassword)
	logrus.Infof("Host : %s", datamoldParams.SrcHost)
	logrus.Infof("Port : %s", datamoldParams.SrcPort)
	src, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", datamoldParams.SrcUsername, datamoldParams.SrcPassword, datamoldParams.SrcHost, datamoldParams.SrcPort))
	if err != nil {
		return nil, err
	}
	return rdbc.New(mysql.New(utils.Provider(datamoldParams.SrcProvider), src), rdbc.WithLogger(logrus.StandardLogger()))
}

func GetDstRDMS(datamoldParams *DatamoldParams) (*rdbc.RDBController, error) {
	logrus.Infof("Provider : %s", datamoldParams.DstProvider)
	logrus.Infof("Username : %s", datamoldParams.DstUsername)
	logrus.Infof("Password : %s", datamoldParams.DstPassword)
	logrus.Infof("Host : %s", datamoldParams.DstHost)
	logrus.Infof("Port : %s", datamoldParams.DstPort)
	dst, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", datamoldParams.DstUsername, datamoldParams.DstPassword, datamoldParams.DstHost, datamoldParams.DstPort))
	if err != nil {
		return nil, err
	}
	return rdbc.New(mysql.New(utils.Provider(datamoldParams.DstProvider), dst), rdbc.WithLogger(logrus.StandardLogger()))
}

func GetSrcNRDMS(datamoldParams *DatamoldParams) (*nrdbc.NRDBController, error) {
	var NRDBC *nrdbc.NRDBController
	logrus.Infof("Provider : %s", datamoldParams.SrcProvider)
	if datamoldParams.SrcProvider == "aws" {
		logrus.Infof("AccessKey : %s", datamoldParams.SrcAccessKey)
		logrus.Infof("SecretKey : %s", datamoldParams.SrcSecretKey)
		logrus.Infof("Region : %s", datamoldParams.SrcRegion)
		awsnrdb, err := config.NewDynamoDBClient(datamoldParams.SrcAccessKey, datamoldParams.SrcSecretKey, datamoldParams.SrcRegion)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(awsdnmdb.New(awsnrdb, datamoldParams.SrcRegion), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	} else if datamoldParams.SrcProvider == "gcp" {
		logrus.Infof("CredentialsFilePath : %s", datamoldParams.SrcGcpCredPath)
		logrus.Infof("ProjectID : %s", datamoldParams.SrcProjectID)
		logrus.Infof("Region : %s", datamoldParams.SrcRegion)
		gcpnrdb, err := config.NewFireStoreClient(datamoldParams.SrcGcpCredPath, datamoldParams.SrcGcpCredJson, datamoldParams.SrcProjectID, datamoldParams.SrcDatabaseID)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(gcpfsdb.New(gcpnrdb, datamoldParams.SrcRegion), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	} else if datamoldParams.SrcProvider == "ncp" {
		logrus.Infof("Username : %s", datamoldParams.SrcUsername)
		logrus.Infof("Password : %s", datamoldParams.SrcPassword)
		logrus.Infof("Host : %s", datamoldParams.SrcHost)
		logrus.Infof("Port : %s", datamoldParams.SrcPort)
		port, err := strconv.Atoi(datamoldParams.SrcPort)
		if err != nil {
			return nil, err
		}

		ncpnrdb, err := config.NewNCPMongoDBClient(datamoldParams.SrcUsername, datamoldParams.SrcPassword, datamoldParams.SrcHost, port)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(ncpmgdb.New(ncpnrdb, datamoldParams.SrcDBName), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	}
	return NRDBC, nil
}

func GetDstNRDMS(datamoldParams *DatamoldParams) (*nrdbc.NRDBController, error) {
	var NRDBC *nrdbc.NRDBController
	logrus.Infof("Provider : %s", datamoldParams.DstProvider)
	if datamoldParams.DstProvider == "aws" {
		logrus.Infof("AccessKey : %s", datamoldParams.DstAccessKey)
		logrus.Infof("SecretKey : %s", datamoldParams.DstSecretKey)
		logrus.Infof("Region : %s", datamoldParams.DstRegion)
		awsnrdb, err := config.NewDynamoDBClient(datamoldParams.DstAccessKey, datamoldParams.DstSecretKey, datamoldParams.DstRegion)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(awsdnmdb.New(awsnrdb, datamoldParams.DstRegion), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	} else if datamoldParams.DstProvider == "gcp" {
		logrus.Infof("CredentialsFilePath : %s", datamoldParams.DstGcpCredPath)
		logrus.Infof("ProjectID : %s", datamoldParams.DstProjectID)
		logrus.Infof("Region : %s", datamoldParams.DstRegion)
		gcpnrdb, err := config.NewFireStoreClient(datamoldParams.DstGcpCredPath, datamoldParams.DstGcpCredJson, datamoldParams.DstProjectID, datamoldParams.DstDatabaseID)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(gcpfsdb.New(gcpnrdb, datamoldParams.DstRegion), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	} else if datamoldParams.DstProvider == "ncp" {
		logrus.Infof("Username : %s", datamoldParams.DstUsername)
		logrus.Infof("Password : %s", datamoldParams.DstPassword)
		logrus.Infof("Host : %s", datamoldParams.DstHost)
		logrus.Infof("Port : %s", datamoldParams.DstPort)
		port, err := strconv.Atoi(datamoldParams.DstPort)
		if err != nil {
			return nil, err
		}

		ncpnrdb, err := config.NewNCPMongoDBClient(datamoldParams.DstUsername, datamoldParams.DstPassword, datamoldParams.DstHost, port)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(ncpmgdb.New(ncpnrdb, datamoldParams.DstDBName), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	}
	return NRDBC, nil
}

func GetConfig(credPath string, ConfigData *map[string]map[string]map[string]string) error {
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

func preRunE(pName string, cmdName string, datamoldParams *DatamoldParams) error {
	logrus.Info("initiate a configuration scan")
	if err := GetConfig(datamoldParams.CredentialPath, &datamoldParams.ConfigData); err != nil {
		return fmt.Errorf("get config error : %s", err)
	}

	if cmdName == "objectstorage" {
		if value, ok := datamoldParams.ConfigData["objectstorage"]; ok {
			if !datamoldParams.TaskTarget {
				if src, ok := value["src"]; ok {
					if err := applyOSValue(src, "src", datamoldParams); err != nil {
						return err
					}
				}
			} else {
				if dst, ok := value["dst"]; ok {
					if err := applyOSValue(dst, "dst", datamoldParams); err != nil {
						return err
					}
				}
			}
		} else {
			return errors.New("does not exist objectstorage")
		}

		if pName != "migration" && pName != "delete" {
			if err := utils.IsDir(datamoldParams.DstPath); err != nil {
				return errors.New("dstPath error")
			}
		} else if pName == "migration" {
			if value, ok := datamoldParams.ConfigData["objectstorage"]; ok {
				if !datamoldParams.TaskTarget {
					if dst, ok := value["dst"]; ok {
						if err := applyOSValue(dst, "dst", datamoldParams); err != nil {
							return err
						}
					}
				} else {
					if src, ok := value["src"]; ok {
						if err := applyOSValue(src, "src", datamoldParams); err != nil {
							return err
						}
					}
				}
			} else {
				return errors.New("does not exist objectstorage dst")
			}
		}
	} else if cmdName == "rdbms" {
		if value, ok := datamoldParams.ConfigData["rdbms"]; ok {
			if !datamoldParams.TaskTarget {
				if src, ok := value["src"]; ok {
					if err := applyRDMValue(src, "src", datamoldParams); err != nil {
						return err
					}
				}
			} else {
				if value, ok := datamoldParams.ConfigData["rdbms"]; ok {
					if dst, ok := value["dst"]; ok {
						return applyRDMValue(dst, "dst", datamoldParams)
					}
				}
			}
		} else {
			return errors.New("does not exist rdbms src")
		}

		if pName != "migration" && pName != "delete" {
			if err := utils.IsDir(datamoldParams.DstPath); err != nil {
				return errors.New("dstPath error")
			}
		} else if pName == "migration" {
			if value, ok := datamoldParams.ConfigData["rdbms"]; ok {
				if !datamoldParams.TaskTarget {
					if value, ok := datamoldParams.ConfigData["rdbms"]; ok {
						if dst, ok := value["dst"]; ok {
							return applyRDMValue(dst, "dst", datamoldParams)
						}
					}
				} else {
					if src, ok := value["src"]; ok {
						if err := applyRDMValue(src, "src", datamoldParams); err != nil {
							return err
						}
					}
				}
			} else {
				return errors.New("does not exist rdbms dst")
			}
		}
	} else if cmdName == "nrdbms" {
		if value, ok := datamoldParams.ConfigData["nrdbms"]; ok {
			if !datamoldParams.TaskTarget {
				if src, ok := value["src"]; ok {
					if err := applyNRDMValue(src, "src", datamoldParams); err != nil {
						return err
					}
				}
			} else {
				if dst, ok := value["dst"]; ok {
					if err := applyNRDMValue(dst, "dst", datamoldParams); err != nil {
						return err
					}
				}
			}
		} else {
			return errors.New("does not exist nrdbms src")
		}

		if pName != "migration" && pName != "delete" {
			if err := utils.IsDir(datamoldParams.DstPath); err != nil {
				return errors.New("dstPath error")
			}
		} else if pName == "migration" {
			if value, ok := datamoldParams.ConfigData["nrdbms"]; ok {
				if !datamoldParams.TaskTarget {
					if value, ok := datamoldParams.ConfigData["nrdbms"]; ok {
						if dst, ok := value["dst"]; ok {
							return applyNRDMValue(dst, "dst", datamoldParams)
						}
					}
				} else {
					if src, ok := value["src"]; ok {
						if err := applyNRDMValue(src, "src", datamoldParams); err != nil {
							return err
						}
					}
				}
			} else {
				return errors.New("does not exist nrdbms dst")
			}
		}
	}
	return nil
}

func applyNRDMValue(src map[string]string, p string, datamoldParams *DatamoldParams) error {
	provider, ok := src["provider"]
	if ok {
		if provider != "aws" && provider != "gcp" && provider != "ncp" {
			return fmt.Errorf("provider[aws,gcp,ncp] error : %s", provider)
		}
	} else {
		return errors.New("does not exist provider")
	}

	if p == "src" {
		datamoldParams.SrcProvider = provider
	} else {
		datamoldParams.DstProvider = provider
	}

	if provider == "aws" {
		access, ok := src["assessKey"]
		if !ok {
			return errors.New("does not exist assessKey")
		}

		if p == "src" {
			datamoldParams.SrcAccessKey = access
		} else {
			datamoldParams.DstAccessKey = access
		}

		secret, ok := src["secretKey"]
		if !ok {
			return errors.New("does not exist secretKey")
		}

		if p == "src" {
			datamoldParams.SrcSecretKey = secret
		} else {
			datamoldParams.DstSecretKey = secret
		}

		region, ok := src["region"]
		if !ok {
			return errors.New("does not exist region")
		}

		if p == "src" {
			datamoldParams.SrcRegion = region
		} else {
			datamoldParams.DstRegion = region
		}
	} else if provider == "gcp" {
		cred, ok := src["gcpCredPath"]
		if !ok {
			return errors.New("does not exist gcpCredPath")
		}
		if p == "src" {
			datamoldParams.SrcGcpCredPath = cred
		} else {
			datamoldParams.DstGcpCredPath = cred
		}

		projectID, ok := src["projectID"]
		if !ok {
			return errors.New("does not exist projectID")
		}
		if p == "src" {
			datamoldParams.SrcProjectID = projectID
		} else {
			datamoldParams.DstProjectID = projectID
		}

		region, ok := src["region"]
		if !ok {
			return errors.New("does not exist region")
		}

		if p == "src" {
			datamoldParams.SrcRegion = region
		} else {
			datamoldParams.DstRegion = region
		}
	} else if provider == "ncp" {
		username, ok := src["username"]
		if !ok {
			return errors.New("does not exist username")
		}

		if p == "src" {
			datamoldParams.SrcUsername = username
		} else {
			datamoldParams.DstUsername = username
		}

		password, ok := src["password"]
		if !ok {
			return errors.New("does not exist password")
		}

		if p == "src" {
			datamoldParams.SrcPassword = password
		} else {
			datamoldParams.DstPassword = password
		}

		host, ok := src["host"]
		if !ok {
			return errors.New("does not exist host")
		}

		if p == "src" {
			datamoldParams.SrcHost = host
		} else {
			datamoldParams.DstHost = host
		}

		port, ok := src["port"]
		if !ok {
			return errors.New("does not exist port")
		}

		if p == "src" {
			datamoldParams.SrcPort = port
		} else {
			datamoldParams.DstPort = port
		}

		DBName, ok := src["databaseName"]
		if !ok {
			return errors.New("does not exist databaseName")
		}

		if p == "src" {
			datamoldParams.SrcDBName = DBName
		} else {
			datamoldParams.DstDBName = DBName
		}
	}
	return nil
}

func applyRDMValue(src map[string]string, p string, datamoldParams *DatamoldParams) error {
	provider, ok := src["provider"]
	if ok {
		if provider != "aws" && provider != "gcp" && provider != "ncp" {
			return fmt.Errorf("provider[aws,gcp,ncp] error : %s", provider)
		}
	} else {
		return errors.New("does not exist provider")
	}

	if p == "src" {
		datamoldParams.SrcProvider = provider
	} else {
		datamoldParams.DstProvider = provider
	}

	username, ok := src["username"]
	if !ok {
		return errors.New("does not exist username")
	}

	if p == "src" {
		datamoldParams.SrcUsername = username
	} else {
		datamoldParams.DstUsername = username
	}

	password, ok := src["password"]
	if !ok {
		return errors.New("does not exist password")
	}

	if p == "src" {
		datamoldParams.SrcPassword = password
	} else {
		datamoldParams.DstPassword = password
	}

	host, ok := src["host"]
	if !ok {
		return errors.New("does not exist host")
	}

	if p == "src" {
		datamoldParams.SrcHost = host
	} else {
		datamoldParams.DstHost = host
	}

	port, ok := src["port"]
	if !ok {
		return errors.New("does not exist port")
	}

	if p == "src" {
		datamoldParams.SrcPort = port
	} else {
		datamoldParams.DstPort = port
	}

	return nil
}

func applyOSValue(src map[string]string, p string, datamoldParams *DatamoldParams) error {
	provider, ok := src["provider"]
	if ok {
		if provider != "aws" && provider != "gcp" && provider != "ncp" {
			return fmt.Errorf("provider[aws,gcp,ncp] error : %s", provider)
		}
	} else {
		return errors.New("does not exist provider")
	}

	if p == "src" {
		datamoldParams.SrcProvider = provider
	} else {
		datamoldParams.DstProvider = provider
	}

	if provider == "aws" || provider == "ncp" {
		access, ok := src["assessKey"]
		if !ok {
			return errors.New("does not exist assessKey")
		}

		if p == "src" {
			datamoldParams.SrcAccessKey = access
		} else {
			datamoldParams.DstAccessKey = access
		}

		secret, ok := src["secretKey"]
		if !ok {
			return errors.New("does not exist secretKey")
		}

		if p == "src" {
			datamoldParams.SrcSecretKey = secret
		} else {
			datamoldParams.DstSecretKey = secret
		}

		region, ok := src["region"]
		if !ok {
			return errors.New("does not exist region")
		}

		if p == "src" {
			datamoldParams.SrcRegion = region
		} else {
			datamoldParams.DstRegion = region
		}

		bktName, ok := src["bucketName"]
		if !ok {
			return errors.New("does not exist bucketName")
		}

		if p == "src" {
			datamoldParams.SrcBucketName = bktName
		} else {
			datamoldParams.DstBucketName = bktName
		}

		if provider == "ncp" {
			endpoint, ok := src["endpoint"]
			if !ok {
				return errors.New("does not exist endpoint")
			}
			if p == "src" {
				datamoldParams.SrcEndpoint = endpoint
			} else {
				datamoldParams.DstEndpoint = endpoint
			}
		}
	}

	if provider == "gcp" {
		cred, ok := src["gcpCredPath"]
		if !ok {
			return errors.New("does not exist gcpCredPath")
		}
		if p == "src" {
			datamoldParams.SrcGcpCredPath = cred
		} else {
			datamoldParams.DstGcpCredPath = cred
		}

		projectID, ok := src["projectID"]
		if !ok {
			return errors.New("does not exist projectID")
		}
		if p == "src" {
			datamoldParams.SrcProjectID = projectID
		} else {
			datamoldParams.DstProjectID = projectID
		}

		region, ok := src["region"]
		if !ok {
			return errors.New("does not exist region")
		}
		if p == "src" {
			datamoldParams.SrcRegion = region
		} else {
			datamoldParams.DstRegion = region
		}

		bktName, ok := src["bucketName"]
		if !ok {
			return errors.New("does not exist bucketName")
		}
		if p == "src" {
			datamoldParams.SrcBucketName = bktName
		} else {
			datamoldParams.DstBucketName = bktName
		}
	}
	return nil
}
