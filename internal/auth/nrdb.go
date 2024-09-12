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
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/service/nrdbc"
	"github.com/rs/zerolog/log"
)

func ImportNRDMFunc(params *models.CommandTask) error {
	var NRDBC *nrdbc.NRDBController
	var err error
	NRDBC, err = GetNRDMS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("NRDBController error importing into nrdbms : %v", err)
		return err
	}

	jsonList := []string{}
	err = filepath.Walk(params.Directory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".json" {
			jsonList = append(jsonList, path)
		}
		return nil
	})

	if err != nil {
		log.Error().Msgf("Walk error : %v", err)
		return err
	}

	var srcData []map[string]interface{}
	for _, jsonFile := range jsonList {
		srcData = []map[string]interface{}{}

		file, err := os.Open(jsonFile)
		if err != nil {
			log.Error().Msgf("file open error : %v", err)
			return err
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&srcData); err != nil {
			log.Error().Msgf("file decoding error : %v", err)
			return err
		}

		fileName := filepath.Base(jsonFile)
		tableName := fileName[:len(fileName)-len(filepath.Ext(fileName))]

		log.Info().Msgf("Import start: %s", fileName)
		if err := NRDBC.Put(tableName, &srcData); err != nil {
			log.Error().Msgf("Put error importing into nrdbms")
			return err
		}
		log.Info().Msgf("successfully imported : %s", params.Directory)
	}
	return nil
}

func ExportNRDMFunc(params *models.CommandTask) error {
	var NRDBC *nrdbc.NRDBController
	var err error
	NRDBC, err = GetNRDMS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("NRDBController error importing into nrdbms : %v", err)
		return err
	}

	tableList, err := NRDBC.ListTables()
	if err != nil {
		log.Error().Msgf("ListTables error : %v", err)
		return err
	}

	var dstData []map[string]interface{}
	for _, table := range tableList {
		log.Info().Msgf("Export start: %s", table)
		dstData = []map[string]interface{}{}

		if err := NRDBC.Get(table, &dstData); err != nil {
			log.Error().Msgf("Get error : %v", err)
			return err
		}

		file, err := os.Create(filepath.Join(params.Directory, fmt.Sprintf("%s.json", table)))
		if err != nil {
			log.Error().Msgf("File create error : %v", err)
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(dstData); err != nil {
			log.Error().Msgf("data encoding error : %v", err)
			return err
		}
		log.Info().Msgf("successfully exported : %s", file.Name())
	}
	log.Info().Msgf("successfully exported : %s", params.Directory)
	return nil
}

func MigrationNRDMFunc(params *models.CommandTask) error {
	var srcNRDBC *nrdbc.NRDBController
	var srcErr error
	var dstNRDBC *nrdbc.NRDBController
	var dstErr error
	log.Info().Msgf("Source Information")
	srcNRDBC, srcErr = GetNRDMS(&params.SourcePoint)
	if srcErr != nil {
		log.Error().Msgf("NRDBController error migration into nrdbms : %v", srcErr)
		return srcErr
	}
	log.Info().Msgf("Target Information")
	dstNRDBC, dstErr = GetNRDMS(&params.TargetPoint)
	if dstErr != nil {
		log.Error().Msgf("NRDBController error migration into nrdbms : %v", dstErr)
		return dstErr
	}

	log.Info().Msgf("Launch NRDBController Copy")
	if err := srcNRDBC.Copy(dstNRDBC); err != nil {
		log.Error().Msgf("Copy error copying into nrdbms : %v", err)
		return err
	}
	log.Info().Msgf("successfully migrationed")
	return nil
}

func DeleteNRDMFunc(params *models.CommandTask) error {
	var NRDBC *nrdbc.NRDBController
	var err error
	NRDBC, err = GetNRDMS(&params.TargetPoint)

	if err != nil {
		log.Error().Msgf("NRDBController error deleting into nrdbms : %v", err)
		return err
	}

	log.Info().Msgf("Launch NRDBController Delete")
	if err := NRDBC.DeleteTables(params.DeleteTableList...); err != nil {
		log.Error().Msgf("Delete error deleting into nrdbms : %v", err)
		return err
	}
	log.Info().Msgf("successfully deleted")
	return nil
}
