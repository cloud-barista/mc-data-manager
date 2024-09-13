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
	"fmt"
	"regexp"
	"strings"

	"github.com/cloud-barista/mc-data-manager/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

// mysqlDBMS struct
type MysqlDBMS struct {
	provider        models.Provider
	db              *sql.DB
	tartgetProvider models.Provider
	ctx             context.Context
}

type MysqlDBOption func(*MysqlDBMS)

func (d *MysqlDBMS) GetProvdier() models.Provider {
	return d.provider
}

func (d *MysqlDBMS) SetProvdier(provider models.Provider) {
	d.provider = provider
}

func (d *MysqlDBMS) GetTargetProvdier() models.Provider {
	return d.tartgetProvider
}

func (d *MysqlDBMS) SetTargetProvdier(provider models.Provider) {
	d.tartgetProvider = provider
}

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

	_, err := d.db.Exec(query)
	if err != nil {
		log.Error().Err(err).Str("query", query).Msg("Failed to execute SQL query")
	}
	log.Debug().Str("query", query).Msg("SQL query executed successfully")
	return err
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
		log.Error().Err(err).Msgf("SQL query executed failed %v", rows)
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			log.Error().Err(err).Msgf("SQL query executed failed %v", rows)
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
		log.Error().Err(err).Msgf("SQL query executed failed")
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

// ShowCreateDBSql modifies the CREATE DATABASE SQL and returns it
func (d *MysqlDBMS) ShowCreateDBSql(dbName string, dbCreateSql *string) error {
	err := d.db.QueryRow(fmt.Sprintf("SHOW CREATE DATABASE %s;", dbName)).Scan(&dbName, dbCreateSql)
	if err != nil {
		log.Error().Err(err).Msgf("SQL query executed failed")
		return err
	}

	// Add "IF NOT EXISTS" to the CREATE DATABASE statement
	*dbCreateSql = strings.Replace(*dbCreateSql, "CREATE DATABASE", "CREATE DATABASE /*!32312 IF NOT EXISTS*/", 1)

	// Ensure charset and collate are utf8mb4 and utf8mb4_general_ci
	*dbCreateSql = addCharsetIfMissing(*dbCreateSql)
	*dbCreateSql = addCollateIfMissing(*dbCreateSql)
	*dbCreateSql = EnsureCharsetAndCollate(*dbCreateSql, extractCharacterSet(*dbCreateSql), extractCollation(*dbCreateSql))

	// If the target provider is NCP, modify the SQL to use NCP's specific procedure
	if d.tartgetProvider == models.NCP {
		dbName, charSet, collate := extractDatabaseInfo(*dbCreateSql)
		*dbCreateSql = fmt.Sprintf("CALL sys.ncp_create_db('%s', '%s', '%s');", dbName, charSet, collate)
	}

	return nil
}

// Get table create sql
func (d *MysqlDBMS) ShowCreateTableSql(dbName, tableName string, tableCreateSql *string) error {
	if err := d.Exec(fmt.Sprintf("USE %s;", dbName)); err != nil {
		log.Error().Err(err).Msgf("SQL query executed failed")
		return err
	}
	if err := d.db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s;", tableName)).Scan(&tableName, tableCreateSql); err != nil {
		log.Error().Err(err).Msgf("SQL query executed failed")
		return err
	}
	*tableCreateSql = removeSequenceOption(*tableCreateSql)
	*tableCreateSql = adjustColumnsToTimestamp(*tableCreateSql)
	*tableCreateSql = ReplaceCharsetAndCollate(*tableCreateSql)
	return nil
}

