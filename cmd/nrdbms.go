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
package cmd

import (
	"os"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/spf13/cobra"
)

// nrdbmsCmd represents the nrdbms command
var importNRDBCmd = &cobra.Command{
	Use:     "nrdbms",
	Aliases: []string{"ndb"},
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("nrdbms", &commandTask, cmd.Parent().Use)
		if err := auth.ImportNRDMFunc(&commandTask); err != nil {
			os.Exit(1)
		}
	},
}

var exportNRDBCmd = &cobra.Command{
	Use:     "nrdbms",
	Aliases: []string{"ndb"},
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("nrdbms", &commandTask, cmd.Parent().Use)
		if err := auth.ExportNRDMFunc(&commandTask); err != nil {
			os.Exit(1)
		}
	},
}

var migrationNRDBCmd = &cobra.Command{
	Use:     "nrdbms",
	Aliases: []string{"ndb"},
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("nrdbms", &commandTask, cmd.Parent().Use)
		if err := auth.MigrationNRDMFunc(&commandTask); err != nil {
			os.Exit(1)
		}
	},
}

var deleteNRDBMSCmd = &cobra.Command{
	Use:     "nrdbms",
	Aliases: []string{"ndb"},
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("nrdbms", &commandTask, cmd.Parent().Use)
		if err := auth.DeleteNRDMFunc(&commandTask); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	importCmd.AddCommand(importNRDBCmd)
	exportCmd.AddCommand(exportNRDBCmd)
	migrationCmd.AddCommand(migrationNRDBCmd)
	deleteCmd.AddCommand(deleteNRDBMSCmd)

	deleteNRDBMSCmd.PersistentFlags().StringVarP(&commandTask.TaskFilePath, "task-file-path", "f", "task.json", "Json file path containing the user's task")
	deleteNRDBMSCmd.Flags().StringArrayVarP(&commandTask.DeleteTableList, "delete-table-list", "D", []string{}, "List of table names to delete")
	deleteNRDBMSCmd.MarkFlagsRequiredTogether("task-file-path", "delete-table-list")
}
