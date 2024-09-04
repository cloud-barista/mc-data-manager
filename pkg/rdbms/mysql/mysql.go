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
package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloud-barista/mc-data-manager/models"
	_ "github.com/go-sql-driver/mysql"
)

// mysqlDBMS struct
type MysqlDBMS struct {
	provider models.Provider

	db  *sql.DB
	ctx context.Context
}

type MysqlDBOption func(*MysqlDBMS)

func New(provider models.Provider, sqlDB *sql.DB, opts ...MysqlDBOption) *MysqlDBMS {
	dms := &MysqlDBMS{
		provider: provider,
		db:       sqlDB,
		ctx:      context.TODO(),
	}

	for _, opt := range opts {
		opt(dms)
	}

	return dms
}

// Functions that execute EXEC commands in sql
func (d *MysqlDBMS) Exec(query string) error {
	if strings.HasPrefix(query, "CREATE DATABASE") {
		query = removeSQLComments(query)
		query = addCharsetIfMissing(query)
		query = addCollateIfMissing(query)
		dbName, charSet, collate := extractDatabaseInfo(query)
		if dbName == "" {
			return errors.New("exec error")
		}
		// Ensure charset is utf8mb4
		query = EnsureCharsetAndCollate(query, charSet, collate)

		// Create db with CALL system.ncp_cp_create_db command when provider is ncp and CREATE DATABASE is called
		if d.provider == models.NCP {
			query = fmt.Sprintf("CALL sys.ncp_create_db('%s', '%s', '%s');", dbName, charSet, collate)
		}

	} else if strings.HasPrefix(query, "CREATE TABLE") {
		query = ReplaceCharsetAndCollate(query)
	}

	_, err := d.db.Exec(query)
	return err
}

func removeSQLComments(line string) string {
	// Remove single line comments
	line = regexp.MustCompile(`--.*`).ReplaceAllString(line, "")
	// Remove multi-line comments
	line = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(line, "")
	return line
}

// addCollateIfMissing adds COLLATE to DEFAULT CHARACTER SET if it's missing
func addCollateIfMissing(sql string) string {
	if !strings.Contains(sql, "COLLATE") {
		sql = sql + " " + "COLLATE utf8mb4_general_ci"
	}
	return sql
}

// addCollateIfMissing adds COLLATE to DEFAULT CHARACTER SET if it's missing
func addCharsetIfMissing(sql string) string {
	if !strings.Contains(sql, "DEFAULT CHARACTER SET") && !strings.Contains(sql, "DEFAULT CHARSET") {
		sql = sql + " " + "DEFAULT CHARACTER SET utf8mb4"
	}
	return sql
}

// ReplaceCharsetAndCollate replaces any charset and collate in the SQL statement with utf8mb4 and utf8mb4_general_ci.
func ReplaceCharsetAndCollate(sql string) string {
	// Regular expression to match DEFAULT CHARSET and COLLATE settings
	reCharset := regexp.MustCompile(`(?i)DEFAULT CHARSET=\w+`)
	reCollate := regexp.MustCompile(`(?i)COLLATE=\w+`)

	// Replace with utf8mb4 and utf8mb4_general_ci
	sql = reCharset.ReplaceAllString(sql, "DEFAULT CHARSET=utf8mb4")
	sql = reCollate.ReplaceAllString(sql, "COLLATE=utf8mb4_general_ci")

	return sql
}

// Extract database information
func extractDatabaseInfo(sql string) (string, string, string) {
	match := []string{}
	if strings.Contains(sql, ";") {
		re := regexp.MustCompile(`CREATE\s+DATABASE\s+(IF\s+NOT\s+EXISTS\s+)?([^\s;]+)[^;]*;`)
		match = re.FindStringSubmatch(sql)
	} else {
		re := regexp.MustCompile("`([^`]*)`[^']*DEFAULT CHARACTER SET ([^ ]*) COLLATE ([^ ]*)")
		match = re.FindStringSubmatch(sql)
	}

	if len(match) == 4 {
		dbName := match[1]
		charSet := match[2]
		collate := match[3]
		return dbName, charSet, collate
	} else if len(match) == 3 {
		return match[2], "", ""
	}

	return "", "", ""
}

