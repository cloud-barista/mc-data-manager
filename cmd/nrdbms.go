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
	"os"

	"github.com/cloud-barista/cm-data-mold/internal/auth"
	"github.com/spf13/cobra"
)

// nrdbmsCmd represents the nrdbms command
var importNRDBCmd = &cobra.Command{
	Use:     "nrdbms",
	Aliases: []string{"ndb"},
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("nrdbms", &datamoldParams, cmd.Parent().Use)
		if err := auth.ImportNRDMFunc(&datamoldParams); err != nil {
			os.Exit(1)
		}
	},
}

var exportNRDBCmd = &cobra.Command{
	Use: "nrdbms",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("nrdbms", &datamoldParams, cmd.Parent().Use)
		if err := auth.ExportNRDMFunc(&datamoldParams); err != nil {
			os.Exit(1)
		}
	},
}

var migrationNRDBCmd = &cobra.Command{
	Use: "nrdbms",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("nrdbms", &datamoldParams, cmd.Parent().Use)
		if err := auth.MigrationNRDMFunc(&datamoldParams); err != nil {
			os.Exit(1)
		}
	},
}

var deleteNRDBMSCmd = &cobra.Command{
	Use: "nrdbms",
	Run: func(cmd *cobra.Command, args []string) {
		auth.PreRun("nrdbms", &datamoldParams, cmd.Parent().Use)
		if err := auth.DeleteNRDMFunc(&datamoldParams); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	importCmd.AddCommand(importNRDBCmd)
	exportCmd.AddCommand(exportNRDBCmd)
	migrationCmd.AddCommand(migrationNRDBCmd)
	deleteCmd.AddCommand(deleteNRDBMSCmd)

	deleteNRDBMSCmd.Flags().StringVarP(&datamoldParams.CredentialPath, "credential-path", "C", "", "Json file path containing the user's credentials")
	deleteNRDBMSCmd.Flags().StringArrayVarP(&datamoldParams.DeleteTableList, "delete-table-list", "D", []string{}, "List of table names to delete")
	deleteNRDBMSCmd.MarkFlagsRequiredTogether("credential-path", "delete-table-list")
}
