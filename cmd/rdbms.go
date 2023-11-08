/*
Copyright © 2023 cychoi, tykim <dev@zconverter.com>

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
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cloud-barista/cm-data-mold/service/rdbc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rdbmsCmd represents the rdbms command
var importRDBCmd = &cobra.Command{
	Use:    "rdbms",
	PreRun: preRun("rdbms"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := importRDMFunc(); err != nil {
			os.Exit(1)
		}
	},
}

var exportRDBCmd = &cobra.Command{
	Use:    "rdbms",
	PreRun: preRun("rdbms"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := exportRDMFunc(); err != nil {
			os.Exit(1)
		}
	},
}

var replicationRDBCmd = &cobra.Command{
	Use:    "rdbms",
	PreRun: preRun("rdbms"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := replicationRDMFunc(); err != nil {
			os.Exit(1)
		}
	},
}

var deleteRDBMSCmd = &cobra.Command{
	Use:    "rdbms",
	PreRun: preRun("rdbms"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteRDMFunc(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	importCmd.AddCommand(importRDBCmd)
	exportCmd.AddCommand(exportRDBCmd)
	replicationCmd.AddCommand(replicationRDBCmd)
	deleteCmd.AddCommand(deleteRDBMSCmd)

	deleteRDBMSCmd.Flags().StringVarP(&credentialPath, "credential-path", "C", "", "Json file path containing the user's credentials")
	deleteRDBMSCmd.Flags().StringArrayVarP(&deleteDBList, "delete-db-list", "D", []string{}, "List of db names to delete")
	deleteRDBMSCmd.MarkFlagsRequiredTogether("credential-path", "delete-db-list")
}

func importRDMFunc() error {
	var RDBC *rdbc.RDBController
	var err error
	logrus.Infof("User Information")
	if !taskTarget {
		RDBC, err = getSrcRDMS()
	} else {
		RDBC, err = getDstRDMS()
	}
	if err != nil {
		logrus.Errorf("RDBController error importing into rdbms : %v", err)
		return err
	}

	sqlList := []string{}
	err = filepath.Walk(dstPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".sql" {
			sqlList = append(sqlList, path)
		}
		return nil
	})
	if err != nil {
		logrus.Errorf("Walk error : %v", err)
		return err
	}

	for _, sqlPath := range sqlList {
		data, err := os.ReadFile(sqlPath)
		if err != nil {
			logrus.Errorf("ReadFile error : %v", err)
			return err
		}
		logrus.Infof("Import start: %s", sqlPath)
		if err := RDBC.Put(string(data)); err != nil {
			logrus.Error("Put error importing into rdbms")
			return err
		}
		logrus.Infof("Import success: %s", sqlPath)
	}
	logrus.Infof("successfully imported : %s", dstPath)
	return nil
}

func exportRDMFunc() error {
	var RDBC *rdbc.RDBController
	var err error
	logrus.Infof("User Information")
	if !taskTarget {
		RDBC, err = getSrcRDMS()
	} else {
		RDBC, err = getDstRDMS()
	}
	if err != nil {
		logrus.Errorf("RDBController error exporting into rdbms : %v", err)
		return err
	}

	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		logrus.Errorf("MkdirAll error : %v", err)
		return err
	}

	dbList := []string{}
	if err := RDBC.ListDB(&dbList); err != nil {
		logrus.Errorf("ListDB error : %v", err)
		return err
	}

	var sqlData string
	for _, db := range dbList {
		sqlData = ""
		logrus.Infof("Export start: %s", db)
		if err := RDBC.Get(db, &sqlData); err != nil {
			logrus.Errorf("Get error : %v", err)
			return err
		}

		file, err := os.Create(filepath.Join(dstPath, fmt.Sprintf("%s.sql", db)))
		if err != nil {
			logrus.Errorf("File create error : %v", err)
			return err
		}
		defer file.Close()

		_, err = file.WriteString(sqlData)
		if err != nil {
			logrus.Errorf("File write error : %v", err)
			return err
		}
		logrus.Infof("successfully exported : %s", file.Name())
		file.Close()
	}
	logrus.Infof("successfully exported : %s", dstPath)
	return nil
}

func replicationRDMFunc() error {
	var srcRDBC *rdbc.RDBController
	var srcErr error
	var dstRDBC *rdbc.RDBController
	var dstErr error
	if !taskTarget {
		logrus.Infof("Source Information")
		srcRDBC, srcErr = getSrcRDMS()
		if srcErr != nil {
			logrus.Errorf("RDBController error replication into rdbms : %v", srcErr)
			return srcErr
		}
		logrus.Infof("Target Information")
		dstRDBC, dstErr = getDstRDMS()
		if dstErr != nil {
			logrus.Errorf("RDBController error replication into rdbms : %v", dstErr)
			return dstErr
		}
	} else {
		logrus.Infof("Source Information")
		srcRDBC, srcErr = getDstRDMS()
		if srcErr != nil {
			logrus.Errorf("RDBController error replication into rdbms : %v", srcErr)
			return srcErr
		}
		logrus.Infof("Target Information")
		dstRDBC, dstErr = getSrcRDMS()
		if dstErr != nil {
			logrus.Errorf("RDBController error replication into rdbms : %v", dstErr)
			return dstErr
		}
	}

	logrus.Info("Launch RDBController Copy")
	if err := srcRDBC.Copy(dstRDBC); err != nil {
		logrus.Errorf("Copy error copying into rdbms : %v", err)
		return err
	}
	logrus.Info("successfully replicationed")
	return nil
}

func deleteRDMFunc() error {
	var RDBC *rdbc.RDBController
	var err error
	if !taskTarget {
		RDBC, err = getSrcRDMS()
	} else {
		RDBC, err = getDstRDMS()
	}
	if err != nil {
		logrus.Errorf("RDBController error deleting into rdbms : %v", err)
		return err
	}

	logrus.Info("Launch RDBController Delete")
	if err := RDBC.DeleteDB(deleteDBList...); err != nil {
		logrus.Errorf("Delete error deleting into rdbms : %v", err)
		return err
	}
	logrus.Info("successfully deleted")
	return nil
}
