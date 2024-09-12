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
	"github.com/rs/zerolog/log"
)

func ImportRDMFunc(datamoldParams *models.CommandTask) error {
	var RDBC *rdbc.RDBController
	var err error
	log.Info().Msgf("User Information")
	RDBC, err = GetRDMS(&datamoldParams.TargetPoint)
	if err != nil {
		log.Error().Msgf("RDBController error importing into rdbms : %v", err)
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
		log.Error().Msgf("Walk error : %v", err)
		return err
	}

	for _, sqlPath := range sqlList {
		data, err := os.ReadFile(sqlPath)
		if err != nil {
			log.Error().Msgf("ReadFile error : %v", err)
			return err
		}
		log.Info().Msgf("Import start: %s", sqlPath)
		if err := RDBC.Put(string(data)); err != nil {
			log.Error().Msgf("Put error importing into rdbms")
			return err
		}
		log.Info().Msgf("Import success: %s", sqlPath)
	}
	log.Info().Msgf("successfully imported : %s", datamoldParams.Directory)
	return nil
}

func ExportRDMFunc(datamoldParams *models.CommandTask) error {
	var RDBC *rdbc.RDBController
	var err error
	log.Info().Msgf("User Information")
	RDBC, err = GetRDMS(&datamoldParams.TargetPoint)
	if err != nil {
		log.Error().Msgf("RDBController error importing into rdbms : %v", err)
		return err
	}

	err = os.MkdirAll(datamoldParams.Directory, 0755)
	if err != nil {
		log.Error().Msgf("MkdirAll error : %v", err)
		return err
	}

	dbList := []string{}
	if err := RDBC.ListDB(&dbList); err != nil {
		log.Error().Msgf("ListDB error : %v", err)
		return err
	}

	var sqlData string
	for _, db := range dbList {
		sqlData = ""
		log.Info().Msgf("Export start: %s", db)
		if err := RDBC.Get(db, &sqlData); err != nil {
			log.Error().Msgf("Get error : %v", err)
			return err
		}

		file, err := os.Create(filepath.Join(datamoldParams.Directory, fmt.Sprintf("%s.sql", db)))
		if err != nil {
			log.Error().Msgf("File create error : %v", err)
			return err
		}
		defer file.Close()

		_, err = file.WriteString(sqlData)
		if err != nil {
			log.Error().Msgf("File write error : %v", err)
			return err
		}
		log.Info().Msgf("successfully exported : %s", file.Name())
		file.Close()
	}
	log.Info().Msgf("successfully exported : %s", datamoldParams.Directory)
	return nil
}

func MigrationRDMFunc(datamoldParams *models.CommandTask) error {
	var srcRDBC *rdbc.RDBController
	var srcErr error
	var dstRDBC *rdbc.RDBController
	var dstErr error
	log.Info().Msgf("Source Information")
	srcRDBC, srcErr = GetRDMS(&datamoldParams.SourcePoint)
	if srcErr != nil {
		log.Error().Msgf("RDBController error migration into rdbms : %v", srcErr)
		return srcErr
	}
	log.Info().Msgf("Target Information")
	dstRDBC, dstErr = GetRDMS(&datamoldParams.TargetPoint)
	if dstErr != nil {
		log.Error().Msgf("RDBController error migration into rdbms : %v", dstErr)
		return dstErr
	}

	log.Info().Msgf("Launch RDBController Copy")
	if err := srcRDBC.Copy(dstRDBC); err != nil {
		log.Error().Msgf("Copy error copying into rdbms : %v", err)
		return err
	}
	log.Info().Msgf("successfully migrationed")
	return nil
}

func DeleteRDMFunc(datamoldParams *models.CommandTask) error {
	var RDBC *rdbc.RDBController
	var err error
	RDBC, err = GetRDMS(&datamoldParams.TargetPoint)

	if err != nil {
		log.Error().Msgf("RDBController error deleting into rdbms : %v", err)
		return err
	}

	log.Info().Msgf("Launch RDBController Delete")
	if err := RDBC.DeleteDB(datamoldParams.DeleteDBList...); err != nil {
		log.Error().Msgf("Delete error deleting into rdbms : %v", err)
		return err
	}
	log.Info().Msgf("successfully deleted")
	return nil
}
