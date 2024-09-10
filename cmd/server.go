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
	log "github.com/cloud-barista/mc-data-manager/internal/zerolog"
	"github.com/rs/zerolog"

	dmsv "github.com/cloud-barista/mc-data-manager/websrc/serve"
	"github.com/spf13/cobra"
)

var listenPort string
var allowIP []string

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start Web Server",
	Long:  `Start Web Server`,
	Run: func(cmd *cobra.Command, args []string) {
		log.GetInstance().NewLogEntry().WithCmdName("server").WithJobName("web Server").WithLevel(zerolog.InfoLevel).WithMessage("Start Web Server")
		dmsv.Run(dmsv.InitServer(listenPort, allowIP...), listenPort)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVarP(&listenPort, "port", "P", "3300", "Listen port")
	serverCmd.Flags().StringArrayVarP(&allowIP, "allow-ip", "I", []string{}, "IP addresses and CIDR blocks to allow; example: 192.168.0.1 or 0.0.0.0/0, 10.0.0.0/8")
}
