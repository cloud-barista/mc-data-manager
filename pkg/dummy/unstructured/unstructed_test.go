package unstructured_test

import (
	"fmt"
	"testing"

	"github.com/cloud-barista/cm-data-mold/pkg/dummy/unstructured"
)

func TestIMG(t *testing.T) {
	// Enter the directory path and total data size, in GB, to store the img dummy data
	if err := unstructured.GenerateRandomPNGImage("img-dummy-directory-path", 1); err != nil {
		fmt.Printf("test img error : %v", err)
		panic(err)
	}
}

func TestGIF(t *testing.T) {
	// Enter the directory path and total data size in GB to store gif dummy data
	if err := unstructured.GenerateRandomPNGImage("gif-dummy-directory-path", 1); err != nil {
		fmt.Printf("test gif error : %v", err)
		panic(err)
	}
}

func TestTXT(t *testing.T) {
	// Enter the directory path and total data size, in GB, to store txt dummy data
	if err := unstructured.GenerateRandomTXT("txt-dummy-directory-path", 1); err != nil {
		fmt.Printf("test txt error : %v", err)
		panic(err)
	}
}

func TestZIP(t *testing.T) {
	// Enter the directory path and total data size in GB to store zip dummy data
	if err := unstructured.GenerateRandomTXT("zip-dummy-directory-path", 1); err != nil {
		fmt.Printf("test zip error : %v", err)
		panic(err)
	}
}