// Get Insert sql
func (d *MysqlDBMS) GetInsert(dbName, tableName string, insertSql *[]string) error {
	colRows, err := d.db.Query("SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?", dbName, tableName)
	if err != nil {
		log.Error().Err(err).Msgf("SQL query executed failed")
		return err
	}
	defer colRows.Close()

	var columns []string
	for colRows.Next() {
		var columnName string
		if err := colRows.Scan(&columnName); err != nil {
			log.Error().Err(err).Msgf("SQL query executed failed")
			return err
		}
		columns = append(columns, columnName)
	}

	tableName = escapeColumnName(tableName)
	escapedColumns := make([]string, len(columns))
	for i, column := range columns {
		escapedColumns[i] = escapeColumnName(column)
	}

	selectQuery := "SELECT " + strings.Join(escapedColumns, ", ") + " FROM " + tableName
	selRows, err := d.db.Query(selectQuery)
	if err != nil {
		log.Error().Err(err).Msgf("SQL query executed failed")
		return err
	}
	defer selRows.Close()

	data := []map[string]sql.NullString{}

	for selRows.Next() {
		values := make([]sql.NullString, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		err := selRows.Scan(valuePtrs...)
		if err != nil {
			log.Error().Err(err).Msgf("SQL query executed failed")
			return err
		}

		entry := make(map[string]sql.NullString)
		for i, column := range columns {
			entry[column] = values[i]
		}

		data = append(data, entry)
	}

	for _, entry := range data {
		values := []string{}
		for _, column := range columns {
			val := entry[column]
			if val.Valid {
				escapedValue := ReplaceEscapeString(val.String)
				values = append(values, fmt.Sprintf("'%v'", escapedValue))
			} else {
				values = append(values, "NULL")
			}
		}

		insertStatement := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", tableName, strings.Join(escapedColumns, ", "), strings.Join(values, ", "))
		*insertSql = append(*insertSql, insertStatement)
	}

	return nil
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

func ReplaceEscapeString(input string) string {
	return strings.ReplaceAll(input, "'", "''")
}

func adjustColumnsToTimestamp(sql string) string {
	// Use a regular expression to find all columns that use DEFAULT current_timestamp()
	re := regexp.MustCompile("`[^`]+`\\s+[^,]+DEFAULT\\s+current_timestamp\\(\\)")

	// Replace these columns with TIMESTAMP DEFAULT current_timestamp()
	modifiedSQL := re.ReplaceAllStringFunc(sql, func(match string) string {
		// Retain the column name and change the rest of the definition to TIMESTAMP
		columnName := strings.Split(match, " ")[0] // The first element is the column name
		return fmt.Sprintf("%s TIMESTAMP DEFAULT current_timestamp()", columnName)
	})

	return modifiedSQL
}

// Extract database information
func extractDatabaseInfo(sql string) (string, string, string) {
	dbName := extractDatabaseName(sql)
	charSet := extractCharacterSet(sql)
	collate := extractCollation(sql)
	return dbName, charSet, collate
}

// extract DBname
func extractDatabaseName(sql string) string {
	re := regexp.MustCompile(`CREATE\s+DATABASE\s+(?:/\*.*?\*/\s*)?(IF\s+NOT\s+EXISTS\s+)?\s*` + "`([^`]*)`")
	match := re.FindStringSubmatch(sql)
	if len(match) >= 3 {
		return match[2]
	}
	return ""
}

// extract Charset
func extractCharacterSet(sql string) string {
	re := regexp.MustCompile(`DEFAULT\s+CHARACTER\s+SET\s+([^\s]+)`)
	match := re.FindStringSubmatch(sql)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

// extract Collation
func extractCollation(sql string) string {
	re := regexp.MustCompile(`(?:/\*.*?\*/\s*)?COLLATE\s+([^\s]+)`)
	match := re.FindStringSubmatch(sql)
	if len(match) >= 2 {
		return match[1]
	}
	return ""
}

// remove Sequence
func removeSequenceOption(sql string) string {
	return strings.Replace(sql, " SEQUENCE=1", "", -1)
}

// escape Reserve Word
func escapeColumnName(columnName string) string {
	return fmt.Sprintf("`%s`", columnName)
}
