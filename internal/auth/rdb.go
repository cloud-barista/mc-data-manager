package auth

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cloud-barista/cm-data-mold/service/rdbc"
	"github.com/sirupsen/logrus"
)

func ImportRDMFunc(datamoldParams *DatamoldParams) error {
	var RDBC *rdbc.RDBController
	var err error
	logrus.Infof("User Information")
	if !datamoldParams.TaskTarget {
		RDBC, err = GetSrcRDMS(datamoldParams)
	} else {
		RDBC, err = GetDstRDMS(datamoldParams)
	}
	if err != nil {
		logrus.Errorf("RDBController error importing into rdbms : %v", err)
		return err
	}

	sqlList := []string{}
	err = filepath.Walk(datamoldParams.DstPath, func(path string, info fs.FileInfo, err error) error {
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
	logrus.Infof("successfully imported : %s", datamoldParams.DstPath)
	return nil
}

func ExportRDMFunc(datamoldParams *DatamoldParams) error {
	var RDBC *rdbc.RDBController
	var err error
	logrus.Infof("User Information")
	if !datamoldParams.TaskTarget {
		RDBC, err = GetSrcRDMS(datamoldParams)
	} else {
		RDBC, err = GetDstRDMS(datamoldParams)
	}
	if err != nil {
		logrus.Errorf("RDBController error exporting into rdbms : %v", err)
		return err
	}

	err = os.MkdirAll(datamoldParams.DstPath, 0755)
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

		file, err := os.Create(filepath.Join(datamoldParams.DstPath, fmt.Sprintf("%s.sql", db)))
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
	logrus.Infof("successfully exported : %s", datamoldParams.DstPath)
	return nil
}

func MigrationRDMFunc(datamoldParams *DatamoldParams) error {
	var srcRDBC *rdbc.RDBController
	var srcErr error
	var dstRDBC *rdbc.RDBController
	var dstErr error
	if !datamoldParams.TaskTarget {
		logrus.Infof("Source Information")
		srcRDBC, srcErr = GetSrcRDMS(datamoldParams)
		if srcErr != nil {
			logrus.Errorf("RDBController error migration into rdbms : %v", srcErr)
			return srcErr
		}
		logrus.Infof("Target Information")
		dstRDBC, dstErr = GetDstRDMS(datamoldParams)
		if dstErr != nil {
			logrus.Errorf("RDBController error migration into rdbms : %v", dstErr)
			return dstErr
		}
	} else {
		logrus.Infof("Source Information")
		srcRDBC, srcErr = GetDstRDMS(datamoldParams)
		if srcErr != nil {
			logrus.Errorf("RDBController error migration into rdbms : %v", srcErr)
			return srcErr
		}
		logrus.Infof("Target Information")
		dstRDBC, dstErr = GetSrcRDMS(datamoldParams)
		if dstErr != nil {
			logrus.Errorf("RDBController error migration into rdbms : %v", dstErr)
			return dstErr
		}
	}

	logrus.Info("Launch RDBController Copy")
	if err := srcRDBC.Copy(dstRDBC); err != nil {
		logrus.Errorf("Copy error copying into rdbms : %v", err)
		return err
	}
	logrus.Info("successfully migrationed")
	return nil
}

func DeleteRDMFunc(datamoldParams *DatamoldParams) error {
	var RDBC *rdbc.RDBController
	var err error
	if !datamoldParams.TaskTarget {
		RDBC, err = GetSrcRDMS(datamoldParams)
	} else {
		RDBC, err = GetDstRDMS(datamoldParams)
	}
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
