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
