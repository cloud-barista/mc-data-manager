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
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	logger *logrus.Logger
	// credential
	credentialPath string
	configData     map[string]map[string]map[string]string

	//src
	cSrcProvider    string
	cSrcAccessKey   string
	cSrcSecretKey   string
	cSrcRegion      string
	cSrcBucketName  string
	cSrcGcpCredPath string
	cSrcProjectID   string
	cSrcEndpoint    string
	cSrcUsername    string
	cSrcPassword    string
	cSrcHost        string
	cSrcPort        string
	cSrcDBName      string

	//dst
	cDstProvider    string
	cDstAccessKey   string
	cDstSecretKey   string
	cDstRegion      string
	cDstBucketName  string
	cDstGcpCredPath string
	cDstProjectID   string
	cDstEndpoint    string
	cDstUsername    string
	cDstPassword    string
	cDstHost        string
	cDstPort        string
	cDstDBName      string

	// dummy
	dstPath  string
	sqlSize  int
	csvSize  int
	jsonSize int
	xmlSize  int
	txtSize  int
	pngSize  int
	gifSize  int
	zipSize  int

	deleteDBList    []string
	deleteTableList []string
)

func logFile() {
	logFile, err := os.Create("app.log")
	if err != nil {
		logrus.Fatal("Failed to create log file")
	}

	logger = logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&CustomTextFormatter{})
	logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cm-data-mold",
	Short: "Data migration validation environment deployment and test data generation tools",
	Long: `It is a tool that builds an environment for verification of data migration technology and 
generates test data necessary for data migration.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logFile()
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}
