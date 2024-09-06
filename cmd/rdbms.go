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

// rdbmsCmd represents the rdbms command
var importRDBCmd = &cobra.Command{
	Use: "rdbms",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("rdbms", &commandTask, cmd.Parent().Use)
		if err := auth.ImportRDMFunc(&commandTask); err != nil {
			os.Exit(1)
		}
	},
}

var exportRDBCmd = &cobra.Command{
	Use: "rdbms",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("rdbms", &commandTask, cmd.Parent().Use)
		if err := auth.ExportRDMFunc(&commandTask); err != nil {
			os.Exit(1)
		}
	},
}

var migrationRDBCmd = &cobra.Command{
	Use: "rdbms",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("rdbms", &commandTask, cmd.Parent().Use)
		if err := auth.MigrationRDMFunc(&commandTask); err != nil {
			os.Exit(1)
		}
	},
}

var deleteRDBMSCmd = &cobra.Command{
	Use: "rdbms",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("rdbms", &commandTask, cmd.Parent().Use)
		if err := auth.DeleteRDMFunc(&commandTask); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	importCmd.AddCommand(importRDBCmd)
	exportCmd.AddCommand(exportRDBCmd)
	migrationCmd.AddCommand(migrationRDBCmd)
	deleteCmd.AddCommand(deleteRDBMSCmd)

	deleteRDBMSCmd.PersistentFlags().StringVarP(&commandTask.TaskFilePath, "task-file-path", "f", "task.json", "Json file path containing the user's task")
	deleteRDBMSCmd.Flags().StringArrayVarP(&commandTask.DeleteDBList, "delete-db-list", "D", []string{}, "List of db names to delete")
	deleteRDBMSCmd.MarkFlagsRequiredTogether("task-file-path", "delete-db-list")
}
