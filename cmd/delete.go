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
		err := os.RemoveAll(dstPath)
		if err != nil {
			logger.Error("Failed to delete")
			return err
		}
		logger.Info("Deletion success: %s\n", dstPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteDummyCmd)

	deleteDummyCmd.Flags().StringVarP(&dstPath, "dst-path", "d", "", "Delete data in directory paths")
	deleteDummyCmd.MarkFlagRequired("dst-path")
}
