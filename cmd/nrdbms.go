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

	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/spf13/cobra"
)

// nrdbmsCmd represents the nrdbms command
var importNRDBCmd = &cobra.Command{
	Use:     "nrdbms",
	Aliases: []string{"ndb"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "nrdbms")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "importNRDMFunc", "import start", nil)
		return importNRDMFunc()
	},
}

var exportNRDBCmd = &cobra.Command{
	Use: "nrdbms",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "nrdbms")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "exportNRDMFunc", "export start", nil)
		return exportNRDMFunc()
	},
}

var replicationNRDBCmd = &cobra.Command{
	Use: "nrdbms",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "nrdbms")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "replicationNRDMFunc", "replication start", nil)
		return replicationNRDMFunc()
	},
}

var deleteNRDBMSCmd = &cobra.Command{
	Use: "nrdbms",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "nrdbms")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "deleteNRDMFunc", "delete start", nil)
		return deleteNRDMFunc()
	},
}

func init() {
	importCmd.AddCommand(importNRDBCmd)
	exportCmd.AddCommand(exportNRDBCmd)
	replicationCmd.AddCommand(replicationNRDBCmd)
	deleteCmd.AddCommand(deleteNRDBMSCmd)

	deleteNRDBMSCmd.Flags().StringVarP(&credentialPath, "credential-path", "C", "", "Json file path containing the user's credentials")
	deleteNRDBMSCmd.Flags().StringArrayVarP(&deleteTableList, "delete-table-list", "D", []string{}, "List of table names to delete")
	deleteNRDBMSCmd.MarkFlagsRequiredTogether("credential-path", "delete-table-list")
}

func importNRDMFunc() error {
	nrdb, err := getSrcNRDMS()
	if err != nil {
		utils.LogWirte(logger, "Error", "getSrcNRDMS", "failed to get source nrdbms information", err)
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
		utils.LogWirte(logger, "Error", "Walk", "Failed to get file list from dstPath", err)
		return err
	}

	var srcData []map[string]interface{}
	for _, jsonFile := range jsonList {
		srcData = []map[string]interface{}{}

		file, err := os.Open(jsonFile)
		if err != nil {
			utils.LogWirte(logger, "Error", "Open", "Failed to open json file", err)
			return err
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&srcData); err != nil {
			utils.LogWirte(logger, "Error", "Decode", "json decoding failed", err)
			return err
		}

		fileName := filepath.Base(jsonFile)
		tableName := fileName[:len(fileName)-len(filepath.Ext(fileName))]

		if err := nrdb.Put(tableName, &srcData); err != nil {
			utils.LogWirte(logger, "Error", "Put", fmt.Sprintf("%s import failed", tableName), err)
			return err
		}
		utils.LogWirte(logger, "Info", "Put", fmt.Sprintf("%s imported", tableName), nil)
	}
	utils.LogWirte(logger, "Info", "Put", "Import Done", nil)
	return nil
}

func exportNRDMFunc() error {
	nrdb, err := getSrcNRDMS()
	if err != nil {
		utils.LogWirte(logger, "Error", "getSrcNRDMS", "Failed to get source nrdbms information", err)
		return err
	}

	tableList, err := nrdb.ListTables()
	if err != nil {
		utils.LogWirte(logger, "Error", "ListTables", "Table list lookup failed", err)
		return err
	}

	var dstData []map[string]interface{}
	for _, table := range tableList {
		dstData = []map[string]interface{}{}

		if err := nrdb.Get(table, &dstData); err != nil {
			utils.LogWirte(logger, "Error", "Get", fmt.Sprintf("%s export failed", table), err)
			return err
		}

		file, err := os.Create(filepath.Join(dstPath, fmt.Sprintf("%s.json", table)))
		if err != nil {
			utils.LogWirte(logger, "Error", "Create", "File creation failed", err)
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(dstData); err != nil {
			utils.LogWirte(logger, "Error", "Encode", "json encoding failed", err)
			return err
		}
		utils.LogWirte(logger, "Info", "Get", fmt.Sprintf("%s exported", table), err)
	}

	return nil
}

func replicationNRDMFunc() error {
	src, err := getSrcNRDMS()
	if err != nil {
		utils.LogWirte(logger, "Error", "getSrcNRDMS", "Failed to get source nrdbms information", err)
		return err
	}

	dst, err := getDstNRDMS()
	if err != nil {
		utils.LogWirte(logger, "Error", "getDstNRDMS", "Failed to get target nrdbms information", err)
		return err
	}

	return src.Copy(dst)
}

func deleteNRDMFunc() error {
	src, err := getSrcNRDMS()
	if err != nil {
		utils.LogWirte(logger, "Error", "getSrcNRDMS", "Failed to get source nrdbms information", err)
		return err
	}

	err = src.DeleteTables(deleteTableList...)
	utils.LogWirte(logger, "Info", "DeleteTables", "Delete Done", nil)
	return err
}
