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

	"github.com/cloud-barista/cm-data-mold/internal/logformatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete dummy data",
	Long: `Delete unstructured, semi-structured, and structured data, 
which are CSP or local dummy data`,
}

var deleteDummyCmd = &cobra.Command{
	Use: "dummy",

	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.SetFormatter(&logformatter.CustomTextFormatter{CmdName: "delete"})
		logrus.WithFields(logrus.Fields{"jobName": "dummy delete"}).Info("start deleting dummy")
		err := os.RemoveAll(dstPath)
		if err != nil {
			logrus.WithFields(logrus.Fields{"jobName": "dummy delete"}).Errorf("failed to delete dummy : %v", err)
			return err
		}
		logrus.Infof("successfully deleted : %s\n", dstPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteDummyCmd)

	deleteCmd.PersistentFlags().BoolVarP(&taskTarget, "task", "T", false, "Select a destination(src, dst) to work with in the credential-path")
	deleteDummyCmd.Flags().StringVarP(&dstPath, "dst-path", "d", "", "Delete data in directory paths")
	deleteDummyCmd.MarkFlagRequired("dst-path")
}
