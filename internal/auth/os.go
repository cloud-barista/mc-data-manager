package auth

import (
	"github.com/cloud-barista/cm-data-mold/service/osc"
	"github.com/sirupsen/logrus"
)

func ImportOSFunc(datamoldParams *DatamoldParams) error {
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	if !datamoldParams.TaskTarget {
		OSC, err = GetSrcOS(datamoldParams)
	} else {
		OSC, err = GetDstOS(datamoldParams)
	}

	if err != nil {
		logrus.Errorf("OSController error importing into objectstorage : %v", err)
		return err
	}

	logrus.Info("Launch OSController MPut")
	if err := OSC.MPut(datamoldParams.DstPath); err != nil {
		logrus.Error("MPut error importing into objectstorage")
		return err
	}
	logrus.Infof("successfully imported : %s", datamoldParams.DstPath)
	return nil
}

func ExportOSFunc(datamoldParams *DatamoldParams) error {
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	if !datamoldParams.TaskTarget {
		OSC, err = GetSrcOS(datamoldParams)
	} else {
		OSC, err = GetDstOS(datamoldParams)
	}
	if err != nil {
		logrus.Errorf("OSController error exporting into objectstorage : %v", err)
		return err
	}

	logrus.Info("Launch OSController MGet")
	if err := OSC.MGet(datamoldParams.DstPath); err != nil {
		logrus.Errorf("MGet error exporting into objectstorage : %v", err)
		return err
	}
	logrus.Infof("successfully exported : %s", datamoldParams.DstPath)
	return nil
}

func MigrationOSFunc(datamoldParams *DatamoldParams) error {
	var src *osc.OSController
	var srcErr error
	var dst *osc.OSController
	var dstErr error
	if !datamoldParams.TaskTarget {
		logrus.Infof("Source Information")
		src, srcErr = GetSrcOS(datamoldParams)
		if srcErr != nil {
			logrus.Errorf("OSController error migration into objectstorage : %v", srcErr)
			return srcErr
		}
		logrus.Infof("Target Information")
		dst, dstErr = GetDstOS(datamoldParams)
		if dstErr != nil {
			logrus.Errorf("OSController error migration into objectstorage : %v", dstErr)
			return dstErr
		}
	} else {
		logrus.Infof("Source Information")
		src, srcErr = GetDstOS(datamoldParams)
		if srcErr != nil {
			logrus.Errorf("OSController error migration into objectstorage : %v", srcErr)
			return srcErr
		}
		logrus.Infof("Target Information")
		dst, dstErr = GetSrcOS(datamoldParams)
		if dstErr != nil {
			logrus.Errorf("OSController error migration into objectstorage : %v", dstErr)
			return dstErr
		}
	}

	logrus.Info("Launch OSController Copy")
	if err := src.Copy(dst); err != nil {
		logrus.Errorf("Copy error copying into objectstorage : %v", err)
		return err
	}
	logrus.Info("successfully migrationed")
	return nil
}

func DeleteOSFunc(datamoldParams *DatamoldParams) error {
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	if !datamoldParams.TaskTarget {
		OSC, err = GetSrcOS(datamoldParams)
	} else {
		OSC, err = GetDstOS(datamoldParams)
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
