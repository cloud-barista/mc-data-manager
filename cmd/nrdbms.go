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
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cloud-barista/cm-data-mold/service/nrdbc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// nrdbmsCmd represents the nrdbms command
var importNRDBCmd = &cobra.Command{
	Use:     "nrdbms",
	Aliases: []string{"ndb"},
	PreRun:  preRun("nrdbms"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := importNRDMFunc(); err != nil {
			os.Exit(1)
		}
	},
}

var exportNRDBCmd = &cobra.Command{
	Use:    "nrdbms",
	PreRun: preRun("nrdbms"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := exportNRDMFunc(); err != nil {
			os.Exit(1)
		}
	},
}

var migrationNRDBCmd = &cobra.Command{
	Use:    "nrdbms",
	PreRun: preRun("nrdbms"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := migrationNRDMFunc(); err != nil {
			os.Exit(1)
		}
	},
}

var deleteNRDBMSCmd = &cobra.Command{
	Use:    "nrdbms",
	PreRun: preRun("nrdbms"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteNRDMFunc(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	importCmd.AddCommand(importNRDBCmd)
	exportCmd.AddCommand(exportNRDBCmd)
	migrationCmd.AddCommand(migrationNRDBCmd)
	deleteCmd.AddCommand(deleteNRDBMSCmd)

	deleteNRDBMSCmd.Flags().StringVarP(&credentialPath, "credential-path", "C", "", "Json file path containing the user's credentials")
	deleteNRDBMSCmd.Flags().StringArrayVarP(&deleteTableList, "delete-table-list", "D", []string{}, "List of table names to delete")
	deleteNRDBMSCmd.MarkFlagsRequiredTogether("credential-path", "delete-table-list")
}

func importNRDMFunc() error {
	var NRDBC *nrdbc.NRDBController
	var err error
	if !taskTarget {
		NRDBC, err = getSrcNRDMS()
	} else {
		NRDBC, err = getDstNRDMS()
	}

	if err != nil {
		logrus.Errorf("NRDBController error importing into nrdbms : %v", err)
		return err
	}

	jsonList := []string{}
	err = filepath.Walk(dstPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".json" {
			jsonList = append(jsonList, path)
		}
		return nil
	})

	if err != nil {
		logrus.Errorf("Walk error : %v", err)
		return err
	}

	var srcData []map[string]interface{}
	for _, jsonFile := range jsonList {
		srcData = []map[string]interface{}{}

		file, err := os.Open(jsonFile)
		if err != nil {
			logrus.Errorf("file open error : %v", err)
			return err
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&srcData); err != nil {
			logrus.Errorf("file decoding error : %v", err)
			return err
		}

		fileName := filepath.Base(jsonFile)
		tableName := fileName[:len(fileName)-len(filepath.Ext(fileName))]

		logrus.Infof("Import start: %s", fileName)
		if err := NRDBC.Put(tableName, &srcData); err != nil {
			logrus.Error("Put error importing into nrdbms")
			return err
		}
		logrus.Infof("successfully imported : %s", dstPath)
	}
	return nil
}

func exportNRDMFunc() error {
	var NRDBC *nrdbc.NRDBController
	var err error
	logrus.Infof("User Information")
	if !taskTarget {
		NRDBC, err = getSrcNRDMS()
	} else {
		NRDBC, err = getDstNRDMS()
	}
	if err != nil {
		logrus.Errorf("NRDBController error exporting into rdbms : %v", err)
		return err
	}

	tableList, err := NRDBC.ListTables()
	if err != nil {
		logrus.Errorf("ListTables error : %v", err)
		return err
	}

	var dstData []map[string]interface{}
	for _, table := range tableList {
		logrus.Infof("Export start: %s", table)
		dstData = []map[string]interface{}{}

		if err := NRDBC.Get(table, &dstData); err != nil {
			logrus.Errorf("Get error : %v", err)
			return err
		}

		file, err := os.Create(filepath.Join(dstPath, fmt.Sprintf("%s.json", table)))
		if err != nil {
			logrus.Errorf("File create error : %v", err)
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(dstData); err != nil {
			logrus.Errorf("data encoding error : %v", err)
			return err
		}
		logrus.Infof("successfully exported : %s", file.Name())
	}
	logrus.Infof("successfully exported : %s", dstPath)
	return nil
}

func migrationNRDMFunc() error {
	var srcNRDBC *nrdbc.NRDBController
	var srcErr error
	var dstNRDBC *nrdbc.NRDBController
	var dstErr error
	if !taskTarget {
		logrus.Infof("Source Information")
		srcNRDBC, srcErr = getSrcNRDMS()
		if srcErr != nil {
			logrus.Errorf("NRDBController error migration into nrdbms : %v", srcErr)
			return srcErr
		}
		logrus.Infof("Target Information")
		dstNRDBC, dstErr = getDstNRDMS()
		if dstErr != nil {
			logrus.Errorf("NRDBController error migration into nrdbms : %v", dstErr)
			return dstErr
		}
	} else {
		logrus.Infof("Source Information")
		srcNRDBC, srcErr = getDstNRDMS()
		if srcErr != nil {
			logrus.Errorf("NRDBController error migration into nrdbms : %v", srcErr)
			return srcErr
		}
		logrus.Infof("Target Information")
		dstNRDBC, dstErr = getSrcNRDMS()
		if dstErr != nil {
			logrus.Errorf("NRDBController error migration into nrdbms : %v", dstErr)
			return dstErr
		}
	}
	logrus.Info("Launch NRDBController Copy")
	if err := srcNRDBC.Copy(dstNRDBC); err != nil {
		logrus.Errorf("Copy error copying into nrdbms : %v", err)
		return err
	}
	logrus.Info("successfully migrationed")
	return nil
}

func deleteNRDMFunc() error {
	var NRDBC *nrdbc.NRDBController
	var err error
	if !taskTarget {
		NRDBC, err = getSrcNRDMS()
	} else {
		NRDBC, err = getDstNRDMS()
	}
	if err != nil {
		logrus.Errorf("NRDBController error deleting into nrdbms : %v", err)
		return err
	}

	logrus.Info("Launch NRDBController Delete")
	if err := NRDBC.DeleteTables(deleteTableList...); err != nil {
		logrus.Errorf("Delete error deleting into nrdbms : %v", err)
		return err
	}
	logrus.Info("successfully deleted")
	return nil
}
