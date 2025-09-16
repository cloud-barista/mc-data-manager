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

var importOSCmd = &cobra.Command{
	Use:     "objectstorage",
	Aliases: []string{"obj"},
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("objectstorage", &commandTask, cmd.Parent().Use)
		if err := auth.ImportOSFunc(&commandTask); err != nil {
			os.Exit(1)
		}
	},
}

// var exportOSCmd = &cobra.Command{
// 	Use:     "objectstorage",
// 	Aliases: []string{"obj"},
// 	Run: func(cmd *cobra.Command, args []string) {
// 		auth.PreRun("objectstorage", &commandTask, cmd.Parent().Use)
// 		if err := auth.ExportOSFunc(&commandTask); err != nil {
// 			os.Exit(1)
// 		}
// 	},
// }

// var migrationOSCmd = &cobra.Command{
// 	Use:     "objectstorage",
// 	Aliases: []string{"obj"},
// 	Run: func(cmd *cobra.Command, args []string) {
// 		auth.PreRun("objectstorage", &commandTask, cmd.Parent().Use)
// 		if err := auth.MigrationOSFunc(&commandTask); err != nil {
// 			os.Exit(1)
// 		}
// 	},
// }

var deleteOSCmd = &cobra.Command{
	Use:     "objectstorage",
	Aliases: []string{"obj"},
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("objectstorage", &commandTask, cmd.Parent().Use)
		if err := auth.DeleteOSFunc(&commandTask); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	importCmd.AddCommand(importOSCmd)
	// exportCmd.AddCommand(exportOSCmd)
	// migrationCmd.AddCommand(migrationOSCmd)
	deleteCmd.AddCommand(deleteOSCmd)

	deleteOSCmd.PersistentFlags().StringVarP(&commandTask.TaskFilePath, "task-file-path", "f", "task.json", "Json file path containing the user's task")
	deleteOSCmd.MarkFlagRequired("task-file-path")
}
