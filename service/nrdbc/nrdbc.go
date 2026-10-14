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
package nrdbc

import (
	"fmt"
	"slices"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type NRDBMS interface {
	ListTables() ([]string, error)
	ListDatabases() ([]string, error)
	CreateTable(tableName string) error
	DeleteTables(tableName string) error
	ClearTable(tableName string) error
	ImportTable(tableName string, srcData *[]map[string]interface{}) error
	ExportTable(tableName string, dstData *[]map[string]interface{}) error
}

type NRDBController struct {
	client NRDBMS

	logger *zerolog.Logger
}

type Option func(*NRDBController)

func WithLogger(logger *zerolog.Logger) Option {
	return func(n *NRDBController) {
		n.logger = logger
	}
}

func New(nrdb NRDBMS, opts ...Option) (*NRDBController, error) {
	nrdbc := &NRDBController{
		client: nrdb,
	}
	for _, opt := range opts {
		opt(nrdbc)
	}

	return nrdbc, nil
}

// list table
func (nrdbc *NRDBController) ListTables() ([]string, error) {
	tableList, err := nrdbc.client.ListTables()
	if err != nil {
		return tableList, err
	}
	return tableList, nil
}

// list database
func (nrdbc *NRDBController) ListDatabases() ([]string, error) {
	dbList, err := nrdbc.client.ListDatabases()
	if err != nil {
		return dbList, err
	}
	return dbList, nil
}

// create table
func (nrdbc *NRDBController) CreateTable(tableName string) error {
	err := nrdbc.client.CreateTable(tableName)
	if err != nil {
		return err
	}
	return nil
}

// delete table
func (nrdbc *NRDBController) DeleteTables(tableName ...string) error {
	for _, table := range tableName {
		if err := nrdbc.client.DeleteTables(table); err != nil {
			return err
		}
	}
	return nil
}

func (nrdbc *NRDBController) ClearTable(tableName string) error {
	return nrdbc.client.ClearTable(tableName)
}

// put
func (nrdbc *NRDBController) Put(tableName string, srcData *[]map[string]interface{}) error {
	tableList, err := nrdbc.client.ListTables()
	if err != nil {
		nrdbc.logWrite("Error", "ListTables error", err)
		return err
	}

	// 기존 테이블 overwrite: 테이블은 그대로 두고 안에 있는 데이터만 비움
	if slices.Contains(tableList, tableName) {
		if err := nrdbc.client.ClearTable(tableName); err != nil {
			nrdbc.logWrite("Error", "ClearTable error", err)
			return err
		}
		nrdbc.logWrite("Info", fmt.Sprintf("Table cleared: %s", tableName), nil)
	} else {
		if err := nrdbc.client.CreateTable(tableName); err != nil {
			nrdbc.logWrite("Error", "CreateTable error", err)
			return err
		}
		nrdbc.logWrite("Info", fmt.Sprintf("Table creation successful: %s", tableName), nil)
	}

	if err := nrdbc.client.ImportTable(tableName, srcData); err != nil {
		nrdbc.logWrite("Error", "ImportTable error", err)
		return err
	}
	nrdbc.logWrite("Info", fmt.Sprintf("Table import success: %s", tableName), err)
	return nil
}

// get
func (nrdbc *NRDBController) Get(tableName string, dstData *[]map[string]interface{}) error {
	err := nrdbc.client.ExportTable(tableName, dstData)
	if err != nil {
		return err
	}
	return nil
}

// copy
func (src *NRDBController) Copy(dst *NRDBController) error {
	tableList, err := src.client.ListTables()
	if err != nil {
		src.logWrite("Error", "ListTables error", err)
		return err
	}
	log.Debug().Msgf("%+v", tableList)
	for _, table := range tableList {
		log.Debug().Msgf("%+v", table)
		src.logWrite("Info", fmt.Sprintf("Migration start: %s", table), nil)
		data := []map[string]interface{}{}
		src.logWrite("Info", fmt.Sprintf("Extract start: %s", table), nil)
		if err := src.Get(table, &data); err != nil {
			src.logWrite("Error", "Get error", err)
			return err
		}
		src.logWrite("Info", fmt.Sprintf("Import start: %s", table), nil)
		if err := dst.Put(table, &data); err != nil {
			src.logWrite("Error", "Put error", err)
			return err
		}
		src.logWrite("Info", fmt.Sprintf("Migration success: src:/%s -> dst:/%s", table, table), nil)
	}
	return nil
}

func (nrdbc *NRDBController) logWrite(logLevel, msg string, err error) {
	if nrdbc.logger != nil {
		switch logLevel {
		case "Info":
			nrdbc.logger.Info().Msg(msg)
		case "Error":
			nrdbc.logger.Error().Msgf("%s : %v", msg, err)
		}
	}
}
