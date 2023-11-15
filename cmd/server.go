/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/cloud-barista/cm-data-mold/internal/log"
	dmsv "github.com/cloud-barista/cm-data-mold/websrc/serve"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listenPort string

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start Web Server",
	Long:  `Start Web Server`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&log.CustomTextFormatter{CmdName: "server", JobName: "web server"})
		logrus.Info("Start Web Server")
		dmsv.Run(dmsv.InitServer(), listenPort)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serverCmd.Flags().StringVarP(&listenPort, "port", "P", "80", "Listen port")
}
