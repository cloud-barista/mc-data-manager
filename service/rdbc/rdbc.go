package rdbc

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/sirupsen/logrus"
)

type EngineType string

const (
	Mysql EngineType = "mysql"
)

// rdbms interface
//
// Configure the interface to make it easier for other DBs to apply in the future
type RDBMS interface {
	Exec(query string) error
	ListDB(dst *[]string) error
	DeleteDB(dbName string) error
	ListTable(dbName string, dst *[]string) error
	ShowCreateDBSql(dbName string, dbCreateSql *string) error
	ShowCreateTableSql(dbName, tableName string, tableCreateSql *string) error
	GetInsert(dbName, tableName string, insertSql *[]string) error
}

type RDBController struct {
	client RDBMS

	logger *logrus.Logger
}

type Option func(*RDBController)

func WithLogger(logger *logrus.Logger) Option {
	return func(r *RDBController) {
		r.logger = logger
	}
}

func New(rdb RDBMS, opts ...Option) (*RDBController, error) {
	rdbc := &RDBController{
		client: rdb,
		logger: nil,
	}

	for _, opt := range opts {
		opt(rdbc)
	}

	return rdbc, nil
}

// Return db list
func (rdb *RDBController) ListDB(dst *[]string) error {
	err := rdb.client.ListDB(dst)
	if err != nil {
		utils.LogWirte(rdb.logger, "Error", "ListDB", "get listDB failed", err)
		return err
	}
	utils.LogWirte(rdb.logger, "Info", "ListDB", "get listDB success", nil)
	return nil
}

// sql import
func (rdb *RDBController) Put(sql string) error {
	scanner := bufio.NewScanner(strings.NewReader(sql))
	scanner.Split(splitLine)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ReplaceAll(line, "\n", "")
		if line != "" {
			err := rdb.client.Exec(line)
			if err != nil {
				utils.LogWirte(rdb.logger, "Error", "Put", "db exec failed", err)
				return err
			}
		}
	}

	err := scanner.Err()
	if err != nil {
		utils.LogWirte(rdb.logger, "Error", "Put", "scan failed", err)
		return err
	}
	return nil
}

// Replication using put and get
func (rdb *RDBController) Copy(dst *RDBController) error {
	var dbList []string
	var sql string
	if err := rdb.ListDB(&dbList); err != nil {
		return err
	}

	for _, db := range dbList {
		sql = ""
		if err := rdb.Get(db, &sql); err != nil {
			utils.LogWirte(rdb.logger, "Error", "Copy", fmt.Sprintf("%s copy failed", db), nil)
			return err
		}

		if err := dst.Put(sql); err != nil {
			utils.LogWirte(rdb.logger, "Error", "Copy", fmt.Sprintf("%s copy failed", db), nil)
			return err
		}
		utils.LogWirte(rdb.logger, "Info", "Copy", fmt.Sprintf("%s copied", db), nil)
	}
	return nil
}

// Export all data in database
func (rdb *RDBController) Get(dbName string, sql *string) error {
	var sqlTemp string
	if err := rdb.client.ShowCreateDBSql(dbName, &sqlTemp); err != nil {
		utils.LogWirte(rdb.logger, "Error", "ShowCreateDBSql", "get db create sql failed", err)
		return err
	}
	sqlWrite(sql, sqlTemp)
	sqlWrite(sql, fmt.Sprintf("USE %s;", dbName))

	var tableList []string
	if err := rdb.client.ListTable(dbName, &tableList); err != nil {
		utils.LogWirte(rdb.logger, "Error", "ListTable", "get listTable failed", err)
		return err
	}

	for _, table := range tableList {
		sqlWrite(sql, fmt.Sprintf("DROP TABLE IF EXISTS %s;", table))

		if err := rdb.client.ShowCreateTableSql(dbName, table, &sqlTemp); err != nil {
			utils.LogWirte(rdb.logger, "Error", "ListTable", "get table create sql failed", err)
			return err
		}
		sqlWrite(sql, sqlTemp)
	}

	for _, table := range tableList {
		var insertData []string
		if err := rdb.client.GetInsert(dbName, table, &insertData); err != nil {
			utils.LogWirte(rdb.logger, "Error", "GetInsert", "get insert sql failed", err)
			return err
		}

		for _, data := range insertData {
			sqlWrite(sql, data)
		}
	}
	return nil
}

// Function to create a dividing line
func sqlWrite(sql *string, data string) {
	*sql += fmt.Sprintf("%s\n\n", data)
}

// Split by line
func splitLine(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := strings.Index(string(data), "\n\n"); i >= 0 {
		return i + 2, data[:i+2], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func (rdb *RDBController) DeleteDB(dbName ...string) error {
	for _, db := range dbName {
		if err := rdb.client.DeleteDB(db); err != nil {
			utils.LogWirte(rdb.logger, "Error", "DeleteDB", fmt.Sprintf("%s delete failed", db), err)
			return err
		}
		utils.LogWirte(rdb.logger, "Info", "DeleteDB", fmt.Sprintf("%s deleted", db), nil)
	}
	return nil
}
