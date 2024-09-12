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
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/rs/zerolog/log"
)

type ImageType string

// PNG generation function using gofakeit
//
// CapacitySize is in GB and generates png files
// within the entered dummyDir path.
func GenerateRandomPNGImage(dummyDir string, capacitySize int) error {
	dummyDir = filepath.Join(dummyDir, "png")
	if err := utils.IsDir(dummyDir); err != nil {
		log.Error().Msgf("IsDir function error : %v", err)
		return err
	}

	size := capacitySize * 10 * 145

	countNum := make(chan int, size)
	resultChan := make(chan error, size)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			randomImageWorker(countNum, dummyDir, resultChan)
		}()
	}

	for i := 0; i < size; i++ {
		countNum <- i
	}
	close(countNum)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for ret := range resultChan {
		if ret != nil {
			log.Error().Msgf("result error : %v", ret)
			return ret
		}
	}

	return nil
}

// png worker
func randomImageWorker(countNum chan int, dirPath string, resultChan chan<- error) {
	for num := range countNum {
		file, err := os.Create(fmt.Sprintf("%s/randomImage_%d.png", dirPath, num))
		if err != nil {
			resultChan <- err
		}
		defer file.Close()

		if _, err := file.Write(gofakeit.ImagePng(500, 500)); err != nil {
			resultChan <- err
		}
		log.Info().Msgf("Creation success: %v", file.Name())

		file.Close()
	}
}
