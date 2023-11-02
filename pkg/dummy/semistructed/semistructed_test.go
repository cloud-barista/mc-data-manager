package semistructed_test

import (
	"testing"

	"fmt"

	"github.com/cloud-barista/cm-data-mold/pkg/dummy/semistructed"
)

func TestJSON(t *testing.T) {
	// json dummy data를 저장할 directory 경로 및 총 데이터 크기(GB단위로) 입력
	if err := semistructed.GenerateRandomJSON("json-dummy-directory-path", 1); err != nil {
		fmt.Printf("test json error : %v", err)
		panic(err)
	}
}

func TestXML(t *testing.T) {
	// xml dummy data를 저장할 directory 경로 및 총 데이터 크기(GB단위로) 입력
	if err := semistructed.GenerateRandomXML("xml-dummy-directory-path", 1); err != nil {
		fmt.Printf("test xml error : %v", err)
		panic(err)
	}
}
