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
	"github.com/cloud-barista/cm-data-mold/internal/execfunc"
	"github.com/cloud-barista/cm-data-mold/internal/log"
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
		if err := execfunc.DummyCreate(datamoldParams); err != nil {
			logrus.Errorf("dummy create failed : %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&datamoldParams.DstPath, "dst-path", "d", "", "Directory path to create dummy data")
	createCmd.MarkFlagRequired("dst-path")

	createCmd.Flags().IntVarP(&datamoldParams.SqlSize, "sql-size", "s", 0, "Total size of sql files")
	createCmd.Flags().IntVarP(&datamoldParams.CsvSize, "csv-size", "c", 0, "Total size of csv files")
	createCmd.Flags().IntVarP(&datamoldParams.JsonSize, "json-size", "j", 0, "Total size of json files")
	createCmd.Flags().IntVarP(&datamoldParams.XmlSize, "xml-size", "x", 0, "Total size of xml files")
	createCmd.Flags().IntVarP(&datamoldParams.TxtSize, "txt-size", "t", 0, "Total size of txt files")
	createCmd.Flags().IntVarP(&datamoldParams.PngSize, "png-size", "p", 0, "Total size of png files")
	createCmd.Flags().IntVarP(&datamoldParams.GifSize, "gif-size", "g", 0, "Total size of gif files")
	createCmd.Flags().IntVarP(&datamoldParams.ZipSize, "zip-size", "z", 0, "Total size of zip files")
}
