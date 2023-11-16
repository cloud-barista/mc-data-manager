package structured_test

import (
	"path/filepath"
	"testing"

	"fmt"

	"github.com/cloud-barista/cm-data-mold/pkg/dummy/structured"
)

func TestCSV(t *testing.T) {
	// Enter the directory path and total data size in GB to store csv dummy data
	if err := structured.GenerateRandomCSV(filepath.Join("csv-dummy-directory-path", "csv"), 100); err != nil {
		fmt.Printf("test csv error : %v", err)
		panic(err)
	}

}

func TestSQL(t *testing.T) {
	// Enter the directory path and total data size in GB to store sql dummy data
	if err := structured.GenerateRandomSQL(filepath.Join("sql-dummy-directory-path", "sql"), 100); err != nil {
		fmt.Printf("test sql error : %v", err)
		panic(err)
	}
}
