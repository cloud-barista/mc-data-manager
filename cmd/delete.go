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

	"github.com/cloud-barista/mc-data-manager/internal/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete local data",
	Long: `Delete unstructured, semi-structured, and structured data, 
which are CSP or local dummy data`,
}

var deleteLocalCmd = &cobra.Command{
	Use: "local",

	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&log.CustomTextFormatter{CmdName: "delete"})
		logrus.WithFields(logrus.Fields{"jobName": "local delete"}).Info("start deleting local data")

		if err := os.RemoveAll(commandTask.Directory); err != nil {
			logrus.WithFields(logrus.Fields{"jobName": "local delete"}).Errorf("failed to delete local : %v", err)
			return
		}
		logrus.Infof("successfully deleted : %s\n", commandTask.Directory)

	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteLocalCmd)

	deleteLocalCmd.Flags().StringVarP(&commandTask.Directory, "dst-path", "d", "", "Delete data in directory paths")
	deleteLocalCmd.MarkFlagRequired("dst-path")
}
