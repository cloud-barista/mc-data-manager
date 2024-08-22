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
package unstructured_test

import (
	"fmt"
	"testing"

	"github.com/cloud-barista/mc-data-manager/pkg/dummy/unstructured"
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
