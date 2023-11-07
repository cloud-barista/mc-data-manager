package nrdbc

import (
	"fmt"

	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/sirupsen/logrus"
)

type NRDBMS interface {
	ListTables() ([]string, error)
	CreateTable(tableName string) error
	DeleteTables(tableName string) error
	ImportTable(tableName string, srcData *[]map[string]interface{}) error
	ExportTable(tableName string, dstData *[]map[string]interface{}) error
}

type NRDBController struct {
	client NRDBMS

	logger *logrus.Logger
}

type Option func(*NRDBController)

func WithLogger(logger *logrus.Logger) Option {
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
		utils.LogWirte(nrdbc.logger, "Error", "ListTables", "Failed to get table list", err)
		return tableList, err
	}
	utils.LogWirte(nrdbc.logger, "Info", "ListTables", "Get Table List Successfully", nil)
	return tableList, nil
}

// create table
func (nrdbc *NRDBController) CreateTable(tableName string) error {
	err := nrdbc.client.CreateTable(tableName)
	if err != nil {
		utils.LogWirte(nrdbc.logger, "Error", "CreateTable", "Failed to create table", err)
		return err
	}
	utils.LogWirte(nrdbc.logger, "Info", "CreateTable", "Table creation successful", nil)
	return nil
}

// delete table
func (nrdbc *NRDBController) DeleteTables(tableName ...string) error {
	for _, table := range tableName {
		if err := nrdbc.client.DeleteTables(table); err != nil {
			utils.LogWirte(nrdbc.logger, "Error", "DeleteTables", "Failed to delete table", err)
			return err
		}
		utils.LogWirte(nrdbc.logger, "Info", "DeleteTables", fmt.Sprintf("%s Deletion Successful", table), nil)
	}
	return nil
}

// put
func (nrdbc *NRDBController) Put(tableName string, srcData *[]map[string]interface{}) error {
	tableList, err := nrdbc.client.ListTables()
	if err != nil {
		return err
	}

	isTable := false
	for _, table := range tableList {
		if table == tableName {
			isTable = true
			break
		}
	}

	if !isTable {
		if err := nrdbc.client.CreateTable(tableName); err != nil {
			return err
		}
	}

	if err := nrdbc.client.ImportTable(tableName, srcData); err != nil {
		utils.LogWirte(nrdbc.logger, "Error", "ImportTable", "Table import failed", err)
		return err
	}
	utils.LogWirte(nrdbc.logger, "Info", "ImportTable", "Table Import Successful", nil)
	return nil
}

// get
func (nrdbc *NRDBController) Get(tableName string, dstData *[]map[string]interface{}) error {
	err := nrdbc.client.ExportTable(tableName, dstData)
	if err != nil {
		utils.LogWirte(nrdbc.logger, "Error", "ExportTable", "Export table failed", err)
		return err
	}
	utils.LogWirte(nrdbc.logger, "Info", "ExportTable", "Table Export Successful", nil)
	return nil
}

// copy
func (src *NRDBController) Copy(dst *NRDBController) error {
	tableList, err := src.client.ListTables()
	if err != nil {
		return err
	}

	for _, table := range tableList {

		data := []map[string]interface{}{}
		if err := src.Get(table, &data); err != nil {
			return err
		}

		if err := dst.Put(table, &data); err != nil {
			return err
		}

		utils.LogWirte(src.logger, "Info", "Copy", fmt.Sprintf("%s Copied", table), nil)
	}
	utils.LogWirte(src.logger, "Info", "Copy", "Replication Done", nil)
	return nil
}
