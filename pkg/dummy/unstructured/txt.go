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
package unstructured

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/sirupsen/logrus"
)

// TXT generation function using gofakeit
//
// CapacitySize is in GB and generates txt files
// within the entered dummyDir path.
func GenerateRandomTXT(dummyDir string, capacitySize int) error {
	dummyDir = filepath.Join(dummyDir, "txt")
	if err := utils.IsDir(dummyDir); err != nil {
		logrus.Errorf("IsDir function error : %v", err)
		return err
	}

	countNum := make(chan int, capacitySize*10)
	resultChan := make(chan error, capacitySize*10)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomTxtWorker(countNum, dummyDir, resultChan)
		}()
	}

	for i := 0; i < capacitySize*10; i++ {
		countNum <- i
	}
	close(countNum)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for ret := range resultChan {
		if ret != nil {
			logrus.Errorf("result error : %v", ret)
			return ret
		}
	}

	return nil
}

// txt worker
func randomTxtWorker(countNum chan int, dirPath string, resultChan chan<- error) {
	for num := range countNum {
		file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("randomTxt_%d.txt", num)))
		if err != nil {
			resultChan <- err
		}

		for i := 0; i < 1000; i++ {
			if _, err := file.WriteString(fmt.Sprintf("%s\n", gofakeit.HipsterParagraph(10, 10, 120, " "))); err != nil {
				resultChan <- err
			}
		}

		logrus.Infof("successfully generated : %s", file.Name())

		if err := file.Close(); err != nil {
			resultChan <- err
		}

		resultChan <- nil
	}
}
