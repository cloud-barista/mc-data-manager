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
	"github.com/cloud-barista/mc-data-manager/internal/execfunc"
	"github.com/cloud-barista/mc-data-manager/internal/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// testCmd represents the create command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test-command",
	Long:  `test-command`,
	Run: func(_ *cobra.Command, _ []string) {
		logrus.SetFormatter(&log.CustomTextFormatter{CmdName: "test", JobName: "test dummy create"})
		if err := execfunc.DummyCreate(commandTask); err != nil {
			logrus.Errorf("test dummy create failed : %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().StringVarP(&commandTask.DummyPath, "dst-path", "d", "", "Directory path to create dummy data")
	testCmd.MarkFlagRequired("dst-path")

	testCmd.Flags().StringVarP(&commandTask.SizeSQL, "sql-size", "s", "0", "Total size of sql files")
	testCmd.Flags().StringVarP(&commandTask.SizeCSV, "csv-size", "c", "0", "Total size of csv files")
	testCmd.Flags().StringVarP(&commandTask.SizeJSON, "json-size", "j", "0", "Total size of json files")
	testCmd.Flags().StringVarP(&commandTask.SizeXML, "xml-size", "x", "0", "Total size of xml files")
	testCmd.Flags().StringVarP(&commandTask.SizeTXT, "txt-size", "t", "0", "Total size of txt files")
	testCmd.Flags().StringVarP(&commandTask.SizePNG, "png-size", "p", "0", "Total size of png files")
	testCmd.Flags().StringVarP(&commandTask.SizeGIF, "gif-size", "g", "0", "Total size of gif files")
	testCmd.Flags().StringVarP(&commandTask.SizeZIP, "zip-size", "z", "0", "Total size of zip files")
}
