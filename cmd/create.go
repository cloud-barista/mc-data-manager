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

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creating dummy data of structured/unstructured/semi-structured",
	Long: `Creates structured/unstructured/semi-structured dummy data.

Structured data: creating files for csv, sql

Unstructured data: png,gif,txt,zip

Semi-structured data: json, xml

You must enter the data size in GB.`,
	Run: func(_ *cobra.Command, _ []string) {
		logrus.SetFormatter(&log.CustomTextFormatter{CmdName: "create", JobName: "dummy create"})
		if err := execfunc.DummyCreate(commandTask); err != nil {
			logrus.Errorf("dummy create failed : %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&commandTask.DummyPath, "dst-path", "d", "", "Directory path to create dummy data")
	createCmd.MarkFlagRequired("dst-path")

	createCmd.Flags().StringVarP(&commandTask.SizeSQL, "sql-size", "s", "0", "Total size of sql files")
	createCmd.Flags().StringVarP(&commandTask.SizeCSV, "csv-size", "c", "0", "Total size of csv files")
	createCmd.Flags().StringVarP(&commandTask.SizeJSON, "json-size", "j", "0", "Total size of json files")
	createCmd.Flags().StringVarP(&commandTask.SizeXML, "xml-size", "x", "0", "Total size of xml files")
	createCmd.Flags().StringVarP(&commandTask.SizeTXT, "txt-size", "t", "0", "Total size of txt files")
	createCmd.Flags().StringVarP(&commandTask.SizePNG, "png-size", "p", "0", "Total size of png files")
	createCmd.Flags().StringVarP(&commandTask.SizeGIF, "gif-size", "g", "0", "Total size of gif files")
	createCmd.Flags().StringVarP(&commandTask.SizeZIP, "zip-size", "z", "0", "Total size of zip files")
}