// EnsureCharsetAndCollate ensures that the charset is utf8mb4 and collate is utf8mb4_general_ci in the SQL query.
func EnsureCharsetAndCollate(query, charSet, collate string) string {
	// Ensure charset is utf8mb4
	if charSet != "utf8mb4" {
		query = strings.Replace(query, charSet, "utf8mb4", 1)
	}
	// Ensure collate is utf8mb4_general_ci
	if collate != "utf8mb4_general_ci" {
		if strings.Contains(query, "COLLATE") {
			re := regexp.MustCompile(`(?i)COLLATE\s+[^\s]+`)
			query = re.ReplaceAllString(query, "COLLATE utf8mb4_general_ci")
		} else {
			query = query + " COLLATE utf8mb4_general_ci"
		}
	}
	return query
}

// Delete database
func (d *MysqlDBMS) DeleteDB(dbName string) error {
	_, err := d.db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	return err
}

// Get database list
func (d *MysqlDBMS) ListDB(dst *[]string) error {
	rows, err := d.db.Query("SHOW DATABASES")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return err
		}

		if dbName != "information_schema" && dbName != "mysql" && dbName != "performance_schema" && dbName != "sys" {
			*dst = append(*dst, dbName)
		}
	}
	return nil
}

// Get table list
func (d *MysqlDBMS) ListTable(dbName string, dst *[]string) error {
	_, err := d.db.Exec(fmt.Sprintf("USE %s;", dbName))
	if err != nil {
		return err
	}

	rows, err := d.db.Query("SHOW TABLES")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		*dst = append(*dst, tableName)
	}
	return nil
}

// Get database create sql
func (d *MysqlDBMS) ShowCreateDBSql(dbName string, dbCreateSql *string) error {
	err := d.db.QueryRow(fmt.Sprintf("SHOW CREATE DATABASE %s;", dbName)).Scan(&dbName, dbCreateSql)
	if err != nil {
		return err
	}
	*dbCreateSql = strings.Replace(*dbCreateSql, "CREATE DATABASE", "CREATE DATABASE /*!32312 IF NOT EXISTS*/", 1)
	return nil
}

// Get table create sql
func (d *MysqlDBMS) ShowCreateTableSql(dbName, tableName string, tableCreateSql *string) error {
	if err := d.Exec(fmt.Sprintf("USE %s;", dbName)); err != nil {
		return err
	}
	if err := d.db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s;", tableName)).Scan(&tableName, tableCreateSql); err != nil {
		return err
	}
	return nil
}

// Get Insert sql
func (d *MysqlDBMS) GetInsert(dbName, tableName string, insertSql *[]string) error {
	colRows, err := d.db.Query("SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?", dbName, tableName)
	if err != nil {
		return err
	}
	defer colRows.Close()

	var columns []string
	for colRows.Next() {
		var columnName string
		if err := colRows.Scan(&columnName); err != nil {
			return err
		}
		columns = append(columns, columnName)
	}

	selectQuery := "SELECT " + strings.Join(columns, ", ") + " FROM " + tableName
	selRows, err := d.db.Query(selectQuery)
	if err != nil {
		return err
	}
	defer selRows.Close()

	data := []map[string]string{}

	for selRows.Next() {
		values := make([]string, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		err := selRows.Scan(valuePtrs...)
		if err != nil {
			return err
		}

		entry := make(map[string]string)
		for i, column := range columns {
			val := values[i]
			entry[column] = val
		}

		data = append(data, entry)
	}

	for _, entry := range data {
		values := []string{}
		for _, column := range columns {
			values = append(values, fmt.Sprintf("'%v'", entry[column]))
		}

		insertStatement := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", tableName, strings.Join(columns, ", "), strings.Join(values, ", "))
		*insertSql = append(*insertSql, insertStatement)
	}

	return nil
}
