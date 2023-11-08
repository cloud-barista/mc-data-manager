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
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/semistructed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/structed"
	"github.com/cloud-barista/cm-data-mold/pkg/dummy/unstructed"
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
	RunE: func(_ *cobra.Command, _ []string) error {
		logrus.SetFormatter(&CustomTextFormatter{cmdName: "create", jobName: "dummy create"})
		logrus.Info("check directory paths")
		if sqlSize != 0 {
			logrus.Info("start sql generation")
			if err := structed.GenerateRandomSQL(dstPath, sqlSize); err != nil {
				logrus.Error("failed to generate sql")
				return err
			}
			logrus.Infof("successfully generated sql : %s", dstPath)
		}

		if csvSize != 0 {
			logrus.Info("start csv generation")
			if err := structed.GenerateRandomCSV(dstPath, csvSize); err != nil {
				logrus.Error("failed to generate csv")
				return err
			}
			logrus.Infof("successfully generated csv : %s", dstPath)
		}

		if jsonSize != 0 {
			logrus.Info("start json generation")
			if err := semistructed.GenerateRandomJSON(dstPath, jsonSize); err != nil {
				logrus.Error("failed to generate json")
				return err
			}
			logrus.Infof("successfully generated json : %s", dstPath)
		}

		if xmlSize != 0 {
			logrus.Info("start xml generation")
			if err := semistructed.GenerateRandomXML(dstPath, xmlSize); err != nil {
				logrus.Error("failed to generate xml")
				return err
			}
			logrus.Infof("successfully generated xml : %s", dstPath)
		}

		if txtSize != 0 {
			logrus.Info("start txt generation")
			if err := unstructed.GenerateRandomTXT(dstPath, txtSize); err != nil {
				logrus.Error("failed to generate txt")
				return err
			}
			logrus.Infof("successfully generated txt : %s", dstPath)
		}

		if pngSize != 0 {
			logrus.Info("start png generation")
			if err := unstructed.GenerateRandomPNGImage(dstPath, pngSize); err != nil {
				logrus.Error("failed to generate png")
				return err
			}
			logrus.Infof("successfully generated png : %s", dstPath)
		}

		if gifSize != 0 {
			logrus.Info("start gif generation")
			if err := unstructed.GenerateRandomGIF(dstPath, gifSize); err != nil {
				logrus.Error("failed to generate gif")
				return err
			}
			logrus.Infof("successfully generated gif : %s", dstPath)
		}

		if zipSize != 0 {
			logrus.Info("start zip generation")
			if err := unstructed.GenerateRandomZIP(dstPath, zipSize); err != nil {
				logrus.Error("failed to generate zip")
				return err
			}
			logrus.Infof("successfully generated zip : %s", dstPath)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&dstPath, "dst-path", "d", "", "Directory path to create dummy data")
	createCmd.MarkFlagRequired("dst-path")

	createCmd.Flags().IntVarP(&sqlSize, "sql-size", "s", 0, "Total size of sql files")
	createCmd.Flags().IntVarP(&csvSize, "csv-size", "c", 0, "Total size of csv files")
	createCmd.Flags().IntVarP(&jsonSize, "json-size", "j", 0, "Total size of json files")
	createCmd.Flags().IntVarP(&xmlSize, "xml-size", "x", 0, "Total size of xml files")
	createCmd.Flags().IntVarP(&txtSize, "txt-size", "t", 0, "Total size of txt files")
	createCmd.Flags().IntVarP(&pngSize, "png-size", "p", 0, "Total size of png files")
	createCmd.Flags().IntVarP(&gifSize, "gif-size", "g", 0, "Total size of gif files")
	createCmd.Flags().IntVarP(&zipSize, "zip-size", "z", 0, "Total size of zip files")
}
