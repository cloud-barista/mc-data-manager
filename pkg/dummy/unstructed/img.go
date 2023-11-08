package unstructed

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/cloud-barista/cm-data-mold/pkg/utils"
	"github.com/sirupsen/logrus"
)

type ImageType string

// PNG generation function using gofakeit
//
// CapacitySize is in GB and generates png files
// within the entered dummyDir path.
func GenerateRandomPNGImage(dummyDir string, capacitySize int) error {
	dummyDir = filepath.Join(dummyDir, "png")
	if err := utils.IsDir(dummyDir); err != nil {
		logrus.WithFields(logrus.Fields{"jobName": "png create"}).Errorf("IsDir function error : %v", err)
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
			logrus.WithFields(logrus.Fields{"jobName": "png create"}).Errorf("result error : %v", ret)
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
		logrus.WithFields(logrus.Fields{"jobName": "png create"}).Infof("Creation success: %v", file.Name())

		file.Close()
	}
}
