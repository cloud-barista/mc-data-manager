package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cloud-barista/cm-data-mold/config"
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
	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type CustomTextFormatter struct {
	cmdName string
	jobName string
}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timeFormatted := entry.Time.Format("2006-01-02T15:04:05-07:00")
	cn := f.cmdName
	jn := f.jobName
	if _, ok := entry.Data["cmdbName"]; ok {
		cn = entry.Data["cmdbName"].(string)
	}
	if _, ok := entry.Data["jobName"]; ok {
		jn = entry.Data["jobName"].(string)
	}
	return []byte(fmt.Sprintf("[%s] [%s] [%s] [%s] %s\n", timeFormatted, entry.Level, cn, jn, strings.ToUpper(entry.Message[:1])+entry.Message[1:])), nil
}

func preRun(task string) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&CustomTextFormatter{cmdName: cmd.Parent().Use, jobName: task})
		logrus.Infof("launch an %s to %s", cmd.Parent().Use, task)
		err := preRunE(cmd.Parent().Use, task)
		if err != nil {
			logrus.Errorf("Pre-check for %s operation errors : %v", task, err)
			os.Exit(1)
		}
		logrus.Infof("successful pre-check %s into %s", cmd.Parent().Use, task)
	}
}

func getSrcOS() (*osc.OSController, error) {
	var OSC *osc.OSController
	logrus.Infof("Provider : %s", cSrcProvider)
	if cSrcProvider == "aws" {
		logrus.Infof("AccessKey : %s", cSrcAccessKey)
		logrus.Infof("SecretKey : %s", cSrcSecretKey)
		logrus.Infof("Region : %s", cSrcRegion)
		logrus.Infof("BucketName : %s", cSrcBucketName)
		s3c, err := config.NewS3Client(cSrcAccessKey, cSrcSecretKey, cSrcRegion)
		if err != nil {
			return nil, fmt.Errorf("NewS3Client error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(utils.AWS, s3c, cSrcBucketName, cSrcRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	} else if cSrcProvider == "gcp" {
		logrus.Infof("CredentialsFilePath : %s", cSrcGcpCredPath)
		logrus.Infof("ProjectID : %s", cSrcProjectID)
		logrus.Infof("Region : %s", cSrcRegion)
		logrus.Infof("BucketName : %s", cSrcBucketName)
		gc, err := config.NewGCSClient(cSrcGcpCredPath)
		if err != nil {
			return nil, fmt.Errorf("NewGCSClient error : %v", err)
		}

		OSC, err = osc.New(gcsfs.New(gc, cSrcProjectID, cSrcBucketName, cSrcRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	} else if cSrcProvider == "ncp" {
		logrus.Infof("AccessKey : %s", cSrcAccessKey)
		logrus.Infof("SecretKey : %s", cSrcSecretKey)
		logrus.Infof("Endpoint : %s", cSrcEndpoint)
		logrus.Infof("Region : %s", cSrcRegion)
		logrus.Infof("BucketName : %s", cSrcBucketName)
		s3c, err := config.NewS3ClientWithEndpoint(cSrcAccessKey, cSrcSecretKey, cSrcRegion, cSrcEndpoint)
		if err != nil {
			return nil, fmt.Errorf("NewS3ClientWithEndpint error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(utils.AWS, s3c, cSrcBucketName, cSrcRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	}
	return OSC, nil
}

func getDstOS() (*osc.OSController, error) {
	var OSC *osc.OSController
	logrus.Infof("Provider : %s", cDstProvider)
	if cDstProvider == "aws" {
		logrus.Infof("AccessKey : %s", cDstAccessKey)
		logrus.Infof("SecretKey : %s", cDstSecretKey)
		logrus.Infof("Region : %s", cDstRegion)
		logrus.Infof("BucketName : %s", cDstBucketName)
		s3c, err := config.NewS3Client(cDstAccessKey, cDstSecretKey, cDstRegion)
		if err != nil {
			return nil, fmt.Errorf("NewS3Client error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(utils.AWS, s3c, cDstBucketName, cDstRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	} else if cDstProvider == "gcp" {
		logrus.Infof("CredentialsFilePath : %s", cDstGcpCredPath)
		logrus.Infof("ProjectID : %s", cDstProjectID)
		logrus.Infof("Region : %s", cDstRegion)
		logrus.Infof("BucketName : %s", cDstBucketName)
		gc, err := config.NewGCSClient(cDstGcpCredPath)
		if err != nil {
			return nil, fmt.Errorf("NewGCSClient error : %v", err)
		}

		OSC, err = osc.New(gcsfs.New(gc, cDstProjectID, cDstBucketName, cDstRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	} else if cDstProvider == "ncp" {
		logrus.Infof("AccessKey : %s", cDstAccessKey)
		logrus.Infof("SecretKey : %s", cDstSecretKey)
		logrus.Infof("Endpoint : %s", cDstEndpoint)
		logrus.Infof("Region : %s", cDstRegion)
		logrus.Infof("BucketName : %s", cDstBucketName)
		s3c, err := config.NewS3ClientWithEndpoint(cDstAccessKey, cDstSecretKey, cDstRegion, cDstEndpoint)
		if err != nil {
			return nil, fmt.Errorf("NewS3ClientWithEndpint error : %v", err)
		}

		OSC, err = osc.New(s3fs.New(utils.AWS, s3c, cDstBucketName, cDstRegion), osc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, fmt.Errorf("osc error : %v", err)
		}
	}
	return OSC, nil
}

func getSrcRDMS() (*rdbc.RDBController, error) {
	logrus.Infof("Provider : %s", cSrcProvider)
	logrus.Infof("Username : %s", cSrcUsername)
	logrus.Infof("Password : %s", cSrcPassword)
	logrus.Infof("Host : %s", cSrcHost)
	logrus.Infof("Port : %s", cSrcPort)
	src, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", cSrcUsername, cSrcPassword, cSrcHost, cSrcPort))
	if err != nil {
		return nil, err
	}
	return rdbc.New(mysql.New(utils.Provider(cSrcProvider), src), rdbc.WithLogger(logrus.StandardLogger()))
}

func getDstRDMS() (*rdbc.RDBController, error) {
	logrus.Infof("Provider : %s", cDstProvider)
	logrus.Infof("Username : %s", cDstUsername)
	logrus.Infof("Password : %s", cDstPassword)
	logrus.Infof("Host : %s", cDstHost)
	logrus.Infof("Port : %s", cDstPort)
	dst, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", cDstUsername, cDstPassword, cDstHost, cDstPort))
	if err != nil {
		return nil, err
	}
	return rdbc.New(mysql.New(utils.Provider(cDstProvider), dst), rdbc.WithLogger(logrus.StandardLogger()))
}

func getSrcNRDMS() (*nrdbc.NRDBController, error) {

	var NRDBC *nrdbc.NRDBController
	logrus.Infof("Provider : %s", cSrcProvider)
	if cSrcProvider == "aws" {
		logrus.Infof("AccessKey : %s", cSrcAccessKey)
		logrus.Infof("SecretKey : %s", cSrcSecretKey)
		logrus.Infof("Region : %s", cSrcRegion)
		awsnrdb, err := config.NewDynamoDBClient(cSrcAccessKey, cSrcSecretKey, cSrcRegion)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(awsdnmdb.New(awsnrdb, cSrcRegion), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	} else if cSrcProvider == "gcp" {
		logrus.Infof("CredentialsFilePath : %s", cSrcGcpCredPath)
		logrus.Infof("ProjectID : %s", cSrcProjectID)
		logrus.Infof("Region : %s", cSrcRegion)
		gcpnrdb, err := config.NewFireStoreClient(cSrcGcpCredPath, cSrcProjectID)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(gcpfsdb.New(gcpnrdb, cSrcRegion), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	} else if cSrcProvider == "ncp" {
		logrus.Infof("Username : %s", cSrcUsername)
		logrus.Infof("Password : %s", cSrcPassword)
		logrus.Infof("Host : %s", cSrcHost)
		logrus.Infof("Port : %s", cSrcPort)
		port, err := strconv.Atoi(cSrcPort)
		if err != nil {
			return nil, err
		}

		ncpnrdb, err := config.NewNCPMongoDBClient(cSrcUsername, cSrcPassword, cSrcHost, port)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(ncpmgdb.New(ncpnrdb, cSrcDBName), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	}
	return NRDBC, nil
}

func getDstNRDMS() (*nrdbc.NRDBController, error) {
	var NRDBC *nrdbc.NRDBController
	logrus.Infof("Provider : %s", cDstProvider)
	if cDstProvider == "aws" {
		logrus.Infof("AccessKey : %s", cDstAccessKey)
		logrus.Infof("SecretKey : %s", cDstSecretKey)
		logrus.Infof("Region : %s", cDstRegion)
		awsnrdb, err := config.NewDynamoDBClient(cDstAccessKey, cDstSecretKey, cDstRegion)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(awsdnmdb.New(awsnrdb, cDstRegion), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	} else if cDstProvider == "gcp" {
		logrus.Infof("CredentialsFilePath : %s", cDstGcpCredPath)
		logrus.Infof("ProjectID : %s", cDstProjectID)
		logrus.Infof("Region : %s", cDstRegion)
		gcpnrdb, err := config.NewFireStoreClient(cDstGcpCredPath, cDstProjectID)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(gcpfsdb.New(gcpnrdb, cDstRegion), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	} else if cDstProvider == "ncp" {
		logrus.Infof("Username : %s", cDstUsername)
		logrus.Infof("Password : %s", cDstPassword)
		logrus.Infof("Host : %s", cDstHost)
		logrus.Infof("Port : %s", cDstPort)
		port, err := strconv.Atoi(cDstPort)
		if err != nil {
			return nil, err
		}

		ncpnrdb, err := config.NewNCPMongoDBClient(cDstUsername, cDstPassword, cDstHost, port)
		if err != nil {
			return nil, err
		}

		NRDBC, err = nrdbc.New(ncpmgdb.New(ncpnrdb, cDstDBName), nrdbc.WithLogger(logrus.StandardLogger()))
		if err != nil {
			return nil, err
		}
	}
	return NRDBC, nil
}

func getConfig(credPath string) error {
	data, err := os.ReadFile(credPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &configData)
	if err != nil {
		return err
	}
	return nil
}

func preRunE(pName string, cmdName string) error {
	logrus.Info("initiate a configuration scan")
	if err := getConfig(credentialPath); err != nil {
		return fmt.Errorf("get config error : %s", err)
	}

	if cmdName == "objectstorage" {
		if value, ok := configData["objectstorage"]; ok {
			if !taskTarget {
				if src, ok := value["src"]; ok {
					if err := applyOSValue(src, "src"); err != nil {
						return err
					}
				}
			} else {
				if dst, ok := value["dst"]; ok {
					if err := applyOSValue(dst, "dst"); err != nil {
						return err
					}
				}
			}
		} else {
			return errors.New("does not exist objectstorage")
		}

		if pName != "replication" && pName != "delete" {
			if err := utils.IsDir(dstPath); err != nil {
				return errors.New("dstPath error")
			}
		} else if pName == "replication" {
			if value, ok := configData["objectstorage"]; ok {
				if !taskTarget {
					if dst, ok := value["dst"]; ok {
						if err := applyOSValue(dst, "dst"); err != nil {
							return err
						}
					}
				} else {
					if src, ok := value["src"]; ok {
						if err := applyOSValue(src, "src"); err != nil {
							return err
						}
					}
				}
			} else {
				return errors.New("does not exist objectstorage dst")
			}
		}
	} else if cmdName == "rdbms" {
		if value, ok := configData["rdbms"]; ok {
			if !taskTarget {
				if src, ok := value["src"]; ok {
					if err := applyRDMValue(src, "src"); err != nil {
						return err
					}
				}
			} else {
				if value, ok := configData["rdbms"]; ok {
					if dst, ok := value["dst"]; ok {
						return applyRDMValue(dst, "dst")
					}
				}
			}
		} else {
			return errors.New("does not exist rdbms src")
		}

		if pName != "replication" && pName != "delete" {
			if err := utils.IsDir(dstPath); err != nil {
				return errors.New("dstPath error")
			}
		} else if pName == "replication" {
			if value, ok := configData["rdbms"]; ok {
				if !taskTarget {
					if value, ok := configData["rdbms"]; ok {
						if dst, ok := value["dst"]; ok {
							return applyRDMValue(dst, "dst")
						}
					}
				} else {
					if src, ok := value["src"]; ok {
						if err := applyRDMValue(src, "src"); err != nil {
							return err
						}
					}
				}
			} else {
				return errors.New("does not exist rdbms dst")
			}
		}
	} else if cmdName == "nrdbms" {
		if value, ok := configData["nrdbms"]; ok {
			if !taskTarget {
				if src, ok := value["src"]; ok {
					if err := applyNRDMValue(src, "src"); err != nil {
						return err
					}
				}
			} else {
				if dst, ok := value["dst"]; ok {
					if err := applyNRDMValue(dst, "dst"); err != nil {
						return err
					}
				}
			}
		} else {
			return errors.New("does not exist nrdbms src")
		}

		if pName != "replication" && pName != "delete" {
			if err := utils.IsDir(dstPath); err != nil {
				return errors.New("dstPath error")
			}
		} else if pName == "replication" {
			if value, ok := configData["nrdbms"]; ok {
				if !taskTarget {
					if value, ok := configData["nrdbms"]; ok {
						if dst, ok := value["dst"]; ok {
							return applyNRDMValue(dst, "dst")
						}
					}
				} else {
					if src, ok := value["src"]; ok {
						if err := applyNRDMValue(src, "src"); err != nil {
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

func applyNRDMValue(src map[string]string, p string) error {
	provider, ok := src["provider"]
	if ok {
		if provider != "aws" && provider != "gcp" && provider != "ncp" {
			return fmt.Errorf("provider[aws,gcp,ncp] error : %s", provider)
		}
	} else {
		return errors.New("does not exist provider")
	}

	if p == "src" {
		cSrcProvider = provider
	} else {
		cDstProvider = provider
	}

	if provider == "aws" {
		access, ok := src["assessKey"]
		if !ok {
			return errors.New("does not exist assessKey")
		}

		if p == "src" {
			cSrcAccessKey = access
		} else {
			cDstAccessKey = access
		}

		secret, ok := src["secretKey"]
		if !ok {
			return errors.New("does not exist secretKey")
		}

		if p == "src" {
			cSrcSecretKey = secret
		} else {
			cDstSecretKey = secret
		}

		region, ok := src["region"]
		if !ok {
			return errors.New("does not exist region")
		}

		if p == "src" {
			cSrcRegion = region
		} else {
			cDstRegion = region
		}
	} else if provider == "gcp" {
		cred, ok := src["gcpCredPath"]
		if !ok {
			return errors.New("does not exist gcpCredPath")
		}
		if p == "src" {
			cSrcGcpCredPath = cred
		} else {
			cDstGcpCredPath = cred
		}

		projectID, ok := src["projectID"]
		if !ok {
			return errors.New("does not exist projectID")
		}
		if p == "src" {
			cSrcProjectID = projectID
		} else {
			cDstProjectID = projectID
		}

		region, ok := src["region"]
		if !ok {
			return errors.New("does not exist region")
		}

		if p == "src" {
			cSrcRegion = region
		} else {
			cDstRegion = region
		}
	} else if provider == "ncp" {
		username, ok := src["username"]
		if !ok {
			return errors.New("does not exist username")
		}

		if p == "src" {
			cSrcUsername = username
		} else {
			cDstUsername = username
		}

		password, ok := src["password"]
		if !ok {
			return errors.New("does not exist password")
		}

		if p == "src" {
			cSrcPassword = password
		} else {
			cDstPassword = password
		}

		host, ok := src["host"]
		if !ok {
			return errors.New("does not exist host")
		}

		if p == "src" {
			cSrcHost = host
		} else {
			cDstHost = host
		}

		port, ok := src["port"]
		if !ok {
			return errors.New("does not exist port")
		}

		if p == "src" {
			cSrcPort = port
		} else {
			cDstPort = port
		}

		DBName, ok := src["databaseName"]
		if !ok {
			return errors.New("does not exist databaseName")
		}

		if p == "src" {
			cSrcDBName = DBName
		} else {
			cDstDBName = DBName
		}
	}
	return nil
}

func applyRDMValue(src map[string]string, p string) error {
	provider, ok := src["provider"]
	if ok {
		if provider != "aws" && provider != "gcp" && provider != "ncp" {
			return fmt.Errorf("provider[aws,gcp,ncp] error : %s", provider)
		}
	} else {
		return errors.New("does not exist provider")
	}

	if p == "src" {
		cSrcProvider = provider
	} else {
		cDstProvider = provider
	}

	username, ok := src["username"]
	if !ok {
		return errors.New("does not exist username")
	}

	if p == "src" {
		cSrcUsername = username
	} else {
		cDstUsername = username
	}

	password, ok := src["password"]
	if !ok {
		return errors.New("does not exist password")
	}

	if p == "src" {
		cSrcPassword = password
	} else {
		cDstPassword = password
	}

	host, ok := src["host"]
	if !ok {
		return errors.New("does not exist host")
	}

	if p == "src" {
		cSrcHost = host
	} else {
		cDstHost = host
	}

	port, ok := src["port"]
	if !ok {
		return errors.New("does not exist port")
	}

	if p == "src" {
		cSrcPort = port
	} else {
		cDstPort = port
	}

	return nil
}

func applyOSValue(src map[string]string, p string) error {
	provider, ok := src["provider"]
	if ok {
		if provider != "aws" && provider != "gcp" && provider != "ncp" {
			return fmt.Errorf("provider[aws,gcp,ncp] error : %s", provider)
		}
	} else {
		return errors.New("does not exist provider")
	}

	if p == "src" {
		cSrcProvider = provider
	} else {
		cDstProvider = provider
	}

	if provider == "aws" || provider == "ncp" {
		access, ok := src["assessKey"]
		if !ok {
			return errors.New("does not exist assessKey")
		}

		if p == "src" {
			cSrcAccessKey = access
		} else {
			cDstAccessKey = access
		}

		secret, ok := src["secretKey"]
		if !ok {
			return errors.New("does not exist secretKey")
		}

		if p == "src" {
			cSrcSecretKey = secret
		} else {
			cDstSecretKey = secret
		}

		region, ok := src["region"]
		if !ok {
			return errors.New("does not exist region")
		}

		if p == "src" {
			cSrcRegion = region
		} else {
			cDstRegion = region
		}

		bktName, ok := src["bucketName"]
		if !ok {
			return errors.New("does not exist bucketName")
		}

		if p == "src" {
			cSrcBucketName = bktName
		} else {
			cDstBucketName = bktName
		}

		if provider == "ncp" {
			endpoint, ok := src["endpoint"]
			if !ok {
				return errors.New("does not exist endpoint")
			}
			if p == "src" {
				cSrcEndpoint = endpoint
			} else {
				cDstEndpoint = endpoint
			}
		}
	}

	if provider == "gcp" {
		cred, ok := src["gcpCredPath"]
		if !ok {
			return errors.New("does not exist gcpCredPath")
		}
		if p == "src" {
			cSrcGcpCredPath = cred
		} else {
			cDstGcpCredPath = cred
		}

		projectID, ok := src["projectID"]
		if !ok {
			return errors.New("does not exist projectID")
		}
		if p == "src" {
			cSrcProjectID = projectID
		} else {
			cDstProjectID = projectID
		}

		region, ok := src["region"]
		if !ok {
			return errors.New("does not exist region")
		}
		if p == "src" {
			cSrcRegion = region
		} else {
			cDstRegion = region
		}

		bktName, ok := src["bucketName"]
		if !ok {
			return errors.New("does not exist bucketName")
		}
		if p == "src" {
			cSrcBucketName = bktName
		} else {
			cDstBucketName = bktName
		}
	}
	return nil
}
