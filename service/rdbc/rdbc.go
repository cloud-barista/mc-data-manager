package rdbc

import (
	"bufio"
	"fmt"
	"strings"
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
}

// Return db list
func (rdbc *RDBController) ListDB(dst *[]string) error {
	return rdbc.client.ListDB(dst)
}

// sql import
func (rdbc *RDBController) Put(sql string) error {
	scanner := bufio.NewScanner(strings.NewReader(sql))
	scanner.Split(splitLine)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ReplaceAll(line, "\n", "")
		if line != "" {
			err := rdbc.client.Exec(line)
			if err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

// Replication using put and get
func (src *RDBController) Copy(dst *RDBController) error {
	var dbList []string
	var sql string
	if err := src.ListDB(&dbList); err != nil {
		return err
	}

	for _, db := range dbList {
		sql = ""
		if err := src.Get(db, &sql); err != nil {
			return err
		}

		if err := dst.Put(sql); err != nil {
			return err
		}
		fmt.Println(db)
	}
	return nil
}

// Export all data in database
func (rdbc *RDBController) Get(dbName string, sql *string) error {
	var sqlTemp string
	if err := rdbc.client.ShowCreateDBSql(dbName, &sqlTemp); err != nil {
		return err
	}
	sqlWrite(sql, sqlTemp)
	sqlWrite(sql, fmt.Sprintf("USE %s;", dbName))

	var tableList []string
	if err := rdbc.client.ListTable(dbName, &tableList); err != nil {
		return err
	}

	for _, table := range tableList {
		sqlWrite(sql, fmt.Sprintf("DROP TABLE IF EXISTS %s;", table))

		if err := rdbc.client.ShowCreateTableSql(dbName, table, &sqlTemp); err != nil {
			return err
		}
		sqlWrite(sql, sqlTemp)
	}

	for _, table := range tableList {
		var insertData []string
		if err := rdbc.client.GetInsert(dbName, table, &insertData); err != nil {
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

type Option func(*RDBController)

func New(rdb RDBMS, opts ...Option) (*RDBController, error) {
	rdbc := &RDBController{
		client: rdb,
	}

	for _, opt := range opts {
		opt(rdbc)
	}

	return rdbc, nil
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
