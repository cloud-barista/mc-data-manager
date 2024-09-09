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
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/service/nrdbc"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// BackupOSPostHandler godoc
//
//	@Summary		Export data from objectstorage
//	@Description	Export data from a objectstorage  to files.
//	@Tags			[Data Export], [Object Storage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.BackupTask	true	"Parameters required for backup"
//	@Success		200			{object}	models.BasicResponse	"Successfully backup data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/backup/objectstorage [post]
func BackupOSPostHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit("backup-objectstorage", "Export data from objectstorage", start)
	params := models.BackupTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusOK, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	switch params.SourcePoint.Provider {
	case string(models.AWS):
		return MigrationS3ToLinuxPostHandler(ctx)
	case string(models.NCP):
		return MigrationNCPToLinuxPostHandler(ctx)
	case string(models.GCP):
		return MigrationGCPToLinuxPostHandler(ctx)
	default:
		logger.Errorf("Unsupported provider: %v", params.SourcePoint.Provider)
		errorMsg := fmt.Sprintf("unsupported provider: %v", params.SourcePoint.Provider)
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errorMsg,
		})
	}
}

// BackupMySQLPostHandler godoc
//
//	@Summary		Export data from MySQL
//	@Description	Export data from a MySQL database to SQL files.
//	@Tags			[Data Export], [RDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.BackupTask	true	"Parameters required for backup"
//	@Success		200			{object}	models.BasicResponse	"Successfully backup data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/backup/rdb [post]
func BackupRDBPostHandler(ctx echo.Context) error {

	var err error

	start := time.Now()

	logger, logstrings := pageLogInit("migmysql", "Export data from mysql", start)

	params := models.BackupTask{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusOK, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	rdbc := getMysqlRDBC(logger, start, "smig", params.SourcePoint)
	if rdbc == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	err = os.MkdirAll(params.TargetPoint.Path, 0755)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	dbList := []string{}
	if err := rdbc.ListDB(&dbList); err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	var sqlData string
	for _, db := range dbList {
		sqlData = ""
		if err := rdbc.Get(db, &sqlData); err != nil {
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  nil,
			})
		}

		file, err := os.Create(filepath.Join(params.TargetPoint.Path, fmt.Sprintf("%s.sql", db)))
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  nil,
			})
		}
		defer file.Close()

		_, err = file.WriteString(sqlData)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  nil,
			})
		}
		logrus.Infof("successfully exported : %s", file.Name())
		file.Close()
	}

	jobEnd(logger, "Successfully exported data from mysql", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})

}

// BackupMySQLPostHandler godoc
//
//	@Summary		Export data from MySQL
//	@Description	Export data from a MySQL database to SQL files.
//	@Tags			[Data Export], [RDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.BackupTask	true	"Parameters required for backup"
//	@Success		200			{object}	models.BasicResponse	"Successfully backup data"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/backup/nrdb [post]
func BackupNRDBPostHandler(ctx echo.Context) error {

	var NRDBC *nrdbc.NRDBController
	var err error
	start := time.Now()

	logger, logstrings := pageLogInit("backup-nrdb", "backup data from nrdb", start)

	params := models.BackupTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusOK, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	NRDBC, err = auth.GetNRDMS(&params.SourcePoint)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	err = os.MkdirAll(params.TargetPoint.Path, 0755)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	tableList, err := NRDBC.ListTables()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	var dstData []map[string]interface{}
	for _, table := range tableList {
		logrus.Infof("Export start: %s", table)
		dstData = []map[string]interface{}{}

		if err := NRDBC.Get(table, &dstData); err != nil {
			logrus.Errorf("Get error : %v", err)
			return err
		}

		file, err := os.Create(filepath.Join(params.TargetPoint.Path, fmt.Sprintf("%s.json", table)))
		if err != nil {
			logrus.Errorf("File create error : %v", err)
			return err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(dstData); err != nil {
			logrus.Errorf("data encoding error : %v", err)
			return err
		}
		logrus.Infof("successfully exported : %s", file.Name())
	}

	jobEnd(logger, "Successfully exported data from mysql", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})

}
