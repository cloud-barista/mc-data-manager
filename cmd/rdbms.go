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

	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/spf13/cobra"
)

// rdbmsCmd represents the rdbms command
var importRDBCmd = &cobra.Command{
	Use: "rdbms",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "rdbms")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "importRDMFunc", "import start", nil)
		return importRDMFunc()
	},
}

var exportRDBCmd = &cobra.Command{
	Use: "rdbms",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "rdbms")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "exportRDMFunc", "export start", nil)
		return exportRDMFunc()
	},
}

var replicationRDBCmd = &cobra.Command{
	Use: "rdbms",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "rdbms")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "replicationRDMFunc", "replication start", nil)
		return replicationRDMFunc()
	},
}

var deleteRDBMSCmd = &cobra.Command{
	Use: "rdbms",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "rdbms")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "deleteRDMFunc", "delete start", nil)
		return deleteRDMFunc()
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
	rdbc, err := getSrcRDMS()
	if err != nil {
		utils.LogWirte(logger, "Error", "getSrcRDMS", "failed to get source rdbms information", err)
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
		utils.LogWirte(logger, "Error", "Walk", "failed to get file list in dstPath", err)
		return err
	}

	for _, sqlPath := range sqlList {
		data, err := os.ReadFile(sqlPath)
		if err != nil {
			utils.LogWirte(logger, "Error", "ReadFile", "sql file read failed", err)
			return err
		}

		if err := rdbc.Put(string(data)); err != nil {
			return err
		}

		utils.LogWirte(logger, "Info", "Put", fmt.Sprintf("%s imported", sqlPath), nil)
	}
	return nil
}

func exportRDMFunc() error {
	rdbc, err := getSrcRDMS()
	if err != nil {
		utils.LogWirte(logger, "Error", "getSrcRDMS", "failed to get source rdbms information", err)
		return err
	}

	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		utils.LogWirte(logger, "Error", "MkdirAll", "mkdir failed", err)
		return err
	}

	dbList := []string{}
	if err := rdbc.ListDB(&dbList); err != nil {
		return err
	}

	var sqlData string
	for _, db := range dbList {
		sqlData = ""
		if err := rdbc.Get(db, &sqlData); err != nil {
			return err
		}

		file, err := os.Create(filepath.Join(dstPath, fmt.Sprintf("%s.sql", db)))
		if err != nil {
			utils.LogWirte(logger, "Error", "Create", "create file failed", err)
			return err
		}
		defer file.Close()

		_, err = file.WriteString(sqlData)
		if err != nil {
			utils.LogWirte(logger, "Error", "WriteString", "writeString failed", err)
			return err
		}

		utils.LogWirte(logger, "Info", "Get", fmt.Sprintf("%s exported", db), err)
		file.Close()
	}
	return nil
}

func replicationRDMFunc() error {
	src, err := getSrcRDMS()
	if err != nil {
		utils.LogWirte(logger, "Error", "getSrcRDMS", "failed to get source rdbms information", err)
		return err
	}

	dst, err := getDstRDMS()
	if err != nil {
		utils.LogWirte(logger, "Error", "getDstRDMS", "failed to get target rdbms information", err)
		return err
	}

	err = src.Copy(dst)
	utils.LogWirte(logger, "Info", "Copy", "Replication Done", nil)
	return err
}

func deleteRDMFunc() error {
	src, err := getSrcRDMS()
	if err != nil {
		utils.LogWirte(logger, "Error", "getSrcRDMS", "failed to get source rdbms information", err)
		return err
	}

	if err := src.DeleteDB(deleteDBList...); err != nil {
		return err
	}
	return nil
}
