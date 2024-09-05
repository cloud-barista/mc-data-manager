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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/service/rdbc"
	"github.com/sirupsen/logrus"
)

func ImportRDMFunc(datamoldParams *models.CommandTask) error {
	var RDBC *rdbc.RDBController
	var err error
	logrus.Infof("User Information")
	RDBC, err = GetRDMS(&datamoldParams.TargetPoint)
	if err != nil {
		logrus.Errorf("RDBController error importing into rdbms : %v", err)
		return err
	}

	sqlList := []string{}
	err = filepath.Walk(datamoldParams.Directory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".sql" {
			sqlList = append(sqlList, path)
		}
		return nil
	})
	if err != nil {
		logrus.Errorf("Walk error : %v", err)
		return err
	}

	for _, sqlPath := range sqlList {
		data, err := os.ReadFile(sqlPath)
		if err != nil {
			logrus.Errorf("ReadFile error : %v", err)
			return err
		}
		logrus.Infof("Import start: %s", sqlPath)
		if err := RDBC.Put(string(data)); err != nil {
			logrus.Error("Put error importing into rdbms")
			return err
		}
		logrus.Infof("Import success: %s", sqlPath)
	}
	logrus.Infof("successfully imported : %s", datamoldParams.Directory)
	return nil
}

func ExportRDMFunc(datamoldParams *models.CommandTask) error {
	var RDBC *rdbc.RDBController
	var err error
	logrus.Infof("User Information")
	RDBC, err = GetRDMS(&datamoldParams.TargetPoint)
	if err != nil {
		logrus.Errorf("RDBController error importing into rdbms : %v", err)
		return err
	}

	err = os.MkdirAll(datamoldParams.Directory, 0755)
	if err != nil {
		logrus.Errorf("MkdirAll error : %v", err)
		return err
	}

	dbList := []string{}
	if err := RDBC.ListDB(&dbList); err != nil {
		logrus.Errorf("ListDB error : %v", err)
		return err
	}

	var sqlData string
	for _, db := range dbList {
		sqlData = ""
		logrus.Infof("Export start: %s", db)
		if err := RDBC.Get(db, &sqlData); err != nil {
			logrus.Errorf("Get error : %v", err)
			return err
		}

		file, err := os.Create(filepath.Join(datamoldParams.Directory, fmt.Sprintf("%s.sql", db)))
		if err != nil {
			logrus.Errorf("File create error : %v", err)
			return err
		}
		defer file.Close()

		_, err = file.WriteString(sqlData)
		if err != nil {
			logrus.Errorf("File write error : %v", err)
			return err
		}
		logrus.Infof("successfully exported : %s", file.Name())
		file.Close()
	}
	logrus.Infof("successfully exported : %s", datamoldParams.Directory)
	return nil
}

func MigrationRDMFunc(datamoldParams *models.CommandTask) error {
	var srcRDBC *rdbc.RDBController
	var srcErr error
	var dstRDBC *rdbc.RDBController
	var dstErr error
	logrus.Infof("Source Information")
	srcRDBC, srcErr = GetRDMS(&datamoldParams.SourcePoint)
	if srcErr != nil {
		logrus.Errorf("RDBController error migration into rdbms : %v", srcErr)
		return srcErr
	}
	logrus.Infof("Target Information")
	dstRDBC, dstErr = GetRDMS(&datamoldParams.TargetPoint)
	if dstErr != nil {
		logrus.Errorf("RDBController error migration into rdbms : %v", dstErr)
		return dstErr
	}

	logrus.Info("Launch RDBController Copy")
	if err := srcRDBC.Copy(dstRDBC); err != nil {
		logrus.Errorf("Copy error copying into rdbms : %v", err)
		return err
	}
	logrus.Info("successfully migrationed")
	return nil
}

func DeleteRDMFunc(datamoldParams *models.CommandTask) error {
	var RDBC *rdbc.RDBController
	var err error
	RDBC, err = GetRDMS(&datamoldParams.TargetPoint)

	if err != nil {
		logrus.Errorf("RDBController error deleting into rdbms : %v", err)
		return err
	}

	logrus.Info("Launch RDBController Delete")
	if err := RDBC.DeleteDB(datamoldParams.DeleteDBList...); err != nil {
		logrus.Errorf("Delete error deleting into rdbms : %v", err)
		return err
	}
	logrus.Info("successfully deleted")
	return nil
}
