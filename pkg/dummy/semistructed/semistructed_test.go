package semistructed_test

import (
	"testing"

	"fmt"

	"github.com/cloud-barista/cm-data-mold/pkg/dummy/semistructed"
)

func TestJSON(t *testing.T) {
	// Enter the directory path and total data size (in GB) to store json dummy data
	if err := semistructed.GenerateRandomJSON("json-dummy-directory-path", 1); err != nil {
		fmt.Printf("test json error : %v", err)
		panic(err)
	}
}

func TestXML(t *testing.T) {
	// Enter the directory path and total data size in GB to store xml dummy data
	if err := semistructed.GenerateRandomXML("xml-dummy-directory-path", 1); err != nil {
		fmt.Printf("test xml error : %v", err)
		panic(err)
	}
}
