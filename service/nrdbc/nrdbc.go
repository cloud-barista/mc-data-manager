package nrdbc

import (
	"fmt"

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
		return tableList, err
	}
	return tableList, nil
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

// put
func (nrdbc *NRDBController) Put(tableName string, srcData *[]map[string]interface{}) error {
	tableList, err := nrdbc.client.ListTables()
	if err != nil {
		nrdbc.logWrite("Error", "ListTables error", err)
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

	for _, table := range tableList {
		src.logWrite("Info", fmt.Sprintf("Replication start: %s", table), nil)
		data := []map[string]interface{}{}
		if err := src.Get(table, &data); err != nil {
			src.logWrite("Error", "Get error", err)
			return err
		}

		if err := dst.Put(table, &data); err != nil {
			src.logWrite("Error", "Put error", err)
			return err
		}
		src.logWrite("Info", fmt.Sprintf("Replication success: src:/%s -> dst:/%s", table, table), nil)
	}
	return nil
}

func (nrdbc *NRDBController) logWrite(logLevel, msg string, err error) {
	if nrdbc.logger != nil {
		switch logLevel {
		case "Info":
			nrdbc.logger.Info(msg)
		case "Error":
			nrdbc.logger.Errorf("%s : %v", msg, err)
		}
	}
}
