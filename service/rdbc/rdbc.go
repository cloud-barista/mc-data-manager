package rdbc

import (
	"bufio"
	"fmt"
	"strings"

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
		return err
	}
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
				rdb.logWrite("Error", "sql exec error", err)
				return err
			}
		}
	}

	err := scanner.Err()
	if err != nil {
		return err
	}
	return nil
}

// Replication using put and get
func (rdb *RDBController) Copy(dst *RDBController) error {
	var dbList []string
	var sql string
	if err := rdb.ListDB(&dbList); err != nil {
		rdb.logWrite("Error", "ListDB error", err)
		return err
	}

	for _, db := range dbList {
		sql = ""
		if err := rdb.Get(db, &sql); err != nil {
			rdb.logWrite("Error", "Get error", err)
			return err
		}

		if err := dst.Put(sql); err != nil {
			rdb.logWrite("Error", "Get error", err)
			return err
		}
		rdb.logWrite("Info", fmt.Sprintf("Replication success: src:/%s -> dst:/%s", db, db), nil)
	}
	return nil
}

// Export all data in database
func (rdb *RDBController) Get(dbName string, sql *string) error {
	var sqlTemp string
	if err := rdb.client.ShowCreateDBSql(dbName, &sqlTemp); err != nil {
		return err
	}
	sqlWrite(sql, sqlTemp)
	sqlWrite(sql, fmt.Sprintf("USE %s;", dbName))

	var tableList []string
	if err := rdb.client.ListTable(dbName, &tableList); err != nil {
		return err
	}

	for _, table := range tableList {
		sqlWrite(sql, fmt.Sprintf("DROP TABLE IF EXISTS %s;", table))

		if err := rdb.client.ShowCreateTableSql(dbName, table, &sqlTemp); err != nil {
			return err
		}
		sqlWrite(sql, sqlTemp)
	}

	for _, table := range tableList {
		var insertData []string
		if err := rdb.client.GetInsert(dbName, table, &insertData); err != nil {
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
			return err
		}
	}
	return nil
}

func (rdbc *RDBController) logWrite(logLevel, msg string, err error) {
	if rdbc.logger != nil {
		switch logLevel {
		case "Info":
			rdbc.logger.Info(msg)
		case "Error":
			rdbc.logger.Errorf("%s : %v", msg, err)
		}
	}
}
