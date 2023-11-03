package nrdbc

import "fmt"

type NRDBMS interface {
	ListTables() ([]string, error)
	CreateTable(tableName string) error
	DeleteTables(tableName string) error
	ImportTable(tableName string, srcData *[]map[string]interface{}) error
	ExportTable(tableName string, dstData *[]map[string]interface{}) error
}

type NRDBController struct {
	client NRDBMS
}

// list table
func (nrdbc *NRDBController) ListTables() ([]string, error) {
	return nrdbc.client.ListTables()
}

// create table
func (nrdbc *NRDBController) CreateTable(tableName string) error {
	return nrdbc.client.CreateTable(tableName)
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

	return nrdbc.client.ImportTable(tableName, srcData)
}

// get
func (nrdbc *NRDBController) Get(tableName string, dstData *[]map[string]interface{}) error {
	return nrdbc.client.ExportTable(tableName, dstData)
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

		fmt.Println(table)
	}
	return nil
}

type Option func(*NRDBController)

func New(nrdb NRDBMS, opts ...Option) (*NRDBController, error) {
	nrdbc := &NRDBController{
		client: nrdb,
	}

	for _, opt := range opts {
		opt(nrdbc)
	}

	return nrdbc, nil
}
