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

	"github.com/cloud-barista/cm-data-mold/internal/auth"
	"github.com/spf13/cobra"
)

var importOSCmd = &cobra.Command{
	Use: "objectstorage",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("objectstorage", &datamoldParams, cmd.Parent().Use)
		if err := auth.ImportOSFunc(&datamoldParams); err != nil {
			os.Exit(1)
		}
	},
}

var exportOSCmd = &cobra.Command{
	Use: "objectstorage",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("objectstorage", &datamoldParams, cmd.Parent().Use)
		if err := auth.ExportOSFunc(&datamoldParams); err != nil {
			os.Exit(1)
		}
	},
}

var migrationOSCmd = &cobra.Command{
	Use: "objectstorage",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("objectstorage", &datamoldParams, cmd.Parent().Use)
		if err := auth.MigrationOSFunc(&datamoldParams); err != nil {
			os.Exit(1)
		}
	},
}

var deleteOSCmd = &cobra.Command{
	Use: "objectstorage",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("objectstorage", &datamoldParams, cmd.Parent().Use)
		if err := auth.DeleteOSFunc(&datamoldParams); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	importCmd.AddCommand(importOSCmd)
	exportCmd.AddCommand(exportOSCmd)
	migrationCmd.AddCommand(migrationOSCmd)
	deleteCmd.AddCommand(deleteOSCmd)

	deleteOSCmd.Flags().StringVarP(&datamoldParams.CredentialPath, "credential-path", "C", "", "Json file path containing the user's credentials")
	deleteOSCmd.MarkFlagRequired("credential-path")
}
