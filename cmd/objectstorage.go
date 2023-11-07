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
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/spf13/cobra"
)

var importOSCmd = &cobra.Command{
	Use: "objectstorage",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "objectstorage")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "importOSFunc", "import start", nil)
		return importOSFunc()
	},
}

var exportOSCmd = &cobra.Command{
	Use: "objectstorage",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "objectstorage")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "exportOSFunc", "export start", nil)
		return exportOSFunc()
	},
}

var replicationOSCmd = &cobra.Command{
	Use: "objectstorage",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "objectstorage")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "replicationOSFunc", "replication start", nil)
		return replicationOSFunc()
	},
}

var deleteOSCmd = &cobra.Command{
	Use: "objectstorage",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return preRunE(cmd.Parent().Use, "objectstorage")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.LogWirte(logger, "Info", "deleteOSFunc", "delete start", nil)
		return deleteOSFunc()
	},
}

func init() {
	importCmd.AddCommand(importOSCmd)
	exportCmd.AddCommand(exportOSCmd)
	replicationCmd.AddCommand(replicationOSCmd)
	deleteCmd.AddCommand(deleteOSCmd)

	deleteOSCmd.Flags().StringVarP(&credentialPath, "credential-path", "C", "", "Json file path containing the user's credentials")
	deleteOSCmd.MarkFlagRequired("credential-path")
}

func importOSFunc() error {
	OSC, err := getSrcOS()
	if err != nil {
		return err
	}
	return OSC.MPut(dstPath)
}

func exportOSFunc() error {
	OSC, err := getSrcOS()
	if err != nil {
		return err
	}
	return OSC.MGet(dstPath)
}

func replicationOSFunc() error {
	src, err := getSrcOS()
	if err != nil {
		return err
	}

	dst, err := getDstOS()
	if err != nil {
		return err
	}

	return src.Copy(dst)
}

func deleteOSFunc() error {
	OSC, err := getSrcOS()
	if err != nil {
		return err
	}

	if err := OSC.DeleteBucket(); err != nil {
		return err
	}

	return nil
}
