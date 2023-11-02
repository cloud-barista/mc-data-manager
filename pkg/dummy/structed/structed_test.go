package structed_test

import (
	"path/filepath"
	"testing"

	"fmt"

	"github.com/cloud-barista/cm-data-mold/pkg/dummy/structed"
)

func TestCSV(t *testing.T) {
	// csv dummy data를 저장할 directory 경로 및 총 데이터 크기(GB단위로) 입력
	if err := structed.GenerateRandomCSV(filepath.Join("csv-dummy-directory-path", "csv"), 100); err != nil {
		fmt.Printf("test csv error : %v", err)
		panic(err)
	}

}

func TestSQL(t *testing.T) {
	// sql dummy data를 저장할 directory 경로 및 총 데이터 크기(GB단위로) 입력
	if err := structed.GenerateRandomSQL(filepath.Join("sql-dummy-directory-path", "sql"), 100); err != nil {
		fmt.Printf("test sql error : %v", err)
		panic(err)
	}
}
