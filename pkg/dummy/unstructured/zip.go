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
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/sirupsen/logrus"
)

// ZIP generation function using gofakeit
//
// CapacitySize is in GB and generates zip files
// within the entered dummyDir path.
func GenerateRandomZIP(dummyDir string, capacitySize int) error {
	dummyDir = filepath.Join(dummyDir, "zip")
	if err := utils.IsDir(dummyDir); err != nil {
		logrus.Errorf("IsDir function error : %v", err)
		return err
	}

	tempPath := filepath.Join(dummyDir, "tmpTxt")
	if err := os.MkdirAll(tempPath, 0755); err != nil {
		logrus.Errorf("MkdirAll function error : %v", err)
		return err
	}
	defer os.RemoveAll(tempPath)

	logrus.Info("start txt generation")
	if err := GenerateRandomTXT(tempPath, 1); err != nil {
		logrus.Error("failed to generate txt")
		return err
	}
	logrus.Info("successfully generated txt")

	countNum := make(chan int, capacitySize)
	resultChan := make(chan error, capacitySize)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomZIPWorker(countNum, dummyDir, tempPath, resultChan)
		}()
	}

	for i := 0; i < capacitySize; i++ {
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
func randomZIPWorker(countNum chan int, dummyDir, tempPath string, resultChan chan<- error) {
	for num := range countNum {
		w, err := os.Create(filepath.Join(dummyDir, fmt.Sprintf("datamold-dummy-data_%d.zip", num)))
		if err != nil {
			resultChan <- err
		}
		defer w.Close()

		zipWriter := zip.NewWriter(w)
		defer zipWriter.Close()

		if err := gzip(tempPath, zipWriter); err != nil {
			resultChan <- err
		}
		logrus.Infof("successfully generated : %s", w.Name())
		zipWriter.Close()
		w.Close()
		resultChan <- nil
	}
}

func gzip(srcDir string, zipWriter *zip.Writer) error {
	return filepath.Walk(srcDir, func(fp string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fileToZip, err := os.Open(fp)
			if err != nil {
				return err
			}
			defer fileToZip.Close()

			infoHeader, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			infoHeader.Name = filepath.Join(filepath.Base(srcDir), filepath.Base(fp))

			writer, err := zipWriter.CreateHeader(infoHeader)
			if err != nil {
				return err
			}

			_, err = io.Copy(writer, fileToZip)

			return err
		}
		return nil
	})
}
