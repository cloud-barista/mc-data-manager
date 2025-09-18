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

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbms/mysql/diagnostics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type EngineType string

const (
	Mysql EngineType = "mysql"
)

// rdbms interface
//
// Configure the interface to make it easier for other DBs to apply in the future
type RDBMS interface {
	GetProvdier() models.Provider
	SetProvdier(provider models.Provider)
	GetTargetProvdier() models.Provider
	SetTargetProvdier(provider models.Provider)
	Exec(query string) error
	ListDB(dst *[]string) error
	DeleteDB(dbName string) error
	ListTable(dbName string, dst *[]string) error
	ShowCreateDBSql(dbName string, dbCreateSql *string) error
	ShowCreateTableSql(dbName, tableName string, tableCreateSql *string) error
	GetInsert(dbName, tableName string, insertSql *[]string) error
	Diagnose() (diagnostics.TimedResult, error)
}

type RDBController struct {
	Client RDBMS

	logger *zerolog.Logger
}

type Option func(*RDBController)

func WithLogger(logger *zerolog.Logger) Option {
	return func(r *RDBController) {
		r.logger = logger
	}
}

func New(rdb RDBMS, opts ...Option) (*RDBController, error) {

	rdbc := &RDBController{
		Client: rdb,
		logger: nil,
	}

	for _, opt := range opts {
		opt(rdbc)
	}

	return rdbc, nil
}

// Return db list
func (rdb *RDBController) ListDB(dst *[]string) error {
	err := rdb.Client.ListDB(dst)
	if err != nil {
		log.Error().Err(err).Msgf("RDB %v", *dst)
		return err
	}
	return nil
}

// sql import each line With Transaction
func (rdb *RDBController) Put(sql string) error {

	var err error
	scanner := bufio.NewScanner(strings.NewReader(sql))
	scanner.Split(splitLine)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ReplaceAll(line, "\n", "")
		if line != "" {
			err = rdb.Client.Exec(line)
			if err != nil {
				log.Error().Msgf("err Line : %+v", line)
				rdb.logWrite("Error", "sql exec error", err)
				return err
			}
		}
	}

	err = scanner.Err()
	if err != nil {
		return err
	}
	return nil
}

// sql import by .sql
func (rdb *RDBController) PutDoc(sql string) error {
	err := rdb.Client.Exec(sql)
	if err != nil {
		log.Error().Msgf("err SQL : %+v", sql)
		rdb.logWrite("Error", "sql exec error", err)
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
		rdb.Client.SetTargetProvdier(dst.Client.GetProvdier())
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
	if err := rdb.Client.ShowCreateDBSql(dbName, &sqlTemp); err != nil {
		log.Error().Msgf("ERR DB")
		return err
	}
	sqlWrite(sql, sqlTemp)
	sqlWrite(sql, fmt.Sprintf("USE %s;", dbName))

	var tableList []string
	if err := rdb.Client.ListTable(dbName, &tableList); err != nil {
		log.Error().Msgf("ERR List TB")

		return err
	}

	for _, table := range tableList {
		sqlWrite(sql, fmt.Sprintf("DROP TABLE IF EXISTS %s;", table))

		if err := rdb.Client.ShowCreateTableSql(dbName, table, &sqlTemp); err != nil {
			log.Error().Msgf("ERR Creatte TB")

			return err
		}
		sqlWrite(sql, sqlTemp)
	}

	for _, table := range tableList {
		var insertData []string
		if err := rdb.Client.GetInsert(dbName, table, &insertData); err != nil {
			log.Error().Msgf("Insert quer err")
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
		if err := rdb.Client.DeleteDB(db); err != nil {
			return err
		}
	}
	return nil
}

func (rdbc *RDBController) logWrite(logLevel, msg string, err error) {
	if rdbc.logger != nil {
		switch logLevel {
		case "Info":
			log.Info().Msg(msg)
		case "Error":
			log.Error().Msgf("%s : %v", msg, err)
		}
	}
}
