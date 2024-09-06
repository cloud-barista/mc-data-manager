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
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export dummy data from the service",
	Long:  `Export data locally with the credentials you have entered.`,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.PersistentFlags().StringVarP(&commandTask.TaskFilePath, "task-file-path", "f", "task.json", "Json file path containing the user's task")
	exportCmd.PersistentFlags().StringVarP(&commandTask.Directory, "dst-path", "d", "", "Directory path to export data")
	exportCmd.MarkFlagsRequiredTogether("task-file-path", "dst-path")
}
