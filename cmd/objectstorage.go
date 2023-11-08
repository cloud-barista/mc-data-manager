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

	"github.com/cloud-barista/cm-data-mold/service/osc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var importOSCmd = &cobra.Command{
	Use:    "objectstorage",
	PreRun: preRun("objectstorage"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := importOSFunc(); err != nil {
			os.Exit(1)
		}
	},
}

var exportOSCmd = &cobra.Command{
	Use:    "objectstorage",
	PreRun: preRun("objectstorage"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := exportOSFunc(); err != nil {
			os.Exit(1)
		}
	},
}

var replicationOSCmd = &cobra.Command{
	Use:    "objectstorage",
	PreRun: preRun("objectstorage"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := replicationOSFunc(); err != nil {
			os.Exit(1)
		}
	},
}

var deleteOSCmd = &cobra.Command{
	Use:    "objectstorage",
	PreRun: preRun("objectstorage"),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteOSFunc(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	importCmd.AddCommand(importOSCmd)
	exportCmd.AddCommand(exportOSCmd)
	replicationCmd.AddCommand(replicationOSCmd)
	deleteCmd.AddCommand(deleteOSCmd)

	deleteOSCmd.Flags().StringVarP(&credentialPath, "credential-path", "C", "", "Json file path containing the user's credentials")
	deleteOSCmd.MarkFlagRequired("credential-path")
}

func importOSFunc() error {
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	if !taskTarget {
		OSC, err = getSrcOS()
	} else {
		OSC, err = getDstOS()
	}

	if err != nil {
		logrus.Errorf("OSController error importing into objectstorage : %v", err)
		return err
	}

	logrus.Info("Launch OSController MPut")
	if err := OSC.MPut(dstPath); err != nil {
		logrus.Error("MPut error importing into objectstorage")
		return err
	}
	logrus.Infof("successfully imported : %s", dstPath)
	return nil
}

func exportOSFunc() error {
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	if !taskTarget {
		OSC, err = getSrcOS()
	} else {
		OSC, err = getDstOS()
	}
	if err != nil {
		logrus.Errorf("OSController error exporting into objectstorage : %v", err)
		return err
	}

	logrus.Info("Launch OSController MGet")
	if err := OSC.MGet(dstPath); err != nil {
		logrus.Errorf("MGet error exporting into objectstorage : %v", err)
		return err
	}
	logrus.Infof("successfully exported : %s", dstPath)
	return nil
}

func replicationOSFunc() error {
	var src *osc.OSController
	var srcErr error
	var dst *osc.OSController
	var dstErr error
	if !taskTarget {
		logrus.Infof("Source Information")
		src, srcErr = getSrcOS()
		if srcErr != nil {
			logrus.Errorf("OSController error replication into objectstorage : %v", srcErr)
			return srcErr
		}
		logrus.Infof("Target Information")
		dst, dstErr = getDstOS()
		if dstErr != nil {
			logrus.Errorf("OSController error replication into objectstorage : %v", dstErr)
			return dstErr
		}
	} else {
		logrus.Infof("Source Information")
		src, srcErr = getDstOS()
		if srcErr != nil {
			logrus.Errorf("OSController error replication into objectstorage : %v", srcErr)
			return srcErr
		}
		logrus.Infof("Target Information")
		dst, dstErr = getSrcOS()
		if dstErr != nil {
			logrus.Errorf("OSController error replication into objectstorage : %v", dstErr)
			return dstErr
		}
	}

	logrus.Info("Launch OSController Copy")
	if err := src.Copy(dst); err != nil {
		logrus.Errorf("Copy error copying into objectstorage : %v", err)
		return err
	}
	logrus.Info("successfully replicationed")
	return nil
}

func deleteOSFunc() error {
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	if !taskTarget {
		OSC, err = getSrcOS()
	} else {
		OSC, err = getDstOS()
	}
	if err != nil {
		logrus.Errorf("OSController error deleting into objectstorage : %v", err)
		return err
	}

	logrus.Info("Launch OSController Delete")
	if err := OSC.DeleteBucket(); err != nil {
		logrus.Errorf("Delete error deleting into objectstorage : %v", err)
		return err
	}
	logrus.Info("successfully deleted")

	return nil
}
