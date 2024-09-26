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
package controllers

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/service/nrdbc"
	"github.com/cloud-barista/mc-data-manager/service/rdbc"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// RestoreOSPostHandler godoc
//
//	@ID 			RestoreOSPostHandler
//	@Summary		Import data from objectstorage
//	@Description	Import data from a objectstorage  to files.
//	@Tags			[Data Restore], [Service Object Storage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.RestoreTask	true	"Parameters required for Restore"
//	@Success		200			{object}	models.BasicResponse	"Successfully Restore data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/restore/objectstorage [post]
func RestoreOSPostHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Restore-objectstorage", "Import data to objectstorage", start)
	params := models.RestoreTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		log.Error().Msgf("Req params err")
		return ctx.JSON(http.StatusOK, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	switch params.TargetPoint.Provider {
	case string(models.AWS):
		return MigrationLinuxToS3PostHandler(ctx)
	case string(models.NCP):
		return MigrationLinuxToNCPPostHandler(ctx)
	case string(models.GCP):
		return MigrationLinuxToGCPPostHandler(ctx)
	default:
		logger.Error().Msgf("Unsupported provider: %v", params.TargetPoint.Provider)
		errorMsg := fmt.Sprintf("unsupported provider: %v", params.TargetPoint.Provider)
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errorMsg,
		})
	}
}

// RestoreRDBPostHandler godoc
//
//	@ID 			RestoreRDBPostHandler
//	@Summary		Import data from MySQL
//	@Description	Import data from a MySQL database to SQL files.
//	@Tags			[Data Restore], [Service RDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.RestoreTask	true	"Parameters required for Restore"
//	@Success		200			{object}	models.BasicResponse	"Successfully Restore data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/restore/rdb [post]
func RestoreRDBPostHandler(ctx echo.Context) error {

	var RDBC *rdbc.RDBController
	var err error

	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "restore-sql", "Import data to mysql", start)

	params := models.RestoreTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		log.Error().Msgf("Req params err")
		return ctx.JSON(http.StatusOK, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	RDBC, err = auth.GetRDMS(&params.TargetPoint)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get RDBController: %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errorMsg,
		})
	}

	sqlList := []string{}
	err = filepath.Walk(params.SourcePoint.Path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".sql" {
			sqlList = append(sqlList, path)
		}
		return nil
	})
	if err != nil {
		errorMsg := fmt.Sprintf("Walk error: %v", err)
		logger.Error().Msgf(errorMsg)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errorMsg,
		})
	}

	for _, sqlPath := range sqlList {
		data, err := os.ReadFile(sqlPath)
		if err != nil {
			errorMsg := fmt.Sprintf("ReadFile error: %v", err)
			logger.Error().Msgf(errorMsg)
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  &errorMsg,
			})
		}
		log.Info().Msgf("Import start: %s", sqlPath)
		if err := RDBC.Put(string(data)); err != nil {
			errorMsg := fmt.Sprintf("Put error importing into RDBMS: %v", err)
			logger.Error().Msgf(errorMsg)
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  &errorMsg,
			})
		}
		log.Info().Msgf("Import success: %s", sqlPath)
	}

	jobEnd(logger, "Successfully Imported data from mysql", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})

}

// RestoreNRDBPostHandler godoc
//
//	@ID 			RestoreNRDBPostHandler
//	@Summary		Import data from MySQL
//	@Description	Import data from a MySQL database to SQL files.
//	@Tags			[Data Restore], [Service RDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.RestoreTask	true	"Parameters required for Restore"
//	@Success		200			{object}	models.BasicResponse	"Successfully Restore data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/restore/nrdb [post]
func RestoreNRDBPostHandler(ctx echo.Context) error {

	var NRDBC *nrdbc.NRDBController
	var err error
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "Restore-nrdb", "Import data from nrdb", start)

	params := models.RestoreTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		log.Error().Msgf("Req params err")
		return ctx.JSON(http.StatusOK, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	NRDBC, err = auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("NRDBController error importing into nrdbms : %v", err)
		return err
	}

	jsonList := []string{}
	err = filepath.Walk(params.SourcePoint.Path, func(path string, info fs.FileInfo, err error) error {
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
		log.Info().Msgf("successfully imported : %s", params.SourcePoint.Path)
	}

	jobEnd(logger, "Successfully Imported NRDB from Data", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})

}
