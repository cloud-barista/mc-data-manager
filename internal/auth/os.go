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
package auth

import (
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/service/osc"
	"github.com/sirupsen/logrus"
)

func ImportOSFunc(params *models.CommandTask) error {
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	OSC, err = GetOS(&params.TargetPoint)
	if err != nil {
		logrus.Errorf("OSController error importing into objectstorage : %v", err)
		return err
	}

	logrus.Info("Launch OSController MPut")
	if err := OSC.MPut(params.Directory); err != nil {
		logrus.Error("MPut error importing into objectstorage")
		logrus.Infof("params : %+v", params.TargetPoint)

		return err
	}
	logrus.Infof("successfully imported : %s", params.Directory)
	return nil
}

func ExportOSFunc(params *models.CommandTask) error {
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	OSC, err = GetOS(&params.TargetPoint)
	if err != nil {
		logrus.Errorf("OSController error importing into objectstorage : %v", err)
		return err
	}

	logrus.Info("Launch OSController MGet")
	if err := OSC.MGet(params.Directory); err != nil {
		logrus.Errorf("MGet error exporting into objectstorage : %v", err)
		return err
	}
	logrus.Infof("successfully exported : %s", params.Directory)
	return nil
}

func MigrationOSFunc(params *models.CommandTask) error {
	var src *osc.OSController
	var srcErr error
	var dst *osc.OSController
	var dstErr error
	logrus.Infof("Source Information")
	src, srcErr = GetOS(&params.TargetPoint)
	if srcErr != nil {
		logrus.Errorf("OSController error migration into objectstorage : %v", srcErr)
		return srcErr
	}
	logrus.Infof("Target Information")
	dst, dstErr = GetOS(&params.TargetPoint)
	if dstErr != nil {
		logrus.Errorf("OSController error migration into objectstorage : %v", dstErr)
		return dstErr
	}

	logrus.Info("Launch OSController Copy")
	if err := src.Copy(dst); err != nil {
		logrus.Errorf("Copy error copying into objectstorage : %v", err)
		return err
	}
	logrus.Info("successfully migrationed")
	return nil
}

func DeleteOSFunc(params *models.CommandTask) error {
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	OSC, err = GetOS(&params.TargetPoint)
	if err != nil {
		logrus.Errorf("OSController error importing into objectstorage : %v", err)
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
