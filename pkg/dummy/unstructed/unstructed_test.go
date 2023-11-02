package unstructed_test

import (
	"fmt"
	"testing"

	"github.com/cloud-barista/cm-data-mold/pkg/dummy/unstructed"
)

func TestIMG(t *testing.T) {
	// img dummy data를 저장할 directory 경로 및 총 데이터 크기(GB단위로) 입력
	if err := unstructed.GenerateRandomPNGImage("img-dummy-directory-path", 1); err != nil {
		fmt.Printf("test img error : %v", err)
		panic(err)
	}
}

func TestGIF(t *testing.T) {
	// gif dummy data를 저장할 directory 경로 및 총 데이터 크기(GB단위로) 입력
	if err := unstructed.GenerateRandomPNGImage("gif-dummy-directory-path", 1); err != nil {
		fmt.Printf("test gif error : %v", err)
		panic(err)
	}
}

func TestTXT(t *testing.T) {
	// txt dummy data를 저장할 directory 경로 및 총 데이터 크기(GB단위로) 입력
	if err := unstructed.GenerateRandomTXT("txt-dummy-directory-path", 1); err != nil {
		fmt.Printf("test txt error : %v", err)
		panic(err)
	}
}

func TestZIP(t *testing.T) {
	// zip dummy data를 저장할 directory 경로 및 총 데이터 크기(GB단위로) 입력
	if err := unstructed.GenerateRandomTXT("zip-dummy-directory-path", 1); err != nil {
		fmt.Printf("test zip error : %v", err)
		panic(err)
	}
}
