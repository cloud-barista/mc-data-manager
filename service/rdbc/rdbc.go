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
			// cd_qr := strings.HasPrefix(line, "CREATE DATABASE")
			// if cd_qr {
			// 	rdb.logger.Warnf("C_D_SQL : %v", line)
			// }
			// if err != nil {
			// 	// Handle collation error when migrating from MySQL 8.0 to versions below 8.0
			// 	updatedLine := strings.ReplaceAll(line, "utf8mb4_0900_ai_ci", "utf8mb4_general_ci")
			// 	err = rdb.client.Exec(updatedLine)
			// 	if err == nil {
			// 		rdb.logger.Warnf("Warning Line: \n %+v", line)
			// 		rdb.logger.Warnf("Collation error handled: MySQL 8.0 to versions below 8.0")
			// 		rdb.logger.Warnf("Changed DB collation from utf8mb4_0900_ai_ci to utf8mb4_general_ci")
			// 	}
			// }
			// if err != nil {
			// 	// Remove SQL comments
			// 	updatedLine := HandleSQL(line)
			// 	err = rdb.client.Exec(updatedLine)
			// 	rdb.logger.Warnf("Changed Line : %+v", updatedLine)
			// 	if err == nil {
			// 		rdb.logger.Warnf("Warning Line: \n %+v", line)
			// 		rdb.logger.Warnf("comments error handled")
			// 		rdb.logger.Warnf("Changed Line : %+v", updatedLine)
			// 	}
			// }
			if err != nil {
				rdb.logger.Errorf("err Line : %+v", line)
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

// Migration using put and get
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
		rdb.logWrite("Info", fmt.Sprintf("Migration success: src:/%s -> dst:/%s", db, db), nil)
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
